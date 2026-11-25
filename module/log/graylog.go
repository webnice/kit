package log

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	kitModuleLogGelf "github.com/webnice/kit/v5/module/log/gelf"
	kitModuleLogLevel "github.com/webnice/kit/v5/module/log/level"
	kitModuleTrace "github.com/webnice/kit/v5/module/trace"
	kitModuleUuid "github.com/webnice/kit/v5/module/uuid"
)

// GraylogEnable Включение отправки сообщений лога в сервер Graylog по протоколу GELF.
// Адрес сервера задаётся в формате [proto://]ip.address.or.domain.name:port, где proto - udp (значение по
// умолчанию), tcp, tls или tcp+tls.
// appName - Название приложения, добавляемое в каждое сообщение GELF.
// Если адрес указан пустым, отправка сообщений в сервер Graylog отключается.
func (log *logger) GraylogEnable(address, appName string) (err error) {
	const (
		tplOnInitiate   = "Отправка логов в сервер graylog %q включена."
		protoDefault    = "udp"
		addressTemplate = "%s://%s"
		stackBack       = 3
	)
	var (
		glg         *graylog
		compression kitModuleLogGelf.Compression
		msg         *Message
	)

	address = strings.TrimSpace(address)
	appName = strings.TrimSpace(appName)
	log.graylogMux.Lock()
	defer log.graylogMux.Unlock()
	// Пустой адрес - отправка сообщений в сервер Graylog отключена.
	if address == "" {
		err = log.graylogDisable()
		return
	}
	// Добавление протокола UDP по умолчанию для адресов вида host:port.
	if !strings.Contains(address, "://") {
		address = fmt.Sprintf(addressTemplate, protoDefault, address)
	}
	// Повторное включение с теми же адресом и названием приложения не требует переподключения к серверу.
	if log.graylog != nil && log.graylog.address == address && log.graylog.appName == appName {
		return
	}
	// Сжатие gzip поддерживается только протоколом UDP, протоколы TCP и TLS сжатие не поддерживают.
	compression = kitModuleLogGelf.CompressionGzip
	if !strings.HasPrefix(strings.ToLower(address), protoDefault+"://") {
		compression = kitModuleLogGelf.CompressionNone
	}
	glg = &graylog{address: address, appName: appName, wr: log.wr}
	if glg.cli, err = kitModuleLogGelf.New(address, compression); err != nil {
		return
	}
	if glg.hostname, err = os.Hostname(); err != nil {
		glg.hostname = hostnameUnknown
	}
	// Отключение ранее включённой отправки сообщений.
	if err = log.graylogDisable(); err != nil {
		return
	}
	log.graylog = glg
	if err = log.HandlerSubscribe(log.graylog.Handler, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		return
	}
	// Отладочное сообщение о включении отправки логов в сервер Graylog.
	if log.debug {
		msg = log.MessageGet()
		msg.Timestamp = time.Now().In(time.UTC)
		msg.Level = kitModuleLogLevel.Info
		kitModuleTrace.Short(msg.Trace, stackBack)
		// bytes.Buffer.WriteString() не возвращает ошибок.
		_, _ = msg.Pattern.WriteString(fmt.Sprintf(tplOnInitiate, address))
		log.Message(msg)
	}

	return
}

// GraylogDisable Выключение отправки сообщений лога в сервер Graylog и закрытие соединения с сервером.
func (log *logger) GraylogDisable() (err error) {
	log.graylogMux.Lock()
	defer log.graylogMux.Unlock()
	err = log.graylogDisable()

	return
}

// Отключение обработчика отправки сообщений и закрытие соединения с сервером Graylog.
func (log *logger) graylogDisable() (err error) {
	if log.graylog == nil {
		return
	}
	if err = log.HandlerUnsubscribe(log.graylog.Handler); err != nil {
		return
	}
	if err = log.graylog.cli.Close(); err != nil {
		return
	}
	log.graylog = nil

	return
}

// Handler Обработчик сообщений лога, отправляющий сообщения в сервер Graylog по протоколу GELF.
func (glg *graylog) Handler(msg *Message) {
	const (
		tplEmpty = ""
		tplError = "отправка сообщения в сервер graylog %q прервана ошибкой: %s\n"
	)
	var (
		err     error
		gelfMsg *kitModuleLogGelf.Message
		short   string
		key     string
	)

	if msg.Pattern == nil {
		msg.Pattern = bytes.NewBufferString(tplEmpty)
	}
	short = msg.Pattern.String()
	// Формирование текста сообщения: аргументы подставляются в шаблон, а при пустом шаблоне аргументы
	// склеиваются без форматирования.
	if short == tplEmpty {
		short = fmt.Sprint(msg.Argument...)
	} else if len(msg.Argument) > 0 {
		short = fmt.Sprintf(short, msg.Argument...)
	}
	gelfMsg = kitModuleLogGelf.NewMessage(glg.hostname, short, levelToSeverity(msg.Level))
	// Полный текст сообщения.
	gelfMsg.Full = short
	// Заполнение полей сообщения по схеме предыдущей библиотеки логирования.
	glg.messageFields(gelfMsg, msg, short)
	// Добавление ключей сообщения лога в дополнительные поля сообщения GELF.
	for key = range msg.Keys {
		gelfMsg.Keys[key] = msg.Keys[key]
	}
	// Добавление информации о точке вызова логирования.
	if msg.Trace != nil {
		if msg.Trace.FilenameShort != nil && msg.Trace.FilenameShort.Len() > 0 {
			gelfMsg.Keys["file"] = msg.Trace.FilenameShort.String()
		}
		if msg.Trace.Line > 0 {
			gelfMsg.Keys["line"] = msg.Trace.Line
		}
		if msg.Trace.Function != nil && msg.Trace.Function.Len() > 0 {
			gelfMsg.Keys["function"] = msg.Trace.Function.String()
		}
	}
	// При ошибке отправки, сообщение об ошибке выводится напрямую, без передачи в логгер, для предотвращения рекурсии.
	if err = glg.cli.Send(gelfMsg); err != nil {
		_, _ = fmt.Fprintf(glg.wr, tplError, glg.address, err)
	}

	return
}

// Заполнение полей сообщения GELF по схеме библиотеки логирования 2019 года: идентификатор сообщения,
// идентификатор процесса, название приложения, имя хоста, время и уровень сообщения, текст сообщения
// и информация о точке вызова логирования.
//
//goland:noinspection DuplicatedCode
func (glg *graylog) messageFields(gelfMsg *kitModuleLogGelf.Message, msg *Message, short string) {
	const (
		keyID            = "id"
		keyPID           = "pid"
		keyAppName       = "appName"
		keyHostName      = "hostName"
		keyTodayAndNow   = "todayAndNow"
		keyLevelString   = "levelString"
		keyMessage       = "message"
		keyPackage       = "package"
		keyModule        = "module"
		keyFilenameLong  = "filenameLong"
		keyFilenameShort = "filenameShort"
		keyFileLine      = "fileLine"
		keyFunction      = "function"
		keyCallStack     = "callStack"
	)
	var (
		err error
		uid kitModuleUuid.UUID
		pkg string
	)

	// Идентификатор сообщения - UUID первой версии, основанный на времени.
	if uid, err = kitModuleUuid.Get().V1(); err == nil {
		gelfMsg.Fields[keyID] = uid.String()
	}
	gelfMsg.Fields[keyPID] = os.Getpid()
	gelfMsg.Fields[keyAppName] = glg.appName
	gelfMsg.Fields[keyHostName] = glg.hostname
	gelfMsg.Fields[keyTodayAndNow] = msg.Timestamp.In(time.Local)
	gelfMsg.Fields[keyLevelString] = msg.Level.String()
	gelfMsg.Fields[keyMessage] = short
	if msg.Trace == nil {
		return
	}
	// Название пакета и модуля, совершившего вызов логирования.
	if msg.Trace.Package != nil && msg.Trace.Package.Len() > 0 {
		pkg = msg.Trace.Package.String()
		gelfMsg.Facility = pkg
		gelfMsg.Fields[keyPackage] = pkg
		gelfMsg.Fields[keyModule] = path.Base(pkg)
	}
	// Путь и имя файла, номер строки и функция, совершившие вызов логирования.
	if msg.Trace.FilenameLong != nil && msg.Trace.FilenameLong.Len() > 0 {
		gelfMsg.Fields[keyFilenameLong] = msg.Trace.FilenameLong.String()
	}
	if msg.Trace.FilenameShort != nil && msg.Trace.FilenameShort.Len() > 0 {
		gelfMsg.Fields[keyFilenameShort] = msg.Trace.FilenameShort.String()
	}
	if msg.Trace.Line > 0 {
		gelfMsg.Fields[keyFileLine] = msg.Trace.Line
	}
	if msg.Trace.Function != nil && msg.Trace.Function.Len() > 0 {
		gelfMsg.Fields[keyFunction] = msg.Trace.Function.String()
	}
	// Стек вызовов, обрезанный до функции вызова логирования.
	if msg.Trace.StackTrace != nil && msg.Trace.StackTrace.Len() > 0 {
		gelfMsg.Fields[keyCallStack] = msg.Trace.StackTrace.String()
	}
}

// Преобразование уровня логирования в уровень важности сообщения syslog.
func levelToSeverity(level kitModuleLogLevel.Level) (ret int32) {
	switch level {
	case kitModuleLogLevel.Fatal:
		ret = kitModuleLogGelf.SeverityEmergency
	case kitModuleLogLevel.Alert:
		ret = kitModuleLogGelf.SeverityAlert
	case kitModuleLogLevel.Critical:
		ret = kitModuleLogGelf.SeverityCritical
	case kitModuleLogLevel.Error:
		ret = kitModuleLogGelf.SeverityError
	case kitModuleLogLevel.Warning:
		ret = kitModuleLogGelf.SeverityWarning
	case kitModuleLogLevel.Notice:
		ret = kitModuleLogGelf.SeverityNotice
	case kitModuleLogLevel.Info:
		ret = kitModuleLogGelf.SeverityInfo
	case kitModuleLogLevel.Debug, kitModuleLogLevel.Trace:
		ret = kitModuleLogGelf.SeverityDebug
	default:
		ret = kitModuleLogGelf.SeverityInfo
	}

	return
}

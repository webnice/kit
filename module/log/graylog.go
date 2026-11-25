package log

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	kitModuleLogGelf "github.com/webnice/kit/v5/module/log/gelf"
	kitModuleLogLevel "github.com/webnice/kit/v5/module/log/level"
	kitModuleTrace "github.com/webnice/kit/v5/module/trace"
)

// GraylogEnable Включение отправки сообщений лога в сервер Graylog по протоколу GELF.
// Адрес сервера задаётся в формате [proto://]ip.address.or.domain.name:port, где proto - udp (значение по
// умолчанию), tcp, tls или tcp+tls.
// Если адрес указан пустым, отправка сообщений в сервер Graylog отключается.
func (log *logger) GraylogEnable(address string) (err error) {
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
	// Повторное включение с тем же адресом не требует переподключения к серверу.
	if log.graylog != nil && log.graylog.address == address {
		return
	}
	// Сжатие gzip поддерживается только протоколом UDP, протоколы TCP и TLS сжатие не поддерживают.
	compression = kitModuleLogGelf.CompressionGzip
	if !strings.HasPrefix(strings.ToLower(address), protoDefault+"://") {
		compression = kitModuleLogGelf.CompressionNone
	}
	glg = &graylog{address: address, wr: log.wr}
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

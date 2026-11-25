package log

/*

Менеджер логирования.

Пакет выполняет функции подготовки сообщения лога со всеми атрибутами, ключами и свойствами и передачу сообщения лога
всем зарегистрированным обработчикам логов.
Если не зарегистрировано ни одного обработчика, тогда лог выводится в самой простой форме через собственный
обработчик по умолчанию.

*/

import (
	"container/list"
	"io"
	"regexp"
	"sync"
	"sync/atomic"

	kitModuleBus "github.com/webnice/kit/v5/module/bus"
	kitModuleLogGelf "github.com/webnice/kit/v5/module/log/gelf"
	kitModuleLogLevel "github.com/webnice/kit/v5/module/log/level"
	kitTypes "github.com/webnice/kit/v5/types"
)

const eApplicationFatality uint8 = 252 // 252 - приложение завершилось из-за печати в лог сообщения с уровнем Fatal.
const hostnameUnknown = "unknown"      // Имя хоста по умолчанию, используется при невозможности определить имя хоста операционной системы.

var (
	rexSpaceFirst = regexp.MustCompile(`^[\t\n\f\r ]+`)
	rexSpaceLast  = regexp.MustCompile(`[\t\n\f\r ]+$`)
)

// Logger Интерфейс логирования.
type Logger interface {
	// Initialization Инициализация менеджера логирования.
	Initialization() (err error)

	// Debug Установка флага отладки.
	Debug(d bool) Logger

	// MessageGet Получение объекта сообщений лога из бассейна объектов.
	MessageGet() (msg *Message)

	// MessagePut Функция для возврата объекта сообщений лога в бассейн объектов.
	// Функция используется только для возврата объекта в бассейн без вывода сообщения в лог.
	MessagePut(msg *Message)

	// Message Отправка log сообщения к обработчику сообщений.
	// Отправленные сообщения через эту функцию, отправлять в PutMessage не требуется.
	Message(msg *Message)

	// FlushAndClose Очистка буферизированных каналов, ожидание окончания обработки всех накопленных в памяти сообщений,
	// восстановление состояния стандартного лога.
	FlushAndClose()

	// HandlerSubscribe Регистрация обработчика сообщений лога с указанием интервала уровней сообщения лога,
	// которые будут передаваться подписчику.
	// min - Минимальный уровень сообщений лога, включительно.
	// max - Максимальный уровень сообщений лога, включительно.
	// В режиме отладки приложения, фильтрация по уровню сообщений лога не применяется.
	HandlerSubscribe(handler Handler, min, max kitModuleLogLevel.Level) (err error)

	// HandlerUnsubscribe Удаление обработчика сообщений лога.
	HandlerUnsubscribe(handler Handler) (err error)

	// GRAYLOG

	// GraylogEnable Включение отправки сообщений лога в сервер Graylog по протоколу GELF.
	// address - Адрес сервера Graylog, формат: [proto://]ip.address.or.domain.name:port, где proto - udp (значение
	// по умолчанию), tcp, tls или tcp+tls. Пример: udp://127.0.0.1:12201.
	// appName - Название приложения, добавляемое в каждое сообщение GELF.
	// Если адрес указан пустым, отправка сообщений в сервер Graylog отключается.
	GraylogEnable(address, appName string) (err error)

	// GraylogDisable Выключение отправки сообщений лога в сервер Graylog и закрытие соединения с сервером.
	// Если отправка сообщений не была включена, функция ничего не выполняет.
	GraylogDisable() (err error)

	// ИНТЕРФЕЙСЫ

	// Write Реализация интерфейса io.Writer.
	Write([]byte) (int, error)

	// WriteString Реализация интерфейса io.StringWriter.
	WriteString(string) (int, error)

	// ОШИБКИ

	// Errors Все ошибки известного состояния, которые может вернуть приложение или функция.
	Errors() *Error
}

// Объект пакета логирования
type logger struct {
	debug             bool                   // Флаг отладки приложения.
	wr                kitTypes.SyncWriter    // Интерфейс вывода всех возможных сообщений на печать.
	waitCounter       int64                  // Обеспечение обработки всех поступивших сообщений в лог.
	originalLogWriter io.Writer              // Оригинальный писатель логов до перехвата сообщений стандартного лога.
	messagePool       *sync.Pool             // Бассейн объектов сообщений лога.
	doEnd             atomic.Bool            // Флаг отключения логирования, когда установлен в "истина", прекращается приём новых сообщений лога.
	bus               kitModuleBus.Interface // Объект интерфейса Databus.
	subscribers       *list.List             // Список зарегистрированных обработчиков сообщений лога. *subscriber
	subscribersLock   sync.RWMutex           // Блокировка конкурентного доступа к списку обработчиков сообщений лога.
	graylogMux        *sync.RWMutex          // Блокировка конкурентного доступа к объекту отправки сообщений лога в сервер Graylog.
	graylog           *graylog               // Объект отправки сообщений лога в сервер Graylog, nil - отправка отключена.
}

// Handler Обработчик сообщений лога.
type Handler func(msg *Message)

type subscriber struct {
	Func Handler                 // Функция подписчика.
	Name string                  // Подписчик: пакет, объект.
	Min  kitModuleLogLevel.Level // Минимальный уровень сообщений лога, включительно.
	Max  kitModuleLogLevel.Level // Максимальный уровень сообщений лога, включительно.
}

// Объект отправки сообщений лога в сервер Graylog.
type graylog struct {
	address  string                  // Адрес сервера Graylog.
	hostname string                  // Имя хоста отправителя сообщений.
	appName  string                  // Название приложения отправителя сообщений.
	cli      kitModuleLogGelf.Client // Клиент протокола GELF.
	wr       kitTypes.SyncWriter     // Интерфейс вывода сообщений об ошибках отправки.
}

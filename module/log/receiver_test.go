package log

import (
	"strings"
	"testing"
	"time"

	kitModuleBus "github.com/webnice/kit/v4/module/bus"
	kitModuleLogLevel "github.com/webnice/kit/v4/module/log/level"
)

func TestConsumerWrongType(t *testing.T) {
	const (
		tplWrongType = "функция Consumer() с не верным типом данных, вернула %v, %v."
	)
	var (
		tl   *logger
		ret  []any
		errs []error
	)

	tl, _, _ = newTestLogger()
	ret, errs = tl.Consumer(true, "не структура сообщения")
	if ret != nil || errs != nil {
		t.Errorf(tplWrongType, ret, errs)
	}
}

func TestConsumerDefaultHandler(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		testMessage            = "сообщение по умолчанию"
	)
	var (
		tl  *logger
		tsw *testSyncWriter
		msg *Message
		err error
	)

	tl, _, tsw = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	defer tl.FlushAndClose()
	// Обработчики сообщений лога не зарегистрированы - сообщение печатается обработчиком по умолчанию.
	msg = tl.MessageGet()
	msg.Level = kitModuleLogLevel.Error
	msg.Pattern.WriteString(testMessage)
	tl.Message(msg)
	waitFor(t, func() bool { return strings.Contains(tsw.String(), testMessage) })
}

func TestConsumerLevelFilter(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		tplFatalSubscribe      = "функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s."
		tplLevelMismatch       = "получено сообщение с уровнем %v, ожидался Error."
	)
	var (
		tl      *logger
		receive chan *Message
		msg     *Message
		err     error
	)

	tl, _, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	defer tl.FlushAndClose()
	receive = make(chan *Message, 2)
	// Регистрация обработчика только для уровня Error.
	if err = tl.HandlerSubscribe(func(m *Message) {
		var s Message

		s = messageSnapshot(m)
		receive <- &s
	}, kitModuleLogLevel.Error, kitModuleLogLevel.Error); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	// Сообщение с уровнем Warning отфильтровано.
	msg = tl.MessageGet()
	msg.Level = kitModuleLogLevel.Warning
	tl.Message(msg)
	// Сообщение с уровнем Error доставлено обработчику.
	msg = tl.MessageGet()
	msg.Level = kitModuleLogLevel.Error
	tl.Message(msg)
	if ret := waitMessage(t, receive); ret.Level != kitModuleLogLevel.Error {
		t.Errorf(tplLevelMismatch, ret.Level)
	}
}

func TestConsumerDebugModeNoFilter(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		tplFatalSubscribe      = "функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s."
		tplLevelMismatch       = "получено сообщение с уровнем %v, ожидался Warning."
	)
	var (
		tl      *logger
		receive chan *Message
		msg     *Message
		err     error
	)

	tl, _, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	defer tl.FlushAndClose()
	tl.Debug(true)
	receive = make(chan *Message, 2)
	// Регистрация обработчика только для уровня Error.
	if err = tl.HandlerSubscribe(func(m *Message) {
		var s Message

		s = messageSnapshot(m)
		receive <- &s
	}, kitModuleLogLevel.Error, kitModuleLogLevel.Error); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	// В режиме отладки фильтрация по уровню не применяется.
	msg = tl.MessageGet()
	msg.Level = kitModuleLogLevel.Warning
	tl.Message(msg)
	if ret := waitMessage(t, receive); ret.Level != kitModuleLogLevel.Warning {
		t.Errorf(tplLevelMismatch, ret.Level)
	}
}

func TestConsumerMultipleHandlers(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		tplFatalSubscribe      = "функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s."
	)
	var (
		tl       *logger
		receive1 chan *Message
		receive2 chan *Message
		msg      *Message
		err      error
	)

	receive1 = make(chan *Message, 2)
	receive2 = make(chan *Message, 2)
	tl, _, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	defer tl.FlushAndClose()
	if err = tl.HandlerSubscribe(func(m *Message) {
		var s Message

		s = messageSnapshot(m)
		receive1 <- &s
	}, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	if err = tl.HandlerSubscribe(func(m *Message) {
		var s Message

		s = messageSnapshot(m)
		receive2 <- &s
	}, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	msg = tl.MessageGet()
	msg.Level = kitModuleLogLevel.Info
	tl.Message(msg)
	waitMessage(t, receive1)
	waitMessage(t, receive2)
}

func TestConsumerHandlerPanic(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		tplFatalSubscribe      = "функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s."
		tplPanicErrors         = "функция Consumer() с паникой обработчика, вернула не ожидаемые ошибки: %v."
		testPanicMessage       = "паника обработчика"
	)
	var (
		tl   *logger
		bs   kitModuleBus.Interface
		errs []error
		msg  *Message
		err  error
	)

	tl, bs, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	defer tl.FlushAndClose()
	if err = tl.HandlerSubscribe(testHandlerPanic, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	// Синхронная публикация, обработчик прерывает выполнение паникой.
	msg = tl.MessageGet()
	msg.Level = kitModuleLogLevel.Error
	msg.Pattern.WriteString(testPanicMessage)
	if _, errs = bs.PublishSync(msg); len(errs) == 0 ||
		!strings.Contains(errs[0].Error(), "паникой") {
		t.Errorf(tplPanicErrors, errs)
	}
}

// testHandlerPanic Тестовый обработчик сообщений лога с паникой.
func testHandlerPanic(_ *Message) { panic("паника обработчика") }

// waitFor Ожидание выполнения условия в течение ограниченного времени.
func waitFor(t *testing.T, condition func() bool) {
	const errWaitTimeout = "превышено время ожидания условия."
	var n time.Time

	t.Helper()
	n = time.Now().Add(time.Second * 5)
	for time.Now().Before(n) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal(errWaitTimeout)
}

// TestSubscribersWrongElementType Элемент списка обработчиков не верного типа пропускается циклами обхода Consumer().
func TestSubscribersWrongElementType(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		tplFatalSubscribe      = "функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s."
		tplFatalUnsubscribe    = "функция HandlerUnsubscribe(), вернула не ожидаемую ошибку: %s."
		tplWrongElementErrors  = "функция Consumer() с элементом не верного типа, вернула ошибки: %v."
	)
	var (
		tl   *logger
		bs   kitModuleBus.Interface
		msg  *Message
		errs []error
		err  error
	)

	tl, bs, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	defer tl.FlushAndClose()
	// Элемент списка обработчиков не верного типа пропускается циклами обхода.
	tl.subscribers.PushBack("не обработчик")
	if err = tl.HandlerSubscribe(testHandlerSubscribeOne, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	if err = tl.HandlerUnsubscribe(testHandlerSubscribeOne); err != nil {
		t.Fatalf(tplFatalUnsubscribe, err)
	}
	// Обработка сообщения с элементом не верного типа в списке обработчиков.
	// Синхронная публикация гарантирует завершение Consumer до возврата функции.
	msg = tl.MessageGet()
	msg.Level = kitModuleLogLevel.Error
	if _, errs = bs.PublishSync(msg); len(errs) != 0 {
		t.Errorf(tplWrongElementErrors, errs)
	}
}

package log

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	kitModuleBus "github.com/webnice/kit/v4/module/bus"
	kitModuleLogLevel "github.com/webnice/kit/v4/module/log/level"
	kitTypes "github.com/webnice/kit/v4/types"
)

// testSyncWriter Реализация интерфейса kitTypes.SyncWriter на основе буфера с блокировкой конкурентного доступа.
type testSyncWriter struct {
	bytes.Buffer
	syncProtect sync.RWMutex
	syncError   error
}

// Write Реализация интерфейса io.Writer.
func (tsw *testSyncWriter) Write(p []byte) (int, error) {
	tsw.syncProtect.Lock()
	defer tsw.syncProtect.Unlock()

	return tsw.Buffer.Write(p)
}

// Sync Реализация интерфейса SyncWriter.
func (tsw *testSyncWriter) Sync() error { return tsw.syncError }

// String Реализация интерфейса fmt.Stringer.
func (tsw *testSyncWriter) String() string {
	tsw.syncProtect.RLock()
	defer tsw.syncProtect.RUnlock()

	return tsw.Buffer.String()
}

// Len Длина буфера.
func (tsw *testSyncWriter) Len() int {
	tsw.syncProtect.RLock()
	defer tsw.syncProtect.RUnlock()

	return tsw.Buffer.Len()
}

// Reset Сброс буфера.
func (tsw *testSyncWriter) Reset() {
	tsw.syncProtect.Lock()
	defer tsw.syncProtect.Unlock()
	tsw.Buffer.Reset()
}

// testBusAsyncError Имитация шины данных возвращающей ошибку при асинхронной публикации.
type testBusAsyncError struct {
	kitModuleBus.Interface
}

// IsSubscriber Реализация интерфейса шины данных.
func (tbae *testBusAsyncError) IsSubscriber(_ any) bool { return true }

// PublishAsync Реализация интерфейса шины данных.
func (tbae *testBusAsyncError) PublishAsync(_ any) error {
	return errors.New(errTextPublishAsync)
}

// testBusSubscribeError Имитация шины данных возвращающей ошибку при регистрации потребителя.
type testBusSubscribeError struct {
	kitModuleBus.Interface
}

// Subscribe Реализация интерфейса шины данных.
func (tbse *testBusSubscribeError) Subscribe(_ kitTypes.Databuser) error {
	return errors.New(errTextSubscribeError)
}

// Строки ошибок тестовых имитаций шины данных.
const (
	errTextPublishAsync   = "тестовая ошибка публикации."
	errTextSubscribeError = "тестовая ошибка регистрации."
)

// Создание тестового объекта логирования без регистрации в шине данных.
func newTestLogger() (tl *logger, bus kitModuleBus.Interface, tsw *testSyncWriter) {
	tsw = &testSyncWriter{}
	bus = kitModuleBus.New(16, 4, true)
	tl = New(tsw, bus).(*logger)

	return
}

// Ожидание получения сообщения лога обработчиком.
func waitMessage(t *testing.T, ch chan *Message) (msg *Message) {
	const errWaitMessageTimeout = "превышено время ожидания сообщения лога."

	t.Helper()
	select {
	case msg = <-ch:
	case <-time.After(time.Second * 5):
		t.Fatal(errWaitMessageTimeout)
	}

	return
}

func TestNew(t *testing.T) {
	const (
		errNewNil        = "функция New(), вернулся nil, ожидался не nil объект."
		errErrorsNil     = "функция Errors(), вернулся nil, ожидался не nil объект."
		errDebugWrong    = "функция Debug(), вернулся не верный объект."
		errDebugNotSet   = "функция Debug(true), флаг отладки не установлен."
		errDebugNotReset = "функция Debug(false), флаг отладки не сброшен."
		errBusNil        = "тестовая шина данных не создана."
	)
	var (
		tl  *logger
		bus kitModuleBus.Interface
	)

	tl, bus, _ = newTestLogger()
	if tl == nil {
		t.Fatal(errNewNil)
	}
	if tl.Errors() == nil {
		t.Error(errErrorsNil)
	}
	if tl.Debug(true) != Logger(tl) {
		t.Error(errDebugWrong)
	}
	if !tl.debug {
		t.Error(errDebugNotSet)
	}
	tl.Debug(false)
	if tl.debug {
		t.Error(errDebugNotReset)
	}
	if bus == nil {
		t.Error(errBusNil)
	}
}

func TestMessageGetPut(t *testing.T) {
	const (
		errGetNil        = "функция MessageGet(), вернулся nil, ожидался не nil объект."
		errFieldsNotInit = "функция MessageGet(), объект сообщения содержит не инициализированные поля."
		errNotReset      = "функция MessagePut(), объект сообщения не сброшен."
	)
	var (
		tl  *logger
		msg *Message
	)

	tl, _, _ = newTestLogger()
	// Получение объекта из бассейна.
	if msg = tl.MessageGet(); msg == nil {
		t.Fatal(errGetNil)
	}
	if msg.Pattern == nil || msg.Argument == nil || msg.Keys == nil || msg.Trace == nil || msg.Done == nil {
		t.Error(errFieldsNotInit)
	}
	// Заполнение объекта и возврат в бассейн.
	msg.Level = kitModuleLogLevel.Error
	msg.Pattern.WriteString("тест %s")
	msg.Argument = append(msg.Argument, "аргумент")
	msg.Keys["ключ"] = "значение"
	msg.Fatality = true
	tl.MessagePut(msg)
	msg = tl.MessageGet()
	if msg.Level != 0 || msg.Pattern.Len() != 0 || len(msg.Argument) != 0 || len(msg.Keys) != 0 || msg.Fatality {
		t.Error(errNotReset)
	}
	// Возврат nil в бассейн - паники нет.
	tl.MessagePut(nil)
}

func TestInitialization(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		errNotSubscribed       = "после Initialization(), лог не зарегистрирован в шине данных."
		errStillSubscribed     = "после FlushAndClose(), лог остался зарегистрирован в шине данных."
	)
	var (
		tl  *logger
		bus kitModuleBus.Interface
		err error
	)

	tl, bus, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	// Лог зарегистрирован в шине данных как потребитель сообщений.
	if !bus.IsSubscriber(new(Message)) {
		t.Error(errNotSubscribed)
	}
	tl.FlushAndClose()
	// После FlushAndClose() регистрация в шине данных удалена.
	if bus.IsSubscriber(new(Message)) {
		t.Error(errStillSubscribed)
	}
}

func TestMessageAfterClose(t *testing.T) {
	const (
		errMessageAfterClose = "функция Message() после FlushAndClose(), сообщение было обработано."
	)
	var (
		tl   *logger
		tsw  *testSyncWriter
		msg  *Message
		size int
	)

	tl, _, tsw = newTestLogger()
	tl.FlushAndClose()
	msg = tl.MessageGet()
	msg.Pattern.WriteString("после закрытия")
	size = tsw.Len()
	tl.Message(msg)
	if tsw.Len() != size {
		t.Error(errMessageAfterClose)
	}
}

func TestMessageDefaultHandler(t *testing.T) {
	const (
		tplNoTemplatePrint = "сообщение без шаблона не напечатано, лог: %q."
		tplNoPatternPrint  = "сообщение с шаблоном не напечатано, лог: %q."
	)
	var (
		tl  *logger
		tsw *testSyncWriter
		msg *Message
	)

	tl, _, tsw = newTestLogger()
	// Сообщение без шаблона - аргументы печатаются через Fprintln().
	msg = tl.MessageGet()
	msg.Argument = append(msg.Argument, "тестовое сообщение")
	tl.Message(msg)
	if !strings.Contains(tsw.String(), "тестовое сообщение") {
		t.Errorf(tplNoTemplatePrint, tsw.String())
	}
	// Сообщение с шаблоном и аргументами.
	tsw.Reset()
	msg = tl.MessageGet()
	msg.Pattern.WriteString("шаблон %q")
	msg.Argument = append(msg.Argument, "аргумент")
	tl.Message(msg)
	if !strings.Contains(tsw.String(), `шаблон "аргумент"`) {
		t.Errorf(tplNoPatternPrint, tsw.String())
	}
}

func TestHandlerSubscribeUnsubscribe(t *testing.T) {
	const (
		tplFatalSubscribe      = "функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s."
		tplSubscribeAgainWrong = "функция HandlerSubscribe() повторной регистрации, вернула не ожидаемую ошибку: %v."
		tplHandlersCount       = "количество обработчиков %d, ожидалось %d."
		tplFatalUnsubscribe    = "функция HandlerUnsubscribe(), вернула не ожидаемую ошибку: %s."
		tplUnsubscribeAgain    = "функция HandlerUnsubscribe() повторного вызова, вернула не ожидаемую ошибку: %v."
		errUnsubscribeNotFound = "функция HandlerUnsubscribe() не зарегистрированного обработчика, ошибка не возвращена."
	)
	var (
		tl  *logger
		err error
	)

	tl, _, _ = newTestLogger()
	if err = tl.HandlerSubscribe(testHandlerSubscribeOne, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	// Повторная регистрация того же обработчика - ошибка.
	if err = tl.HandlerSubscribe(testHandlerSubscribeOne, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err == nil ||
		!strings.Contains(err.Error(), "уже зарегистрирован") {
		t.Errorf(tplSubscribeAgainWrong, err)
	}
	if err = tl.HandlerSubscribe(testHandlerSubscribeTwo, kitModuleLogLevel.Error, kitModuleLogLevel.Error); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	if tl.subscribers.Len() != 2 {
		t.Errorf(tplHandlersCount, tl.subscribers.Len(), 2)
	}
	// Удаление регистрации обработчика.
	if err = tl.HandlerUnsubscribe(testHandlerSubscribeOne); err != nil {
		t.Fatalf(tplFatalUnsubscribe, err)
	}
	if tl.subscribers.Len() != 1 {
		t.Errorf(tplHandlersCount, tl.subscribers.Len(), 1)
	}
	// Повторное удаление регистрации - ошибка.
	if err = tl.HandlerUnsubscribe(testHandlerSubscribeOne); err == nil ||
		!strings.Contains(err.Error(), "не найдена") {
		t.Errorf(tplUnsubscribeAgain, err)
	}
	// Удаление не зарегистрированного обработчика - ошибка.
	if err = tl.HandlerUnsubscribe(testHandlerSubscribeOne); err == nil {
		t.Error(errUnsubscribeNotFound)
	}
	if err = tl.HandlerUnsubscribe(testHandlerSubscribeTwo); err != nil {
		t.Fatalf(tplFatalUnsubscribe, err)
	}
}

// Тестовый обработчик сообщений лога.
func testHandlerSubscribeOne(_ *Message) {}

// Тестовый обработчик сообщений лога.
func testHandlerSubscribeTwo(_ *Message) {}

func TestFlushAndCloseUnsubscribeError(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		tplFatalUnsubscribe    = "функция Unsubscribe(), вернула не ожидаемую ошибку: %s."
		errUnsubscribeNotPrint = "ошибка отписывания от шины данных не напечатана в лог."
	)
	var (
		tl  *logger
		bus kitModuleBus.Interface
		tsw *testSyncWriter
		err error
	)

	tl, bus, tsw = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	// Удаление регистрации в шине данных до FlushAndClose(), ошибка отписывания
	// печатается в лог, функция завершается без паники.
	if err = bus.Unsubscribe(tl); err != nil {
		t.Fatalf(tplFatalUnsubscribe, err)
	}
	tl.FlushAndClose()
	if tsw.Len() == 0 {
		t.Error(errUnsubscribeNotPrint)
	}
}

func TestInitializationSubscribeError(t *testing.T) {
	const (
		tplSubscribeErrorWrong = "функция Initialization() с ошибкой регистрации, вернула не ожидаемую ошибку: %v."
	)
	var (
		tl  Logger
		err error
	)

	tl = New(&testSyncWriter{}, &testBusSubscribeError{})
	if err = tl.Initialization(); err == nil ||
		!strings.Contains(err.Error(), "тестовая ошибка регистрации") {
		t.Errorf(tplSubscribeErrorWrong, err)
	}
}

// TestSubscribersWrongElementType moved to receiver_test.go: обработка элементов списка обработчиков
// не верного типа выполняется в Consumer().

// TestMessageFatalityExit Фатальное сообщение приводит к завершению приложения с кодом 252.
func TestMessageFatalityExit(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		errChildNoError        = "дочерний процесс завершился без ошибки, ожидалось завершение с кодом 252."
		tplChildExitWrong      = "дочерний процесс завершился с ошибкой %v, ожидался код выхода %d."
		envFatalExit           = "GO_WANT_LOG_FATAL_EXIT"
		testName               = "TestMessageFatalityExit"
	)
	var (
		tl      *logger
		err     error
		cmd     *exec.Cmd
		exitErr *exec.ExitError
	)

	if os.Getenv(envFatalExit) == "1" {
		tl, _, _ = newTestLogger()
		if err = tl.Initialization(); err != nil {
			t.Fatalf(tplFatalInitialization, err)
		}
		// Обработчик прерывается паникой - ошибки синхронной публикации печатаются перед завершением.
		_ = tl.HandlerSubscribe(func(_ *Message) { panic("паника обработчика") },
			kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace)
		// Вызов Write из функции с именем содержащим log.fatal - уровень Fatal и флаг фатальности.
		log{}.fatal(tl)

		return
	}
	cmd = exec.Command(os.Args[0], "-test.run="+testName)
	cmd.Env = append(os.Environ(), envFatalExit+"=1")
	if err = cmd.Run(); err == nil {
		t.Fatal(errChildNoError)
	}
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != int(eApplicationFatality) {
		t.Errorf(tplChildExitWrong, err, eApplicationFatality)
	}
}

// TestMessagePublishAsyncErrorExit Ошибка асинхронной публикации в шину данных приводит к завершению приложения.
func TestMessagePublishAsyncErrorExit(t *testing.T) {
	const (
		errChildNoError   = "дочерний процесс завершился без ошибки, ожидалось завершение с кодом 252."
		tplChildExitWrong = "дочерний процесс завершился с ошибкой %v, ожидался код выхода %d."
		envAsyncErrorExit = "GO_WANT_LOG_ASYNC_ERROR_EXIT"
		testName          = "TestMessagePublishAsyncErrorExit"
	)
	var (
		tl      Logger
		err     error
		cmd     *exec.Cmd
		exitErr *exec.ExitError
	)

	if os.Getenv(envAsyncErrorExit) == "1" {
		tl = New(&testSyncWriter{}, &testBusAsyncError{})
		tl.Message(tl.MessageGet())

		return
	}
	cmd = exec.Command(os.Args[0], "-test.run="+testName)
	cmd.Env = append(os.Environ(), envAsyncErrorExit+"=1")
	if err = cmd.Run(); err == nil {
		t.Fatal(errChildNoError)
	}
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != int(eApplicationFatality) {
		t.Errorf(tplChildExitWrong, err, eApplicationFatality)
	}
}

// TestWriteStdLogInterception Перехват стандартного лога.
func TestWriteStdLogInterception(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		testMessage            = "сообщение стандартного лога."
	)
	var (
		tl  *logger
		tsw *testSyncWriter
		err error
	)

	tl, _, tsw = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	standardLogPrint(testMessage)
	waitFor(t, func() bool {
		return strings.Contains(tsw.String(), testMessage)
	})
	tl.FlushAndClose()
}

func TestFlushAndCloseWaitInflight(t *testing.T) {
	const (
		errWaitTimeout = "превышено время ожидания FlushAndClose."
	)
	var (
		tl   *logger
		done chan struct{}
	)

	tl, _, _ = newTestLogger()
	// Имитация сообщения находящегося в обработке - тело цикла ожидания выполняется.
	atomic.AddInt64(&tl.waitCounter, 1)
	done = make(chan struct{})
	go func() {
		tl.FlushAndClose()
		close(done)
	}()
	time.Sleep(time.Millisecond * 50)
	atomic.AddInt64(&tl.waitCounter, -1)
	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Fatal(errWaitTimeout)
	}
}

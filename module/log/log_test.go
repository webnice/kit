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

func (tsw *testSyncWriter) Write(p []byte) (int, error) {
	tsw.syncProtect.Lock()
	defer tsw.syncProtect.Unlock()

	return tsw.Buffer.Write(p)
}

func (tsw *testSyncWriter) Sync() error { return tsw.syncError }

func (tsw *testSyncWriter) String() string {
	tsw.syncProtect.RLock()
	defer tsw.syncProtect.RUnlock()

	return tsw.Buffer.String()
}

func (tsw *testSyncWriter) Len() int {
	tsw.syncProtect.RLock()
	defer tsw.syncProtect.RUnlock()

	return tsw.Buffer.Len()
}

func (tsw *testSyncWriter) Reset() {
	tsw.syncProtect.Lock()
	defer tsw.syncProtect.Unlock()
	tsw.Buffer.Reset()
}

// testBusAsyncError Имитация шины данных возвращающей ошибку при асинхронной публикации.
type testBusAsyncError struct {
	kitModuleBus.Interface
}

func (tbae *testBusAsyncError) IsSubscriber(_ any) bool { return true }

func (tbae *testBusAsyncError) PublishAsync(_ any) error {
	return errors.New("тестовая ошибка публикации")
}

// testBusSubscribeError Имитация шины данных возвращающей ошибку при регистрации потребителя.
type testBusSubscribeError struct {
	kitModuleBus.Interface
}

func (tbse *testBusSubscribeError) Subscribe(_ kitTypes.Databuser) error {
	return errors.New("тестовая ошибка регистрации")
}

// newTestLogger Создание тестового объекта логирования без регистрации в шине данных.
func newTestLogger() (tl *logger, bus kitModuleBus.Interface, tsw *testSyncWriter) {
	tsw = &testSyncWriter{}
	bus = kitModuleBus.New(16, 4, true)
	tl = New(tsw, bus).(*logger)

	return
}

// waitMessage Ожидание получения сообщения лога обработчиком.
func waitMessage(t *testing.T, ch chan *Message) (msg *Message) {
	t.Helper()
	select {
	case msg = <-ch:
	case <-time.After(time.Second * 5):
		t.Fatal("превышено время ожидания сообщения лога")
	}

	return
}

func TestNew(t *testing.T) {
	var (
		tl  *logger
		bus kitModuleBus.Interface
	)

	tl, bus, _ = newTestLogger()
	if tl == nil {
		t.Fatal("функция New(), вернулся nil, ожидался не nil объект")
	}
	if tl.Errors() == nil {
		t.Error("функция Errors(), вернулся nil, ожидался не nil объект")
	}
	if tl.Debug(true) != Logger(tl) {
		t.Error("функция Debug(), вернулся не верный объект")
	}
	if !tl.debug {
		t.Error("функция Debug(true), флаг отладки не установлен")
	}
	tl.Debug(false)
	if tl.debug {
		t.Error("функция Debug(false), флаг отладки не сброшен")
	}
	if bus == nil {
		t.Error("тестовая шина данных не создана")
	}
}

func TestMessageGetPut(t *testing.T) {
	var (
		tl  *logger
		msg *Message
	)

	tl, _, _ = newTestLogger()
	// Получение объекта из бассейна.
	if msg = tl.MessageGet(); msg == nil {
		t.Fatal("функция MessageGet(), вернулся nil, ожидался не nil объект")
	}
	if msg.Pattern == nil || msg.Argument == nil || msg.Keys == nil || msg.Trace == nil || msg.Done == nil {
		t.Error("функция MessageGet(), объект сообщения содержит не инициализированные поля")
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
		t.Error("функция MessagePut(), объект сообщения не сброшен")
	}
	// Возврат nil в бассейн - паники нет.
	tl.MessagePut(nil)
}

func TestInitialization(t *testing.T) {
	var (
		tl  *logger
		bus kitModuleBus.Interface
		err error
	)

	tl, bus, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf("функция Initialization(), вернула не ожидаемую ошибку: %s", err)
	}
	// Лог зарегистрирован в шине данных как потребитель сообщений.
	if !bus.IsSubscriber(new(Message)) {
		t.Error("после Initialization(), лог не зарегистрирован в шине данных")
	}
	tl.FlushAndClose()
	// После FlushAndClose() регистрация в шине данных удалена.
	if bus.IsSubscriber(new(Message)) {
		t.Error("после FlushAndClose(), лог остался зарегистрирован в шине данных")
	}
}

func TestMessageAfterClose(t *testing.T) {
	var (
		tl  *logger
		tsw *testSyncWriter
		msg *Message
	)

	tl, _, tsw = newTestLogger()
	tl.FlushAndClose()
	msg = tl.MessageGet()
	msg.Pattern.WriteString("после закрытия")
	size := tsw.Len()
	tl.Message(msg)
	if tsw.Len() != size {
		t.Error("функция Message() после FlushAndClose(), сообщение было обработано")
	}
}

func TestMessageDefaultHandler(t *testing.T) {
	var (
		tl  *logger
		msg *Message
	)

	tl, _, tsw := newTestLogger()
	// Сообщение без шаблона - аргументы печатаются через Fprintln().
	msg = tl.MessageGet()
	msg.Argument = append(msg.Argument, "тестовое сообщение")
	tl.Message(msg)
	if !strings.Contains(tsw.String(), "тестовое сообщение") {
		t.Errorf("сообщение без шаблона не напечатано, лог: %q", tsw.String())
	}
	// Сообщение с шаблоном и аргументами.
	tsw.Reset()
	msg = tl.MessageGet()
	msg.Pattern.WriteString("шаблон %q")
	msg.Argument = append(msg.Argument, "аргумент")
	tl.Message(msg)
	if !strings.Contains(tsw.String(), `шаблон "аргумент"`) {
		t.Errorf("сообщение с шаблоном не напечатано, лог: %q", tsw.String())
	}
}

func TestHandlerSubscribeUnsubscribe(t *testing.T) {
	var (
		tl  *logger
		err error
	)

	tl, _, _ = newTestLogger()
	if err = tl.HandlerSubscribe(testHandlerSubscribeOne, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf("функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s", err)
	}
	// Повторная регистрация того же обработчика - ошибка.
	if err = tl.HandlerSubscribe(testHandlerSubscribeOne, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err == nil ||
		!strings.Contains(err.Error(), "уже зарегистрирован") {
		t.Errorf("функция HandlerSubscribe() повторной регистрации, вернула не ожидаемую ошибку: %v", err)
	}
	if err = tl.HandlerSubscribe(testHandlerSubscribeTwo, kitModuleLogLevel.Error, kitModuleLogLevel.Error); err != nil {
		t.Fatalf("функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s", err)
	}
	if tl.subscribers.Len() != 2 {
		t.Errorf("количество обработчиков %d, ожидалось 2", tl.subscribers.Len())
	}
	// Удаление регистрации обработчика.
	if err = tl.HandlerUnsubscribe(testHandlerSubscribeOne); err != nil {
		t.Fatalf("функция HandlerUnsubscribe(), вернула не ожидаемую ошибку: %s", err)
	}
	if tl.subscribers.Len() != 1 {
		t.Errorf("количество обработчиков %d, ожидалось 1", tl.subscribers.Len())
	}
	// Повторное удаление регистрации - ошибка.
	if err = tl.HandlerUnsubscribe(testHandlerSubscribeOne); err == nil ||
		!strings.Contains(err.Error(), "не найдена") {
		t.Errorf("функция HandlerUnsubscribe() повторного вызова, вернула не ожидаемую ошибку: %v", err)
	}
	// Удаление не зарегистрированного обработчика - ошибка.
	if err = tl.HandlerUnsubscribe(testHandlerSubscribeOne); err == nil {
		t.Error("функция HandlerUnsubscribe() не зарегистрированного обработчика, ошибка не возвращена")
	}
	if err = tl.HandlerUnsubscribe(testHandlerSubscribeTwo); err != nil {
		t.Fatalf("функция HandlerUnsubscribe(), вернула не ожидаемую ошибку: %s", err)
	}
}

// testHandlerSubscribeOne Тестовый обработчик сообщений лога.
func testHandlerSubscribeOne(_ *Message) {}

// testHandlerSubscribeTwo Тестовый обработчик сообщений лога.
func testHandlerSubscribeTwo(_ *Message) {}

func TestFlushAndCloseUnsubscribeError(t *testing.T) {
	var (
		tl  *logger
		bus kitModuleBus.Interface
		tsw *testSyncWriter
		err error
	)

	tl, bus, tsw = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf("функция Initialization(), вернула не ожидаемую ошибку: %s", err)
	}
	// Удаление регистрации в шине данных до FlushAndClose(), ошибка отписывания
	// печатается в лог, функция завершается без паники.
	if err = bus.Unsubscribe(tl); err != nil {
		t.Fatalf("функция Unsubscribe(), вернула не ожидаемую ошибку: %s", err)
	}
	tl.FlushAndClose()
	if tsw.Len() == 0 {
		t.Error("ошибка отписывания от шины данных не напечатана в лог")
	}
}

func TestInitializationSubscribeError(t *testing.T) {
	var (
		tl  Logger
		err error
	)

	tl = New(&testSyncWriter{}, &testBusSubscribeError{})
	if err = tl.Initialization(); err == nil ||
		!strings.Contains(err.Error(), "тестовая ошибка регистрации") {
		t.Errorf("функция Initialization() с ошибкой регистрации, вернула не ожидаемую ошибку: %v", err)
	}
}

// TestSubscribersWrongElementType moved to receiver_test.go: обработка элементов списка обработчиков
// не верного типа выполняется в Consumer().

// TestMessageFatalityExit Фатальное сообщение приводит к завершению приложения с кодом 252.
func TestMessageFatalityExit(t *testing.T) {
	if os.Getenv("GO_WANT_LOG_FATAL_EXIT") == "1" {
		var tl *logger

		tl, _, _ = newTestLogger()
		if err := tl.Initialization(); err != nil {
			t.Fatalf("функция Initialization(), вернула не ожидаемую ошибку: %s", err)
		}
		// Обработчик прерывается паникой - ошибки синхронной публикации печатаются перед завершением.
		_ = tl.HandlerSubscribe(func(_ *Message) { panic("паника обработчика") },
			kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace)
		// Вызов Write из функции с именем содержащим log.fatal - уровень Fatal и флаг фатальности.
		log{}.fatal(tl)
		return
	}
	var err error
	cmd := exec.Command(os.Args[0], "-test.run=TestMessageFatalityExit")
	cmd.Env = append(os.Environ(), "GO_WANT_LOG_FATAL_EXIT=1")
	if err = cmd.Run(); err == nil {
		t.Fatal("дочерний процесс завершился без ошибки, ожидалось завершение с кодом 252")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != int(eApplicationFatality) {
		t.Errorf("дочерний процесс завершился с ошибкой %v, ожидался код выхода %d", err, eApplicationFatality)
	}
}

// TestMessagePublishAsyncErrorExit Ошибка асинхронной публикации в шину данных приводит к завершению приложения.
func TestMessagePublishAsyncErrorExit(t *testing.T) {
	if os.Getenv("GO_WANT_LOG_ASYNC_ERROR_EXIT") == "1" {
		var tl = New(&testSyncWriter{}, &testBusAsyncError{})

		tl.Message(tl.MessageGet())
		return
	}
	var err error
	cmd := exec.Command(os.Args[0], "-test.run=TestMessagePublishAsyncErrorExit")
	cmd.Env = append(os.Environ(), "GO_WANT_LOG_ASYNC_ERROR_EXIT=1")
	if err = cmd.Run(); err == nil {
		t.Fatal("дочерний процесс завершился без ошибки, ожидалось завершение с кодом 252")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != int(eApplicationFatality) {
		t.Errorf("дочерний процесс завершился с ошибкой %v, ожидался код выхода %d", err, eApplicationFatality)
	}
}

// TestWriteStdLogInterception Перехват стандартного лога.
func TestWriteStdLogInterception(t *testing.T) {
	var (
		tl  *logger
		tsw *testSyncWriter
		err error
	)

	tl, _, tsw = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf("функция Initialization(), вернула не ожидаемую ошибку: %s", err)
	}
	standardLogPrint("сообщение стандартного лога")
	waitFor(t, func() bool {
		return strings.Contains(tsw.String(), "сообщение стандартного лога")
	})
	tl.FlushAndClose()
}

func TestFlushAndCloseWaitInflight(t *testing.T) {
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
		t.Fatal("превышено время ожидания FlushAndClose")
	}
}

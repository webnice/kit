package bus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type busTestData struct {
	Value string
}

type busTestOtherData struct {
	Value string
}

type busTestPanicData struct {
	Value string
}

type testSubscriber struct {
	types    []any
	known    func() []any
	consumer func(isSync bool, data any) (ret []any, errs []error)
}

func (tsb *testSubscriber) Consumer(isSync bool, data any) (ret []any, errs []error) {
	if tsb.consumer != nil {
		return tsb.consumer(isSync, data)
	}

	return nil, nil
}

func (tsb *testSubscriber) KnownType() (ret []any) {
	if tsb.known != nil {
		return tsb.known()
	}

	return tsb.types
}

type testAnySubscriber struct{ testSubscriber }

type testPanicSubscriber struct{ testSubscriber }

type testSlowSubscriber struct{ testSubscriber }

type testNilSubscriber struct{}

func (tns *testNilSubscriber) Consumer(bool, any) ([]any, []error) { return nil, nil }

func (tns *testNilSubscriber) KnownType() []any { return nil }

func newTestBus(t *testing.T, workerCount int, debug bool) Interface {
	t.Helper()
	var bus Interface

	bus = New(16, workerCount, debug)
	t.Cleanup(func() { bus.Gist().WorkerStop() })

	return bus
}

func waitTestSignal(t *testing.T, ch <-chan struct{}, timeout time.Duration, message string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatal(message)
	}
}

// TestBusSubscribeUnsubscribeErrors Ошибки подписки и отписки.
func TestBusSubscribeUnsubscribeErrors(t *testing.T) {
	const (
		errGistNil              = "Gist() вернул nil."
		errErrorsNotPackage     = "Errors() вернул не справочник пакета."
		errSubscribeNilNoError  = "Subscribe(nil) вернул nil ошибку."
		tplSubscribeError       = "Subscribe() вернул ошибку: %v."
		errIsSubscriberNotFound = "IsSubscriber() не нашёл подписчика."
		tplUnsubscribeError     = "Unsubscribe() вернул ошибку: %v."
		errIsSubscriberFound    = "IsSubscriber() нашёл удалённого подписчика."
		errUnsubscribeNilNoErr  = "Unsubscribe() для отсутствующего подписчика вернул nil ошибку."
		tplSubscribePanic       = "Subscribe() с паникой KnownType вернул %v."
		tplSubscribeNilType     = "Subscribe() с nil типом вернул %v."
		tplUnsubscribeTypedNil  = "Unsubscribe() с typed nil вернул %v."
		errTextKnownPanic       = "known panic"
	)
	var (
		bus               Interface
		err               error
		subscriber        *testSubscriber
		panicSubscriber   *testSubscriber
		nilTypeSubscriber *testSubscriber
		nilSubscriber     *testNilSubscriber
	)

	bus = newTestBus(t, 2, true)
	if bus.Gist() == nil {
		t.Fatal(errGistNil)
	}
	if bus.Errors() != Errors() {
		t.Fatal(errErrorsNotPackage)
	}
	if err = bus.Subscribe(nil); err == nil {
		t.Fatal(errSubscribeNilNoError)
	}
	subscriber = &testSubscriber{types: []any{busTestData{}}}
	if err = bus.Subscribe(subscriber); err != nil {
		t.Fatalf(tplSubscribeError, err)
	}
	if !bus.IsSubscriber(busTestData{}) {
		t.Fatal(errIsSubscriberNotFound)
	}
	if err = bus.Unsubscribe(&testSubscriber{}); err != nil {
		t.Fatalf(tplUnsubscribeError, err)
	}
	if bus.IsSubscriber(busTestData{}) {
		t.Fatal(errIsSubscriberFound)
	}
	if err = bus.Unsubscribe(&testSubscriber{}); err == nil {
		t.Fatal(errUnsubscribeNilNoErr)
	}
	panicSubscriber = &testSubscriber{known: func() []any { panic(errTextKnownPanic) }}
	if err = bus.Subscribe(panicSubscriber); err == nil || !strings.Contains(err.Error(), "паникой") {
		t.Fatalf(tplSubscribePanic, err)
	}
	nilTypeSubscriber = &testSubscriber{types: []any{nil}}
	if err = bus.Subscribe(nilTypeSubscriber); err == nil || !strings.Contains(err.Error(), "паникой") {
		t.Fatalf(tplSubscribeNilType, err)
	}
	nilSubscriber = nil
	if err = bus.Unsubscribe(nilSubscriber); err == nil || !strings.Contains(err.Error(), "паникой") {
		t.Fatalf(tplUnsubscribeTypedNil, err)
	}
}

// TestBusPublishAsync Асинхронная публикация данных.
func TestBusPublishAsync(t *testing.T) {
	const (
		errTextUnexpectedData  = "неожиданные данные"
		errTextConsumerPanic   = "consumer panic"
		errNoSubscriberNoError = "PublishAsync() без подписчика вернул nil ошибку."
		tplSubscribeError      = "Subscribe() вернул ошибку: %v."
		tplSubscribeAnyError   = "Subscribe(any) вернул ошибку: %v."
		tplSubscribePanicErr   = "Subscribe(panic) вернул ошибку: %v."
		tplPublishError        = "PublishAsync() вернул ошибку: %v."
		errAsyncNoSignal       = "async subscriber не получил данные."
		tplPublishAnyError     = "PublishAsync() any тип вернул ошибку: %v."
		errPublishNilNoError   = "PublishAsync(nil) вернул nil ошибку."
		tplPublishPanicError   = "PublishAsync() panic subscriber вернул ошибку: %v."
	)
	var (
		bus       Interface
		err       error
		called    chan struct{}
		anyCalled chan struct{}
		sub       *testSubscriber
		anySub    *testAnySubscriber
		panicSub  *testPanicSubscriber
	)

	bus = newTestBus(t, 2, false)
	called = make(chan struct{}, 1)
	sub = &testSubscriber{
		types: []any{busTestData{}},
		consumer: func(_ bool, data any) ([]any, []error) {
			if data.(busTestData).Value != "async" {
				return nil, []error{errors.New(errTextUnexpectedData)}
			}
			called <- struct{}{}

			return nil, nil
		},
	}
	anyCalled = make(chan struct{}, 10)
	anySub = &testAnySubscriber{testSubscriber{
		consumer: func(_ bool, _ any) ([]any, []error) {
			anyCalled <- struct{}{}

			return nil, nil
		},
	}}
	panicSub = &testPanicSubscriber{testSubscriber{
		types: []any{busTestPanicData{}},
		consumer: func(_ bool, _ any) ([]any, []error) {
			panic(errTextConsumerPanic)
		},
	}}
	if err = bus.PublishAsync(busTestOtherData{}); err == nil {
		t.Fatal(errNoSubscriberNoError)
	}
	if err = bus.Subscribe(sub); err != nil {
		t.Fatalf(tplSubscribeError, err)
	}
	if err = bus.Subscribe(anySub); err != nil {
		t.Fatalf(tplSubscribeAnyError, err)
	}
	if err = bus.Subscribe(panicSub); err != nil {
		t.Fatalf(tplSubscribePanicErr, err)
	}
	if err = bus.PublishAsync(busTestData{Value: "async"}); err != nil {
		t.Fatalf(tplPublishError, err)
	}
	waitTestSignal(t, called, time.Second, errAsyncNoSignal)
	if err = bus.PublishAsync(123); err != nil {
		t.Fatalf(tplPublishAnyError, err)
	}
	waitTestSignal(t, anyCalled, time.Second, "any subscriber не получил данные.")
	if err = bus.PublishAsync(nil); err == nil {
		t.Fatal(errPublishNilNoError)
	}
	if err = bus.PublishAsync(busTestPanicData{}); err != nil {
		t.Fatalf(tplPublishPanicError, err)
	}
	time.Sleep(50 * time.Millisecond)
}

// TestBusPublishSync Синхронная публикация данных.
func TestBusPublishSync(t *testing.T) {
	const (
		errTextExpectedSync      = "ожидалась синхронная публикация"
		errTextConsumerError     = "consumer error"
		errTextSyncConsumerPanic = "sync consumer panic"
		tplRespValue             = "resp:%s"
		errNoSubscriberNoErr     = "PublishSync() без подписчика не вернул ошибку."
		tplSubscribeError        = "Subscribe() вернул ошибку: %v."
		tplSubscribePanicErr     = "Subscribe(panic) вернул ошибку: %v."
		tplSubscribeEmptyErr     = "Subscribe(empty) вернул ошибку: %v."
		tplPublishSyncResult     = "PublishSync() = %v, %v."
		tplPublishCtxResult      = "PublishSyncWithContext() = %v, %v."
		tplPublishTimeoutRes     = "PublishSyncWithTimeout() = %v, %v."
		tplPublishEmptyResult    = "PublishSyncWithContext() с пустым результатом = %v, %v."
		errPublishNilNoErr       = "PublishSync(nil) не вернул ошибку."
		tplPublishPanicResult    = "PublishSync() panic subscriber = %v."
	)
	var (
		bus      Interface
		err      error
		sub      *testSubscriber
		panicSub *testPanicSubscriber
		emptySub *testSubscriber
		ret      []any
		errs     []error
	)

	bus = newTestBus(t, 2, true)
	sub = &testSubscriber{
		types: []any{busTestData{}},
		consumer: func(isSync bool, data any) ([]any, []error) {
			if !isSync {
				return nil, []error{errors.New(errTextExpectedSync)}
			}

			return []any{fmt.Sprintf(tplRespValue, data.(busTestData).Value)}, []error{errors.New(errTextConsumerError)}
		},
	}
	panicSub = &testPanicSubscriber{testSubscriber{
		types: []any{busTestPanicData{}},
		consumer: func(_ bool, _ any) ([]any, []error) {
			panic(errTextSyncConsumerPanic)
		},
	}}
	emptySub = &testSubscriber{types: []any{busTestOtherData{}}}
	if _, errs = bus.PublishSync(busTestOtherData{}); len(errs) == 0 {
		t.Fatal(errNoSubscriberNoErr)
	}
	if err = bus.Subscribe(sub); err != nil {
		t.Fatalf(tplSubscribeError, err)
	}
	if err = bus.Subscribe(panicSub); err != nil {
		t.Fatalf(tplSubscribePanicErr, err)
	}
	if err = bus.Subscribe(emptySub); err != nil {
		t.Fatalf(tplSubscribeEmptyErr, err)
	}
	ret, errs = bus.PublishSync(busTestData{Value: "one"})
	if len(ret) != 1 || ret[0] != "resp:one" || len(errs) != 1 || errs[0].Error() != errTextConsumerError {
		t.Fatalf(tplPublishSyncResult, ret, errs)
	}
	ret, errs = bus.PublishSyncWithContext(context.Background(), busTestData{Value: "ctx"})
	if len(ret) != 1 || ret[0] != "resp:ctx" || len(errs) != 1 {
		t.Fatalf(tplPublishCtxResult, ret, errs)
	}
	ret, errs = bus.PublishSyncWithTimeout(time.Second, busTestData{Value: "timeout"})
	if len(ret) != 1 || ret[0] != "resp:timeout" || len(errs) != 1 {
		t.Fatalf(tplPublishTimeoutRes, ret, errs)
	}
	ret, errs = bus.PublishSyncWithContext(context.Background(), busTestOtherData{})
	if len(ret) != 0 || len(errs) != 0 {
		t.Fatalf(tplPublishEmptyResult, ret, errs)
	}
	_, errs = bus.PublishSync(nil)
	if len(errs) == 0 {
		t.Fatal(errPublishNilNoErr)
	}
	_, errs = bus.PublishSync(busTestPanicData{})
	if len(errs) == 0 || !strings.Contains(errs[0].Error(), "паникой") {
		t.Fatalf(tplPublishPanicResult, errs)
	}
}

// TestBusPublishSyncTimeoutSlowConsumer Тайм-аут синхронной публикации при медленном потребителе.
func TestBusPublishSyncTimeoutSlowConsumer(t *testing.T) {
	const (
		tplSubscribeError = "Subscribe() вернул ошибку: %v."
		tplSlowConsumer   = "PublishSyncWithTimeout() с медленным потребителем = %v."
	)
	var (
		bus  Interface
		err  error
		errs []error
		sub  *testSlowSubscriber
	)

	bus = newTestBus(t, 1, false)
	sub = &testSlowSubscriber{testSubscriber{
		types: []any{busTestData{}},
		consumer: func(_ bool, _ any) ([]any, []error) {
			time.Sleep(250 * time.Millisecond)

			return []any{"late"}, nil
		},
	}}
	if err = bus.Subscribe(sub); err != nil {
		t.Fatalf(tplSubscribeError, err)
	}
	_, errs = bus.PublishSyncWithTimeout(20*time.Millisecond, busTestData{Value: "slow"})
	if len(errs) == 0 || !errors.Is(errs[0], context.DeadlineExceeded) {
		t.Fatalf(tplSlowConsumer, errs)
	}
	// Даём медленной горутине-потребителю завершить отправку ответа после тайм-аута.
	// Старый код закрывал ssRsp до этого момента и падал с panic: send on closed channel.
	time.Sleep(300 * time.Millisecond)
}

// TestBusNewWithDefaultBuffer Конструктор с нулевым размером буфера по умолчанию.
func TestBusNewWithDefaultBuffer(t *testing.T) {
	const errNotFullyInitialized = "New(0, 1, true) вернул не полностью инициализированный объект."
	var bus Interface

	bus = New(0, 1, true)
	t.Cleanup(func() { bus.Gist().WorkerStop() })
	if bus == nil || bus.Gist() == nil {
		t.Fatal(errNotFullyInitialized)
	}
}

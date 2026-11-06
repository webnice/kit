package bus

import (
	"context"
	"errors"
	"reflect"
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
	var bus = New(16, workerCount, debug)
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

func TestSubscribers(t *testing.T) {
	var (
		sss       = newSubscribers()
		bad       = "bad element"
		dataType  = reflect.TypeOf(busTestData{})
		otherType = reflect.TypeOf(busTestOtherData{})
		typed     = &subscriber{
			Name: "typed",
			Item: &testSubscriber{},
			Types: []*subscriberTypeInfo{{
				Original:     busTestData{},
				OriginalType: dataType,
			}},
		}
		anySubscriber = &subscriber{Name: "any", Item: &testAnySubscriber{}}
	)

	if sss.Len() != 0 {
		t.Fatal("newSubscribers() вернул не пустой объект")
	}
	sss.items.PushBack(bad)
	if sss.Len() != 1 {
		t.Fatalf("Len() после добавления служебного элемента = %d, ожидался 1", sss.Len())
	}
	sss.Store(typed)
	if sss.Len() != 2 {
		t.Fatalf("Len() = %d, ожидался 2", sss.Len())
	}
	if !sss.IsExistSubscriber(dataType) {
		t.Fatal("IsExistSubscriber() не нашёл подписчика типа")
	}
	if sss.IsExistSubscriber(otherType) {
		t.Fatal("IsExistSubscriber() нашёл подписчика для незарегистрированного типа")
	}
	if got := sss.GetSubscriber(dataType); len(got) != 1 || got[0] != typed {
		t.Fatalf("GetSubscriber(dataType) = %+v", got)
	}
	if got := sss.GetSubscriber(otherType); len(got) != 0 {
		t.Fatalf("GetSubscriber(otherType) = %+v", got)
	}

	sss.Store(anySubscriber)
	if !sss.IsExistSubscriber(otherType) {
		t.Fatal("IsExistSubscriber() не учёл any-подписчика")
	}
	if got := sss.GetSubscriber(dataType); len(got) != 2 {
		t.Fatalf("GetSubscriber(dataType) после any = %+v", got)
	}
	if got := sss.GetSubscriber(otherType); len(got) != 1 || got[0] != anySubscriber {
		t.Fatalf("GetSubscriber(otherType) после any = %+v", got)
	}

	got := sss.GetSubscriber(dataType)
	got = append(got, anySubscriber)
	if len(sss.GetSubscriber(dataType)) != 2 {
		t.Fatal("GetSubscriber() вернул срез-алиас внутренней карты")
	}

	if err := sss.Delete("unknown"); err == nil {
		t.Fatal("Delete() для отсутствующего подписчика вернул nil ошибку")
	}
	if err := sss.Delete("TYPED"); err != nil {
		t.Fatalf("Delete() вернул ошибку: %v", err)
	}
	if got := sss.GetSubscriber(dataType); len(got) != 1 || got[0] != anySubscriber {
		t.Fatalf("GetSubscriber(dataType) после удаления typed = %+v", got)
	}
	if err := sss.Delete("any"); err != nil {
		t.Fatalf("Delete() any вернул ошибку: %v", err)
	}
	if sss.IsExistSubscriber(dataType) {
		t.Fatal("IsExistSubscriber() нашёл удалённых подписчиков")
	}
	if sss.IsExistSubscriber(otherType) {
		t.Fatal("IsExistSubscriber() нашёл удалённого any-подписчика")
	}
}

func TestUtilities(t *testing.T) {
	var data busTestData
	if name := getFuncFullName(&data); !strings.Contains(name, "busTestData") {
		t.Fatalf("getFuncFullName(&data) = %q", name)
	}
	if name := getFuncFullName(&struct{}{}); name != "" {
		t.Fatalf("getFuncFullName(&struct{}{}) = %q", name)
	}
	if name := getFuncFullName(&[]int{}); name == "" {
		t.Fatal("getFuncFullName(&[]int{}) вернул пустое имя")
	}
	fn := func() {}
	if name := getFuncFullName(fn); !strings.Contains(name, "TestUtilities") {
		t.Fatalf("getFuncFullName(func) = %q", name)
	}
	if got := indirect(reflect.ValueOf(&data)); got.Interface().(busTestData) != data {
		t.Fatalf("indirect() = %+v", got)
	}
	if got := indirectType(reflect.TypeOf(&data)); got != reflect.TypeOf(data) {
		t.Fatalf("indirectType() = %v", got)
	}
	sti, err := makeSubscriberTypeInfo(reflect.TypeOf(&data))
	if err != nil {
		t.Fatalf("makeSubscriberTypeInfo() вернул ошибку: %v", err)
	}
	if sti.BaseType != reflect.TypeOf(data) || sti.OriginalType != reflect.TypeOf(&data) || sti.IndirectionCount != 1 {
		t.Fatalf("makeSubscriberTypeInfo() = %+v", sti)
	}
	for _, rt := range []reflect.Type{reflect.TypeOf(data), reflect.TypeOf(struct{}{}), reflect.TypeOf(&struct{}{})} {
		if sti, err = makeSubscriberTypeInfo(rt); err != nil || sti.OriginalType != rt {
			t.Fatalf("makeSubscriberTypeInfo(%v) = %+v, %v", rt, sti, err)
		}
	}
	ch := make(chan struct{})
	safeCloseSignalChannel(ch)
	safeCloseSignalChannel(ch)
}

func TestBusSubscribeUnsubscribeErrors(t *testing.T) {
	var bus = newTestBus(t, 2, true)

	if bus.Gist() == nil {
		t.Fatal("Gist() вернул nil")
	}
	if bus.Errors() != Errors() {
		t.Fatal("Errors() вернул не справочник пакета")
	}
	if err := bus.Subscribe(nil); err == nil {
		t.Fatal("Subscribe(nil) вернул nil ошибку")
	}
	subscriber := &testSubscriber{types: []any{busTestData{}}}
	if err := bus.Subscribe(subscriber); err != nil {
		t.Fatalf("Subscribe() вернул ошибку: %v", err)
	}
	if !bus.IsSubscriber(busTestData{}) {
		t.Fatal("IsSubscriber() не нашёл подписчика")
	}
	if err := bus.Unsubscribe(&testSubscriber{}); err != nil {
		t.Fatalf("Unsubscribe() вернул ошибку: %v", err)
	}
	if bus.IsSubscriber(busTestData{}) {
		t.Fatal("IsSubscriber() нашёл удалённого подписчика")
	}
	if err := bus.Unsubscribe(&testSubscriber{}); err == nil {
		t.Fatal("Unsubscribe() для отсутствующего подписчика вернул nil ошибку")
	}

	panicSubscriber := &testSubscriber{known: func() []any { panic("known panic") }}
	if err := bus.Subscribe(panicSubscriber); err == nil || !strings.Contains(err.Error(), "паникой") {
		t.Fatalf("Subscribe() с паникой KnownType вернул %v", err)
	}
	nilTypeSubscriber := &testSubscriber{types: []any{nil}}
	if err := bus.Subscribe(nilTypeSubscriber); err == nil || !strings.Contains(err.Error(), "паникой") {
		t.Fatalf("Subscribe() с nil типом вернул %v", err)
	}
	var nilSubscriber *testNilSubscriber
	if err := bus.Unsubscribe(nilSubscriber); err == nil || !strings.Contains(err.Error(), "паникой") {
		t.Fatalf("Unsubscribe() с typed nil вернул %v", err)
	}
}

func TestBusPublishAsync(t *testing.T) {
	var (
		bus    = newTestBus(t, 2, false)
		called = make(chan struct{}, 1)
		sub    = &testSubscriber{
			types: []any{busTestData{}},
			consumer: func(_ bool, data any) ([]any, []error) {
				if data.(busTestData).Value != "async" {
					return nil, []error{errors.New("unexpected data")}
				}
				called <- struct{}{}
				return nil, nil
			},
		}
		anyCalled = make(chan struct{}, 10)
		anySub    = &testAnySubscriber{testSubscriber{
			consumer: func(_ bool, _ any) ([]any, []error) {
				anyCalled <- struct{}{}
				return nil, nil
			},
		}}
		panicSub = &testPanicSubscriber{testSubscriber{
			types: []any{busTestPanicData{}},
			consumer: func(_ bool, _ any) ([]any, []error) {
				panic("consumer panic")
			},
		}}
	)

	if err := bus.PublishAsync(busTestOtherData{}); err == nil {
		t.Fatal("PublishAsync() без подписчика вернул nil ошибку")
	}
	if err := bus.Subscribe(sub); err != nil {
		t.Fatalf("Subscribe() вернул ошибку: %v", err)
	}
	if err := bus.Subscribe(anySub); err != nil {
		t.Fatalf("Subscribe(any) вернул ошибку: %v", err)
	}
	if err := bus.Subscribe(panicSub); err != nil {
		t.Fatalf("Subscribe(panic) вернул ошибку: %v", err)
	}
	if err := bus.PublishAsync(busTestData{Value: "async"}); err != nil {
		t.Fatalf("PublishAsync() вернул ошибку: %v", err)
	}
	waitTestSignal(t, called, time.Second, "async subscriber не получил данные")
	if err := bus.PublishAsync(123); err != nil {
		t.Fatalf("PublishAsync() any тип вернул ошибку: %v", err)
	}
	waitTestSignal(t, anyCalled, time.Second, "any subscriber не получил данные")
	if err := bus.PublishAsync(nil); err == nil {
		t.Fatal("PublishAsync(nil) вернул nil ошибку")
	}
	if err := bus.PublishAsync(busTestPanicData{}); err != nil {
		t.Fatalf("PublishAsync() panic subscriber вернул ошибку: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
}

func TestBusPublishSync(t *testing.T) {
	var (
		bus = newTestBus(t, 2, true)
		sub = &testSubscriber{
			types: []any{busTestData{}},
			consumer: func(isSync bool, data any) ([]any, []error) {
				if !isSync {
					return nil, []error{errors.New("expected sync")}
				}
				return []any{"resp:" + data.(busTestData).Value}, []error{errors.New("consumer error")}
			},
		}
		panicSub = &testPanicSubscriber{testSubscriber{
			types: []any{busTestPanicData{}},
			consumer: func(_ bool, _ any) ([]any, []error) {
				panic("sync consumer panic")
			},
		}}
		emptySub = &testSubscriber{types: []any{busTestOtherData{}}}
	)

	if _, errs := bus.PublishSync(busTestOtherData{}); len(errs) == 0 {
		t.Fatal("PublishSync() без подписчика не вернул ошибку")
	}
	if err := bus.Subscribe(sub); err != nil {
		t.Fatalf("Subscribe() вернул ошибку: %v", err)
	}
	if err := bus.Subscribe(panicSub); err != nil {
		t.Fatalf("Subscribe(panic) вернул ошибку: %v", err)
	}
	if err := bus.Subscribe(emptySub); err != nil {
		t.Fatalf("Subscribe(empty) вернул ошибку: %v", err)
	}

	ret, errs := bus.PublishSync(busTestData{Value: "one"})
	if len(ret) != 1 || ret[0] != "resp:one" || len(errs) != 1 || errs[0].Error() != "consumer error" {
		t.Fatalf("PublishSync() = %v, %v", ret, errs)
	}
	ret, errs = bus.PublishSyncWithContext(context.Background(), busTestData{Value: "ctx"})
	if len(ret) != 1 || ret[0] != "resp:ctx" || len(errs) != 1 {
		t.Fatalf("PublishSyncWithContext() = %v, %v", ret, errs)
	}
	ret, errs = bus.PublishSyncWithTimeout(time.Second, busTestData{Value: "timeout"})
	if len(ret) != 1 || ret[0] != "resp:timeout" || len(errs) != 1 {
		t.Fatalf("PublishSyncWithTimeout() = %v, %v", ret, errs)
	}
	ret, errs = bus.PublishSyncWithContext(context.Background(), busTestOtherData{})
	if len(ret) != 0 || len(errs) != 0 {
		t.Fatalf("PublishSyncWithContext() с пустым результатом = %v, %v", ret, errs)
	}

	_, errs = bus.PublishSync(nil)
	if len(errs) == 0 {
		t.Fatal("PublishSync(nil) не вернул ошибку")
	}
	_, errs = bus.PublishSync(busTestPanicData{})
	if len(errs) == 0 || !strings.Contains(errs[0].Error(), "паникой") {
		t.Fatalf("PublishSync() panic subscriber = %v", errs)
	}
}

func TestBusPublishSyncTimeoutSlowConsumer(t *testing.T) {
	var (
		bus = newTestBus(t, 1, false)
		sub = &testSlowSubscriber{testSubscriber{
			types: []any{busTestData{}},
			consumer: func(_ bool, _ any) ([]any, []error) {
				time.Sleep(250 * time.Millisecond)
				return []any{"late"}, nil
			},
		}}
	)

	if err := bus.Subscribe(sub); err != nil {
		t.Fatalf("Subscribe() вернул ошибку: %v", err)
	}
	_, errs := bus.PublishSyncWithTimeout(20*time.Millisecond, busTestData{Value: "slow"})
	if len(errs) == 0 || !errors.Is(errs[0], context.DeadlineExceeded) {
		t.Fatalf("PublishSyncWithTimeout() с медленным потребителем = %v", errs)
	}
	// Даём медленной горутине-потребителю завершить отправку ответа после тайм-аута.
	// Старый код закрывал ssRsp до этого момента и падал с panic: send on closed channel.
	time.Sleep(300 * time.Millisecond)
}

func TestBusWorkerStopAndStatistic(t *testing.T) {
	var (
		bus    = New(8, 1, true)
		called = make(chan struct{}, 1)
		sub    = &testSubscriber{
			types: []any{busTestData{}},
			consumer: func(_ bool, _ any) ([]any, []error) {
				called <- struct{}{}
				return nil, nil
			},
		}
	)
	if err := bus.Subscribe(sub); err != nil {
		t.Fatalf("Subscribe() вернул ошибку: %v", err)
	}
	if stat := bus.Gist().Statistic(); stat == nil {
		t.Fatal("Statistic() для debug bus вернул nil")
	}
	bus.Gist().Debug(false)
	if stat := bus.Gist().Statistic(); stat != nil {
		t.Fatalf("Statistic() после Debug(false) вернул %+v", stat)
	}
	bus.Gist().WorkerStart(1) // Повторный запуск должен быть проигнорирован.
	if err := bus.PublishAsync(busTestData{}); err != nil {
		t.Fatalf("PublishAsync() вернул ошибку: %v", err)
	}
	bus.Gist().WorkerStop()
	waitTestSignal(t, called, time.Second, "WorkerStop() не дождался обработки данных")
	bus.Gist().WorkerStop() // Повторная остановка должна быть проигнорирована.

	var manual = &impl{}
	newEssence(manual).WorkerStop()

	noDebug := New(8, 1, false)
	t.Cleanup(func() { noDebug.Gist().WorkerStop() })
	if stat := noDebug.Gist().Statistic(); stat != nil {
		t.Fatalf("Statistic() для no-debug bus вернул %+v", stat)
	}
}

func TestBusNewWithDefaultBuffer(t *testing.T) {
	var bus = New(0, 1, true)
	t.Cleanup(func() { bus.Gist().WorkerStop() })
	if bus == nil || bus.Gist() == nil {
		t.Fatal("New(0, 1, true) вернул не полностью инициализированный объект")
	}
}

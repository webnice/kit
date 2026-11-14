package bus

import (
	kitModulePdw "github.com/webnice/kit/v4/module/pdw"
	"testing"
	"time"
)

// TestBusWorkerStopAndStatistic Остановка воркеров и статистика бассейна данных.
func TestBusWorkerStopAndStatistic(t *testing.T) {
	const (
		tplSubscribeError      = "Subscribe() вернул ошибку: %v."
		errStatisticNil        = "Statistic() для debug bus вернул nil."
		tplStatisticAfterDebug = "Statistic() после Debug(false) вернул %+v."
		tplPublishError        = "PublishAsync() вернул ошибку: %v."
		errWorkerNoWait        = "WorkerStop() не дождался обработки данных."
		tplStatisticNoDebug    = "Statistic() для no-debug bus вернул %+v."
	)
	var (
		bus     Interface
		called  chan struct{}
		sub     *testSubscriber
		manual  *impl
		noDebug Interface
		err     error
		stat    *kitModulePdw.Statistic
	)

	bus = New(8, 1, true)
	t.Cleanup(func() { bus.Gist().WorkerStop() })
	called = make(chan struct{}, 1)
	sub = &testSubscriber{
		types: []any{busTestData{}},
		consumer: func(_ bool, _ any) ([]any, []error) {
			called <- struct{}{}

			return nil, nil
		},
	}
	if err = bus.Subscribe(sub); err != nil {
		t.Fatalf(tplSubscribeError, err)
	}
	stat = bus.Gist().Statistic()
	if stat == nil {
		t.Fatal(errStatisticNil)
	}
	bus.Gist().Debug(false)
	if stat = bus.Gist().Statistic(); stat != nil {
		t.Fatalf(tplStatisticAfterDebug, stat)
	}
	bus.Gist().WorkerStart(1) // Повторный запуск должен быть проигнорирован.
	if err = bus.PublishAsync(busTestData{}); err != nil {
		t.Fatalf(tplPublishError, err)
	}
	bus.Gist().WorkerStop()
	waitTestSignal(t, called, time.Second, errWorkerNoWait)
	bus.Gist().WorkerStop() // Повторная остановка должна быть проигнорирована.
	manual = &impl{}
	newEssence(manual).WorkerStop()
	noDebug = New(8, 1, false)
	t.Cleanup(func() { noDebug.Gist().WorkerStop() })
	if stat = noDebug.Gist().Statistic(); stat != nil {
		t.Fatalf(tplStatisticNoDebug, stat)
	}
}

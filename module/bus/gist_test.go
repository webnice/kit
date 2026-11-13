package bus

import (
	"testing"
	"time"
)

// TestBusWorkerStopAndStatistic Остановка воркеров и статистика бассейна данных.
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

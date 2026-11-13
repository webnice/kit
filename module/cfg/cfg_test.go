package cfg

import (
	"testing"
	"time"
)

// testMainFn Тестовая main функция.
func testMainFn() (uint8, error) { return 0, nil }

// TestNew Конструктор объекта конфигурации.
func TestNew(t *testing.T) {
	var (
		obj Interface
		ok  bool
	)

	obj = Get()
	if obj == nil {
		t.Error("Function New() return nil")
		return
	}
	if _, ok = obj.(*impl); !ok {
		t.Error("Function New() return invalid object")
		return
	}
}

// TestRegistrationMainPanicUnlocksMutex Panic в RegistrationMain не оставляет мьютекс заблокированным.
func TestRegistrationMainPanicUnlocksMutex(t *testing.T) {
	var (
		oldFn    = singleton.main.Fn
		oldError = singleton.error
	)
	t.Cleanup(func() {
		singleton.main.Fn = oldFn
		singleton.error = oldError
	})

	singleton.main.Fn = testMainFn
	func() {
		defer func() { _ = recover() }()
		RegistrationMain(func() (uint8, error) { return 1, nil })
	}()

	done := make(chan struct{})
	go func() {
		singleton.main.FnMutex.Lock()
		singleton.main.FnMutex.Unlock()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("FnMutex остался заблокированным после panic в RegistrationMain")
	}
}

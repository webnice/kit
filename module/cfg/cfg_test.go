package cfg

import (
	"testing"
	"time"
)

// Тестовая main функция.
func testMainFn() (uint8, error) { return 0, nil }

// TestNew Конструктор объекта конфигурации.
func TestNew(t *testing.T) {
	const (
		errNewNil     = "функция New() вернула nil."
		errNewInvalid = "функция New() вернула неверный объект."
	)
	var (
		obj Interface
		ok  bool
	)

	obj = Get()
	if obj == nil {
		t.Error(errNewNil)

		return
	}
	if _, ok = obj.(*impl); !ok {
		t.Error(errNewInvalid)

		return
	}
}

// TestRegistrationMainPanicUnlocksMutex Panic в RegistrationMain не оставляет мьютекс заблокированным.
func TestRegistrationMainPanicUnlocksMutex(t *testing.T) {
	const (
		errMutexLocked = "FnMutex остался заблокированным после panic в RegistrationMain."
	)
	var (
		oldFn    = singleton.main.Fn
		oldError = singleton.error
		done     chan struct{}
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
	done = make(chan struct{})
	go func() {
		singleton.main.FnMutex.Lock()
		singleton.main.FnMutex.Unlock()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal(errMutexLocked)
	}
}

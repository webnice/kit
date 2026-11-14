package server

import (
	"strings"
	"testing"
)

func TestSafeCall(t *testing.T) {
	const (
		errSafeCallNoPanic   = "функция safeCall() без паники, вернула ошибку: %s."
		errSafeCallNotCalled = "функция safeCall() без паники, функция не была вызвана."
		errSafeCallPanic     = "функция safeCall() с паникой, вернула не ожидаемую ошибку: %v."
	)
	var (
		srv    Interface
		iwl    *implWebLib
		called bool
		err    error
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	// Вызов функции без паники.
	if err = iwl.safeCall(func() { called = true }); err != nil {
		t.Errorf(errSafeCallNoPanic, err)
	}
	if !called {
		t.Error(errSafeCallNotCalled)
	}
	// Вызов функции с паникой.
	err = iwl.safeCall(func() { panic("тестовая паника") })
	if err == nil || !strings.Contains(err.Error(), "тестовая паника") {
		t.Errorf(errSafeCallPanic, err)
	}
}

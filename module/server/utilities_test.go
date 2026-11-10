package server

import (
	"strings"
	"testing"
)

func TestSafeCall(t *testing.T) {
	var (
		srv  = New(&testLogger{})
		iwl  = srv.(*impl).serverWeb.lib
		called bool
	)

	// Вызов функции без паники.
	if err := iwl.safeCall(func() { called = true }); err != nil {
		t.Errorf("функция safeCall() без паники, вернула ошибку: %s", err)
	}
	if !called {
		t.Error("функция safeCall() без паники, функция не была вызвана")
	}
	// Вызов функции с паникой.
	err := iwl.safeCall(func() { panic("тестовая паника") })
	if err == nil || !strings.Contains(err.Error(), "тестовая паника") {
		t.Errorf("функция safeCall() с паникой, вернула не ожидаемую ошибку: %v", err)
	}
}

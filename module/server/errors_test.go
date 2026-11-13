package server

import (
	"strings"
	"testing"
)

// TestErrorsDictionary Справочник ошибок пакета.
func TestErrorsDictionary(t *testing.T) {
	var e = Errors()

	if e == nil {
		t.Fatal("функция Errors(), вернулся nil, ожидался не nil объект")
	}
	if e.ServerByIdNotFound == nil || e.ServersAddedNotEqualLaunched == nil ||
		e.ModulePanicException == nil || e.BeforeExitWithError == nil ||
		e.ServerByIdAlreadyStarted == nil || e.ServerByIdNotStarted == nil ||
		e.ServersNotRegistered == nil {
		t.Error("справочник ошибок содержит не инициализированные ошибки")
	}
	if err := e.ServersNotRegistered.Bind(); err == nil || !strings.Contains(err.Error(), "сервер") {
		t.Errorf("ошибка ServersNotRegistered.Bind(), вернулась не корректная ошибка: %v", err)
	}
	if err := e.ServersAddedNotEqualLaunched.Bind(1, 0); err == nil ||
		!strings.Contains(err.Error(), "1") || !strings.Contains(err.Error(), "0") {
		t.Errorf("ошибка ServersAddedNotEqualLaunched.Bind(), вернулась не корректная ошибка: %v", err)
	}
}

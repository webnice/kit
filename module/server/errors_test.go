package server

import (
	"strings"
	"testing"
)

// TestErrorsDictionary Справочник ошибок пакета.
func TestErrorsDictionary(t *testing.T) {
	const (
		errNilDictionary = "функция Errors(), вернулся nil, ожидался не nil объект."
		errNotInit       = "справочник ошибок содержит не инициализированные ошибки."
		tplBindWrong     = "ошибка %s, вернулась не корректная ошибка: %v."
	)
	var e *Error

	e = Errors()
	if e == nil {
		t.Fatal(errNilDictionary)
	}
	if e.ServerByIdNotFound == nil || e.ServersAddedNotEqualLaunched == nil ||
		e.ModulePanicException == nil || e.BeforeExitWithError == nil ||
		e.ServerByIdAlreadyStarted == nil || e.ServerByIdNotStarted == nil ||
		e.ServersNotRegistered == nil {
		t.Error(errNotInit)
	}
	if err := e.ServersNotRegistered.Bind(); err == nil || !strings.Contains(err.Error(), "сервер") {
		t.Errorf(tplBindWrong, "ServersNotRegistered.Bind()", err)
	}
	if err := e.ServersAddedNotEqualLaunched.Bind(1, 0); err == nil ||
		!strings.Contains(err.Error(), "1") || !strings.Contains(err.Error(), "0") {
		t.Errorf(tplBindWrong, "ServersAddedNotEqualLaunched.Bind(1, 0)", err)
	}
}

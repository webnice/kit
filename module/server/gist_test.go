package server

import "testing"

// TestGistDebug Управление флагом отладки через сущность пакета.
func TestGistDebug(t *testing.T) {
	const (
		errDebugNil      = "функция Debug(), вернулся nil, ожидался не nil объект."
		errDebugNotSet   = "функция Debug(true), флаг режима отладки не установлен."
		errDebugNotReset = "функция Debug(false), флаг режима отладки не сброшен."
	)
	var (
		srv Interface
		obj *impl
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	if essence := srv.Gist().Debug(true); essence == nil {
		t.Error(errDebugNil)
	}
	if !obj.debug {
		t.Error(errDebugNotSet)
	}
	srv.Gist().Debug(false)
	if obj.debug {
		t.Error(errDebugNotReset)
	}
}

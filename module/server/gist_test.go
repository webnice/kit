package server

import "testing"

// TestGistDebug Управление флагом отладки через сущность пакета.
func TestGistDebug(t *testing.T) {
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
	)

	if essence := srv.Gist().Debug(true); essence == nil {
		t.Error("функция Debug(), вернулся nil, ожидался не nil объект")
	}
	if !obj.debug {
		t.Error("функция Debug(true), флаг режима отладки не установлен")
	}
	srv.Gist().Debug(false)
	if obj.debug {
		t.Error("функция Debug(false), флаг режима отладки не сброшен")
	}
}

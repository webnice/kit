package log

import "testing"

// TestErrorsDictionary Справочник ошибок пакета.
func TestErrorsDictionary(t *testing.T) {
	var e = Errors()

	if e == nil {
		t.Fatal("функция Errors(), вернулся nil, ожидался не nil объект")
	}
	if e.LogPanicException == nil || e.HandlerAlreadySubscribed == nil || e.HandlerSubscriptionNotFound == nil {
		t.Error("справочник ошибок содержит не инициализированные ошибки")
	}
}

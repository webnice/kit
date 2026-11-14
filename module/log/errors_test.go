package log

import "testing"

// TestErrorsDictionary Справочник ошибок пакета.
func TestErrorsDictionary(t *testing.T) {
	const (
		errNilDictionary = "функция Errors(), вернулся nil, ожидался не nil объект."
		errNotInit       = "справочник ошибок содержит не инициализированные ошибки."
	)
	var e *Error

	e = Errors()
	if e == nil {
		t.Fatal(errNilDictionary)
	}
	if e.LogPanicException == nil || e.HandlerAlreadySubscribed == nil || e.HandlerSubscriptionNotFound == nil {
		t.Error(errNotInit)
	}
}

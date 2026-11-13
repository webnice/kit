package migration

import "testing"

// TestErrorsDictionary Справочник ошибок пакета.
func TestErrorsDictionary(t *testing.T) {
	var mmn = New()

	if mmn.Errors() != Errors() {
		t.Fatal("Errors() вернул не справочник ошибок пакета")
	}
	if mmn.Errors().DatabaseIsNotInUse == nil {
		t.Fatal("DatabaseIsNotInUse не определена")
	}
	if mmn.Errors().DatabaseUnexpectedError == nil {
		t.Fatal("DatabaseUnexpectedError не определена")
	}
}

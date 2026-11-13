package migration

import "testing"

// TestImpl_DbVersionTableName Имя таблицы версий базы данных.
func TestImpl_DbVersionTableName(t *testing.T) {
	if (&DbVersion{}).TableName() != "goose_db_version" {
		t.Fatal("TableName() != goose_db_version")
	}
}

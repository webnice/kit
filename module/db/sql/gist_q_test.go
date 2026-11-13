package sql

import (
	"errors"
	"strings"
	"testing"
)

func TestImplementation_Q(t *testing.T) {
	var implementation Implementation
	model := testModelGorm{}

	if got := implementation.QErr(errors.New("boom")); got == nil || !strings.Contains(got.Error(), "ошибка базы данных") {
		t.Fatalf("QErr() = %v", got)
	}
	if got := implementation.QFn(model, " field "); got != "`test_table`.`field`" {
		t.Fatalf("QFn() = %q", got)
	}
	if got := implementation.QAll(model); got != "`test_table`.*" {
		t.Fatalf("QAll() = %q", got)
	}
	if got := implementation.QAs(model, "field", "alias"); got != "`test_table`.`field` AS `alias`" {
		t.Fatalf("QAs() = %q", got)
	}
	if got := implementation.QJoin(model, model, "left_id", model, "right_id"); got != "LEFT OUTER JOIN `test_table` ON `test_table`.`left_id` = `test_table`.`right_id`" {
		t.Fatalf("QJoin() = %q", got)
	}
	if got := implementation.QJoinAs(model, "alias", model, "left_id", model, "right_id"); got != "LEFT OUTER JOIN `test_table` AS `alias` ON `test_table`.`left_id` = `test_table`.`right_id`" {
		t.Fatalf("QJoinAs() = %q", got)
	}
	if got := implementation.QJoinAs(model, "", model, "left_id", model, "right_id"); got != "LEFT OUTER JOIN `test_table` ON `test_table`.`left_id` = `test_table`.`right_id`" {
		t.Fatalf("QJoinAs(empty alias) = %q", got)
	}
	if got := implementation.QWh(model, "field", "IS NULL"); got != "`test_table`.`field` IS NULL" {
		t.Fatalf("QWh() = %q", got)
	}
}

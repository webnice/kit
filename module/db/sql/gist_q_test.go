package sql

import (
	"errors"
	"strings"
	"testing"
)

func TestImplementation_Q(t *testing.T) {
	const (
		errTextBoom     = "boom"
		tplQErr         = "QErr() = %v."
		tplQFn          = "QFn() = %q."
		tplQAll         = "QAll() = %q."
		tplQAs          = "QAs() = %q."
		tplQJoin        = "QJoin() = %q."
		tplQJoinAs      = "QJoinAs() = %q."
		tplQJoinAsEmpty = "QJoinAs(empty alias) = %q."
		tplQWh          = "QWh() = %q."
	)
	var implementation Implementation

	model := testModelGorm{}
	if got := implementation.QErr(errors.New(errTextBoom)); got == nil || !strings.Contains(got.Error(), "ошибка базы данных") {
		t.Fatalf(tplQErr, got)
	}
	if got := implementation.QFn(model, " field "); got != "`test_table`.`field`" {
		t.Fatalf(tplQFn, got)
	}
	if got := implementation.QAll(model); got != "`test_table`.*" {
		t.Fatalf(tplQAll, got)
	}
	if got := implementation.QAs(model, "field", "alias"); got != "`test_table`.`field` AS `alias`" {
		t.Fatalf(tplQAs, got)
	}
	if got := implementation.QJoin(model, model, "left_id", model, "right_id"); got != "LEFT OUTER JOIN `test_table` ON `test_table`.`left_id` = `test_table`.`right_id`" {
		t.Fatalf(tplQJoin, got)
	}
	if got := implementation.QJoinAs(model, "alias", model, "left_id", model, "right_id"); got != "LEFT OUTER JOIN `test_table` AS `alias` ON `test_table`.`left_id` = `test_table`.`right_id`" {
		t.Fatalf(tplQJoinAs, got)
	}
	if got := implementation.QJoinAs(model, "", model, "left_id", model, "right_id"); got != "LEFT OUTER JOIN `test_table` ON `test_table`.`left_id` = `test_table`.`right_id`" {
		t.Fatalf(tplQJoinAsEmpty, got)
	}
	if got := implementation.QWh(model, "field", "IS NULL"); got != "`test_table`.`field` IS NULL" {
		t.Fatalf(tplQWh, got)
	}
}

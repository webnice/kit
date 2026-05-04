package sql

import (
	"fmt"
	"strings"

	"github.com/webnice/dic"
)

// QErr Создание единообразной непредвиденной ошибки
func (db *Implementation) QErr(err error) error {
	const qError = "ошибка базы данных: %s"
	return dic.NewError(qError, "ошибка").Bind(err)
}

// QFn Создание строки с полным названием колонки таблицы вместе с названием таблицы.
//
// Пример: Qfn(&table_object, "field") -> `table_name`.`field`
func (db *Implementation) QFn(obj ModelGorm, field string) (ret string) {
	const tpl = "`%s`.`%s`"
	ret = fmt.Sprintf(tpl, obj.TableName(), strings.TrimSpace(field))
	return
}

// QAll Создание строки с полным названием таблицы и всеми колонками этой таблицы.
//
// Пример: QAll(&table_object) -> `table_name`.*
func (db *Implementation) QAll(obj ModelGorm) (ret string) {
	const tpl = "`%s`.*"
	ret = fmt.Sprintf(tpl, obj.TableName())
	return
}

// QAs Создание строки с полным названием колонки таблицы вместе с названием таблицы и новым названием.
//
// Пример: QAs(&table_object, "field", "new_name") -> `table_name`.`field` AS `new_name`
func (db *Implementation) QAs(obj ModelGorm, name string, as string) (ret string) {
	const tpl = "`%s`.`%s` AS `%s`"
	ret = fmt.Sprintf(tpl, obj.TableName(), name, as)
	return
}

// QJoin Создание строки запроса LEFT OUTER JOIN.
//
// Пример: QJoin(&table1_object, &table_object_1, "field_1", &table_object_2, "field_2") ->
// LEFT OUTER JOIN `table1_name` ON `table_object_1_name`.`field_1` = `table_object_2_name`.`field_2`
func (db *Implementation) QJoin(obj ModelGorm, o1 ModelGorm, f1 string, o2 ModelGorm, f2 string) (ret string) {
	const tplJn = "LEFT OUTER JOIN `%s` ON %s = %s"
	ret = fmt.Sprintf(tplJn, obj.TableName(), db.QFn(o1, f1), db.QFn(o2, f2))
	return
}

// QJoinAs Создание строки запроса LEFT OUTER JOIN с синонимом.
//
// Пример: QJoinAs(&table1_object, &table_object_1, "field_1", &table_object_2, "field_2") ->
// LEFT OUTER JOIN `table1_name` AS `table1_object_name` ON `table_object_1_name`.`field_1` =
// `table_object_2_name`.`field_2`
func (db *Implementation) QJoinAs(obj ModelGorm, as string, o1 ModelGorm, f1 string, o2 ModelGorm, f2 string) (ret string) {
	const (
		tplJn = "LEFT OUTER JOIN `%s`%s ON %s = %s"
		tplAs = " AS `%s`"
	)
	var (
		left  string
		right string
		alias string
	)

	left, right = db.QFn(o1, f1), db.QFn(o2, f2)
	if as != "" {
		alias = fmt.Sprintf(tplAs, as)
	}
	ret = fmt.Sprintf(tplJn, obj.TableName(), alias, left, right)

	return
}

// QWh Создание строки запроса для WHERE.
//
// Пример: QWh(&table_object, "field", "IS NULL") -> `table_object_name`.`field` IS NULL
func (db *Implementation) QWh(obj ModelGorm, name string, where string) (ret string) {
	ret = fmt.Sprintf("%s %s", db.QFn(obj, name), where)
	return
}

package sql

import (
	"fmt"
	"strings"

	"github.com/webnice/dic"
)

// QErr Создание единообразной непредвиденной ошибки
func (db *Implementation) QErr(err error) error {
	const qError = "ошибка базы данных: %w"
	return dic.NewError(qError, "ошибка").Bind(err)
}

// QFn Создание строки с полным именем колонки таблицы вместе с именем таблицы.
func (db *Implementation) QFn(obj ModelGorm, field string) (ret string) {
	const tpl = "`%s`.`%s`"
	ret = fmt.Sprintf(tpl, obj.TableName(), strings.TrimSpace(field))
	return
}

// QAs Создание строки с полным именем колонки таблицы вместе с именем таблицы и новым временным именем.
func (db *Implementation) QAs(obj ModelGorm, name string, as string) (ret string) {
	const tpl = "`%s`.`%s` AS `%s`"
	ret = fmt.Sprintf(tpl, obj.TableName(), name, as)
	return
}

// QJoin Создание строки запроса LEFT OUTER JOIN.
func (db *Implementation) QJoin(obj ModelGorm, o1 ModelGorm, f1 string, o2 ModelGorm, f2 string) (ret string) {
	const tplJn = "LEFT OUTER JOIN `%s` ON %s = %s"
	ret = fmt.Sprintf(tplJn, obj.TableName(), db.QFn(o1, f1), db.QFn(o2, f2))
	return
}

// QJoinAs Создание строки запроса LEFT OUTER JOIN.
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
func (db *Implementation) QWh(obj ModelGorm, name string, where string) (ret string) {
	ret = fmt.Sprintf("%s %s", db.QFn(obj, name), where)
	return
}

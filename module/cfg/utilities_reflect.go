// Package cfg
package cfg

import (
	"fmt"
	"reflect"
	runtimeDebug "runtime/debug"
	"strings"
)

func indirect(rv reflect.Value) reflect.Value {
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	return rv
}

func indirectType(rt reflect.Type) reflect.Type {
	for rt.Kind() == reflect.Ptr || rt.Kind() == reflect.Slice {
		rt = rt.Elem()
	}

	return rt
}

func reflectObject(c interface{}) (crv reflect.Value, crt reflect.Type, err error) {
	const (
		tplErrorNil        = "в функцию RegistrationConfiguration() передан nil объект"
		tplErrorNotValid   = "объект с типом %q не инициализирован"
		tplErrorNotAddress = "объект с типом %q передан не корректно. Необходимо передать адрес объекта"
	)

	defer func() {
		if e := recover(); e != nil {
			err = Errors().ConfigurationApplicationPanic(0, e, runtimeDebug.Stack())
		}
	}()
	// Проверка на nil.
	if c == nil {
		err = fmt.Errorf(tplErrorNil)
		return
	}
	crv = indirect(reflect.ValueOf(c))
	// Проверка на не инициализированный объект равный nil.
	if !crv.IsValid() {
		err = fmt.Errorf(tplErrorNotValid, reflect.TypeOf(c).String())
		return
	}
	crt = indirectType(crv.Type())
	// Проверка того что объект является адресом.
	if !crv.CanAddr() {
		err = fmt.Errorf(tplErrorNotAddress, crv.Type().String())
		return
	}

	return
}

func reflectStructObject(c interface{}) (crv reflect.Value, crt reflect.Type, err error) {
	const tplErrorNotStructure = "переданный объект %q не является структурой"

	if crv, crt, err = reflectObject(c); err != nil {
		return
	}
	// В качестве конфигураций ожидаются только структуры.
	if crt.Kind() != reflect.Struct {
		err = fmt.Errorf(tplErrorNotStructure, reflect.TypeOf(c).String())
		return
	}

	return
}

// Удаление всех тегов, кроме перечисленных
func reflectCleanStructTag(src reflect.StructTag, tags ...string) (ret reflect.StructTag) {
	const tpl = ` %s:"%s"`
	var (
		val string
		buf *strings.Builder
		n   int
		ok  bool
	)

	buf = &strings.Builder{}
	for n = range tags {
		if val, ok = src.Lookup(tags[n]); ok {
			buf.WriteString(fmt.Sprintf(tpl, tags[n], strings.TrimSpace(val)))
		}
	}
	ret = reflect.StructTag(strings.TrimSpace(buf.String()))

	return
}

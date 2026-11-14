package log

import (
	"reflect"
	"strings"
	"testing"

	kitModuleLogLevel "github.com/webnice/kit/v4/module/log/level"
)

// testUtilitiesNamed Именованный тип для тестирования getFuncFullName.
type testUtilitiesNamed struct{}

func TestGetFuncFullName(t *testing.T) {
	const (
		tplFuncNameMismatch = "getFuncFullName(функция) = %q, ожидалось имя содержащее %q."
		tplTypeNameMismatch = "getFuncFullName(указатель на тип) = %q, ожидалось имя содержащее %q."
		tplIntNameMismatch  = "getFuncFullName(указатель на int) = %q, ожидалось %q."
		tplAnonNotEmpty     = "getFuncFullName(анонимный тип) = %q, ожидалась пустая строка."
		tplClosureMismatch  = "getFuncFullName(замыкание) = %q, ожидалось имя содержащее %q."
	)
	var (
		i  int
		fn func(*Message)
	)

	// Функция - полное имя функции.
	if ret := getFuncFullName(testHandlerSubscribeOne); !strings.Contains(ret, "testHandlerSubscribeOne") {
		t.Errorf(tplFuncNameMismatch, ret, "testHandlerSubscribeOne")
	}
	// Указатель на именованный тип - имя пакета и типа.
	if ret := getFuncFullName(&testUtilitiesNamed{}); !strings.Contains(ret, "testUtilitiesNamed") {
		t.Errorf(tplTypeNameMismatch, ret, "testUtilitiesNamed")
	}
	// Предопределённый тип без пакета.
	if ret := getFuncFullName(&i); ret != "int" {
		t.Errorf(tplIntNameMismatch, ret, "int")
	}
	// Анонимный тип - пустое имя.
	if ret := getFuncFullName(&struct{}{}); ret != "" {
		t.Errorf(tplAnonNotEmpty, ret)
	}
	fn = func(_ *Message) {}
	if ret := getFuncFullName(fn); !strings.Contains(ret, "TestGetFuncFullName") {
		t.Errorf(tplClosureMismatch, ret, "TestGetFuncFullName")
	}
}

func TestIndirect(t *testing.T) {
	const (
		errIndirectWrong = "функция indirect() вернула не корректное значение."
		errIndirectValue = "функция indirect() для значения вернула не корректное значение."
	)
	var (
		v  int
		rv reflect.Value
	)

	rv = indirect(reflect.ValueOf(&v))
	if rv.Kind() != reflect.Int || !rv.CanAddr() {
		t.Error(errIndirectWrong)
	}
	if rv = indirect(reflect.ValueOf(v)); rv.Kind() != reflect.Int {
		t.Error(errIndirectValue)
	}
}

func TestIndirectType(t *testing.T) {
	const (
		tplIndirectType = "функция indirectType() вернула %v, ожидался int."
	)
	var (
		sl []int
		rt reflect.Type
	)

	if rt = indirectType(reflect.TypeOf(&sl)); rt.Kind() != reflect.Int {
		t.Errorf(tplIndirectType, rt)
	}
}

func TestFindLogLevelDefault(t *testing.T) {
	const (
		tplFindLevel = "findLoglevel() = %v, ожидался Notice."
	)

	// Вызов из функции без признаков log.fatal/log.panic - уровень Notice.
	if ret := findLoglevel(); ret != kitModuleLogLevel.Notice {
		t.Errorf(tplFindLevel, ret)
	}
}

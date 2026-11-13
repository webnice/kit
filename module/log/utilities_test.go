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
	var (
		i  int
		fn func(*Message)
	)

	// Функция - полное имя функции.
	if ret := getFuncFullName(testHandlerSubscribeOne); !strings.Contains(ret, "testHandlerSubscribeOne") {
		t.Errorf("getFuncFullName(функция) = %q, ожидалось имя содержащее %q", ret, "testHandlerSubscribeOne")
	}
	// Указатель на именованный тип - имя пакета и типа.
	if ret := getFuncFullName(&testUtilitiesNamed{}); !strings.Contains(ret, "testUtilitiesNamed") {
		t.Errorf("getFuncFullName(указатель на тип) = %q, ожидалось имя содержащее %q", ret, "testUtilitiesNamed")
	}
	// Предопределённый тип без пакета.
	if ret := getFuncFullName(&i); ret != "int" {
		t.Errorf("getFuncFullName(указатель на int) = %q, ожидалось %q", ret, "int")
	}
	// Анонимный тип - пустое имя.
	if ret := getFuncFullName(&struct{}{}); ret != "" {
		t.Errorf("getFuncFullName(анонимный тип) = %q, ожидалась пустая строка", ret)
	}
	fn = func(_ *Message) {}
	if ret := getFuncFullName(fn); !strings.Contains(ret, "TestGetFuncFullName") {
		t.Errorf("getFuncFullName(замыкание) = %q, ожидалось имя содержащее %q", ret, "TestGetFuncFullName")
	}
}

func TestIndirect(t *testing.T) {
	var (
		v  int
		rv reflect.Value
	)

	rv = indirect(reflect.ValueOf(&v))
	if rv.Kind() != reflect.Int || !rv.CanAddr() {
		t.Error("функция indirect() вернула не корректное значение")
	}
	if rv = indirect(reflect.ValueOf(v)); rv.Kind() != reflect.Int {
		t.Error("функция indirect() для значения вернула не корректное значение")
	}
}

func TestIndirectType(t *testing.T) {
	var (
		sl []int
		rt reflect.Type
	)

	if rt = indirectType(reflect.TypeOf(&sl)); rt.Kind() != reflect.Int {
		t.Errorf("функция indirectType() вернула %v, ожидался int", rt)
	}
}

func TestFindLogLevelDefault(t *testing.T) {
	// Вызов из функции без признаков log.fatal/log.panic - уровень Notice.
	if ret := findLoglevel(); ret != kitModuleLogLevel.Notice {
		t.Errorf("findLoglevel() = %v, ожидался Notice", ret)
	}
}

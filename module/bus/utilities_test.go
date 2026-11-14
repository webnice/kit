package bus

import (
	"reflect"
	"strings"
	"testing"
)

// TestUtilities Вспомогательные функции пакета.
func TestUtilities(t *testing.T) {
	const (
		tplFuncNameData   = "getFuncFullName(&data) = %q."
		tplFuncNameStruct = "getFuncFullName(&struct{}{}) = %q."
		errFuncNameSlice  = "getFuncFullName(&[]int{}) вернул пустое имя."
		tplFuncNameFunc   = "getFuncFullName(func) = %q."
		tplIndirect       = "indirect() = %+v."
		tplIndirectType   = "indirectType() = %v."
		tplTypeInfoError  = "makeSubscriberTypeInfo() вернул ошибку: %v."
		tplTypeInfo       = "makeSubscriberTypeInfo() = %+v."
		tplTypeInfoFor    = "makeSubscriberTypeInfo(%v) = %+v, %v."
	)
	var (
		data  busTestData
		name  string
		fn    func()
		got   reflect.Value
		rt    reflect.Type
		sti   *subscriberTypeInfo
		err   error
		types []reflect.Type
		n     int
		ch    chan struct{}
	)

	if name = getFuncFullName(&data); !strings.Contains(name, "busTestData") {
		t.Fatalf(tplFuncNameData, name)
	}
	if name = getFuncFullName(&struct{}{}); name != "" {
		t.Fatalf(tplFuncNameStruct, name)
	}
	if name = getFuncFullName(&[]int{}); name == "" {
		t.Fatal(errFuncNameSlice)
	}
	fn = func() {}
	if name = getFuncFullName(fn); !strings.Contains(name, "TestUtilities") {
		t.Fatalf(tplFuncNameFunc, name)
	}
	if got = indirect(reflect.ValueOf(&data)); got.Interface().(busTestData) != data {
		t.Fatalf(tplIndirect, got)
	}
	if rt = indirectType(reflect.TypeOf(&data)); rt != reflect.TypeOf(data) {
		t.Fatalf(tplIndirectType, rt)
	}
	sti, err = makeSubscriberTypeInfo(reflect.TypeOf(&data))
	if err != nil {
		t.Fatalf(tplTypeInfoError, err)
	}
	if sti.BaseType != reflect.TypeOf(data) || sti.OriginalType != reflect.TypeOf(&data) || sti.IndirectionCount != 1 {
		t.Fatalf(tplTypeInfo, sti)
	}
	types = []reflect.Type{reflect.TypeOf(data), reflect.TypeOf(struct{}{}), reflect.TypeOf(&struct{}{})}
	for n = range types {
		if sti, err = makeSubscriberTypeInfo(types[n]); err != nil || sti.OriginalType != types[n] {
			t.Fatalf(tplTypeInfoFor, types[n], sti, err)
		}
	}
	ch = make(chan struct{})
	safeCloseSignalChannel(ch)
	safeCloseSignalChannel(ch)
}

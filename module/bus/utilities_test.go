package bus

import (
	"reflect"
	"strings"
	"testing"
)

// TestUtilities Вспомогательные функции пакета.
func TestUtilities(t *testing.T) {
	var data busTestData
	if name := getFuncFullName(&data); !strings.Contains(name, "busTestData") {
		t.Fatalf("getFuncFullName(&data) = %q", name)
	}
	if name := getFuncFullName(&struct{}{}); name != "" {
		t.Fatalf("getFuncFullName(&struct{}{}) = %q", name)
	}
	if name := getFuncFullName(&[]int{}); name == "" {
		t.Fatal("getFuncFullName(&[]int{}) вернул пустое имя")
	}
	fn := func() {}
	if name := getFuncFullName(fn); !strings.Contains(name, "TestUtilities") {
		t.Fatalf("getFuncFullName(func) = %q", name)
	}
	if got := indirect(reflect.ValueOf(&data)); got.Interface().(busTestData) != data {
		t.Fatalf("indirect() = %+v", got)
	}
	if got := indirectType(reflect.TypeOf(&data)); got != reflect.TypeOf(data) {
		t.Fatalf("indirectType() = %v", got)
	}
	sti, err := makeSubscriberTypeInfo(reflect.TypeOf(&data))
	if err != nil {
		t.Fatalf("makeSubscriberTypeInfo() вернул ошибку: %v", err)
	}
	if sti.BaseType != reflect.TypeOf(data) || sti.OriginalType != reflect.TypeOf(&data) || sti.IndirectionCount != 1 {
		t.Fatalf("makeSubscriberTypeInfo() = %+v", sti)
	}
	for _, rt := range []reflect.Type{reflect.TypeOf(data), reflect.TypeOf(struct{}{}), reflect.TypeOf(&struct{}{})} {
		if sti, err = makeSubscriberTypeInfo(rt); err != nil || sti.OriginalType != rt {
			t.Fatalf("makeSubscriberTypeInfo(%v) = %+v, %v", rt, sti, err)
		}
	}
	ch := make(chan struct{})
	safeCloseSignalChannel(ch)
	safeCloseSignalChannel(ch)
}

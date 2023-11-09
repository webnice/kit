package cfg

import "testing"

func TestNew(t *testing.T) {
	var (
		obj Interface
		ok  bool
	)

	obj = Get()
	if obj == nil {
		t.Error("Function New() return nil")
		return
	}
	if _, ok = obj.(*impl); !ok {
		t.Error("Function New() return invalid object")
		return
	}
}

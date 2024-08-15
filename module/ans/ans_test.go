package ans

import (
	"testing"
)

func TestNew(t *testing.T) {
	var (
		ans Interface
		obj *impl
	)

	ans = New(nil)
	if ans == nil {
		t.Errorf("функция New(), вернулся nil, ожидался не nil объект")
		return
	}
	obj = ans.(*impl)
	if obj.essence == nil {
		t.Errorf("функция New(), вернулся объект у которого essence = nil, ожидался не nil объект")
		return
	}
}

func TestImpl_Gist(t *testing.T) {
	var (
		ans Interface
		obj *impl
	)

	ans = New(nil)
	if ans == nil {
		t.Errorf("функция New(), вернулся nil, ожидался не nil объект")
		return
	}
	obj = ans.(*impl)
	if obj.debug != false {
		t.Errorf("Неожиданное значение debug, ожидалось %v", false)
		return
	}
	if obj.debug != false {
		t.Errorf("Неожиданное значение debug, ожидалось %v", false)
		return
	}
	ans.Gist().Debug(true)
	if obj.debug == false {
		t.Errorf("Функция Gist().Debug() не устанавливает значение.")
		return
	}
}

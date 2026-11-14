package ans

import (
	"testing"
)

func TestNew(t *testing.T) {
	const (
		errNewNil     = "функция New(), вернулся nil, ожидался не nil объект."
		errEssenceNil = "функция New(), вернулся объект у которого essence = nil, ожидался не nil объект."
	)
	var (
		ans Interface
		obj *impl
	)

	ans = New(nil)
	if ans == nil {
		t.Error(errNewNil)

		return
	}
	obj = ans.(*impl)
	if obj.essence == nil {
		t.Error(errEssenceNil)

		return
	}
}

func TestImpl_Gist(t *testing.T) {
	const (
		errNewNil          = "функция New(), вернулся nil, ожидался не nil объект."
		tplDebugUnexpected = "Неожиданное значение debug, ожидалось %v."
		errDebugNotSet     = "Функция Gist().Debug() не устанавливает значение."
	)
	var (
		ans Interface
		obj *impl
	)

	ans = New(nil)
	if ans == nil {
		t.Error(errNewNil)

		return
	}
	obj = ans.(*impl)
	if obj.debug != false {
		t.Errorf(tplDebugUnexpected, false)

		return
	}
	if obj.debug != false {
		t.Errorf(tplDebugUnexpected, false)

		return
	}
	ans.Gist().Debug(true)
	if obj.debug == false {
		t.Error(errDebugNotSet)

		return
	}
}

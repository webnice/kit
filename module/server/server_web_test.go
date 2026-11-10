package server

import (
	"testing"

	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

func TestWebAdd(t *testing.T) {
	var (
		srv = New(&testLogger{})
		cfg = testWebConfiguration(t)
		si  *kitTypesServer.Server
	)

	si = srv.Web().Add(cfg)
	if si == nil {
		t.Fatal("функция Add(), вернулся nil, ожидался не nil объект")
	}
	if si.T != kitTypesServer.TWeb {
		t.Error("функция Add(), тип сервера не равен TWeb")
	}
	if si.Web != cfg {
		t.Error("функция Add(), конфигурация сервера не совпадает с переданной")
	}
	if len(srv.(*impl).server) != 1 {
		t.Error("функция Add(), сервер не добавлен в список серверов")
	}
	// Идентификатор заданный в конфигурации не должен перезаписываться.
	cfg2 := testWebConfiguration(t)
	cfg2.Server.ID = "test-server-id"
	si2 := srv.Web().Add(cfg2)
	if si2.Web.Server.ID != "test-server-id" {
		t.Error("функция Add(), заданный идентификатор сервера перезаписан")
	}
}

func TestWebReg(t *testing.T) {
	var srv = New(&testLogger{})

	if len(srv.(*impl).serverWeb.res) != 0 {
		t.Fatal("список ресурсов не пуст перед регистрацией")
	}
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{Path: "/api"}
	}))
	if len(srv.(*impl).serverWeb.res) != 1 {
		t.Error("функция Reg(), ресурс не зарегистрирован")
	}
}

func TestWebLib(t *testing.T) {
	var (
		srv = New(&testLogger{})
		lib = srv.Web().Lib()
	)

	if lib == nil {
		t.Fatal("функция Lib(), вернулся nil, ожидался не nil объект")
	}
	if lib.Handler() == nil {
		t.Error("функция Handler(), вернулся nil, ожидался не nil объект")
	}
	if lib.Middleware() == nil {
		t.Error("функция Middleware(), вернулся nil, ожидался не nil объект")
	}
}

func TestWebError(t *testing.T) {
	var (
		srv   = New(&testLogger{})
		iwe   = srv.Web().Error()
		value = errTestValue
	)

	if iwe == nil {
		t.Fatal("функция Error(), вернулся nil, ожидался не nil объект")
	}
	iwe.Reset()
	if ret := iwe.InternalServerError(nil); ret != nil {
		t.Errorf("функция InternalServerError(nil), вернула %v, ожидался nil", ret)
	}
	if ret := iwe.InternalServerError(value); ret != value {
		t.Errorf("функция InternalServerError(err), вернула %v, ожидалась %v", ret, value)
	}
	if ret := iwe.MethodNotAllowed(nil); ret != nil {
		t.Errorf("функция MethodNotAllowed(nil), вернула %v, ожидался nil", ret)
	}
	if ret := iwe.MethodNotAllowed(value); ret != value {
		t.Errorf("функция MethodNotAllowed(err), вернула %v, ожидалась %v", ret, value)
	}
	// Сброс всех ошибок.
	iwe.Reset()
	if ret := iwe.InternalServerError(nil); ret != nil {
		t.Errorf("после Reset() функция InternalServerError(nil), вернула %v, ожидался nil", ret)
	}
}

var errTestValue = errorString("тестовая ошибка")

type errorString string

func (es errorString) Error() string { return string(es) }

func TestImplWebLog(t *testing.T) {
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
	)

	// Проверка ссылки на менеджер логирования.
	if obj.serverWeb.log() != obj.logger {
		t.Error("функция log(), вернула не верный менеджер логирования")
	}
	// Режим отладки выключен - в лог ничего не пишется, паники нет.
	obj.serverWeb.info("тестовое сообщение %q", "отладка")
	// Режим отладки включен - сообщение уходит в лог, паники нет.
	srv.Gist().Debug(true)
	obj.serverWeb.info("тестовое сообщение %q", "отладка")
}

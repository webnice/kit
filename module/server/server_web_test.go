package server

import (
	"testing"

	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

func TestWebAdd(t *testing.T) {
	const (
		errAddNil        = "функция Add(), вернулся nil, ожидался не nil объект."
		errTypeNotWeb    = "функция Add(), тип сервера не равен TWeb."
		errConfigWrong   = "функция Add(), конфигурация сервера не совпадает с переданной."
		errNotAdded      = "функция Add(), сервер не добавлен в список серверов."
		errIDOverwritten = "функция Add(), заданный идентификатор сервера перезаписан."
		testServerID     = "test-server-id"
	)
	var (
		srv  Interface
		cfg  *kitTypesServer.WebConfiguration
		cfg2 *kitTypesServer.WebConfiguration
		si   *kitTypesServer.Server
		si2  *kitTypesServer.Server
	)

	srv = New(&testLogger{})
	cfg = testWebConfiguration(t)
	si = srv.Web().Add(cfg)
	if si == nil {
		t.Fatal(errAddNil)
	}
	if si.T != kitTypesServer.TWeb {
		t.Error(errTypeNotWeb)
	}
	if si.Web != cfg {
		t.Error(errConfigWrong)
	}
	if len(srv.(*impl).server) != 1 {
		t.Error(errNotAdded)
	}
	// Идентификатор заданный в конфигурации не должен перезаписываться.
	cfg2 = testWebConfiguration(t)
	cfg2.Server.ID = testServerID
	si2 = srv.Web().Add(cfg2)
	if si2.Web.Server.ID != testServerID {
		t.Error(errIDOverwritten)
	}
}

func TestWebReg(t *testing.T) {
	const (
		errResNotEmpty = "список ресурсов не пуст перед регистрацией."
		errResNotAdded = "функция Reg(), ресурс не зарегистрирован."
	)
	var srv Interface

	srv = New(&testLogger{})
	if len(srv.(*impl).serverWeb.res) != 0 {
		t.Fatal(errResNotEmpty)
	}
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{Path: "/api"}
	}))
	if len(srv.(*impl).serverWeb.res) != 1 {
		t.Error(errResNotAdded)
	}
}

func TestWebLib(t *testing.T) {
	const (
		errLibNil     = "функция Lib(), вернулся nil, ожидался не nil объект."
		errHandlerNil = "функция Handler(), вернулся nil, ожидался не nil объект."
		errMiddleNil  = "функция Middleware(), вернулся nil, ожидался не nil объект."
	)
	var (
		srv Interface
		lib InterfaceWebLib
	)

	srv = New(&testLogger{})
	lib = srv.Web().Lib()
	if lib == nil {
		t.Fatal(errLibNil)
	}
	if lib.Handler() == nil {
		t.Error(errHandlerNil)
	}
	if lib.Middleware() == nil {
		t.Error(errMiddleNil)
	}
}

var errTestValue = errorString("тестовая ошибка.")

type errorString string

// Error Реализация интерфейса error.
func (es errorString) Error() string { return string(es) }

func TestWebError(t *testing.T) {
	const (
		errErrorNil      = "функция Error(), вернулся nil, ожидался не nil объект."
		tplInternalNil   = "функция InternalServerError(nil), вернула %v, ожидался nil."
		tplInternalErr   = "функция InternalServerError(err), вернула %v, ожидалась %v."
		tplMethodNil     = "функция MethodNotAllowed(nil), вернула %v, ожидался nil."
		tplMethodErr     = "функция MethodNotAllowed(err), вернула %v, ожидалась %v."
		tplResetInternal = "после Reset() функция InternalServerError(nil), вернула %v, ожидался nil."
	)
	var (
		srv   Interface
		iwe   InterfaceWebError
		value error
	)

	srv = New(&testLogger{})
	iwe = srv.Web().Error()
	value = errTestValue
	if iwe == nil {
		t.Fatal(errErrorNil)
	}
	iwe.Reset()
	if ret := iwe.InternalServerError(nil); ret != nil {
		t.Errorf(tplInternalNil, ret)
	}
	if ret := iwe.InternalServerError(value); ret != value {
		t.Errorf(tplInternalErr, ret, value)
	}
	if ret := iwe.MethodNotAllowed(nil); ret != nil {
		t.Errorf(tplMethodNil, ret)
	}
	if ret := iwe.MethodNotAllowed(value); ret != value {
		t.Errorf(tplMethodErr, ret, value)
	}
	// Сброс всех ошибок.
	iwe.Reset()
	if ret := iwe.InternalServerError(nil); ret != nil {
		t.Errorf(tplResetInternal, ret)
	}
}

func TestImplWebLog(t *testing.T) {
	const (
		errLogWrong    = "функция log(), вернула не верный менеджер логирования."
		tplInfoMessage = "тестовое сообщение %q."
		infoArgument   = "отладка."
	)
	var (
		srv Interface
		obj *impl
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	// Проверка ссылки на менеджер логирования.
	if obj.serverWeb.log() != obj.logger {
		t.Error(errLogWrong)
	}
	// Режим отладки выключен - в лог ничего не пишется, паники нет.
	obj.serverWeb.info(tplInfoMessage, infoArgument)
	// Режим отладки включен - сообщение уходит в лог, паники нет.
	srv.Gist().Debug(true)
	obj.serverWeb.info(tplInfoMessage, infoArgument)
}

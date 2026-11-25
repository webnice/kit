package server

import (
	"net/http"
	"strings"
	"testing"

	kitTypesServer "github.com/webnice/kit/v5/types/server"
	"github.com/webnice/web/v4"
)

func TestIsStarted(t *testing.T) {
	const (
		errUnknownTrue    = "функция IsStarted() для не известного сервера, вернулась истина."
		errNotStartedTrue = "функция IsStarted() для не запущенного сервера, вернулась истина."
	)
	var (
		srv Interface
		cfg *kitTypesServer.WebConfiguration
	)

	srv = New(&testLogger{})
	cfg = testWebConfiguration(t)
	if srv.Web().IsStarted("unknown-server-id") {
		t.Error(errUnknownTrue)
	}
	srv.Web().Add(cfg)
	if srv.Web().IsStarted(cfg.Server.ID) {
		t.Error(errNotStartedTrue)
	}
}

func TestStartServerByIdNotFound(t *testing.T) {
	const (
		tplStartWrong   = "функция Start() с не известным идентификатором, вернула не ожидаемую ошибку: %v."
		errServerInList = "функция Start() с не известным идентификатором, сервер остался в списке запущенных."
	)
	var (
		srv Interface
		err error
	)

	srv = New(&testLogger{})
	// Запуск сервера с не известным идентификатором.
	if err = srv.(*impl).serverWeb.Start("unknown-server-id"); err == nil ||
		!strings.Contains(err.Error(), "не найдена") {
		t.Errorf(tplStartWrong, err)
	}
	// Запись о сервере не осталась в списке запущенных.
	if srv.Web().IsStarted("unknown-server-id") {
		t.Error(errServerInList)
	}
}

func TestStartAlreadyStarted(t *testing.T) {
	const (
		tplFatalPrepare = "функция Prepare() вернула не ожидаемую ошибку: %s."
		tplFatalStart   = "функция Start() вернула не ожидаемую ошибку: %s."
		tplStartAgain   = "функция Start() повторного запуска, вернула не ожидаемую ошибку: %v."
		tplStartLower   = "функция Start() повторного запуска в нижнем регистре, вернула не ожидаемую ошибку: %v."
	)
	var (
		srv Interface
		cfg *kitTypesServer.WebConfiguration
		si  *kitTypesServer.Server
		err error
	)

	srv = New(&testLogger{})
	cfg = testWebConfiguration(t)
	si = srv.Web().Add(cfg)
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{Path: "/api"}
	}))
	if err = srv.(*impl).serverWeb.Prepare(cfg); err != nil {
		t.Fatalf(tplFatalPrepare, err)
	}
	if err = srv.(*impl).serverWeb.Start(si.Web.Server.ID); err != nil {
		t.Fatalf(tplFatalStart, err)
	}
	defer func() { _ = srv.(*impl).serverWeb.Stop(si.Web.Server.ID) }()
	// Повторный запуск уже запущенного сервера.
	if err = srv.(*impl).serverWeb.Start(si.Web.Server.ID); err == nil ||
		!strings.Contains(err.Error(), "уже запущен") {
		t.Errorf(tplStartAgain, err)
	}
	// Проверка не зависит от регистрара символов идентификатора.
	if err = srv.(*impl).serverWeb.Start(strings.ToLower(si.Web.Server.ID)); err == nil ||
		!strings.Contains(err.Error(), "уже запущен") {
		t.Errorf(tplStartLower, err)
	}
}

func TestStopServerByIdNotStarted(t *testing.T) {
	const (
		tplStopWrong = "функция Stop() не запущенного сервера, вернула не ожидаемую ошибку: %v."
	)
	var err error

	err = New(&testLogger{}).Web().(*implWeb).Stop("unknown-server-id")
	if err == nil || !strings.Contains(err.Error(), "не был запущен") {
		t.Errorf(tplStopWrong, err)
	}
}

func TestStartDomainRefusedLogging(t *testing.T) {
	const (
		tplFatalPrepare  = "функция Prepare() вернула не ожидаемую ошибку: %s."
		tplFatalStart    = "функция Start() вернула не ожидаемую ошибку: %s."
		tplNewRequest    = "функция NewRequest() вернула не ожидаемую ошибку: %v."
		tplDoError       = "функция Do() вернула не ожидаемую ошибку: %v."
		tplWrongCode     = "функция ServeHTTP(), код ответа: %d, ожидалось: %d."
		tplLoggedAllowed = "функция domainRefused(), сообщения лога: %v, ожидалось: отсутствие сообщений."
		tplNotLogged     = "функция domainRefused(), сообщения лога: %v, ожидалось сообщение с доменом %q."
		testHostAllowed  = "allowed.example.com"
		testHostDenied   = "denied.example.com"
	)
	var (
		lg  *testWarningLogger
		srv Interface
		cfg *kitTypesServer.WebConfiguration
		si  *kitTypesServer.Server
		req *http.Request
		rsp *http.Response
		err error
	)

	lg = &testWarningLogger{}
	srv = New(lg)
	cfg = testWebConfiguration(t)
	cfg.Server.Domain = []string{testHostAllowed}
	si = srv.Web().Add(cfg)
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{Path: "/api"}
	}))
	if err = srv.(*impl).serverWeb.Prepare(cfg); err != nil {
		t.Fatalf(tplFatalPrepare, err)
	}
	if err = srv.(*impl).serverWeb.Start(si.Web.Server.ID); err != nil {
		t.Fatalf(tplFatalStart, err)
	}
	defer func() { _ = srv.(*impl).serverWeb.Stop(si.Web.Server.ID) }()
	// Запрос на разрешённый домен, сообщение об отказе не логируется.
	if req, err = http.NewRequest(http.MethodGet, "http://"+cfg.Server.HostPort()+"/", nil); err != nil {
		t.Fatalf(tplNewRequest, err)
	}
	req.Host = testHostAllowed
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatalf(tplDoError, err)
	}
	_ = rsp.Body.Close()
	if len(lg.messages) != 0 {
		t.Errorf(tplLoggedAllowed, lg.messages)
	}
	// Запрос на запрещённый домен: статус 434 и сообщение об отказе в логе.
	if req, err = http.NewRequest(http.MethodGet, "http://"+cfg.Server.HostPort()+"/", nil); err != nil {
		t.Fatalf(tplNewRequest, err)
	}
	req.Host = testHostDenied
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatalf(tplDoError, err)
	}
	_ = rsp.Body.Close()
	if rsp.StatusCode != web.Status.RequestedHostUnavailable.Code() {
		t.Errorf(tplWrongCode, rsp.StatusCode, web.Status.RequestedHostUnavailable.Code())
	}
	if len(lg.messages) != 1 || !strings.Contains(lg.messages[0], testHostDenied) {
		t.Errorf(tplNotLogged, lg.messages, testHostDenied)
	}
}

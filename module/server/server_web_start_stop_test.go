package server

import (
	"strings"
	"testing"

	kitTypesServer "github.com/webnice/kit/v4/types/server"
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

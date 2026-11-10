package server

import (
	"strings"
	"testing"

	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

func TestIsStarted(t *testing.T) {
	var (
		srv = New(&testLogger{})
		cfg = testWebConfiguration(t)
	)

	if srv.Web().IsStarted("unknown-server-id") {
		t.Error("функция IsStarted() для не известного сервера, вернулась истина")
	}
	srv.Web().Add(cfg)
	if srv.Web().IsStarted(cfg.Server.ID) {
		t.Error("функция IsStarted() для не запущенного сервера, вернулась истина")
	}
}

func TestStartServerByIdNotFound(t *testing.T) {
	var (
		srv = New(&testLogger{})
		err error
	)

	// Запуск сервера с не известным идентификатором.
	if err = srv.(*impl).serverWeb.Start("unknown-server-id"); err == nil ||
		!strings.Contains(err.Error(), "не найдена") {
		t.Errorf("функция Start() с не известным идентификатором, вернула не ожидаемую ошибку: %v", err)
	}
	// Запись о сервере не осталась в списке запущенных.
	if srv.Web().IsStarted("unknown-server-id") {
		t.Error("функция Start() с не известным идентификатором, сервер остался в списке запущенных")
	}
}

func TestStartAlreadyStarted(t *testing.T) {
	var (
		srv = New(&testLogger{})
		cfg = testWebConfiguration(t)
		si  *kitTypesServer.Server
		err error
	)

	si = srv.Web().Add(cfg)
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{Path: "/api"}
	}))
	if err = srv.(*impl).serverWeb.Prepare(cfg); err != nil {
		t.Fatalf("функция Prepare() вернула не ожидаемую ошибку: %s", err)
	}
	if err = srv.(*impl).serverWeb.Start(si.Web.Server.ID); err != nil {
		t.Fatalf("функция Start() вернула не ожидаемую ошибку: %s", err)
	}
	defer func() { _ = srv.(*impl).serverWeb.Stop(si.Web.Server.ID) }()
	// Повторный запуск уже запущенного сервера.
	if err = srv.(*impl).serverWeb.Start(si.Web.Server.ID); err == nil ||
		!strings.Contains(err.Error(), "уже запущен") {
		t.Errorf("функция Start() повторного запуска, вернула не ожидаемую ошибку: %v", err)
	}
	// Проверка не зависит от регистрара символов идентификатора.
	if err = srv.(*impl).serverWeb.Start(strings.ToLower(si.Web.Server.ID)); err == nil ||
		!strings.Contains(err.Error(), "уже запущен") {
		t.Errorf("функция Start() повторного запуска в нижнем регистре, вернула не ожидаемую ошибку: %v", err)
	}
}

func TestStopServerByIdNotStarted(t *testing.T) {
	var err = New(&testLogger{}).Web().(*implWeb).Stop("unknown-server-id")

	if err == nil || !strings.Contains(err.Error(), "не был запущен") {
		t.Errorf("функция Stop() не запущенного сервера, вернула не ожидаемую ошибку: %v", err)
	}
}

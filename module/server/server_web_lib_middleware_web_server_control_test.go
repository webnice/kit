package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

func TestWebServerControlHandler(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		ctl = &kitTypesServer.WebServerControl{}
		rq  = httptest.NewRequest(http.MethodGet, "/", nil)
		ret *kitTypesServer.WebServerControl
	)

	var handler = iwl.WebServerControlHandler(ctl)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ret, _ = iwl.WebServerControlGetFromContext(r)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), rq)
	if ret != ctl {
		t.Errorf("функция WebServerControlHandler(), из контекста получен объект %v, ожидался %v", ret, ctl)
	}
}

func TestWebServerControlGetFromContext(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		ctl = &kitTypesServer.WebServerControl{}
		rq  = httptest.NewRequest(http.MethodGet, "/", nil)
		ret *kitTypesServer.WebServerControl
		err error
	)

	// Объект контроля отсутствует в контексте.
	if ret, err = iwl.WebServerControlGetFromContext(rq); err == nil || ret != nil {
		t.Errorf("функция WebServerControlGetFromContext() без данных, ret=%v err=%v, ожидались nil и ошибка", ret, err)
	}
	if err != nil && !strings.Contains(err.Error(), "не найдены") {
		t.Errorf("функция WebServerControlGetFromContext() без данных, не ожидаемая ошибка: %v", err)
	}
	// В контексте данные не верного типа.
	rq = rq.WithContext(contextWith(rq, cKeyWebServerControl, cNameWebServerControl, "wrong value"))
	if ret, err = iwl.WebServerControlGetFromContext(rq); err == nil || ret != nil {
		t.Errorf("функция WebServerControlGetFromContext() с не верным типом, ret=%v err=%v, ожидались nil и ошибка", ret, err)
	}
	if err != nil && !strings.Contains(err.Error(), "не верные данные") {
		t.Errorf("функция WebServerControlGetFromContext() с не верным типом, не ожидаемая ошибка: %v", err)
	}
	// Корректные данные в контексте.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(contextWith(rq, cKeyWebServerControl, cNameWebServerControl, ctl))
	if ret, err = iwl.WebServerControlGetFromContext(rq); err != nil || ret != ctl {
		t.Errorf("функция WebServerControlGetFromContext(), ret=%v err=%v, ожидался объект %v и nil", ret, err, ctl)
	}
}

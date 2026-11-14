package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	kitTypesServer "github.com/webnice/kit/v5/types/server"
)

func TestRequestShadowInfoHandler(t *testing.T) {
	const (
		errInfoNil        = "функция RequestShadowInfoHandler(), объект информации о запросе не получен из контекста."
		tplProjectIDWrong = "ProjectID = %d, ожидалось 42."
		tplUserAgentWrong = "UserAgent = %q, ожидалось %q."
		errServerMismatch = "Server не совпадает с конфигурацией из объекта контроля."
		tplIPWrong        = "IP = %v, ожидался 192.168.1.2."
		testUserAgent     = "test-agent"
	)
	var (
		srv Interface
		iwl *implWebLib
		ctl *kitTypesServer.WebServerControl
		rq  *http.Request
		ret *kitTypesServer.RequestShadowInfo
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	ctl = &kitTypesServer.WebServerControl{Configuration: &kitTypesServer.Server{}}
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rq.Header.Set("X-Project-Id", "42")
	rq.Header.Set("User-Agent", testUserAgent)
	rq.RemoteAddr = "192.168.1.2:1234"
	// Цепочка промежуточных слоёв: контроль сервера, IP адрес, информация о запросе.
	handler := iwl.WebServerControlHandler(ctl)(
		iwl.IpHandler()(
			iwl.RequestShadowInfoHandler()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				ret = iwl.RequestShadowInfoGetFromContext(r)
			}))))
	handler.ServeHTTP(httptest.NewRecorder(), rq)
	if ret == nil {
		t.Fatal(errInfoNil)
	}
	if ret.ProjectID != 42 {
		t.Errorf(tplProjectIDWrong, ret.ProjectID)
	}
	if ret.UserAgent != testUserAgent {
		t.Errorf(tplUserAgentWrong, ret.UserAgent, testUserAgent)
	}
	if ret.Server != ctl.Configuration {
		t.Error(errServerMismatch)
	}
	if ret.IP == nil || ret.IP.String() != "192.168.1.2" {
		t.Errorf(tplIPWrong, ret.IP)
	}
}

func TestRequestShadowInfoExtraction(t *testing.T) {
	const (
		tplNoControlWrong = "функция RequestShadowInfoExtraction() без объекта контроля, вернула %+v."
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rq.RemoteAddr = "192.168.1.2:1234"
	// Извлечение без объекта контроля ВЕБ сервера в контексте.
	if ret := iwl.RequestShadowInfoExtraction(rq); ret == nil || ret.Server != nil {
		t.Errorf(tplNoControlWrong, ret)
	}
}

func TestRequestShadowInfoGetFromContext(t *testing.T) {
	const (
		tplNoDataNotNil    = "функция RequestShadowInfoGetFromContext() без данных, вернула %v, ожидался nil."
		tplWrongTypeNotNil = "функция RequestShadowInfoGetFromContext() с не верным типом, вернула %v, ожидался nil."
		tplInfoWrong       = "функция RequestShadowInfoGetFromContext(), вернула %v, ожидался объект %v."
	)
	var (
		srv  Interface
		iwl  *implWebLib
		rq   *http.Request
		info *kitTypesServer.RequestShadowInfo
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	// Объект информации отсутствует в контексте.
	if ret := iwl.RequestShadowInfoGetFromContext(rq); ret != nil {
		t.Errorf(tplNoDataNotNil, ret)
	}
	// В контексте данные не верного типа.
	rq = rq.WithContext(contextWith(rq, cKeyRequestShadowInfo, cNameRequestShadowInfo, "wrong value"))
	if ret := iwl.RequestShadowInfoGetFromContext(rq); ret != nil {
		t.Errorf(tplWrongTypeNotNil, ret)
	}
	// Корректные данные в контексте.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	info = new(kitTypesServer.RequestShadowInfo)
	rq = rq.WithContext(contextWith(rq, cKeyRequestShadowInfo, cNameRequestShadowInfo, info))
	if ret := iwl.RequestShadowInfoGetFromContext(rq); ret != info {
		t.Errorf(tplInfoWrong, ret, info)
	}
}

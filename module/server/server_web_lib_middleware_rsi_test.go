package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

func TestRequestShadowInfoHandler(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		ctl = &kitTypesServer.WebServerControl{Configuration: &kitTypesServer.Server{}}
		rq  = httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ret *kitTypesServer.RequestShadowInfo
	)

	rq.Header.Set("X-Project-Id", "42")
	rq.Header.Set("User-Agent", "test-agent")
	rq.RemoteAddr = "192.168.1.2:1234"
	// Цепочка промежуточных слоёв: контроль сервера, IP адрес, информация о запросе.
	var handler = iwl.WebServerControlHandler(ctl)(
		iwl.IpHandler()(
			iwl.RequestShadowInfoHandler()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				ret = iwl.RequestShadowInfoGetFromContext(r)
			}))))
	handler.ServeHTTP(httptest.NewRecorder(), rq)
	if ret == nil {
		t.Fatal("функция RequestShadowInfoHandler(), объект информации о запросе не получен из контекста")
	}
	if ret.ProjectID != 42 {
		t.Errorf("ProjectID = %d, ожидалось 42", ret.ProjectID)
	}
	if ret.UserAgent != "test-agent" {
		t.Errorf("UserAgent = %q, ожидалось %q", ret.UserAgent, "test-agent")
	}
	if ret.Server != ctl.Configuration {
		t.Error("Server не совпадает с конфигурацией из объекта контроля")
	}
	if ret.IP == nil || ret.IP.String() != "192.168.1.2" {
		t.Errorf("IP = %v, ожидался 192.168.1.2", ret.IP)
	}
}

func TestRequestShadowInfoExtraction(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	)

	rq.RemoteAddr = "192.168.1.2:1234"
	// Извлечение без объекта контроля ВЕБ сервера в контексте.
	if ret := iwl.RequestShadowInfoExtraction(rq); ret == nil || ret.Server != nil {
		t.Errorf("функция RequestShadowInfoExtraction() без объекта контроля, вернула %+v", ret)
	}
}

func TestRequestShadowInfoGetFromContext(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/", nil)
	)

	// Объект информации отсутствует в контексте.
	if ret := iwl.RequestShadowInfoGetFromContext(rq); ret != nil {
		t.Errorf("функция RequestShadowInfoGetFromContext() без данных, вернула %v, ожидался nil", ret)
	}
	// В контексте данные не верного типа.
	rq = rq.WithContext(contextWith(rq, cKeyRequestShadowInfo, cNameRequestShadowInfo, "wrong value"))
	if ret := iwl.RequestShadowInfoGetFromContext(rq); ret != nil {
		t.Errorf("функция RequestShadowInfoGetFromContext() с не верным типом, вернула %v, ожидался nil", ret)
	}
	// Корректные данные в контексте.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	info := new(kitTypesServer.RequestShadowInfo)
	rq = rq.WithContext(contextWith(rq, cKeyRequestShadowInfo, cNameRequestShadowInfo, info))
	if ret := iwl.RequestShadowInfoGetFromContext(rq); ret != info {
		t.Errorf("функция RequestShadowInfoGetFromContext(), вернула %v, ожидался объект %v", ret, info)
	}
}

package server

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIpHandler(t *testing.T) {
	const (
		tplIPWrong = "функция IpHandler(), из контекста получен IP %v, ожидался 192.168.1.2."
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
		ret net.IP
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.RemoteAddr = "192.168.1.2:1234"
	// Проверка извлечения IP адреса внутри промежуточного слоя.
	handler := iwl.IpHandler()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ret = iwl.IpGetFromContext(r)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), rq)
	if ret == nil || ret.String() != "192.168.1.2" {
		t.Errorf(tplIPWrong, ret)
	}
}

func TestIpExtraction(t *testing.T) {
	const (
		tplIPExpected = "IpExtraction() = %v, ожидался %s."
		tplIPNil      = "IpExtraction() = %v, ожидался nil."
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
		ip  net.IP
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	// Приоритетный заголовок X-Client-Forwarded-For.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Client-Forwarded-For", "10.1.1.1")
	rq.Header.Set("X-Real-Ip", "10.2.2.2")
	rq.Header.Set("X-Forwarded-For", "10.3.3.3")
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.1.1.1" {
		t.Errorf(tplIPExpected, ip, "10.1.1.1")
	}
	// Заголовок X-Real-Ip.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Real-Ip", "10.2.2.2")
	rq.Header.Set("X-Forwarded-For", "10.3.3.3")
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.2.2.2" {
		t.Errorf(tplIPExpected, ip, "10.2.2.2")
	}
	// Заголовок X-Forwarded-For со списком адресов.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Forwarded-For", "10.3.3.3, 10.5.5.5, 10.6.6.6")
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.3.3.3" {
		t.Errorf(tplIPExpected, ip, "10.3.3.3")
	}
	// Заголовок X-Forwarded-For с одним адресом.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Forwarded-For", " 10.3.3.3 ")
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.3.3.3" {
		t.Errorf(tplIPExpected, ip, "10.3.3.3")
	}
	// Адрес клиента из RemoteAddr.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.4.4.4" {
		t.Errorf(tplIPExpected, ip, "10.4.4.4")
	}
	// Пустой RemoteAddr.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.RemoteAddr = " "
	if ip = iwl.IpExtraction(rq); ip != nil {
		t.Errorf(tplIPNil, ip)
	}
	// RemoteAddr без порта.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.RemoteAddr = "10.4.4.4"
	if ip = iwl.IpExtraction(rq); ip != nil {
		t.Errorf(tplIPNil, ip)
	}
	// Не корректный IP адрес в приоритетном заголовке.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Client-Forwarded-For", "wrong ip")
	if ip = iwl.IpExtraction(rq); ip != nil {
		t.Errorf(tplIPNil, ip)
	}
}

func TestIpGetFromContext(t *testing.T) {
	const (
		tplNoDataNotNil    = "функция IpGetFromContext() без данных в контексте, вернула %v, ожидался nil."
		tplWrongTypeNotNil = "функция IpGetFromContext() с не верным типом данных, вернула %v, ожидался nil."
		tplIPWrong         = "функция IpGetFromContext(), вернула %v, ожидался 192.168.1.2."
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	// Объект IP адреса отсутствует в контексте.
	if ip := iwl.IpGetFromContext(rq); ip != nil {
		t.Errorf(tplNoDataNotNil, ip)
	}
	// Объект IP адреса в контексте имеет не верный тип данных.
	rq = rq.WithContext(contextWith(rq, cKeyIP, cNameIP, "wrong value"))
	if ip := iwl.IpGetFromContext(rq); ip != nil {
		t.Errorf(tplWrongTypeNotNil, ip)
	}
	// Корректные данные в контексте.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(contextWith(rq, cKeyIP, cNameIP, net.ParseIP("192.168.1.2")))
	if ip := iwl.IpGetFromContext(rq); ip == nil || ip.String() != "192.168.1.2" {
		t.Errorf(tplIPWrong, ip)
	}
}

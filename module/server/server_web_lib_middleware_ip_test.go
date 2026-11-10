package server

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIpHandler(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/", nil)
		ret net.IP
	)

	rq.RemoteAddr = "192.168.1.2:1234"
	// Проверка извлечения IP адреса внутри промежуточного слоя.
	var handler = iwl.IpHandler()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ret = iwl.IpGetFromContext(r)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), rq)
	if ret == nil || ret.String() != "192.168.1.2" {
		t.Errorf("функция IpHandler(), из контекста получен IP %v, ожидался 192.168.1.2", ret)
	}
}

func TestIpExtraction(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  *http.Request
		ip  net.IP
	)

	// Приоритетный заголовок X-Client-Forwarded-For.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Client-Forwarded-For", "10.1.1.1")
	rq.Header.Set("X-Real-Ip", "10.2.2.2")
	rq.Header.Set("X-Forwarded-For", "10.3.3.3")
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.1.1.1" {
		t.Errorf("IpExtraction() = %v, ожидался 10.1.1.1", ip)
	}
	// Заголовок X-Real-Ip.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Real-Ip", "10.2.2.2")
	rq.Header.Set("X-Forwarded-For", "10.3.3.3")
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.2.2.2" {
		t.Errorf("IpExtraction() = %v, ожидался 10.2.2.2", ip)
	}
	// Заголовок X-Forwarded-For со списком адресов.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Forwarded-For", "10.3.3.3, 10.5.5.5, 10.6.6.6")
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.3.3.3" {
		t.Errorf("IpExtraction() = %v, ожидался 10.3.3.3", ip)
	}
	// Заголовок X-Forwarded-For с одним адресом.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Forwarded-For", " 10.3.3.3 ")
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.3.3.3" {
		t.Errorf("IpExtraction() = %v, ожидался 10.3.3.3", ip)
	}
	// Адрес клиента из RemoteAddr.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.RemoteAddr = "10.4.4.4:1234"
	if ip = iwl.IpExtraction(rq); ip == nil || ip.String() != "10.4.4.4" {
		t.Errorf("IpExtraction() = %v, ожидался 10.4.4.4", ip)
	}
	// Пустый RemoteAddr.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.RemoteAddr = " "
	if ip = iwl.IpExtraction(rq); ip != nil {
		t.Errorf("IpExtraction() = %v, ожидался nil", ip)
	}
	// RemoteAddr без порта.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.RemoteAddr = "10.4.4.4"
	if ip = iwl.IpExtraction(rq); ip != nil {
		t.Errorf("IpExtraction() = %v, ожидался nil", ip)
	}
	// Не корректный IP адрес в приоритетном заголовке.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq.Header.Set("X-Client-Forwarded-For", "wrong ip")
	if ip = iwl.IpExtraction(rq); ip != nil {
		t.Errorf("IpExtraction() = %v, ожидался nil", ip)
	}
}

func TestIpGetFromContext(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/", nil)
	)

	// Объект IP адреса отсутствует в контексте.
	if ip := iwl.IpGetFromContext(rq); ip != nil {
		t.Errorf("функция IpGetFromContext() без данных в контексте, вернула %v, ожидался nil", ip)
	}
	// Объект IP адреса в контексте имеет не верный тип данных.
	rq = rq.WithContext(contextWith(rq, cKeyIP, cNameIP, "wrong value"))
	if ip := iwl.IpGetFromContext(rq); ip != nil {
		t.Errorf("функция IpGetFromContext() с не верным типом данных, вернула %v, ожидался nil", ip)
	}
	// Корректные данные в контексте.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(contextWith(rq, cKeyIP, cNameIP, net.ParseIP("192.168.1.2")))
	if ip := iwl.IpGetFromContext(rq); ip == nil || ip.String() != "192.168.1.2" {
		t.Errorf("функция IpGetFromContext(), вернула %v, ожидался 192.168.1.2", ip)
	}
}

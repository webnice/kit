package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoverHandler(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rsp = httptest.NewRecorder()
	)

	// Восстановление после паники, вызов обработчика ошибки InternalServerError по умолчанию.
	var handler = iwl.RecoverHandler()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("тестовая паника")
	}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf("функция RecoverHandler(), код ответа %d, ожидался код 500", rsp.Code)
	}
	if body := rsp.Body.String(); !strings.Contains(body, "тестовая паника") {
		t.Errorf("функция RecoverHandler(), тело ответа %q не содержит текст паники", body)
	}
	// Ошибка InternalServerError сохранена в объекте ошибок ВЕБ сервера.
	if err := srv.Web().Error().InternalServerError(nil); err == nil ||
		!strings.Contains(err.Error(), "тестовая паника") {
		t.Errorf("функция RecoverHandler(), ошибка InternalServerError = %v", err)
	}
}

func TestRecoverHandlerWithoutPanic(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rsp = httptest.NewRecorder()
		ok  bool
	)

	var handler = iwl.RecoverHandler()(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		ok = true
		wr.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rsp, rq)
	if !ok || rsp.Code != http.StatusOK {
		t.Errorf("функция RecoverHandler() без паники, обработчик не выполнен или код %d", rsp.Code)
	}
}

func TestRecoverHandlerCustomInternalServerError(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rsp = httptest.NewRecorder()
	)

	iwl.InternalServerErrorSet(func(wr http.ResponseWriter, _ *http.Request) {
		wr.WriteHeader(http.StatusTeapot)
		_, _ = wr.Write([]byte("custom error"))
	})
	var handler = iwl.RecoverHandler()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("тестовая паника")
	}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusTeapot || rsp.Body.String() != "custom error" {
		t.Errorf("функция RecoverHandler(), код %d и тело %q, ожидался код 418 и тело %q",
			rsp.Code, rsp.Body.String(), "custom error")
	}
}

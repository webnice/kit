package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoverHandler(t *testing.T) {
	const (
		tplCodeWrong   = "функция RecoverHandler(), код ответа %d, ожидался код 500."
		tplBodyNoPanic = "функция RecoverHandler(), тело ответа %q не содержит текст паники."
		tplErrorWrong  = "функция RecoverHandler(), ошибка InternalServerError = %v."
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
		rsp *httptest.ResponseRecorder
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rsp = httptest.NewRecorder()
	// Восстановление после паники, вызов обработчика ошибки InternalServerError по умолчанию.
	handler := iwl.RecoverHandler()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("тестовая паника")
	}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf(tplCodeWrong, rsp.Code)
	}
	if body := rsp.Body.String(); !strings.Contains(body, "тестовая паника") {
		t.Errorf(tplBodyNoPanic, body)
	}
	// Ошибка InternalServerError сохранена в объекте ошибок ВЕБ сервера.
	if err := srv.Web().Error().InternalServerError(nil); err == nil ||
		!strings.Contains(err.Error(), "тестовая паника") {
		t.Errorf(tplErrorWrong, err)
	}
}

func TestRecoverHandlerWithoutPanic(t *testing.T) {
	const (
		errNotExecuted = "функция RecoverHandler() без паники, обработчик не выполнен или код %d."
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
		rsp *httptest.ResponseRecorder
		ok  bool
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rsp = httptest.NewRecorder()
	handler := iwl.RecoverHandler()(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		ok = true
		wr.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rsp, rq)
	if !ok || rsp.Code != http.StatusOK {
		t.Errorf(errNotExecuted, rsp.Code)
	}
}

func TestRecoverHandlerCustomInternalServerError(t *testing.T) {
	const (
		tplCustomWrong = "функция RecoverHandler(), код %d и тело %q, ожидался код 418 и тело %q."
		customBody     = "custom error"
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
		rsp *httptest.ResponseRecorder
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rsp = httptest.NewRecorder()
	iwl.InternalServerErrorSet(func(wr http.ResponseWriter, _ *http.Request) {
		wr.WriteHeader(http.StatusTeapot)
		_, _ = wr.Write([]byte(customBody))
	})
	handler := iwl.RecoverHandler()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("тестовая паника")
	}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusTeapot || rsp.Body.String() != customBody {
		t.Errorf(tplCustomWrong, rsp.Code, rsp.Body.String(), customBody)
	}
}

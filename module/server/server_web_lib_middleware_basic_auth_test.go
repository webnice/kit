package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/webnice/dic"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

func basicAuthRequest(t *testing.T, username, password string) (rq *http.Request) {
	t.Helper()
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rq.SetBasicAuth(username, password)

	return
}

func TestBasicAuthHandlerWithoutAuthorization(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rsp = httptest.NewRecorder()
		ok  bool
	)

	var handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{IsCorrect: true}
		},
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { ok = true }))
	handler.ServeHTTP(rsp, rq)
	if ok {
		t.Error("функция BasicAuthHandler() без авторизации, обработчик не должен быть вызван")
	}
	if rsp.Code != http.StatusUnauthorized {
		t.Errorf("функция BasicAuthHandler() без авторизации, код %d, ожидался код 401", rsp.Code)
	}
	if auth := rsp.Header().Get(dic.Header().WwwAuthenticate.String()); !strings.Contains(auth, "Пожалуйста, авторизуйтесь") {
		t.Errorf("функция BasicAuthHandler() без авторизации, заголовок WWW-Authenticate = %q", auth)
	}
}

func TestBasicAuthHandlerCustomRequest(t *testing.T) {
	var (
		srv  = New(&testLogger{})
		iwl  = srv.(*impl).serverWeb.lib
		rq   = httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rsp  = httptest.NewRecorder()
		body = bytes.NewBufferString("custom body")
	)

	var handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{IsCorrect: true}
		},
		Request: "Custom realm message",
		Header:  map[dic.IHeader]string{dic.Header().ContentType: "text/plain"},
		Body:    body,
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusUnauthorized {
		t.Errorf("функция BasicAuthHandler(), код %d, ожидался код 401", rsp.Code)
	}
	if auth := rsp.Header().Get(dic.Header().WwwAuthenticate.String()); !strings.Contains(auth, "Custom realm message") {
		t.Errorf("функция BasicAuthHandler(), заголовок WWW-Authenticate = %q", auth)
	}
	if ct := rsp.Header().Get(dic.Header().ContentType.String()); ct != "text/plain" {
		t.Errorf("функция BasicAuthHandler(), заголовок Content-Type = %q", ct)
	}
	if rsp.Body.String() != "custom body" {
		t.Errorf("функция BasicAuthHandler(), тело %q, ожидалось %q", rsp.Body.String(), "custom body")
	}
}

func TestBasicAuthHandlerAuthFuncPanic(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = basicAuthRequest(t, "user", "password")
		rsp = httptest.NewRecorder()
		ok  bool
	)

	var handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse { panic("паника в AuthFunc") },
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { ok = true }))
	handler.ServeHTTP(rsp, rq)
	if ok {
		t.Error("функция BasicAuthHandler() с паникой AuthFunc, обработчик не должен быть вызван")
	}
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf("функция BasicAuthHandler() с паникой AuthFunc, код %d, ожидался код 500", rsp.Code)
	}
}

func TestBasicAuthHandlerIncorrect(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		ok  bool
	)

	// Не верные данные авторизации, код ответа по умолчанию.
	rq := basicAuthRequest(t, "user", "wrong")
	rsp := httptest.NewRecorder()
	var handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(_, password string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{IsCorrect: password == "password"}
		},
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { ok = true }))
	handler.ServeHTTP(rsp, rq)
	if ok || rsp.Code != http.StatusUnauthorized {
		t.Errorf("функция BasicAuthHandler() с не верными данными, ok=%v код=%d", ok, rsp.Code)
	}
	// Не верные данные авторизации, свой код ответа.
	rq, rsp = basicAuthRequest(t, "user", "wrong"), httptest.NewRecorder()
	handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{
				IsCorrect: false,
				Code:      dic.Status().Forbidden,
				Header:    map[dic.IHeader]string{dic.Header().ContentType: "text/plain"},
			}
		},
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { ok = true }))
	handler.ServeHTTP(rsp, rq)
	if ok || rsp.Code != http.StatusForbidden {
		t.Errorf("функция BasicAuthHandler() с не верными данными и кодом 403, ok=%v код=%d", ok, rsp.Code)
	}
	// Не верные данные авторизации, своё тело ответа.
	rq, rsp = basicAuthRequest(t, "user", "wrong"), httptest.NewRecorder()
	handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{
				IsCorrect: false,
				Code:      dic.Status().Forbidden,
				Body:      bytes.NewBufferString("access denied"),
			}
		},
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusForbidden || rsp.Body.String() != "access denied" {
		t.Errorf("функция BasicAuthHandler() с телом ответа, код=%d тело=%q", rsp.Code, rsp.Body.String())
	}
}

func TestBasicAuthHandlerCorrect(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = basicAuthRequest(t, "username", "password")
		rsp = httptest.NewRecorder()
		ok  bool
	)

	var handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(username, password string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{IsCorrect: username == "username" && password == "password"}
		},
	})(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ok = true
		// Имя пользователя доступно в контексте запроса.
		if user := iwl.BasicAuthGetFromContext(r); user != "username" {
			t.Errorf("функция BasicAuthGetFromContext(), вернула %q, ожидалось %q", user, "username")
		}
	}))
	handler.ServeHTTP(rsp, rq)
	if !ok {
		t.Error("функция BasicAuthHandler() с верными данными, обработчик не был вызван")
	}
}

func TestBasicAuthGetFromContext(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/", nil)
	)

	// Имя пользователя отсутствует в контексте.
	if ret := iwl.BasicAuthGetFromContext(rq); ret != "" {
		t.Errorf("функция BasicAuthGetFromContext() без данных, вернула %q, ожидалась пустая строка", ret)
	}
	// В контексте данные не верного типа.
	rq = rq.WithContext(contextWith(rq, cKeyBasicAuth, cNameBasicAuth, 42))
	if ret := iwl.BasicAuthGetFromContext(rq); ret != "" {
		t.Errorf("функция BasicAuthGetFromContext() с не верным типом, вернула %q, ожидалась пустая строка", ret)
	}
	// Корректные данные в контексте.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(contextWith(rq, cKeyBasicAuth, cNameBasicAuth, "username"))
	if ret := iwl.BasicAuthGetFromContext(rq); ret != "username" {
		t.Errorf("функция BasicAuthGetFromContext(), вернула %q, ожидалось %q", ret, "username")
	}
}

func TestBasicAuthHandlerAuthFuncError(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = basicAuthRequest(t, "user", "password")
		rsp = httptest.NewRecorder()
	)

	// Функция проверки авторизации возвращает ошибку через панику с восстановлением.
	var handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse {
			panic(errors.New("ошибка проверки"))
		},
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf("функция BasicAuthHandler() с ошибкой AuthFunc, код %d, ожидался код 500", rsp.Code)
	}
}

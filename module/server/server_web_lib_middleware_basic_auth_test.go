package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/webnice/dic"
	kitTypesServer "github.com/webnice/kit/v5/types/server"
)

func basicAuthRequest(t *testing.T, username, password string) (rq *http.Request) {
	t.Helper()
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rq.SetBasicAuth(username, password)

	return
}

func TestBasicAuthHandlerWithoutAuthorization(t *testing.T) {
	const (
		errHandlerCalled   = "функция BasicAuthHandler() без авторизации, обработчик не должен быть вызван."
		tplCodeWrong       = "функция BasicAuthHandler() без авторизации, код %d, ожидался код 401."
		tplAuthHeaderWrong = "функция BasicAuthHandler() без авторизации, заголовок WWW-Authenticate = %q."
		errTextAuthRequest = "Пожалуйста, авторизуйтесь"
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
	handler := iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{IsCorrect: true}
		},
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { ok = true }))
	handler.ServeHTTP(rsp, rq)
	if ok {
		t.Error(errHandlerCalled)
	}
	if rsp.Code != http.StatusUnauthorized {
		t.Errorf(tplCodeWrong, rsp.Code)
	}
	if auth := rsp.Header().Get(dic.Header().WwwAuthenticate.String()); !strings.Contains(auth, errTextAuthRequest) {
		t.Errorf(tplAuthHeaderWrong, auth)
	}
}

func TestBasicAuthHandlerCustomRequest(t *testing.T) {
	const (
		tplCodeWrong       = "функция BasicAuthHandler(), код %d, ожидался код 401."
		tplAuthHeaderWrong = "функция BasicAuthHandler(), заголовок WWW-Authenticate = %q."
		tplContentType     = "функция BasicAuthHandler(), заголовок Content-Type = %q."
		tplBodyWrong       = "функция BasicAuthHandler(), тело %q, ожидалось %q."
		customRealm        = "Custom realm message"
		customContentType  = "text/plain"
		customBody         = "custom body"
	)
	var (
		srv  Interface
		iwl  *implWebLib
		rq   *http.Request
		rsp  *httptest.ResponseRecorder
		body *bytes.Buffer
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rsp = httptest.NewRecorder()
	body = bytes.NewBufferString(customBody)
	handler := iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{IsCorrect: true}
		},
		Request: customRealm,
		Header:  map[dic.IHeader]string{dic.Header().ContentType: customContentType},
		Body:    body,
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusUnauthorized {
		t.Errorf(tplCodeWrong, rsp.Code)
	}
	if auth := rsp.Header().Get(dic.Header().WwwAuthenticate.String()); !strings.Contains(auth, customRealm) {
		t.Errorf(tplAuthHeaderWrong, auth)
	}
	if ct := rsp.Header().Get(dic.Header().ContentType.String()); ct != customContentType {
		t.Errorf(tplContentType, ct)
	}
	if rsp.Body.String() != customBody {
		t.Errorf(tplBodyWrong, rsp.Body.String(), customBody)
	}
}

func TestBasicAuthHandlerAuthFuncPanic(t *testing.T) {
	const (
		errHandlerCalled = "функция BasicAuthHandler() с паникой AuthFunc, обработчик не должен быть вызван."
		tplCodeWrong     = "функция BasicAuthHandler() с паникой AuthFunc, код %d, ожидался код 500."
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
	rq = basicAuthRequest(t, "user", "password")
	rsp = httptest.NewRecorder()
	handler := iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse { panic("паника в AuthFunc") },
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { ok = true }))
	handler.ServeHTTP(rsp, rq)
	if ok {
		t.Error(errHandlerCalled)
	}
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf(tplCodeWrong, rsp.Code)
	}
}

func TestBasicAuthHandlerIncorrect(t *testing.T) {
	const (
		tplIncorrectWrong = "функция BasicAuthHandler() с не верными данными, ok=%v код=%d."
		tplForbiddenWrong = "функция BasicAuthHandler() с не верными данными и кодом 403, ok=%v код=%d."
		tplBodyWrong      = "функция BasicAuthHandler() с телом ответа, код=%d тело=%q."
		testUser          = "user"
		testPassword      = "password"
		accessDeniedBody  = "access denied"
	)
	var (
		srv     Interface
		iwl     *implWebLib
		rq      *http.Request
		rsp     *httptest.ResponseRecorder
		handler http.Handler
		ok      bool
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	// Не верные данные авторизации, код ответа по умолчанию.
	rq = basicAuthRequest(t, testUser, "wrong")
	rsp = httptest.NewRecorder()
	handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(_, password string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{IsCorrect: password == testPassword}
		},
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { ok = true }))
	handler.ServeHTTP(rsp, rq)
	if ok || rsp.Code != http.StatusUnauthorized {
		t.Errorf(tplIncorrectWrong, ok, rsp.Code)
	}
	// Не верные данные авторизации, свой код ответа.
	rq, rsp = basicAuthRequest(t, testUser, "wrong"), httptest.NewRecorder()
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
		t.Errorf(tplForbiddenWrong, ok, rsp.Code)
	}
	// Не верные данные авторизации, своё тело ответа.
	rq, rsp = basicAuthRequest(t, testUser, "wrong"), httptest.NewRecorder()
	handler = iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{
				IsCorrect: false,
				Code:      dic.Status().Forbidden,
				Body:      bytes.NewBufferString(accessDeniedBody),
			}
		},
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusForbidden || rsp.Body.String() != accessDeniedBody {
		t.Errorf(tplBodyWrong, rsp.Code, rsp.Body.String())
	}
}

func TestBasicAuthHandlerCorrect(t *testing.T) {
	const (
		errHandlerNotCalled = "функция BasicAuthHandler() с верными данными, обработчик не был вызван."
		tplUserWrong        = "функция BasicAuthGetFromContext(), вернула %q, ожидалось %q."
		testUsername        = "username"
		testPassword        = "password"
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
	rq = basicAuthRequest(t, testUsername, testPassword)
	rsp = httptest.NewRecorder()
	handler := iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(username, password string) kitTypesServer.BasicAuthResponse {
			return kitTypesServer.BasicAuthResponse{IsCorrect: username == testUsername && password == testPassword}
		},
	})(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ok = true
		// Имя пользователя доступно в контексте запроса.
		if user := iwl.BasicAuthGetFromContext(r); user != testUsername {
			t.Errorf(tplUserWrong, user, testUsername)
		}
	}))
	handler.ServeHTTP(rsp, rq)
	if !ok {
		t.Error(errHandlerNotCalled)
	}
}

func TestBasicAuthGetFromContext(t *testing.T) {
	const (
		tplNoDataNotEmpty    = "функция BasicAuthGetFromContext() без данных, вернула %q, ожидалась пустая строка."
		tplWrongTypeNotEmpty = "функция BasicAuthGetFromContext() с не верным типом, вернула %q, ожидалась пустая строка."
		tplUserWrong         = "функция BasicAuthGetFromContext(), вернула %q, ожидалось %q."
		testUsername         = "username"
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	// Имя пользователя отсутствует в контексте.
	if ret := iwl.BasicAuthGetFromContext(rq); ret != "" {
		t.Errorf(tplNoDataNotEmpty, ret)
	}
	// В контексте данные не верного типа.
	rq = rq.WithContext(contextWith(rq, cKeyBasicAuth, cNameBasicAuth, 42))
	if ret := iwl.BasicAuthGetFromContext(rq); ret != "" {
		t.Errorf(tplWrongTypeNotEmpty, ret)
	}
	// Корректные данные в контексте.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(contextWith(rq, cKeyBasicAuth, cNameBasicAuth, testUsername))
	if ret := iwl.BasicAuthGetFromContext(rq); ret != testUsername {
		t.Errorf(tplUserWrong, ret, testUsername)
	}
}

func TestBasicAuthHandlerAuthFuncError(t *testing.T) {
	const (
		tplCodeWrong     = "функция BasicAuthHandler() с ошибкой AuthFunc, код %d, ожидался код 500."
		errTextAuthError = "ошибка проверки"
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
		rsp *httptest.ResponseRecorder
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = basicAuthRequest(t, "user", "password")
	rsp = httptest.NewRecorder()
	// Функция проверки авторизации возвращает ошибку через панику с восстановлением.
	handler := iwl.BasicAuthHandler(kitTypesServer.BasicAuthConfiguration{
		AuthFunc: func(string, string) kitTypesServer.BasicAuthResponse {
			panic(errors.New(errTextAuthError))
		},
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	handler.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf(tplCodeWrong, rsp.Code)
	}
}

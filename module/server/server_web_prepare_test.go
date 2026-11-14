package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/webnice/dic"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

func TestPrepareSuccess(t *testing.T) {
	const (
		tplFatalPrepare     = "функция Prepare() вернула не ожидаемую ошибку: %s."
		tplBeforeCount      = "функция Before() вызвана %d раз, ожидался 1 вызов."
		errRouterNil        = "функция Prepare(), роутер не создан."
		tplRoutingWrong     = "роутинг вернул код %d и тело %q, ожидался код 200 и тело %q."
		errMiddlewareNotRun = "промежуточный слой ресурса не был выполнен."
	)
	var (
		srv          Interface
		obj          *impl
		cfg          *kitTypesServer.WebConfiguration
		beforeCalled int
		middleware   bool
		rq           *http.Request
		rsp          *httptest.ResponseRecorder
		err          error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	cfg = testWebConfiguration(t)
	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	srv.Gist().Debug(true)
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Path: "/api",
			Middleware: []kitTypesServer.WebMiddlewareFn{
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(wr http.ResponseWriter, rq *http.Request) {
						middleware = true
						next.ServeHTTP(wr, rq)
					})
				},
			},
			Controller: []kitTypesServer.WebController{{
				Path:   "/ping",
				Method: []dic.IMethod{dic.Method().Get},
				Controller: func(wr http.ResponseWriter, _ *http.Request) {
					wr.WriteHeader(http.StatusOK)
					_, _ = wr.Write([]byte("pong"))
				},
			}},
			Before: func(_ *kitTypesServer.Server) error {
				beforeCalled++
				return nil
			},
		}
	}))
	if err = obj.serverWeb.Prepare(cfg); err != nil {
		t.Fatalf(tplFatalPrepare, err)
	}
	if beforeCalled != 1 {
		t.Errorf(tplBeforeCount, beforeCalled)
	}
	if obj.serverWeb.router == nil {
		t.Fatal(errRouterNil)
	}
	// Проверка созданного роутинга.
	rq = httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rsp = httptest.NewRecorder()
	obj.serverWeb.router.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusOK || rsp.Body.String() != "pong" {
		t.Errorf(tplRoutingWrong, rsp.Code, rsp.Body.String(), "pong")
	}
	if !middleware {
		t.Error(errMiddlewareNotRun)
	}
}

func TestPrepareResetErrors(t *testing.T) {
	const (
		tplFatalPrepare   = "функция Prepare() вернула не ожидаемую ошибку: %s."
		tplErrorsNotReset = "функция Prepare() не сбросила ошибки ВЕБ сервера, вернула %v."
		errTextStaleError = "устаревшая ошибка"
	)
	var (
		srv Interface
		obj *impl
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	cfg = testWebConfiguration(t)
	obj.serverWeb.error.InternalServerError(fmt.Errorf(errTextStaleError))
	if err = obj.serverWeb.Prepare(cfg); err != nil {
		t.Fatalf(tplFatalPrepare, err)
	}
	if ret := obj.serverWeb.error.InternalServerError(nil); ret != nil {
		t.Errorf(tplErrorsNotReset, ret)
	}
}

func TestCallBeforeNilResource(t *testing.T) {
	const (
		tplFatalPrepareNil = "функция Prepare() с nil ресурсом, вернула не ожидаемую ошибку: %s."
	)
	var (
		srv Interface
		obj *impl
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	cfg = testWebConfiguration(t)
	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	obj.serverWeb.res = append(obj.serverWeb.res, nil)
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{Path: "/api"}
	}))
	if err = obj.serverWeb.Prepare(cfg); err != nil {
		t.Fatalf(tplFatalPrepareNil, err)
	}
}

func TestCallBeforeResourceReturnNil(t *testing.T) {
	const (
		tplCallBeforeWrong = "функция callBefore() с nil описанием ресурса, вернула не ожидаемую ошибку: %v."
	)
	var (
		srv Interface
		obj *impl
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	cfg = testWebConfiguration(t)
	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return nil
	}))
	if err = obj.serverWeb.callBefore(); err == nil ||
		!strings.Contains(err.Error(), "прервано паникой") {
		t.Errorf(tplCallBeforeWrong, err)
	}
}

func TestCallBeforeWithoutBefore(t *testing.T) {
	const (
		tplCallBeforeWrong = "функция callBefore() без функции Before, вернула не ожидаемую ошибку: %s."
	)
	var (
		srv Interface
		obj *impl
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	cfg = testWebConfiguration(t)
	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{Path: "/api"}
	}))
	if err = obj.serverWeb.callBefore(); err != nil {
		t.Errorf(tplCallBeforeWrong, err)
	}
}

func TestCallBeforeNotWebServer(t *testing.T) {
	const (
		tplCallBeforeWrong = "функция callBefore(), вернула не ожидаемую ошибку: %s."
		tplBeforeCount     = "функция Before() вызвана %d раз, ожидалось 0 вызовов."
	)
	var (
		srv          Interface
		obj          *impl
		beforeCalled int
		err          error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	// Сервер не ВЕБ типа - функция Before не вызывается.
	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TTcp})
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Path:   "/api",
			Before: func(_ *kitTypesServer.Server) error { beforeCalled++; return nil },
		}
	}))
	if err = obj.serverWeb.callBefore(); err != nil {
		t.Errorf(tplCallBeforeWrong, err)
	}
	if beforeCalled != 0 {
		t.Errorf(tplBeforeCount, beforeCalled)
	}
}

func TestCallBeforeError(t *testing.T) {
	const (
		tplCallBeforeWrong = "функция callBefore() с ошибкой Before, вернула не ожидаемую ошибку: %v."
		errTextBeforeError = "ошибка Before"
	)
	var (
		srv Interface
		obj *impl
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	cfg = testWebConfiguration(t)
	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Path:   "/api",
			Before: func(_ *kitTypesServer.Server) error { return fmt.Errorf(errTextBeforeError) },
		}
	}))
	if err = obj.serverWeb.callBefore(); err == nil ||
		!strings.Contains(err.Error(), "Before()") {
		t.Errorf(tplCallBeforeWrong, err)
	}
}

func TestCallBeforePanic(t *testing.T) {
	const (
		tplCallBeforeWrong = "функция callBefore() с паникой Before, вернула не ожидаемую ошибку: %v."
	)
	var (
		srv Interface
		obj *impl
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	cfg = testWebConfiguration(t)
	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Path: "/api",
			Before: func(_ *kitTypesServer.Server) error {
				panic("паника в Before")
			},
		}
	}))
	if err = obj.serverWeb.callBefore(); err == nil ||
		!strings.Contains(err.Error(), "прервано паникой") {
		t.Errorf(tplCallBeforeWrong, err)
	}
}

func TestMakeRoutingPanic(t *testing.T) {
	const (
		tplMakeRoutingWrong = "функция makeRouting() с паникой Resource, вернула не ожидаемую ошибку: %v."
	)
	var (
		srv Interface
		obj *impl
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	cfg = testWebConfiguration(t)
	obj.serverWeb.Reg(testResourceFunc(func(c *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		if c == nil {
			return &kitTypesServer.Web{Path: "/api"}
		}
		panic("паника в Resource")
	}))
	if err = obj.serverWeb.Prepare(cfg); err == nil ||
		!strings.Contains(err.Error(), "прервано паникой") {
		t.Errorf(tplMakeRoutingWrong, err)
	}
}

func TestMakeRoutingGroups(t *testing.T) {
	const (
		serverMiddlewareHeader    = "X-Server-Middleware"
		tplFatalPrepare           = "функция makeRouting() вернула не ожидаемую ошибку: %s."
		tplRoutingWrong           = "роутинг вернул код %d и тело %q, ожидался код 200 и тело %q."
		errServerMiddlewareNotRun = "промежуточный слой сервера не был выполнен."
		tplRoutingAllWrong        = "роутинг вернул код %d и тело %q, ожидался код 202 и тело %q."
		tplRoutingCodeWrong       = "роутинг вернул код %d, ожидался код 404."
	)
	var (
		srv Interface
		obj *impl
		cfg *kitTypesServer.WebConfiguration
		rq  *http.Request
		rsp *httptest.ResponseRecorder
		err error
	)

	srv = New(&testLogger{})
	obj = srv.(*impl)
	cfg = testWebConfiguration(t)
	// Ресурс без пути - промежуточный слой подключается к серверу целиком.
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Middleware: []kitTypesServer.WebMiddlewareFn{
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(wr http.ResponseWriter, rq *http.Request) {
						wr.Header().Set(serverMiddlewareHeader, "on")
						next.ServeHTTP(wr, rq)
					})
				},
			},
			Path: "",
		}
	}))
	// Два ресурса с одним базовым путем - группировка, контроллер с методами и без методов.
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Path: "/api",
			Controller: []kitTypesServer.WebController{{
				Path:   "/ping",
				Method: []dic.IMethod{dic.Method().Get},
				Middleware: []kitTypesServer.WebMiddlewareFn{
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(wr http.ResponseWriter, rq *http.Request) {
							wr.Header().Set("X-Controller-Middleware", "on")
							next.ServeHTTP(wr, rq)
						})
					},
				},
				Controller: func(wr http.ResponseWriter, _ *http.Request) {
					wr.WriteHeader(http.StatusOK)
					_, _ = wr.Write([]byte("pong"))
				},
			}},
		}
	}))
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Path: "/api",
			Controller: []kitTypesServer.WebController{{
				Path: "/all",
				Controller: func(wr http.ResponseWriter, _ *http.Request) {
					wr.WriteHeader(http.StatusAccepted)
					_, _ = wr.Write([]byte("all"))
				},
			}},
		}
	}))
	if err = obj.serverWeb.Prepare(cfg); err != nil {
		t.Fatalf(tplFatalPrepare, err)
	}
	// Проверка роутинга: контроллер с методом.
	rq = httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rsp = httptest.NewRecorder()
	obj.serverWeb.router.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusOK || rsp.Body.String() != "pong" {
		t.Errorf(tplRoutingWrong, rsp.Code, rsp.Body.String(), "pong")
	}
	if rsp.Header().Get(serverMiddlewareHeader) != "on" {
		t.Error(errServerMiddlewareNotRun)
	}
	// Проверка роутинга: контроллер без методов.
	rq, rsp = httptest.NewRequest(http.MethodDelete, "/api/all", nil), httptest.NewRecorder()
	obj.serverWeb.router.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusAccepted || rsp.Body.String() != "all" {
		t.Errorf(tplRoutingAllWrong, rsp.Code, rsp.Body.String(), "all")
	}
	// Проверка роутинга: не зарегистрированный путь.
	rq, rsp = httptest.NewRequest(http.MethodGet, "/api/unknown", nil), httptest.NewRecorder()
	obj.serverWeb.router.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusNotFound {
		t.Errorf(tplRoutingCodeWrong, rsp.Code)
	}
}

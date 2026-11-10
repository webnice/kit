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
	var (
		srv          = New(&testLogger{})
		obj          = srv.(*impl)
		cfg          = testWebConfiguration(t)
		beforeCalled int
		middleware   = false
		err          error
	)

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
		t.Fatalf("функция Prepare() вернула не ожидаемую ошибку: %s", err)
	}
	if beforeCalled != 1 {
		t.Errorf("функция Before() вызвана %d раз, ожидался 1 вызов", beforeCalled)
	}
	if obj.serverWeb.router == nil {
		t.Fatal("функция Prepare(), роутер не создан")
	}
	// Проверка созданного роутинга.
	var (
		rq  = httptest.NewRequest(http.MethodGet, "/api/ping", nil)
		rsp = httptest.NewRecorder()
	)
	obj.serverWeb.router.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusOK || rsp.Body.String() != "pong" {
		t.Errorf("роутинг вернул код %d и тело %q, ожидался код 200 и тело %q",
			rsp.Code, rsp.Body.String(), "pong")
	}
	if !middleware {
		t.Error("промежуточный слой ресурса не был выполнен")
	}
}

func TestPrepareResetErrors(t *testing.T) {
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
		cfg = testWebConfiguration(t)
		err error
	)

	obj.serverWeb.error.InternalServerError(fmt.Errorf("устаревшая ошибка"))
	if err = obj.serverWeb.Prepare(cfg); err != nil {
		t.Fatalf("функция Prepare() вернула не ожидаемую ошибку: %s", err)
	}
	if ret := obj.serverWeb.error.InternalServerError(nil); ret != nil {
		t.Errorf("функция Prepare() не сбросила ошибки ВЕБ сервера, вернула %v", ret)
	}
}

func TestCallBeforeNilResource(t *testing.T) {
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
		cfg = testWebConfiguration(t)
		err error
	)

	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	obj.serverWeb.res = append(obj.serverWeb.res, nil)
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{Path: "/api"}
	}))
	if err = obj.serverWeb.Prepare(cfg); err != nil {
		t.Fatalf("функция Prepare() с nil ресурсом, вернула не ожидаемую ошибку: %s", err)
	}
}

func TestCallBeforeResourceReturnNil(t *testing.T) {
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
		cfg = testWebConfiguration(t)
		err error
	)

	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return nil
	}))
	if err = obj.serverWeb.callBefore(); err == nil ||
		!strings.Contains(err.Error(), "прервано паникой") {
		t.Errorf("функция callBefore() с nil описанием ресурса, вернула не ожидаемую ошибку: %v", err)
	}
}

func TestCallBeforeWithoutBefore(t *testing.T) {
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
		cfg = testWebConfiguration(t)
		err error
	)

	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{Path: "/api"}
	}))
	if err = obj.serverWeb.callBefore(); err != nil {
		t.Errorf("функция callBefore() без функции Before, вернула не ожидаемую ошибку: %s", err)
	}
}

func TestCallBeforeNotWebServer(t *testing.T) {
	var (
		srv          = New(&testLogger{})
		obj          = srv.(*impl)
		beforeCalled int
		err          error
	)

	// Сервер не ВЕБ типа - функция Before не вызывается.
	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TTcp})
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Path:   "/api",
			Before: func(_ *kitTypesServer.Server) error { beforeCalled++; return nil },
		}
	}))
	if err = obj.serverWeb.callBefore(); err != nil {
		t.Errorf("функция callBefore(), вернула не ожидаемую ошибку: %s", err)
	}
	if beforeCalled != 0 {
		t.Errorf("функция Before() вызвана %d раз, ожидалось 0 вызовов", beforeCalled)
	}
}

func TestCallBeforeError(t *testing.T) {
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
		cfg = testWebConfiguration(t)
		err error
	)

	obj.server = append(obj.server, &kitTypesServer.Server{T: kitTypesServer.TWeb, Web: cfg})
	obj.serverWeb.Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Path:   "/api",
			Before: func(_ *kitTypesServer.Server) error { return fmt.Errorf("ошибка Before") },
		}
	}))
	if err = obj.serverWeb.callBefore(); err == nil ||
		!strings.Contains(err.Error(), "Before()") {
		t.Errorf("функция callBefore() с ошибкой Before, вернула не ожидаемую ошибку: %v", err)
	}
}

func TestCallBeforePanic(t *testing.T) {
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
		cfg = testWebConfiguration(t)
		err error
	)

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
		t.Errorf("функция callBefore() с паникой Before, вернула не ожидаемую ошибку: %v", err)
	}
}

func TestMakeRoutingPanic(t *testing.T) {
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
		cfg = testWebConfiguration(t)
		err error
	)

	obj.serverWeb.Reg(testResourceFunc(func(c *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		if c == nil {
			return &kitTypesServer.Web{Path: "/api"}
		}
		panic("паника в Resource")
	}))
	if err = obj.serverWeb.Prepare(cfg); err == nil ||
		!strings.Contains(err.Error(), "прервано паникой") {
		t.Errorf("функция makeRouting() с паникой Resource, вернула не ожидаемую ошибку: %v", err)
	}
}

func TestMakeRoutingGroups(t *testing.T) {
	const serverMiddlewareHeader = "X-Server-Middleware"
	var (
		srv = New(&testLogger{})
		obj = srv.(*impl)
		cfg = testWebConfiguration(t)
		err error
	)

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
		t.Fatalf("функция makeRouting() вернула не ожидаемую ошибку: %s", err)
	}
	// Проверка роутинга: контроллер с методом.
	var (
		rq  = httptest.NewRequest(http.MethodGet, "/api/ping", nil)
		rsp = httptest.NewRecorder()
	)
	obj.serverWeb.router.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusOK || rsp.Body.String() != "pong" {
		t.Errorf("роутинг вернул код %d и тело %q, ожидался код 200 и тело %q", rsp.Code, rsp.Body.String(), "pong")
	}
	if rsp.Header().Get(serverMiddlewareHeader) != "on" {
		t.Error("промежуточный слой сервера не был выполнен")
	}
	// Проверка роутинга: контроллер без методов.
	rq, rsp = httptest.NewRequest(http.MethodDelete, "/api/all", nil), httptest.NewRecorder()
	obj.serverWeb.router.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusAccepted || rsp.Body.String() != "all" {
		t.Errorf("роутинг вернул код %d и тело %q, ожидался код 202 и тело %q", rsp.Code, rsp.Body.String(), "all")
	}
	// Проверка роутинга: не зарегистрированный путь.
	rq, rsp = httptest.NewRequest(http.MethodGet, "/api/unknown", nil), httptest.NewRecorder()
	obj.serverWeb.router.ServeHTTP(rsp, rq)
	if rsp.Code != http.StatusNotFound {
		t.Errorf("роутинг вернул код %d, ожидался код 404", rsp.Code)
	}
}

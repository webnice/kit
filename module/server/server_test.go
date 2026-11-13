package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/webnice/dic"
	kitModuleLogLevel "github.com/webnice/kit/v4/module/log/level"
	kitTypes "github.com/webnice/kit/v4/types"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// testLogger Заглушка интерфейса логирования.
type testLogger struct {
	kitTypes.Logger
}

func (tl *testLogger) Key(_ ...kitTypes.LoggerKey) kitTypes.Logger { return tl }

func (tl *testLogger) Time(_ time.Time) kitTypes.Logger { return tl }

func (tl *testLogger) StackBackCorrect(_ int) kitTypes.Logger { return tl }

func (tl *testLogger) Fatality(_ bool) kitTypes.Logger { return tl }

func (tl *testLogger) MessageWithLevel(_ kitModuleLogLevel.Level, _ string, _ ...any) {}

func (tl *testLogger) Fatal(_ ...any) {}

func (tl *testLogger) Fatalf(_ string, _ ...any) {}

func (tl *testLogger) Alert(_ ...any) {}

func (tl *testLogger) Alertf(_ string, _ ...any) {}

func (tl *testLogger) Critical(_ ...any) {}

func (tl *testLogger) Criticalf(_ string, _ ...any) {}

func (tl *testLogger) Error(_ ...any) {}

func (tl *testLogger) Errorf(_ string, _ ...any) {}

func (tl *testLogger) Warning(_ ...any) {}

func (tl *testLogger) Warningf(_ string, _ ...any) {}

func (tl *testLogger) Notice(_ ...any) {}

func (tl *testLogger) Noticef(_ string, _ ...any) {}

func (tl *testLogger) Info(_ ...any) {}

func (tl *testLogger) Infof(_ string, _ ...any) {}

func (tl *testLogger) Debug(_ ...any) {}

func (tl *testLogger) Debugf(_ string, _ ...any) {}

func (tl *testLogger) Trace(_ ...any) {}

func (tl *testLogger) Tracef(_ string, _ ...any) {}

// testResourceFunc Реализация интерфейса WebResource на основе функции.
type testResourceFunc func(cfg *kitTypesServer.WebConfiguration) *kitTypesServer.Web

func (trf testResourceFunc) Resource(cfg *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
	return trf(cfg)
}

// freeTCPPort Поиск свободного TCP/IP порта.
func freeTCPPort(t *testing.T) (port uint16) {
	var (
		err error
		ln  net.Listener
	)

	t.Helper()
	if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
		t.Fatalf("поиск свободного порта прерван ошибкой: %s", err)
	}
	port = uint16(ln.Addr().(*net.TCPAddr).Port)
	if err = ln.Close(); err != nil {
		t.Fatalf("освобождение порта прервано ошибкой: %s", err)
	}

	return
}

// testWebConfiguration Конфигурация тестового ВЕБ сервера на свободном порту.
func testWebConfiguration(t *testing.T) (cfg *kitTypesServer.WebConfiguration) {
	t.Helper()
	cfg = &kitTypesServer.WebConfiguration{}
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = freeTCPPort(t)

	return
}

// contextWith Контекст запроса с добавленной обёрткой данных контекста.
func contextWith(rq *http.Request, key any, name string, value any) context.Context {
	return context.WithValue(rq.Context(), key, &contextWrapper{Name: name, Value: value})
}

func TestNew(t *testing.T) {
	var srv = New(&testLogger{})

	if srv == nil {
		t.Fatal("функция New(), вернулся nil, ожидался не nil объект")
	}
	if srv.Gist() == nil {
		t.Error("функция Gist(), вернулся nil, ожидался не nil объект")
	}
	if srv.Web() == nil {
		t.Error("функция Web(), вернулся nil, ожидался не nil объект")
	}
	if srv.Errors() == nil {
		t.Error("функция Errors(), вернулся nil, ожидался не nil объект")
	}
}

func TestStartWithoutServers(t *testing.T) {
	var err = New(&testLogger{}).Start()

	if err == nil {
		t.Error("функция Start() без зарегистрированных серверов, ошибка не возвращена")
	}
	if err != nil && !strings.Contains(err.Error(), "Не зарегистрирован ни один сервер") {
		t.Errorf("функция Start() вернула не ожидаемую ошибку: %v", err)
	}
}

func TestStopWithoutServers(t *testing.T) {
	var err = New(&testLogger{}).Stop()

	if err == nil {
		t.Error("функция Stop() без зарегистрированных серверов, ошибка не возвращена")
	}
	if err != nil && !strings.Contains(err.Error(), "Не зарегистрирован ни один сервер") {
		t.Errorf("функция Stop() вернула не ожидаемую ошибку: %v", err)
	}
}

func TestStartPrepareError(t *testing.T) {
	var (
		srv = New(&testLogger{})
		cfg = testWebConfiguration(t)
		err error
	)

	srv.Web().Add(cfg)
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Before: func(_ *kitTypesServer.Server) error { return fmt.Errorf("ошибка подготовки") },
		}
	}))
	if err = srv.Start(); err == nil {
		t.Fatal("функция Start() с ошибкой подготовки, ошибка не возвращена")
	}
	if !strings.Contains(err.Error(), "подготовка ресурсов сервера прервана ошибкой") {
		t.Errorf("функция Start() вернула не ожидаемую ошибку: %v", err)
	}
}

func TestStartStopWebServer(t *testing.T) {
	const patternPing = "pong"
	var (
		srv  = New(&testLogger{})
		cfg  = testWebConfiguration(t)
		si   *kitTypesServer.Server
		err  error
		rsp  *http.Response
		body []byte
	)

	si = srv.Web().Add(cfg)
	if si == nil {
		t.Fatal("функция Add(), вернулся nil, ожидался не nil объект")
	}
	if si.T != kitTypesServer.TWeb || si.Web != cfg {
		t.Error("функция Add(), вернулся объект с не корректной конфигурацией")
	}
	if cfg.Server.ID == "" {
		t.Error("функция Add(), идентификатор сервера не создан")
	}
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Path: "/api",
			Controller: []kitTypesServer.WebController{{
				Path:   "/ping",
				Method: []dic.IMethod{dic.Method().Get},
				Controller: func(wr http.ResponseWriter, _ *http.Request) {
					wr.WriteHeader(http.StatusOK)
					_, _ = wr.Write([]byte(patternPing))
				},
			}},
		}
	}))
	if err = srv.Start(); err != nil {
		t.Fatalf("функция Start() вернула не ожидаемую ошибку: %s", err)
	}
	defer func() { _ = srv.Stop() }()
	if !srv.Web().IsStarted(cfg.Server.ID) {
		t.Error("функция IsStarted() после Start(), сервер не найден в списке запущенных")
	}
	// Проверка работы запущенного сервера.
	if rsp, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/ping", cfg.Server.Port)); err != nil {
		t.Fatalf("http запрос к запущенному серверу прерван ошибкой: %s", err)
	}
	if body, err = io.ReadAll(rsp.Body); err != nil {
		t.Fatalf("чтение тела ответа прервано ошибкой: %s", err)
	}
	_ = rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK || string(body) != patternPing {
		t.Errorf("http запрос, получен код %q и тело %q, ожидался код 200 и тело %q",
			rsp.Status, string(body), patternPing)
	}
	// Повторный запуск уже запущенного сервера.
	if err = srv.Start(); err == nil ||
		!strings.Contains(err.Error(), "Запуск серверов завершился ошибкой") {
		t.Errorf("функция Start() повторного запуска, вернула не ожидаемую ошибку: %v", err)
	}
	if err = srv.Stop(); err != nil {
		t.Errorf("функция Stop() вернула не ожидаемую ошибку: %s", err)
	}
	if srv.Web().IsStarted(cfg.Server.ID) {
		t.Error("функция IsStarted() после Stop(), сервер найден в списке запущенных")
	}
}

func TestStartAddressBusy(t *testing.T) {
	var (
		srv = New(&testLogger{})
		ln  net.Listener
		err error
	)

	// Занятие порта, чтобы запуск сервера на него завершился ошибкой.
	if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
		t.Fatalf("занятие порта прервано ошибкой: %s", err)
	}
	defer func() { _ = ln.Close() }()
	cfg := &kitTypesServer.WebConfiguration{}
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = uint16(ln.Addr().(*net.TCPAddr).Port)
	srv.Web().Add(cfg)
	if err = srv.Start(); err == nil ||
		!strings.Contains(err.Error(), "Запуск серверов завершился ошибкой") {
		t.Errorf("функция Start() с занятым адресом, вернула не ожидаемую ошибку: %v", err)
	}
}

func TestStopNotStartedServer(t *testing.T) {
	var (
		srv = New(&testLogger{})
		cfg = testWebConfiguration(t)
		err error
	)

	srv.Web().Add(cfg)
	// Остановка сервера, который не был запущен, ошибка останавливаемого сервера
	// логируется, функция завершается без ошибки.
	if err = srv.Stop(); err != nil {
		t.Errorf("функция Stop() не запущенного сервера, вернула не ожидаемую ошибку: %v", err)
	}
}

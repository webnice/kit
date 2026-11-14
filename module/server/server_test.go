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
	kitModuleLogLevel "github.com/webnice/kit/v5/module/log/level"
	kitTypes "github.com/webnice/kit/v5/types"
	kitTypesServer "github.com/webnice/kit/v5/types/server"
)

// testLogger Заглушка интерфейса логирования.
type testLogger struct {
	kitTypes.Logger
}

// Key Реализация интерфейса журналирования.
func (tl *testLogger) Key(_ ...kitTypes.LoggerKey) kitTypes.Logger { return tl }

// Time Реализация интерфейса журналирования.
func (tl *testLogger) Time(_ time.Time) kitTypes.Logger { return tl }

// StackBackCorrect Реализация интерфейса журналирования.
func (tl *testLogger) StackBackCorrect(_ int) kitTypes.Logger { return tl }

// Fatality Реализация интерфейса журналирования.
func (tl *testLogger) Fatality(_ bool) kitTypes.Logger { return tl }

// MessageWithLevel Реализация интерфейса журналирования.
func (tl *testLogger) MessageWithLevel(_ kitModuleLogLevel.Level, _ string, _ ...any) {}

// Fatal Реализация интерфейса журналирования.
func (tl *testLogger) Fatal(_ ...any) {}

// Fatalf Реализация интерфейса журналирования.
func (tl *testLogger) Fatalf(_ string, _ ...any) {}

// Alert Реализация интерфейса журналирования.
func (tl *testLogger) Alert(_ ...any) {}

// Alertf Реализация интерфейса журналирования.
func (tl *testLogger) Alertf(_ string, _ ...any) {}

// Critical Реализация интерфейса журналирования.
func (tl *testLogger) Critical(_ ...any) {}

// Criticalf Реализация интерфейса журналирования.
func (tl *testLogger) Criticalf(_ string, _ ...any) {}

// Error Реализация интерфейса журналирования.
func (tl *testLogger) Error(_ ...any) {}

// Errorf Реализация интерфейса журналирования.
func (tl *testLogger) Errorf(_ string, _ ...any) {}

// Warning Реализация интерфейса журналирования.
func (tl *testLogger) Warning(_ ...any) {}

// Warningf Реализация интерфейса журналирования.
func (tl *testLogger) Warningf(_ string, _ ...any) {}

// Notice Реализация интерфейса журналирования.
func (tl *testLogger) Notice(_ ...any) {}

// Noticef Реализация интерфейса журналирования.
func (tl *testLogger) Noticef(_ string, _ ...any) {}

// Info Реализация интерфейса журналирования.
func (tl *testLogger) Info(_ ...any) {}

// Infof Реализация интерфейса журналирования.
func (tl *testLogger) Infof(_ string, _ ...any) {}

// Debug Реализация интерфейса журналирования.
func (tl *testLogger) Debug(_ ...any) {}

// Debugf Реализация интерфейса журналирования.
func (tl *testLogger) Debugf(_ string, _ ...any) {}

// Trace Реализация интерфейса журналирования.
func (tl *testLogger) Trace(_ ...any) {}

// Tracef Реализация интерфейса журналирования.
func (tl *testLogger) Tracef(_ string, _ ...any) {}

// testResourceFunc Реализация интерфейса WebResource на основе функции.
type testResourceFunc func(cfg *kitTypesServer.WebConfiguration) *kitTypesServer.Web

// Resource Реализация интерфейса WebResource.
func (trf testResourceFunc) Resource(cfg *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
	return trf(cfg)
}

// Поиск свободного TCP/IP порта.
func freeTCPPort(t *testing.T) (port uint16) {
	const (
		tplFatalListen = "поиск свободного порта прерван ошибкой: %s."
		tplFatalClose  = "освобождение порта прервано ошибкой: %s."
	)
	var (
		err error
		ln  net.Listener
	)

	t.Helper()
	if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListen, err)
	}
	port = uint16(ln.Addr().(*net.TCPAddr).Port)
	if err = ln.Close(); err != nil {
		t.Fatalf(tplFatalClose, err)
	}

	return
}

// Конфигурация тестового ВЕБ сервера на свободном порту.
func testWebConfiguration(t *testing.T) (cfg *kitTypesServer.WebConfiguration) {
	t.Helper()
	cfg = &kitTypesServer.WebConfiguration{}
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = freeTCPPort(t)

	return
}

// Контекст запроса с добавленной обёрткой данных контекста.
func contextWith(rq *http.Request, key any, name string, value any) context.Context {
	return context.WithValue(rq.Context(), key, &contextWrapper{Name: name, Value: value})
}

func TestNew(t *testing.T) {
	const (
		errNewNil    = "функция New(), вернулся nil, ожидался не nil объект."
		errGistNil   = "функция Gist(), вернулся nil, ожидался не nil объект."
		errWebNil    = "функция Web(), вернулся nil, ожидался не nil объект."
		errErrorsNil = "функция Errors(), вернулся nil, ожидался не nil объект."
	)
	var srv Interface

	srv = New(&testLogger{})
	if srv == nil {
		t.Fatal(errNewNil)
	}
	if srv.Gist() == nil {
		t.Error(errGistNil)
	}
	if srv.Web() == nil {
		t.Error(errWebNil)
	}
	if srv.Errors() == nil {
		t.Error(errErrorsNil)
	}
}

func TestStartWithoutServers(t *testing.T) {
	const (
		errNoError      = "функция Start() без зарегистрированных серверов, ошибка не возвращена."
		tplStartWrong   = "функция Start() вернула не ожидаемую ошибку: %v."
		errTextNotFound = "Не зарегистрирован ни один сервер"
	)
	var err error

	err = New(&testLogger{}).Start()
	if err == nil {
		t.Error(errNoError)
	}
	if err != nil && !strings.Contains(err.Error(), errTextNotFound) {
		t.Errorf(tplStartWrong, err)
	}
}

func TestStopWithoutServers(t *testing.T) {
	const (
		errNoError      = "функция Stop() без зарегистрированных серверов, ошибка не возвращена."
		tplStopWrong    = "функция Stop() вернула не ожидаемую ошибку: %v."
		errTextNotFound = "Не зарегистрирован ни один сервер"
	)
	var err error

	err = New(&testLogger{}).Stop()
	if err == nil {
		t.Error(errNoError)
	}
	if err != nil && !strings.Contains(err.Error(), errTextNotFound) {
		t.Errorf(tplStopWrong, err)
	}
}

func TestStartPrepareError(t *testing.T) {
	const (
		tplFatalPrepare = "функция Start() с ошибкой подготовки, ошибка не возвращена."
		tplStartWrong   = "функция Start() вернула не ожидаемую ошибку: %v."
		errTextPrepare  = "подготовка ресурсов сервера прервана ошибкой"
		errTextBefore   = "ошибка подготовки"
	)
	var (
		srv Interface
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	cfg = testWebConfiguration(t)
	srv.Web().Add(cfg)
	srv.Web().Reg(testResourceFunc(func(_ *kitTypesServer.WebConfiguration) *kitTypesServer.Web {
		return &kitTypesServer.Web{
			Before: func(_ *kitTypesServer.Server) error { return fmt.Errorf(errTextBefore) },
		}
	}))
	if err = srv.Start(); err == nil {
		t.Fatal(tplFatalPrepare)
	}
	if !strings.Contains(err.Error(), errTextPrepare) {
		t.Errorf(tplStartWrong, err)
	}
}

func TestStartStopWebServer(t *testing.T) {
	const (
		patternPing          = "pong"
		tplFatalAddNil       = "функция Add(), вернулся nil, ожидался не nil объект."
		errAddConfigWrong    = "функция Add(), вернулся объект с не корректной конфигурацией."
		errAddIDEmpty        = "функция Add(), идентификатор сервера не создан."
		tplFatalStart        = "функция Start() вернула не ожидаемую ошибку: %s."
		errIsStartedNotFound = "функция IsStarted() после Start(), сервер не найден в списке запущенных."
		tplFatalHttpGet      = "http запрос к запущенному серверу прерван ошибкой: %s."
		tplFatalReadBody     = "чтение тела ответа прервано ошибкой: %s."
		tplHttpWrong         = "http запрос, получен код %q и тело %q, ожидался код 200 и тело %q."
		tplStartAgainWrong   = "функция Start() повторного запуска, вернула не ожидаемую ошибку: %v."
		tplStopWrong         = "функция Stop() вернула не ожидаемую ошибку: %s."
		errIsStartedFound    = "функция IsStarted() после Stop(), сервер найден в списке запущенных."
		errTextStartFailed   = "Запуск серверов завершился ошибкой"
	)
	var (
		srv  Interface
		cfg  *kitTypesServer.WebConfiguration
		si   *kitTypesServer.Server
		err  error
		rsp  *http.Response
		body []byte
	)

	srv = New(&testLogger{})
	cfg = testWebConfiguration(t)
	si = srv.Web().Add(cfg)
	if si == nil {
		t.Fatal(tplFatalAddNil)
	}
	if si.T != kitTypesServer.TWeb || si.Web != cfg {
		t.Error(errAddConfigWrong)
	}
	if cfg.Server.ID == "" {
		t.Error(errAddIDEmpty)
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
		t.Fatalf(tplFatalStart, err)
	}
	defer func() { _ = srv.Stop() }()
	if !srv.Web().IsStarted(cfg.Server.ID) {
		t.Error(errIsStartedNotFound)
	}
	// Проверка работы запущенного сервера.
	if rsp, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/ping", cfg.Server.Port)); err != nil {
		t.Fatalf(tplFatalHttpGet, err)
	}
	if body, err = io.ReadAll(rsp.Body); err != nil {
		t.Fatalf(tplFatalReadBody, err)
	}
	_ = rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK || string(body) != patternPing {
		t.Errorf(tplHttpWrong, rsp.Status, string(body), patternPing)
	}
	// Повторный запуск уже запущенного сервера.
	if err = srv.Start(); err == nil ||
		!strings.Contains(err.Error(), errTextStartFailed) {
		t.Errorf(tplStartAgainWrong, err)
	}
	if err = srv.Stop(); err != nil {
		t.Errorf(tplStopWrong, err)
	}
	if srv.Web().IsStarted(cfg.Server.ID) {
		t.Error(errIsStartedFound)
	}
}

func TestStartAddressBusy(t *testing.T) {
	const (
		tplFatalListen     = "занятие порта прервано ошибкой: %s."
		tplStartWrong      = "функция Start() с занятым адресом, вернула не ожидаемую ошибку: %v."
		errTextStartFailed = "Запуск серверов завершился ошибкой"
	)
	var (
		srv Interface
		ln  net.Listener
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	// Занятие порта, чтобы запуск сервера на него завершился ошибкой.
	if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
		t.Fatalf(tplFatalListen, err)
	}
	defer func() { _ = ln.Close() }()
	cfg = &kitTypesServer.WebConfiguration{}
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = uint16(ln.Addr().(*net.TCPAddr).Port)
	srv.Web().Add(cfg)
	if err = srv.Start(); err == nil ||
		!strings.Contains(err.Error(), errTextStartFailed) {
		t.Errorf(tplStartWrong, err)
	}
}

func TestStopNotStartedServer(t *testing.T) {
	const (
		tplStopWrong = "функция Stop() не запущенного сервера, вернула не ожидаемую ошибку: %v."
	)
	var (
		srv Interface
		cfg *kitTypesServer.WebConfiguration
		err error
	)

	srv = New(&testLogger{})
	cfg = testWebConfiguration(t)
	srv.Web().Add(cfg)
	// Остановка сервера, который не был запущен, ошибка останавливаемого сервера
	// логируется, функция завершается без ошибки.
	if err = srv.Stop(); err != nil {
		t.Errorf(tplStopWrong, err)
	}
}

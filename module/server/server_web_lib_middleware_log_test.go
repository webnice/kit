package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	kitModuleWrWrap "github.com/webnice/kit/v5/module/wrwrap"
)

func timeNow() time.Time { return time.Now() }

func TestLogHandlerDebugMode(t *testing.T) {
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	srv.Gist().Debug(true)
	handler := iwl.LogHandler()(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		wr.Header().Set("Location", "/api/moved")
		wr.Header().Set("Content-Length", "4")
		wr.WriteHeader(http.StatusOK)
		_, _ = wr.Write([]byte("data"))
	}))
	// В режиме отладки запрос любого кода попадает в журнал, паники нет.
	handler.ServeHTTP(httptest.NewRecorder(), rq)
}

func TestLogHandlerProductionMode(t *testing.T) {
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	// Режим "продакшн": штатные коды 200-299 в журнал не пишутся.
	handlerOK := iwl.LogHandler()(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		wr.WriteHeader(http.StatusOK)
	}))
	handlerOK.ServeHTTP(httptest.NewRecorder(), rq)
	// Режим "продакшн": коды 3xx, 4xx в журнал не пишутся.
	handlerNotFound := iwl.LogHandler()(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		wr.WriteHeader(http.StatusNotFound)
	}))
	handlerNotFound.ServeHTTP(httptest.NewRecorder(), rq)
	// Режим "продакшн": коды 500+ попадают в журнал.
	handlerError := iwl.LogHandler()(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		wr.WriteHeader(http.StatusInternalServerError)
	}))
	handlerError.ServeHTTP(httptest.NewRecorder(), rq)
}

func TestLogHandlerCodes(t *testing.T) {
	var (
		srv   Interface
		iwl   *implWebLib
		codes []int
		code  int
		n     int
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	srv.Gist().Debug(true)
	// Все диапазоны кодов ответов и проекция уровней логирования.
	codes = []int{
		http.StatusContinue,
		http.StatusOK,
		http.StatusMultipleChoices,
		http.StatusBadRequest,
		http.StatusInternalServerError,
		http.StatusNetworkAuthenticationRequired,
	}
	for n = range codes {
		code = codes[n]
		var (
			rq      = httptest.NewRequest(http.MethodGet, "/api/test", nil)
			handler = iwl.LogHandler()(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
				wr.WriteHeader(code)
			}))
		)
		handler.ServeHTTP(httptest.NewRecorder(), rq)
	}
}

// rawWriter ResponseWriter не проверяющий корректность кода статуса.
type rawWriter struct {
	header http.Header
	status int
	body   strings.Builder
}

func newRawWriter() *rawWriter { return &rawWriter{header: make(http.Header)} }

// Header Реализация интерфейса http.ResponseWriter.
func (rw *rawWriter) Header() http.Header { return rw.header }

// WriteHeader Реализация интерфейса http.ResponseWriter.
func (rw *rawWriter) WriteHeader(status int) { rw.status = status }

// Write Реализация интерфейса http.ResponseWriter.
func (rw *rawWriter) Write(p []byte) (int, error) { return rw.body.WriteString(string(p)) }

func TestLogHandlerCodeOutOfRange(t *testing.T) {
	var (
		srv Interface
		iwl *implWebLib
		rw  *rawWriter
		rq  *http.Request
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rw = newRawWriter()
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	srv.Gist().Debug(true)
	// Код ответа вне всех диапазонов - цвет записи по умолчанию.
	var wrw = kitModuleWrWrap.New(rw, true)
	wrw.WriteHeader(99)
	iwl.logHandler(timeNow(), wrw, rq)
}

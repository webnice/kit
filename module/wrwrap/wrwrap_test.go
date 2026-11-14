package wrwrap

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type minimalWriter struct {
	header      http.Header
	status      int
	headerCalls int
	body        bytes.Buffer
}

func newMinimalWriter() *minimalWriter {
	return &minimalWriter{header: make(http.Header)}
}

func (mw *minimalWriter) Header() http.Header { return mw.header }

func (mw *minimalWriter) WriteHeader(status int) {
	mw.headerCalls++
	if mw.status == 0 {
		mw.status = status
	}
}

func (mw *minimalWriter) Write(p []byte) (int, error) { return mw.body.Write(p) }

type failTeeWriter struct{ err error }

func (ftw *failTeeWriter) Write([]byte) (int, error) { return 0, ftw.err }

type pusherWriter struct {
	*minimalWriter
	pushed []string
}

func (pw *pusherWriter) Push(target string, _ *http.PushOptions) error {
	pw.pushed = append(pw.pushed, target)

	return nil
}

type readerFromWriter struct {
	*minimalWriter
	called bool
}

func (rfw *readerFromWriter) ReadFrom(r io.Reader) (int64, error) {
	rfw.called = true

	return io.Copy(&rfw.body, r)
}

// Строки тестовых ошибок.
const (
	errTextTeeError       = "tee error."
	errTextHijackError    = "hijack error."
	errTextWriteError     = "write error."
	errTextStringWriteErr = "string write error."
)

func TestWrWrapStatusLenHeaderAndTee(t *testing.T) {
	const (
		tplInitialWrong    = "начальные StatusCode/Len = %d/%d."
		errHeaderNotPass   = "Header() не пробрасывает заголовки."
		errEssenceWrong    = "Essence()/Unwrap() вернули не оригинальный writer."
		tplWriteWrong      = "Write() = %d, %v."
		tplAfterWriteWrong = "после Write(): status=%d len=%d body=%q tee=%q."
		tplStatusChanged   = "повторный WriteHeader изменил статус: %d/%d."
	)
	var (
		recorder  *httptest.ResponseRecorder
		recorder2 *httptest.ResponseRecorder
		wrw       Interface
		wrw2      Interface
		tee       bytes.Buffer
		n         int
		err       error
	)

	recorder = httptest.NewRecorder()
	wrw = New(recorder, true)
	if wrw.StatusCode() != http.StatusOK || wrw.Len() != 0 {
		t.Fatalf(tplInitialWrong, wrw.StatusCode(), wrw.Len())
	}
	wrw.Header().Set("X-Test", "value")
	if wrw.Header().Get("X-Test") != "value" || recorder.Header().Get("X-Test") != "value" {
		t.Fatal(errHeaderNotPass)
	}
	if wrw.Essence() != recorder || wrw.Unwrap() != recorder {
		t.Fatal(errEssenceWrong)
	}
	wrw.Tee(&tee)
	if n, err = wrw.Write([]byte("hello")); err != nil || n != 5 {
		t.Fatalf(tplWriteWrong, n, err)
	}
	if wrw.StatusCode() != http.StatusOK || wrw.Len() != 5 || recorder.Body.String() != "hello" || tee.String() != "hello" {
		t.Fatalf(tplAfterWriteWrong, wrw.StatusCode(), wrw.Len(), recorder.Body.String(), tee.String())
	}
	recorder2 = httptest.NewRecorder()
	wrw2 = New(recorder2, false)
	wrw2.WriteHeader(http.StatusCreated)
	wrw2.WriteHeader(http.StatusInternalServerError)
	if wrw2.StatusCode() != http.StatusCreated || recorder2.Code != http.StatusCreated {
		t.Fatalf(tplStatusChanged, wrw2.StatusCode(), recorder2.Code)
	}
}

func TestWrWrapMinimalWriterWriteHeaderOnce(t *testing.T) {
	const (
		tplStatusWrong = "minimalWriter status/headerCalls = %d/%d."
		tplWriteWrong  = "Write после WriteHeader изменил статус: %v/%d."
	)
	var (
		mw  *minimalWriter
		wrw Interface
	)

	mw = newMinimalWriter()
	wrw = New(mw, false)
	wrw.WriteHeader(http.StatusAccepted)
	wrw.WriteHeader(http.StatusConflict)
	if mw.status != http.StatusAccepted || mw.headerCalls != 1 {
		t.Fatalf(tplStatusWrong, mw.status, mw.headerCalls)
	}
	if _, err := wrw.Write([]byte("x")); err != nil || wrw.StatusCode() != http.StatusAccepted {
		t.Fatalf(tplWriteWrong, err, wrw.StatusCode())
	}
}

func TestWrWrapTeeError(t *testing.T) {
	const (
		tplTeeErrorWrong = "Write() с tee error = %d, %v."
		tplLenBodyWrong  = "Len/body после tee error = %d/%q."
	)
	var (
		recorder *httptest.ResponseRecorder
		expected error
		wrw      Interface
		n        int
		err      error
	)

	recorder = httptest.NewRecorder()
	expected = errors.New(errTextTeeError)
	wrw = New(recorder, false).Tee(&failTeeWriter{err: expected})
	n, err = wrw.Write([]byte("data"))
	if !errors.Is(err, expected) || n != 4 {
		t.Fatalf(tplTeeErrorWrong, n, err)
	}
	if wrw.Len() != 4 || recorder.Body.String() != "data" {
		t.Fatalf(tplLenBodyWrong, wrw.Len(), recorder.Body.String())
	}
}

func TestWrWrapWriteStringFallback(t *testing.T) {
	const (
		errStringWriterImplemented = "minimalWriter не должен реализовывать io.StringWriter."
		tplWriteStringWrong        = "WriteString() fallback = %d, %v."
		tplFallbackStateWrong      = "WriteString fallback state = %q/%d/%d."
		tplWriteStringResult       = "WriteString() = %d, %v, body=%q, len=%d."
	)
	var (
		mw       *minimalWriter
		recorder *httptest.ResponseRecorder
		wrw      Interface
		n        int
		err      error
	)

	mw = newMinimalWriter()
	wrw = New(mw, false)
	if _, ok := interface{}(mw).(io.StringWriter); ok {
		t.Fatal(errStringWriterImplemented)
	}
	n, err = wrw.WriteString("string")
	if err != nil || n != len("string") {
		t.Fatalf(tplWriteStringWrong, n, err)
	}
	if mw.body.String() != "string" || wrw.Len() != uint64(len("string")) || wrw.StatusCode() != http.StatusOK {
		t.Fatalf(tplFallbackStateWrong, mw.body.String(), wrw.Len(), wrw.StatusCode())
	}
	recorder = httptest.NewRecorder()
	wrw = New(recorder, false)
	if n, err = wrw.WriteString("abc"); err != nil || n != 3 || recorder.Body.String() != "abc" || wrw.Len() != 3 {
		t.Fatalf(tplWriteStringResult, n, err, recorder.Body.String(), wrw.Len())
	}
}

func TestWrWrapFlushAndPush(t *testing.T) {
	const (
		errFlushNotCalled = "Flush() не вызвал Flush у httptest.ResponseRecorder."
		tplPushWrong      = "Push() без поддержки = %v."
		tplPushError      = "Push() с поддержкой вернул ошибку: %v."
		tplPushTarget     = "Push() не передал target в essence: %#v."
	)
	var (
		recorder *httptest.ResponseRecorder
		mw       *minimalWriter
		pw       *pusherWriter
		wrw      Interface
		err      error
	)

	recorder = httptest.NewRecorder()
	wrw = New(recorder, false)
	wrw.Flush()
	if !recorder.Flushed {
		t.Fatal(errFlushNotCalled)
	}
	mw = newMinimalWriter()
	wrw = New(mw, false)
	wrw.Flush() // Отсутствие http.Flusher у essence не должно приводить к panic.
	if err = wrw.Push("/app.css", nil); !errors.Is(err, http.ErrNotSupported) {
		t.Fatalf(tplPushWrong, err)
	}
	pw = &pusherWriter{minimalWriter: newMinimalWriter()}
	wrw = New(pw, false)
	if err = wrw.Push("/app.js", &http.PushOptions{Method: http.MethodGet}); err != nil {
		t.Fatalf(tplPushError, err)
	}
	if len(pw.pushed) != 1 || pw.pushed[0] != "/app.js" {
		t.Fatalf(tplPushTarget, pw.pushed)
	}
}

func TestWrWrapHijackUnsupported(t *testing.T) {
	const (
		tplHijackWrong = "Hijack() без поддержки = %v, %v, %v."
	)
	var (
		wrw Interface
		con net.Conn
		rw  *bufio.ReadWriter
		err error
	)

	wrw = New(newMinimalWriter(), false)
	con, rw, err = wrw.Hijack()
	if !errors.Is(err, http.ErrHijacked) || con != nil || rw != nil {
		t.Fatalf(tplHijackWrong, con, rw, err)
	}
}

func TestWrWrapHijackSuccess(t *testing.T) {
	const (
		tplFatalGet    = "http.Get() вернул ошибку: %v."
		tplStatusWrong = "hijack response status = %d."
		errResponse500 = "HTTP/1.1 500 Internal Server Error\r\nContent-Length: %d\r\n\r\n%s."
		responseOK     = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 2\r\n\r\nOK."
	)
	var (
		server *httptest.Server
		resp   *http.Response
		err    error
	)

	server = httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		var (
			wrw    = New(wr, false)
			con    net.Conn
			rw     *bufio.ReadWriter
			errSub error
		)

		con, rw, errSub = wrw.Hijack()
		if errSub != nil {
			_, _ = fmt.Fprintf(rw, errResponse500, len(errSub.Error()), errSub.Error())
			_ = rw.Flush()
			_ = con.Close()

			return
		}
		_, _ = fmt.Fprint(rw, responseOK)
		_ = rw.Flush()
		_ = con.Close()
	}))
	defer server.Close()
	resp, err = http.Get(server.URL)
	if err != nil {
		t.Fatalf(tplFatalGet, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf(tplStatusWrong, resp.StatusCode)
	}
}

type hijackRecorder struct {
	*httptest.ResponseRecorder
	con net.Conn
	rw  *bufio.ReadWriter
	err error
}

func (hr *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return hr.con, hr.rw, hr.err
}

func TestWrWrapHijackDelegatesValues(t *testing.T) {
	const (
		tplHijackWrong = "Hijack() не вернул значения essence: %v, %v, %v."
	)
	var (
		recorder *httptest.ResponseRecorder
		expected error
		wrw      Interface
		con      net.Conn
		rw       *bufio.ReadWriter
		err      error
	)

	recorder = httptest.NewRecorder()
	expected = errors.New(errTextHijackError)
	wrw = New(&hijackRecorder{ResponseRecorder: recorder, err: expected}, false)
	con, rw, err = wrw.Hijack()
	if con != nil || rw != nil || !errors.Is(err, expected) {
		t.Fatalf(tplHijackWrong, con, rw, err)
	}
}

func TestWrWrapReadFrom(t *testing.T) {
	const (
		tplReadFromWrong    = "ReadFrom() = %d, %v."
		tplBodyLenWrong     = "ReadFrom body/len = %q/%d."
		tplPassthroughWrong = "ReadFrom() passthrough = %d, %v, called=%t."
		tplPassBodyLenWrong = "ReadFrom passthrough body/len = %q/%d."
		tplTeeWrong         = "ReadFrom() с Tee = %d, %v, called=%t."
		tplTeeBodyLenWrong  = "ReadFrom с Tee body/tee/len = %q/%q/%d."
	)
	var (
		recorder *httptest.ResponseRecorder
		rfw      *readerFromWriter
		wrw      Interface
		tee      bytes.Buffer
		n        int64
		err      error
	)

	recorder = httptest.NewRecorder()
	wrw = New(recorder, false)
	n, err = wrw.ReadFrom(strings.NewReader("read-from"))
	if err != nil || n != int64(len("read-from")) {
		t.Fatalf(tplReadFromWrong, n, err)
	}
	if recorder.Body.String() != "read-from" || wrw.Len() != uint64(len("read-from")) {
		t.Fatalf(tplBodyLenWrong, recorder.Body.String(), wrw.Len())
	}
	rfw = &readerFromWriter{minimalWriter: newMinimalWriter()}
	wrw = New(rfw, false)
	n, err = wrw.ReadFrom(strings.NewReader("fast"))
	if err != nil || n != 4 || !rfw.called {
		t.Fatalf(tplPassthroughWrong, n, err, rfw.called)
	}
	if rfw.body.String() != "fast" || wrw.Len() != 4 {
		t.Fatalf(tplPassBodyLenWrong, rfw.body.String(), wrw.Len())
	}
	rfw = &readerFromWriter{minimalWriter: newMinimalWriter()}
	wrw = New(rfw, false).Tee(&tee)
	n, err = wrw.ReadFrom(strings.NewReader("copy"))
	if err != nil || n != 4 || rfw.called {
		t.Fatalf(tplTeeWrong, n, err, rfw.called)
	}
	if rfw.body.String() != "copy" || tee.String() != "copy" || wrw.Len() != 4 {
		t.Fatalf(tplTeeBodyLenWrong, rfw.body.String(), tee.String(), wrw.Len())
	}
}

func TestWrWrapResponseControllerUnwrap(t *testing.T) {
	const (
		tplFlushWrong   = "ResponseController.Flush() вернул ошибку: %v."
		errFlushNoReach = "ResponseController через Unwrap не достиг оригинального Flusher."
	)
	var (
		recorder   *httptest.ResponseRecorder
		wrw        Interface
		controller *http.ResponseController
		err        error
	)

	recorder = httptest.NewRecorder()
	wrw = New(recorder, false)
	controller = http.NewResponseController(wrw)
	if err = controller.Flush(); err != nil {
		t.Fatalf(tplFlushWrong, err)
	}
	if !recorder.Flushed {
		t.Fatal(errFlushNoReach)
	}
}

type errorWriter struct {
	*minimalWriter
	err error
}

func (ew *errorWriter) Write([]byte) (int, error) { return 0, ew.err }

type stringWriter struct {
	*minimalWriter
	err error
}

func (sw *stringWriter) WriteString(s string) (int, error) {
	if sw.err != nil {
		return 0, sw.err
	}

	return sw.body.WriteString(s)
}

func TestWrWrapWriteError(t *testing.T) {
	const (
		tplWriteErrorWrong = "Write() error branch = %d, %v, len=%d, status=%d."
	)
	var (
		expected error
		wrw      Interface
		n        int
		err      error
	)

	expected = errors.New(errTextWriteError)
	wrw = New(&errorWriter{minimalWriter: newMinimalWriter(), err: expected}, false)
	n, err = wrw.Write([]byte("data"))
	if !errors.Is(err, expected) || n != 0 || wrw.Len() != 0 || wrw.StatusCode() != http.StatusOK {
		t.Fatalf(tplWriteErrorWrong, n, err, wrw.Len(), wrw.StatusCode())
	}
}

func TestWrWrapWriteStringErrorsAndTee(t *testing.T) {
	const (
		tplStringErrorWrong = "WriteString() error branch = %d, %v, len=%d."
		tplStringTeeWrong   = "WriteString() tee = %d, %v, tee=%q, len=%d."
		tplStringTeeErr     = "WriteString() tee error = %d, %v, len=%d."
	)
	var (
		expected error
		wrw      Interface
		tee      bytes.Buffer
		n        int
		err      error
	)

	expected = errors.New(errTextStringWriteErr)
	wrw = New(&stringWriter{minimalWriter: newMinimalWriter(), err: expected}, false)
	n, err = wrw.WriteString("data")
	if !errors.Is(err, expected) || n != 0 || wrw.Len() != 0 {
		t.Fatalf(tplStringErrorWrong, n, err, wrw.Len())
	}
	wrw = New(&stringWriter{minimalWriter: newMinimalWriter()}, false).Tee(&tee)
	n, err = wrw.WriteString("copy")
	if err != nil || n != 4 || tee.String() != "copy" || wrw.Len() != 4 {
		t.Fatalf(tplStringTeeWrong, n, err, tee.String(), wrw.Len())
	}
	wrw = New(&stringWriter{minimalWriter: newMinimalWriter()}, false).Tee(&failTeeWriter{err: expected})
	n, err = wrw.WriteString("data")
	if !errors.Is(err, expected) || n != 4 || wrw.Len() != 4 {
		t.Fatalf(tplStringTeeErr, n, err, wrw.Len())
	}
}

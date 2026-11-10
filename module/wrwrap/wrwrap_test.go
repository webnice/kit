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

func TestWrWrapStatusLenHeaderAndTee(t *testing.T) {
	var (
		recorder = httptest.NewRecorder()
		wrw      = New(recorder, true)
		tee      bytes.Buffer
	)

	if wrw.StatusCode() != http.StatusOK || wrw.Len() != 0 {
		t.Fatalf("начальные StatusCode/Len = %d/%d", wrw.StatusCode(), wrw.Len())
	}
	wrw.Header().Set("X-Test", "value")
	if wrw.Header().Get("X-Test") != "value" || recorder.Header().Get("X-Test") != "value" {
		t.Fatal("Header() не пробрасывает заголовки")
	}
	if wrw.Essence() != recorder || wrw.Unwrap() != recorder {
		t.Fatal("Essence()/Unwrap() вернули не оригинальный writer")
	}

	wrw.Tee(&tee)
	if n, err := wrw.Write([]byte("hello")); err != nil || n != 5 {
		t.Fatalf("Write() = %d, %v", n, err)
	}
	if wrw.StatusCode() != http.StatusOK || wrw.Len() != 5 || recorder.Body.String() != "hello" || tee.String() != "hello" {
		t.Fatalf("после Write(): status=%d len=%d body=%q tee=%q", wrw.StatusCode(), wrw.Len(), recorder.Body.String(), tee.String())
	}

	recorder2 := httptest.NewRecorder()
	wrw2 := New(recorder2, false)
	wrw2.WriteHeader(http.StatusCreated)
	wrw2.WriteHeader(http.StatusInternalServerError)
	if wrw2.StatusCode() != http.StatusCreated || recorder2.Code != http.StatusCreated {
		t.Fatalf("повторный WriteHeader изменил статус: %d/%d", wrw2.StatusCode(), recorder2.Code)
	}
}

func TestWrWrapMinimalWriterWriteHeaderOnce(t *testing.T) {
	var (
		mw  = newMinimalWriter()
		wrw = New(mw, false)
	)
	wrw.WriteHeader(http.StatusAccepted)
	wrw.WriteHeader(http.StatusConflict)
	if mw.status != http.StatusAccepted || mw.headerCalls != 1 {
		t.Fatalf("minimalWriter status/headerCalls = %d/%d", mw.status, mw.headerCalls)
	}
	if _, err := wrw.Write([]byte("x")); err != nil || wrw.StatusCode() != http.StatusAccepted {
		t.Fatalf("Write после WriteHeader изменил статус: %v/%d", err, wrw.StatusCode())
	}
}

func TestWrWrapTeeError(t *testing.T) {
	var (
		recorder = httptest.NewRecorder()
		expected = errors.New("tee error")
		wrw      = New(recorder, false).Tee(&failTeeWriter{err: expected})
	)
	n, err := wrw.Write([]byte("data"))
	if !errors.Is(err, expected) || n != 4 {
		t.Fatalf("Write() с tee error = %d, %v", n, err)
	}
	if wrw.Len() != 4 || recorder.Body.String() != "data" {
		t.Fatalf("Len/body после tee error = %d/%q", wrw.Len(), recorder.Body.String())
	}
}

func TestWrWrapWriteStringFallback(t *testing.T) {
	var (
		mw  = newMinimalWriter()
		wrw = New(mw, false)
	)
	if _, ok := interface{}(mw).(io.StringWriter); ok {
		t.Fatal("minimalWriter не должен реализовывать io.StringWriter")
	}
	n, err := wrw.WriteString("string")
	if err != nil || n != len("string") {
		t.Fatalf("WriteString() fallback = %d, %v", n, err)
	}
	if mw.body.String() != "string" || wrw.Len() != uint64(len("string")) || wrw.StatusCode() != http.StatusOK {
		t.Fatalf("WriteString fallback state = %q/%d/%d", mw.body.String(), wrw.Len(), wrw.StatusCode())
	}

	recorder := httptest.NewRecorder()
	wrw = New(recorder, false)
	if n, err = wrw.WriteString("abc"); err != nil || n != 3 || recorder.Body.String() != "abc" || wrw.Len() != 3 {
		t.Fatalf("WriteString() = %d, %v, body=%q, len=%d", n, err, recorder.Body.String(), wrw.Len())
	}
}

func TestWrWrapFlushAndPush(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrw := New(recorder, false)
	wrw.Flush()
	if !recorder.Flushed {
		t.Fatal("Flush() не вызвал Flush у httptest.ResponseRecorder")
	}

	mw := newMinimalWriter()
	wrw = New(mw, false)
	wrw.Flush() // Отсутствие http.Flusher у essence не должно приводить к panic.
	if err := wrw.Push("/app.css", nil); !errors.Is(err, http.ErrNotSupported) {
		t.Fatalf("Push() без поддержки = %v", err)
	}

	pw := &pusherWriter{minimalWriter: newMinimalWriter()}
	wrw = New(pw, false)
	if err := wrw.Push("/app.js", &http.PushOptions{Method: http.MethodGet}); err != nil {
		t.Fatalf("Push() с поддержкой вернул ошибку: %v", err)
	}
	if len(pw.pushed) != 1 || pw.pushed[0] != "/app.js" {
		t.Fatalf("Push() не передал target в essence: %#v", pw.pushed)
	}
}

func TestWrWrapHijackUnsupported(t *testing.T) {
	wrw := New(newMinimalWriter(), false)
	con, rw, err := wrw.Hijack()
	if !errors.Is(err, http.ErrHijacked) || con != nil || rw != nil {
		t.Fatalf("Hijack() без поддержки = %v, %v, %v", con, rw, err)
	}
}

func TestWrWrapHijackSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		wrw := New(wr, false)
		con, rw, err := wrw.Hijack()
		if err != nil {
			_, _ = fmt.Fprintf(rw, "HTTP/1.1 500 Internal Server Error\r\nContent-Length: %d\r\n\r\n%s", len(err.Error()), err.Error())
			_ = rw.Flush()
			_ = con.Close()
			return
		}
		_, _ = fmt.Fprint(rw, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 2\r\n\r\nOK")
		_ = rw.Flush()
		_ = con.Close()
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("http.Get() вернул ошибку: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("hijack response status = %d", resp.StatusCode)
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
	var (
		recorder = httptest.NewRecorder()
		expected = errors.New("hijack error")
		wrw      = New(&hijackRecorder{ResponseRecorder: recorder, err: expected}, false)
	)
	con, rw, err := wrw.Hijack()
	if con != nil || rw != nil || !errors.Is(err, expected) {
		t.Fatalf("Hijack() не вернул значения essence: %v, %v, %v", con, rw, err)
	}
}

func TestWrWrapReadFrom(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrw := New(recorder, false)
	n, err := wrw.ReadFrom(strings.NewReader("read-from"))
	if err != nil || n != int64(len("read-from")) {
		t.Fatalf("ReadFrom() = %d, %v", n, err)
	}
	if recorder.Body.String() != "read-from" || wrw.Len() != uint64(len("read-from")) {
		t.Fatalf("ReadFrom body/len = %q/%d", recorder.Body.String(), wrw.Len())
	}

	rfw := &readerFromWriter{minimalWriter: newMinimalWriter()}
	wrw = New(rfw, false)
	n, err = wrw.ReadFrom(strings.NewReader("fast"))
	if err != nil || n != 4 || !rfw.called {
		t.Fatalf("ReadFrom() passthrough = %d, %v, called=%t", n, err, rfw.called)
	}
	if rfw.body.String() != "fast" || wrw.Len() != 4 {
		t.Fatalf("ReadFrom passthrough body/len = %q/%d", rfw.body.String(), wrw.Len())
	}

	var tee bytes.Buffer
	rfw = &readerFromWriter{minimalWriter: newMinimalWriter()}
	wrw = New(rfw, false).Tee(&tee)
	n, err = wrw.ReadFrom(strings.NewReader("copy"))
	if err != nil || n != 4 || rfw.called {
		t.Fatalf("ReadFrom() с Tee = %d, %v, called=%t", n, err, rfw.called)
	}
	if rfw.body.String() != "copy" || tee.String() != "copy" || wrw.Len() != 4 {
		t.Fatalf("ReadFrom с Tee body/tee/len = %q/%q/%d", rfw.body.String(), tee.String(), wrw.Len())
	}
}

func TestWrWrapResponseControllerUnwrap(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrw := New(recorder, false)
	controller := http.NewResponseController(wrw)
	if err := controller.Flush(); err != nil {
		t.Fatalf("ResponseController.Flush() вернул ошибку: %v", err)
	}
	if !recorder.Flushed {
		t.Fatal("ResponseController через Unwrap не достиг оригинального Flusher")
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
	expected := errors.New("write error")
	wrw := New(&errorWriter{minimalWriter: newMinimalWriter(), err: expected}, false)
	n, err := wrw.Write([]byte("data"))
	if !errors.Is(err, expected) || n != 0 || wrw.Len() != 0 || wrw.StatusCode() != http.StatusOK {
		t.Fatalf("Write() error branch = %d, %v, len=%d, status=%d", n, err, wrw.Len(), wrw.StatusCode())
	}
}

func TestWrWrapWriteStringErrorsAndTee(t *testing.T) {
	expected := errors.New("string write error")
	wrw := New(&stringWriter{minimalWriter: newMinimalWriter(), err: expected}, false)
	n, err := wrw.WriteString("data")
	if !errors.Is(err, expected) || n != 0 || wrw.Len() != 0 {
		t.Fatalf("WriteString() error branch = %d, %v, len=%d", n, err, wrw.Len())
	}

	var tee bytes.Buffer
	wrw = New(&stringWriter{minimalWriter: newMinimalWriter()}, false).Tee(&tee)
	n, err = wrw.WriteString("copy")
	if err != nil || n != 4 || tee.String() != "copy" || wrw.Len() != 4 {
		t.Fatalf("WriteString() tee = %d, %v, tee=%q, len=%d", n, err, tee.String(), wrw.Len())
	}

	wrw = New(&stringWriter{minimalWriter: newMinimalWriter()}, false).Tee(&failTeeWriter{err: expected})
	n, err = wrw.WriteString("data")
	if !errors.Is(err, expected) || n != 4 || wrw.Len() != 4 {
		t.Fatalf("WriteString() tee error = %d, %v, len=%d", n, err, wrw.Len())
	}
}

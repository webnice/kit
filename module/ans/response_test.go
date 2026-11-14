package ans

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/webnice/dic"
)

// failResponseWriter Объект http.ResponseWriter всегда возвращающий ошибку записи.
type failResponseWriter struct {
	header http.Header
	status int
	err    error
}

// Header Реализация интерфейса http.ResponseWriter.
func (fw *failResponseWriter) Header() http.Header {
	if fw.header == nil {
		fw.header = make(http.Header)
	}

	return fw.header
}

// WriteHeader Реализация интерфейса http.ResponseWriter.
func (fw *failResponseWriter) WriteHeader(status int) { fw.status = status }

// Write Реализация интерфейса http.ResponseWriter.
func (fw *failResponseWriter) Write([]byte) (int, error) { return 0, fw.err }

// TestImpl_ResponseStatuses Коды состояния HTTP ответа.
func TestImpl_ResponseStatuses(t *testing.T) {
	const (
		tplStatusCode = "%s() вернул код %d, ожидался %d."
	)
	// Описание одного теста кода состояния.
	type testItem struct {
		name string
		call func(http.ResponseWriter)
		want int
	}
	var (
		ans      Interface
		recorder *httptest.ResponseRecorder
		tests    []testItem
		test     testItem
	)

	ans = New(nil)
	tests = []testItem{
		{name: "Status", call: func(w http.ResponseWriter) { ans.Status(w, dic.Status().Created) }, want: http.StatusCreated},
		{name: "Ok", call: func(w http.ResponseWriter) { ans.Ok(w) }, want: http.StatusOK},
		{name: "NoContent", call: func(w http.ResponseWriter) { ans.NoContent(w) }, want: http.StatusNoContent},
		{name: "Unauthorized", call: func(w http.ResponseWriter) { ans.Unauthorized(w) }, want: http.StatusUnauthorized},
		{name: "Forbidden", call: func(w http.ResponseWriter) { ans.Forbidden(w) }, want: http.StatusForbidden},
		{name: "NotFound", call: func(w http.ResponseWriter) { ans.NotFound(w) }, want: http.StatusNotFound},
	}
	for _, test = range tests {
		recorder = httptest.NewRecorder()
		test.call(recorder)
		if recorder.Code != test.want {
			t.Fatalf(tplStatusCode, test.name, recorder.Code, test.want)
		}
	}
}

// TestImpl_ResponseHeaders Заголовки HTTP ответа.
func TestImpl_ResponseHeaders(t *testing.T) {
	const (
		tplLastModifiedSet = "LastModified() для нулевого времени установил заголовок %q."
		tplContentType     = "ContentType/Header = %q."
		tplContentLength   = "ContentLength() = %q."
		tplRetryAfter      = "RetryAfter() = %q."
		tplLastModified    = "LastModified() = %q."
	)
	var (
		ans      Interface
		recorder *httptest.ResponseRecorder
		moment   time.Time
		got      string
	)

	ans = New(nil)
	recorder = httptest.NewRecorder()
	moment = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	ans.ContentType(recorder, dic.Mime().ApplicationJson)
	ans.ContentTypeString(recorder, "text/plain")
	ans.HeaderString(recorder, dic.Header().ContentType, "application/custom")
	ans.Header(recorder, dic.Header().ContentType, dic.Mime().ApplicationJson)
	ans.ContentLength(recorder, 123)
	ans.RetryAfter(recorder, 3*time.Second)
	ans.LastModified(recorder, time.Time{})
	got = recorder.Header().Get(dic.Header().LastModified.String())
	if got != "" {
		t.Fatalf(tplLastModifiedSet, got)
	}
	ans.LastModified(recorder, moment)
	got = recorder.Header().Get(dic.Header().ContentType.String())
	if got != dic.Mime().ApplicationJson.String() {
		t.Fatalf(tplContentType, got)
	}
	got = recorder.Header().Get(dic.Header().ContentLength.String())
	if got != "123" {
		t.Fatalf(tplContentLength, got)
	}
	got = recorder.Header().Get(dic.Header().RetryAfter.String())
	if got != "3" {
		t.Fatalf(tplRetryAfter, got)
	}
	got = recorder.Header().Get(dic.Header().LastModified.String())
	if got != "Fri, 02 Jan 2026 03:04:05 GMT" {
		t.Fatalf(tplLastModified, got)
	}
}

// TestImpl_ResponseData Передача данных HTTP ответа.
func TestImpl_ResponseData(t *testing.T) {
	const (
		errTextWriteError   = "write error"
		tplResponseBytes    = "ResponseBytes() вернул code=%d body=%q length=%q."
		tplResponseString   = "ResponseString() вернул code=%d body=%q."
		tplResponseNil      = "Response(nil) вернул code=%d body=%q."
		tplResponseEmpty    = "Response(empty) вернул code=%d body=%q."
		tplResponseWriteLog = "Response() с ошибкой записи не залогировал ошибку: %#v."
	)
	var (
		ans      Interface
		logger   *testLogger
		recorder *httptest.ResponseRecorder
		fw       *failResponseWriter
	)

	ans = New(nil)
	recorder = httptest.NewRecorder()
	ans.ResponseBytes(recorder, dic.Status().Created, []byte("data"))
	if recorder.Code != http.StatusCreated || recorder.Body.String() != "data" ||
		recorder.Header().Get(dic.Header().ContentLength.String()) != "4" {
		t.Fatalf(tplResponseBytes, recorder.Code, recorder.Body.String(), recorder.Header().Get(dic.Header().ContentLength.String()))
	}
	recorder = httptest.NewRecorder()
	ans.ResponseString(recorder, dic.Status().Ok, "text")
	if recorder.Code != http.StatusOK || recorder.Body.String() != "text" {
		t.Fatalf(tplResponseString, recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	ans.Response(recorder, dic.Status().Ok, nil)
	if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 {
		t.Fatalf(tplResponseNil, recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	ans.Response(recorder, dic.Status().Ok, &bytes.Buffer{})
	if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 {
		t.Fatalf(tplResponseEmpty, recorder.Code, recorder.Body.String())
	}
	logger = &testLogger{}
	fw = &failResponseWriter{err: errors.New(errTextWriteError)}
	New(logger).Response(fw, dic.Status().Ok, bytes.NewBufferString("x"))
	if len(logger.errors) != 1 || !strings.Contains(logger.errors[0], "передача HTTP ответа") {
		t.Fatalf(tplResponseWriteLog, logger.errors)
	}
}

// TestImpl_BadRequestAndInternalServerError Ответы с ошибками запроса и сервера.
func TestImpl_BadRequestAndInternalServerError(t *testing.T) {
	const (
		errTextWriteError     = "write error"
		errTextBad            = "bad"
		errTextBoom           = "boom"
		tplBadRequestNil      = "BadRequest(nil) вернул code=%d body=%q."
		tplBadRequest         = "BadRequest() вернул code=%d body=%q."
		tplBadRequestBytesNil = "BadRequestBytes(nil) вернул code=%d body=%q."
		tplBadRequestBytes    = "BadRequestBytes() вернул code=%d body=%q."
		tplBadRequestLog      = "BadRequestBytes() с ошибкой записи не залогировал ошибку: %#v."
		tplInternalError      = "InternalServerError() вернул code=%d body=%q."
		tplInternalLogged     = "InternalServerError() залогировал %#v."
		tplInternalNil        = "InternalServerError(nil) вернул code=%d body=%q."
	)
	var (
		logger   *testLogger
		ans      Interface
		recorder *httptest.ResponseRecorder
		fw       *failResponseWriter
	)

	logger = &testLogger{}
	ans = New(logger)
	recorder = httptest.NewRecorder()
	ans.BadRequest(recorder, nil)
	if recorder.Code != http.StatusBadRequest || recorder.Body.Len() != 0 {
		t.Fatalf(tplBadRequestNil, recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	ans.BadRequest(recorder, ans.NewRestError(dic.Status().BadRequest, errors.New(errTextBad)))
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"message":"bad"`) {
		t.Fatalf(tplBadRequest, recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	ans.BadRequestBytes(recorder, nil)
	if recorder.Code != http.StatusBadRequest || recorder.Body.Len() != 0 {
		t.Fatalf(tplBadRequestBytesNil, recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	ans.BadRequestBytes(recorder, []byte("bad"))
	if recorder.Code != http.StatusBadRequest || recorder.Body.String() != "bad" {
		t.Fatalf(tplBadRequestBytes, recorder.Code, recorder.Body.String())
	}
	logger = &testLogger{}
	ans = New(logger)
	fw = &failResponseWriter{err: errors.New(errTextWriteError)}
	ans.BadRequestBytes(fw, []byte("bad"))
	if len(logger.errors) != 1 || !strings.Contains(logger.errors[0], "передача HTTP ответа") {
		t.Fatalf(tplBadRequestLog, logger.errors)
	}
	logger = &testLogger{}
	ans = New(logger)
	recorder = httptest.NewRecorder()
	ans.InternalServerError(recorder, errors.New(errTextBoom))
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"message":"boom"`) {
		t.Fatalf(tplInternalError, recorder.Code, recorder.Body.String())
	}
	if len(logger.errors) != 1 || logger.errors[0] != "boom" {
		t.Fatalf(tplInternalLogged, logger.errors)
	}
	recorder = httptest.NewRecorder()
	ans.InternalServerError(recorder, nil)
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"code":-1`) {
		t.Fatalf(tplInternalNil, recorder.Code, recorder.Body.String())
	}
}

// TestImpl_JsonExtended Сериализация объекта в JSON ответ.
func TestImpl_JsonExtended(t *testing.T) {
	const (
		tplJsonCode           = "Json(%s) вернул код %d."
		tplJsonUnexpected     = "Json(%s) = %q, ожидалось %q."
		tplJsonContentType    = "Json(%s) Content-Type = %q."
		tplJsonUnserializable = "Json() с несериализуемым объектом вернул code=%d body=%q."
		errJsonNoLog          = "Json() с несериализуемым объектом не залогировал ошибку."
	)
	// Описание одного теста сериализации в JSON.
	type testItem struct {
		name string
		obj  any
		want string
	}
	var (
		ans      Interface
		logger   *testLogger
		recorder *httptest.ResponseRecorder
		tests    []testItem
		test     testItem
		got      string
	)

	ans = New(nil)
	tests = []testItem{
		{name: "nil", obj: nil, want: "null"},
		{name: "empty slice", obj: []int{}, want: "[]"},
		{name: "slice", obj: []int{1, 2}, want: "[1,2]"},
		{name: "struct", obj: struct {
			A int `json:"a"`
		}{A: 1}, want: `{"a":1}`},
	}
	for _, test = range tests {
		recorder = httptest.NewRecorder()
		ans.Json(recorder, dic.Status().Ok, test.obj)
		if recorder.Code != http.StatusOK {
			t.Fatalf(tplJsonCode, test.name, recorder.Code)
		}
		got = strings.TrimSpace(recorder.Body.String())
		if got != test.want {
			t.Fatalf(tplJsonUnexpected, test.name, got, test.want)
		}
		got = recorder.Header().Get(dic.Header().ContentType.String())
		if got != dic.Mime().ApplicationJson.String() {
			t.Fatalf(tplJsonContentType, test.name, got)
		}
	}
	logger = &testLogger{}
	recorder = httptest.NewRecorder()
	New(logger).Json(recorder, dic.Status().Ok, func() {})
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"code":-1`) {
		t.Fatalf(tplJsonUnserializable, recorder.Code, recorder.Body.String())
	}
	if len(logger.errors) == 0 {
		t.Fatal(errJsonNoLog)
	}
}

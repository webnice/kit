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

func (fw *failResponseWriter) Header() http.Header {
	if fw.header == nil {
		fw.header = make(http.Header)
	}
	return fw.header
}

func (fw *failResponseWriter) WriteHeader(status int) { fw.status = status }

func (fw *failResponseWriter) Write([]byte) (int, error) { return 0, fw.err }

// TestImpl_ResponseStatuses Коды состояния HTTP ответа.
func TestImpl_ResponseStatuses(t *testing.T) {
	var ans = New(nil)
	var tests = []struct {
		name string
		call func(http.ResponseWriter)
		want int
	}{
		{name: "Status", call: func(w http.ResponseWriter) { ans.Status(w, dic.Status().Created) }, want: http.StatusCreated},
		{name: "Ok", call: func(w http.ResponseWriter) { ans.Ok(w) }, want: http.StatusOK},
		{name: "NoContent", call: func(w http.ResponseWriter) { ans.NoContent(w) }, want: http.StatusNoContent},
		{name: "Unauthorized", call: func(w http.ResponseWriter) { ans.Unauthorized(w) }, want: http.StatusUnauthorized},
		{name: "Forbidden", call: func(w http.ResponseWriter) { ans.Forbidden(w) }, want: http.StatusForbidden},
		{name: "NotFound", call: func(w http.ResponseWriter) { ans.NotFound(w) }, want: http.StatusNotFound},
	}

	for _, test := range tests {
		var recorder = httptest.NewRecorder()
		test.call(recorder)
		if recorder.Code != test.want {
			t.Fatalf("%s() вернул код %d, ожидался %d", test.name, recorder.Code, test.want)
		}
	}
}

// TestImpl_ResponseHeaders Заголовки HTTP ответа.
func TestImpl_ResponseHeaders(t *testing.T) {
	var (
		ans      = New(nil)
		recorder = httptest.NewRecorder()
		moment   = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	)

	ans.ContentType(recorder, dic.Mime().ApplicationJson)
	ans.ContentTypeString(recorder, "text/plain")
	ans.HeaderString(recorder, dic.Header().ContentType, "application/custom")
	ans.Header(recorder, dic.Header().ContentType, dic.Mime().ApplicationJson)
	ans.ContentLength(recorder, 123)
	ans.RetryAfter(recorder, 3*time.Second)
	ans.LastModified(recorder, time.Time{})
	if got := recorder.Header().Get(dic.Header().LastModified.String()); got != "" {
		t.Fatalf("LastModified() для нулевого времени установил заголовок %q", got)
	}
	ans.LastModified(recorder, moment)

	if got := recorder.Header().Get(dic.Header().ContentType.String()); got != dic.Mime().ApplicationJson.String() {
		t.Fatalf("ContentType/Header = %q", got)
	}
	if got := recorder.Header().Get(dic.Header().ContentLength.String()); got != "123" {
		t.Fatalf("ContentLength() = %q", got)
	}
	if got := recorder.Header().Get(dic.Header().RetryAfter.String()); got != "3" {
		t.Fatalf("RetryAfter() = %q", got)
	}
	if got := recorder.Header().Get(dic.Header().LastModified.String()); got != "Fri, 02 Jan 2026 03:04:05 GMT" {
		t.Fatalf("LastModified() = %q", got)
	}
}

// TestImpl_ResponseData Передача данных HTTP ответа.
func TestImpl_ResponseData(t *testing.T) {
	var ans = New(nil)

	var recorder = httptest.NewRecorder()
	ans.ResponseBytes(recorder, dic.Status().Created, []byte("data"))
	if recorder.Code != http.StatusCreated || recorder.Body.String() != "data" ||
		recorder.Header().Get(dic.Header().ContentLength.String()) != "4" {
		t.Fatalf("ResponseBytes() вернул code=%d body=%q length=%q", recorder.Code, recorder.Body.String(), recorder.Header().Get(dic.Header().ContentLength.String()))
	}

	recorder = httptest.NewRecorder()
	ans.ResponseString(recorder, dic.Status().Ok, "text")
	if recorder.Code != http.StatusOK || recorder.Body.String() != "text" {
		t.Fatalf("ResponseString() вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	ans.Response(recorder, dic.Status().Ok, nil)
	if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 {
		t.Fatalf("Response(nil) вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	ans.Response(recorder, dic.Status().Ok, &bytes.Buffer{})
	if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 {
		t.Fatalf("Response(empty) вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}

	var (
		logger = &testLogger{}
		fw     = &failResponseWriter{err: errors.New("write error")}
	)
	New(logger).Response(fw, dic.Status().Ok, bytes.NewBufferString("x"))
	if len(logger.errors) != 1 || !strings.Contains(logger.errors[0], "передача HTTP ответа") {
		t.Fatalf("Response() с ошибкой записи не залогировал ошибку: %#v", logger.errors)
	}
}

// TestImpl_BadRequestAndInternalServerError Ответы с ошибками запроса и сервера.
func TestImpl_BadRequestAndInternalServerError(t *testing.T) {
	var (
		logger = &testLogger{}
		ans    = New(logger)
	)

	recorder := httptest.NewRecorder()
	ans.BadRequest(recorder, nil)
	if recorder.Code != http.StatusBadRequest || recorder.Body.Len() != 0 {
		t.Fatalf("BadRequest(nil) вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	ans.BadRequest(recorder, ans.NewRestError(dic.Status().BadRequest, errors.New("bad")))
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"message":"bad"`) {
		t.Fatalf("BadRequest() вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	ans.BadRequestBytes(recorder, nil)
	if recorder.Code != http.StatusBadRequest || recorder.Body.Len() != 0 {
		t.Fatalf("BadRequestBytes(nil) вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	ans.BadRequestBytes(recorder, []byte("bad"))
	if recorder.Code != http.StatusBadRequest || recorder.Body.String() != "bad" {
		t.Fatalf("BadRequestBytes() вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}

	logger = &testLogger{}
	ans = New(logger)
	var fw = &failResponseWriter{err: errors.New("write error")}
	ans.BadRequestBytes(fw, []byte("bad"))
	if len(logger.errors) != 1 || !strings.Contains(logger.errors[0], "передача HTTP ответа") {
		t.Fatalf("BadRequestBytes() с ошибкой записи не залогировал ошибку: %#v", logger.errors)
	}

	logger = &testLogger{}
	ans = New(logger)
	recorder = httptest.NewRecorder()
	ans.InternalServerError(recorder, errors.New("boom"))
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"message":"boom"`) {
		t.Fatalf("InternalServerError() вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}
	if len(logger.errors) != 1 || logger.errors[0] != "boom" {
		t.Fatalf("InternalServerError() залогировал %#v", logger.errors)
	}

	recorder = httptest.NewRecorder()
	ans.InternalServerError(recorder, nil)
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"code":-1`) {
		t.Fatalf("InternalServerError(nil) вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}
}

// TestImpl_JsonExtended Сериализация объекта в JSON ответ.
func TestImpl_JsonExtended(t *testing.T) {
	var ans = New(nil)

	var tests = []struct {
		name string
		obj  any
		want string
	}{
		{name: "nil", obj: nil, want: "null"},
		{name: "empty slice", obj: []int{}, want: "[]"},
		{name: "slice", obj: []int{1, 2}, want: "[1,2]"},
		{name: "struct", obj: struct {
			A int `json:"a"`
		}{A: 1}, want: `{"a":1}`},
	}
	for _, test := range tests {
		var recorder = httptest.NewRecorder()
		ans.Json(recorder, dic.Status().Ok, test.obj)
		if recorder.Code != http.StatusOK {
			t.Fatalf("Json(%s) вернул код %d", test.name, recorder.Code)
		}
		if got := strings.TrimSpace(recorder.Body.String()); got != test.want {
			t.Fatalf("Json(%s) = %q, ожидалось %q", test.name, got, test.want)
		}
		if got := recorder.Header().Get(dic.Header().ContentType.String()); got != dic.Mime().ApplicationJson.String() {
			t.Fatalf("Json(%s) Content-Type = %q", test.name, got)
		}
	}

	var (
		logger   = &testLogger{}
		recorder = httptest.NewRecorder()
	)
	New(logger).Json(recorder, dic.Status().Ok, func() {})
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"code":-1`) {
		t.Fatalf("Json() с несериализуемым объектом вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}
	if len(logger.errors) == 0 {
		t.Fatal("Json() с несериализуемым объектом не залогировал ошибку")
	}
}

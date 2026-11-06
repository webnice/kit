package ans

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/webnice/dic"
	kitTypes "github.com/webnice/kit/v4/types"

	"github.com/go-chi/chi/v5"
)

type testLogger struct {
	kitTypes.Logger
	warnings []string
	errors   []string
}

func (tl *testLogger) Warningf(pattern string, args ...any) {
	tl.warnings = append(tl.warnings, fmt.Sprintf(pattern, args...))
}

func (tl *testLogger) Errorf(pattern string, args ...any) {
	tl.errors = append(tl.errors, fmt.Sprintf(pattern, args...))
}

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

type testBody struct {
	reader   io.Reader
	closeErr error
}

func (tb *testBody) Read(p []byte) (int, error) { return tb.reader.Read(p) }

func (tb *testBody) Close() error { return tb.closeErr }

type errorReader struct{ err error }

func (er *errorReader) Read([]byte) (int, error) { return 0, er.err }

func newRequestWithBody(body string, contentType string) *http.Request {
	var req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(dic.Header().ContentType.String(), contentType)

	return req
}

func newChiRequest(params map[string]string) *http.Request {
	var (
		req  = httptest.NewRequest(http.MethodGet, "/", nil)
		rctx = chi.NewRouteContext()
	)
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestImpl_RqBytesExtended(t *testing.T) {
	var ans = New(nil)

	if _, err := ans.RqBytes(nil); err == nil || !strings.Contains(err.Error(), "*http.Request") {
		t.Fatalf("RqBytes(nil) вернул ошибку %v", err)
	}
	if _, err := ans.RqBytes(&http.Request{}); err == nil || !strings.Contains(err.Error(), "тела HTTP запроса") {
		t.Fatalf("RqBytes(&http.Request{}) вернул ошибку %v", err)
	}

	buf, err := ans.RqBytes(newRequestWithBody("hello", dic.Mime().ApplicationJson.String()))
	if err != nil {
		t.Fatalf("RqBytes() вернул ошибку: %v", err)
	}
	if buf.String() != "hello" {
		t.Fatalf("RqBytes() = %q, ожидалось hello", buf.String())
	}

	var (
		logger = &testLogger{}
		ansLog = New(logger)
	)
	buf, err = ansLog.RqBytes(&http.Request{
		Body: &testBody{reader: strings.NewReader("data"), closeErr: errors.New("close error")},
	})
	if err != nil {
		t.Fatalf("RqBytes() с ошибкой Close вернул ошибку: %v", err)
	}
	if buf.String() != "data" {
		t.Fatalf("RqBytes() с ошибкой Close = %q, ожидалось data", buf.String())
	}
	if len(logger.warnings) != 1 || !strings.Contains(logger.warnings[0], "Закрытие тела HTTP запроса") {
		t.Fatalf("logWarningf() не зафиксировал ошибку Close: %#v", logger.warnings)
	}

	_, _ = ans.RqBytes(&http.Request{
		Body: &testBody{reader: strings.NewReader("x"), closeErr: errors.New("close error")},
	})

	_, err = ans.RqBytes(&http.Request{
		Body: &testBody{reader: io.MultiReader(strings.NewReader("12"), &errorReader{err: errors.New("read error")})},
	})
	if err == nil || !strings.Contains(err.Error(), "чтение тела HTTP запроса") {
		t.Fatalf("RqBytes() с ошибкой чтения вернул ошибку %v", err)
	}

	var mismatch = newRequestWithBody("123", dic.Mime().ApplicationJson.String())
	mismatch.ContentLength = 10
	if _, err = ans.RqBytes(mismatch); err == nil || !strings.Contains(err.Error(), "размер тела HTTP запроса") {
		t.Fatalf("RqBytes() с неверной длиной вернул ошибку %v", err)
	}
}

func TestImpl_RqDataExtended(t *testing.T) {
	var ans = New(nil)

	var jsonVariable struct {
		A int `json:"a"`
	}
	if err := ans.RqData(nil, &jsonVariable); err == nil {
		t.Fatal("RqData(nil) вернул nil ошибку")
	}

	if err := ans.RqData(newRequestWithBody(`{"a":1}`, dic.Mime().ApplicationJson.String()), &jsonVariable); err != nil {
		t.Fatalf("RqData(JSON) вернул ошибку: %v", err)
	}
	if jsonVariable.A != 1 {
		t.Fatalf("RqData(JSON) декодировал %+v", jsonVariable)
	}

	var xmlVariable struct {
		Name string `xml:"name"`
	}
	if err := ans.RqData(newRequestWithBody(`<data><name>test</name></data>`, dic.Mime().TextXml.String()), &xmlVariable); err != nil {
		t.Fatalf("RqData(XML) вернул ошибку: %v", err)
	}
	if xmlVariable.Name != "test" {
		t.Fatalf("RqData(XML) декодировал %+v", xmlVariable)
	}

	if err := ans.RqData(newRequestWithBody(`{"a":`, dic.Mime().ApplicationJson.String()), &jsonVariable); err == nil ||
		!strings.Contains(err.Error(), "ошибка декодирования") {
		t.Fatalf("RqData() с невалидным JSON вернул ошибку %v", err)
	}
	if err := ans.RqData(newRequestWithBody(`<data>`, dic.Mime().TextXml.String()), &xmlVariable); err == nil ||
		!strings.Contains(err.Error(), "ошибка декодирования") {
		t.Fatalf("RqData() с невалидным XML вернул ошибку %v", err)
	}
	if err := ans.RqData(newRequestWithBody(`{}`, "application/octet-stream"), &jsonVariable); err == nil ||
		!strings.Contains(err.Error(), "Content-Type") {
		t.Fatalf("RqData() с неизвестным Content-Type вернул ошибку %v", err)
	}
}

func TestImpl_RqIdsAndRqId(t *testing.T) {
	var ans = New(nil)

	ids, err := ans.RqIds(httptest.NewRecorder(), newChiRequest(map[string]string{"ids": "1, 2,, 3"}), "ids")
	if err != nil {
		t.Fatalf("RqIds() вернул ошибку: %v", err)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Fatalf("RqIds() = %v, ожидалось [1 2 3]", ids)
	}

	var recorder = httptest.NewRecorder()
	_, err = ans.RqIds(recorder, newChiRequest(map[string]string{"ids": "1,x"}), "ids")
	if err == nil {
		t.Fatal("RqIds() с невалидным значением вернул nil ошибку")
	}
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"field":"ids"`) {
		t.Fatalf("RqIds() сформировал неверный ответ: code=%d body=%s", recorder.Code, recorder.Body.String())
	}

	id, err := ans.RqId(httptest.NewRecorder(), newChiRequest(map[string]string{"id": "42"}), "id")
	if err != nil || id != 42 {
		t.Fatalf("RqId() = %d, %v", id, err)
	}
	id, err = ans.RqId(httptest.NewRecorder(), newChiRequest(map[string]string{}), "id")
	if err != nil || id != 0 {
		t.Fatalf("RqId() для пустого параметра = %d, %v", id, err)
	}
	recorder = httptest.NewRecorder()
	_, err = ans.RqId(recorder, newChiRequest(map[string]string{"id": "x"}), "id")
	if err == nil || recorder.Code != http.StatusBadRequest {
		t.Fatalf("RqId() с невалидным значением вернул ошибку %v и код %d", err, recorder.Code)
	}
}

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

func TestImpl_RestErrorExtended(t *testing.T) {
	var (
		ans = New(nil)
		reo = ans.NewRestError(dic.Status().BadRequest, nil)
	)

	if reo.CodeGet() != -1 || reo.MessageGet() != "" || reo.I18nKeyGet() != "" || len(reo.Field()) != 0 {
		t.Fatalf("NewRestError() вернул некорректное начальное состояние: %+v", reo)
	}
	reo.CodeSet(7).MessageSet("message").I18nKeySet("key")
	if reo.CodeGet() != 7 || reo.MessageGet() != "message" || reo.I18nKeyGet() != "key" {
		t.Fatalf("Set/Get методы RestError вернули некорректное состояние: %+v", reo)
	}
	reo.Add("field", "value", "message").AddWithKey("field2", "value2", "message2", "key2")
	if fields := reo.Field(); len(fields) != 2 || fields[0].Field != "field" || fields[1].I18nKey != "key2" {
		t.Fatalf("Add/AddWithKey/Field вернули %+v", fields)
	}

	buf, err := reo.JsonBytes()
	if err != nil {
		t.Fatalf("JsonBytes() вернул ошибку: %v", err)
	}
	if !strings.Contains(buf.String(), `"field2"`) {
		t.Fatalf("JsonBytes() вернул %q", buf.String())
	}

	var recorder = httptest.NewRecorder()
	reo.Json(recorder)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"message2"`) {
		t.Fatalf("Json() вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}

	if err = ans.NewRestError(dic.Status().BadRequest, errors.New("as error")).AsError(); err == nil || err.Error() != "as error" {
		t.Fatalf("AsError() по message вернул %v", err)
	}
	err = ans.NewRestError(dic.Status().BadRequest, nil).CodeSet(0).Add("field", "value", "message").AsError()
	if err == nil || err.Error() != "ошибки значений в полях структуры данных" {
		t.Fatalf("AsError() по fields вернул %v", err)
	}
	err = ans.NewRestError(dic.Status().BadRequest, nil).CodeSet(7).AsError()
	if err == nil || err.Error() != "ошибка с кодом 7" {
		t.Fatalf("AsError() по code вернул %v", err)
	}
	if err = ans.NewRestError(dic.Status().BadRequest, nil).CodeSet(0).AsError(); err != nil {
		t.Fatalf("AsError() для пустой ошибки вернул %v", err)
	}
}

func TestImpl_VerifyExtended(t *testing.T) {
	var ans = New(nil)

	type valid struct {
		Name string `validate:"required"`
	}
	if reo := ans.(*impl).Verify(&valid{Name: "ok"}); reo != nil {
		t.Fatalf("Verify() для валидного объекта вернул %+v", reo)
	}

	type invalid struct {
		Name string `validate:"required"`
	}
	if reo := ans.(*impl).Verify(&invalid{}); reo == nil || len(reo.Field()) != 1 || reo.Field()[0].Field != "Name" {
		t.Fatalf("Verify() для невалидного объекта вернул %+v", reo)
	}

	var slice = []invalid{{Name: "ok"}, {}}
	if reo := ans.(*impl).Verify(&slice); reo == nil || len(reo.Field()) != 1 {
		t.Fatalf("Verify() для невалидного среза вернул %+v", reo)
	}

	if reo := ans.(*impl).Verify(1); reo == nil || reo.AsError() == nil {
		t.Fatal("Verify() для неструктурного значения не вернул ошибку")
	}

	type invalidTag struct {
		Name string `validate:"unknown-tag"`
	}
	if reo := ans.(*impl).Verify(&invalidTag{Name: "ok"}); reo == nil || reo.AsError() == nil {
		t.Fatal("Verify() с невалидным тегом валидации не вернул ошибку")
	}

	if reo := ans.(*impl).Verify(nil); reo == nil || reo.AsError() == nil {
		t.Fatal("Verify(nil) не вернул ошибку")
	}
}

func TestImpl_RqVar(t *testing.T) {
	if New(nil).RqVar() == nil {
		t.Fatal("RqVar() вернул nil")
	}
}

func TestImpl_RestErrorJsonEncodeError(t *testing.T) {
	var (
		ans = New(nil)
		reo = ans.NewRestError(dic.Status().BadRequest, errors.New("bad"))
		old = jsonEncode
	)
	jsonEncode = func(wr io.Writer, obj any) error {
		if obj == reo {
			return errors.New("encode error")
		}
		return old(wr, obj)
	}
	t.Cleanup(func() { jsonEncode = old })

	if _, err := reo.JsonBytes(); err == nil || !strings.Contains(err.Error(), "сериализация в JSON") {
		t.Fatalf("JsonBytes() с ошибкой кодирования вернул %v", err)
	}

	var recorder = httptest.NewRecorder()
	reo.Json(recorder)
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"code":-1`) {
		t.Fatalf("Json() с ошибкой кодирования вернул code=%d body=%q", recorder.Code, recorder.Body.String())
	}
}

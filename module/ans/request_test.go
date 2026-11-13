package ans

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/webnice/dic"
	kitTypes "github.com/webnice/kit/v4/types"
)

// testLogger Логгер фиксирующий сообщения для проверки.
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

// testBody Тело HTTP запроса с управляемой ошибкой закрытия.
type testBody struct {
	reader   io.Reader
	closeErr error
}

func (tb *testBody) Read(p []byte) (int, error) { return tb.reader.Read(p) }

func (tb *testBody) Close() error { return tb.closeErr }

// errorReader Объект io.Reader всегда возвращающий ошибку.
type errorReader struct{ err error }

func (er *errorReader) Read([]byte) (int, error) { return 0, er.err }

// newRequestWithBody Формирование тестового запроса с телом и типом содержимого.
func newRequestWithBody(body string, contentType string) *http.Request {
	var req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(dic.Header().ContentType.String(), contentType)

	return req
}

// TestImpl_RqBytesExtended Чтение тела HTTP запроса.
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

// TestImpl_RqDataExtended Декодирование тела HTTP запроса в переменную.
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

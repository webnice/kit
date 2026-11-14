package ans

import (
	"bytes"
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

// Warningf Реализация интерфейса журналирования.
func (tl *testLogger) Warningf(pattern string, args ...any) {
	tl.warnings = append(tl.warnings, fmt.Sprintf(pattern, args...))
}

// Errorf Реализация интерфейса журналирования.
func (tl *testLogger) Errorf(pattern string, args ...any) {
	tl.errors = append(tl.errors, fmt.Sprintf(pattern, args...))
}

// testBody Тело HTTP запроса с управляемой ошибкой закрытия.
type testBody struct {
	reader   io.Reader
	closeErr error
}

// Read Реализация интерфейса io.Reader.
func (tb *testBody) Read(p []byte) (int, error) { return tb.reader.Read(p) }

// Close Реализация интерфейса io.Closer.
func (tb *testBody) Close() error { return tb.closeErr }

// errorReader Объект io.Reader всегда возвращающий ошибку.
type errorReader struct{ err error }

// Read Реализация интерфейса io.Reader.
func (er *errorReader) Read([]byte) (int, error) { return 0, er.err }

// Формирование тестового запроса с телом и типом содержимого.
func newRequestWithBody(body string, contentType string) (req *http.Request) {
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(dic.Header().ContentType.String(), contentType)

	return
}

// TestImpl_RqBytesExtended Чтение тела HTTP запроса.
func TestImpl_RqBytesExtended(t *testing.T) {
	const (
		errTextCloseError           = "close error"
		errTextReadError            = "read error"
		tplRqBytesNilError          = "RqBytes(nil) вернул ошибку %v."
		tplRqBytesEmptyRequestError = "RqBytes(&http.Request{}) вернул ошибку %v."
		tplRqBytesUnexpectedError   = "RqBytes() вернул ошибку: %v."
		tplRqBytesUnexpected        = "RqBytes() = %q, ожидалось hello."
		tplRqBytesCloseError        = "RqBytes() с ошибкой Close вернул ошибку: %v."
		tplRqBytesCloseUnexpected   = "RqBytes() с ошибкой Close = %q, ожидалось data."
		tplLogWarningNotRecorded    = "logWarningf() не зафиксировал ошибку Close: %#v."
		tplRqBytesReadError         = "RqBytes() с ошибкой чтения вернул ошибку %v."
		tplRqBytesLengthError       = "RqBytes() с неверной длиной вернул ошибку %v."
	)
	var (
		ans      Interface
		ansLog   Interface
		logger   *testLogger
		buf      *bytes.Buffer
		mismatch *http.Request
		err      error
	)

	ans = New(nil)
	if _, err = ans.RqBytes(nil); err == nil || !strings.Contains(err.Error(), "*http.Request") {
		t.Fatalf(tplRqBytesNilError, err)
	}
	if _, err = ans.RqBytes(&http.Request{}); err == nil || !strings.Contains(err.Error(), "тела HTTP запроса") {
		t.Fatalf(tplRqBytesEmptyRequestError, err)
	}
	buf, err = ans.RqBytes(newRequestWithBody("hello", dic.Mime().ApplicationJson.String()))
	if err != nil {
		t.Fatalf(tplRqBytesUnexpectedError, err)
	}
	if buf.String() != "hello" {
		t.Fatalf(tplRqBytesUnexpected, buf.String())
	}
	logger = &testLogger{}
	ansLog = New(logger)
	buf, err = ansLog.RqBytes(&http.Request{
		Body: &testBody{reader: strings.NewReader("data"), closeErr: errors.New(errTextCloseError)},
	})
	if err != nil {
		t.Fatalf(tplRqBytesCloseError, err)
	}
	if buf.String() != "data" {
		t.Fatalf(tplRqBytesCloseUnexpected, buf.String())
	}
	if len(logger.warnings) != 1 || !strings.Contains(logger.warnings[0], "Закрытие тела HTTP запроса") {
		t.Fatalf(tplLogWarningNotRecorded, logger.warnings)
	}
	_, _ = ans.RqBytes(&http.Request{
		Body: &testBody{reader: strings.NewReader("x"), closeErr: errors.New(errTextCloseError)},
	})
	_, err = ans.RqBytes(&http.Request{
		Body: &testBody{reader: io.MultiReader(strings.NewReader("12"), &errorReader{err: errors.New(errTextReadError)})},
	})
	if err == nil || !strings.Contains(err.Error(), "чтение тела HTTP запроса") {
		t.Fatalf(tplRqBytesReadError, err)
	}
	mismatch = newRequestWithBody("123", dic.Mime().ApplicationJson.String())
	mismatch.ContentLength = 10
	if _, err = ans.RqBytes(mismatch); err == nil || !strings.Contains(err.Error(), "размер тела HTTP запроса") {
		t.Fatalf(tplRqBytesLengthError, err)
	}
}

// TestImpl_RqDataExtended Декодирование тела HTTP запроса в переменную.
func TestImpl_RqDataExtended(t *testing.T) {
	const (
		errRqDataNilNoError  = "RqData(nil) вернул nil ошибку."
		tplRqDataJSONError   = "RqData(JSON) вернул ошибку: %v."
		tplRqDataJSONDecoded = "RqData(JSON) декодировал %+v."
		tplRqDataXMLError    = "RqData(XML) вернул ошибку: %v."
		tplRqDataXMLDecoded  = "RqData(XML) декодировал %+v."
		tplRqDataBadJSON     = "RqData() с невалидным JSON вернул ошибку %v."
		tplRqDataBadXML      = "RqData() с невалидным XML вернул ошибку %v."
		tplRqDataUnknownType = "RqData() с неизвестным Content-Type вернул ошибку %v."
	)
	var (
		ans          Interface
		jsonVariable struct {
			A int `json:"a"`
		}
		xmlVariable struct {
			Name string `xml:"name"`
		}
		err error
	)

	ans = New(nil)
	if err = ans.RqData(nil, &jsonVariable); err == nil {
		t.Fatal(errRqDataNilNoError)
	}
	if err = ans.RqData(newRequestWithBody(`{"a":1}`, dic.Mime().ApplicationJson.String()), &jsonVariable); err != nil {
		t.Fatalf(tplRqDataJSONError, err)
	}
	if jsonVariable.A != 1 {
		t.Fatalf(tplRqDataJSONDecoded, jsonVariable)
	}
	if err = ans.RqData(newRequestWithBody(`<data><name>test</name></data>`, dic.Mime().TextXml.String()), &xmlVariable); err != nil {
		t.Fatalf(tplRqDataXMLError, err)
	}
	if xmlVariable.Name != "test" {
		t.Fatalf(tplRqDataXMLDecoded, xmlVariable)
	}
	if err = ans.RqData(newRequestWithBody(`{"a":`, dic.Mime().ApplicationJson.String()), &jsonVariable); err == nil ||
		!strings.Contains(err.Error(), "ошибка декодирования") {
		t.Fatalf(tplRqDataBadJSON, err)
	}
	if err = ans.RqData(newRequestWithBody(`<data>`, dic.Mime().TextXml.String()), &xmlVariable); err == nil ||
		!strings.Contains(err.Error(), "ошибка декодирования") {
		t.Fatalf(tplRqDataBadXML, err)
	}
	if err = ans.RqData(newRequestWithBody(`{}`, "application/octet-stream"), &jsonVariable); err == nil ||
		!strings.Contains(err.Error(), "Content-Type") {
		t.Fatalf(tplRqDataUnknownType, err)
	}
}

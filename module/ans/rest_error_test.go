package ans

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/webnice/dic"
)

// TestImpl_RestErrorExtended Объект REST ошибки.
func TestImpl_RestErrorExtended(t *testing.T) {
	const (
		errTextAsError      = "as error"
		tplInitialState     = "NewRestError() вернул некорректное начальное состояние: %+v."
		tplSetGetState      = "Set/Get методы RestError вернули некорректное состояние: %+v."
		tplFieldsState      = "Add/AddWithKey/Field вернули %+v."
		tplJsonBytesError   = "JsonBytes() вернул ошибку: %v."
		tplJsonBytesResult  = "JsonBytes() вернул %q."
		tplJsonResult       = "Json() вернул code=%d body=%q."
		tplAsErrorByMessage = "AsError() по message вернул %v."
		tplAsErrorByFields  = "AsError() по fields вернул %v."
		tplAsErrorByCode    = "AsError() по code вернул %v."
		tplAsErrorEmpty     = "AsError() для пустой ошибки вернул %v."
	)
	var (
		ans      Interface
		reo      RestErrorInterface
		buf      *bytes.Buffer
		recorder *httptest.ResponseRecorder
		fields   []RestErrorField
		err      error
	)

	ans = New(nil)
	reo = ans.NewRestError(dic.Status().BadRequest, nil)
	if reo.CodeGet() != -1 || reo.MessageGet() != "" || reo.I18nKeyGet() != "" || len(reo.Field()) != 0 {
		t.Fatalf(tplInitialState, reo)
	}
	reo.CodeSet(7).MessageSet("message").I18nKeySet("key")
	if reo.CodeGet() != 7 || reo.MessageGet() != "message" || reo.I18nKeyGet() != "key" {
		t.Fatalf(tplSetGetState, reo)
	}
	reo.Add("field", "value", "message").AddWithKey("field2", "value2", "message2", "key2")
	fields = reo.Field()
	if len(fields) != 2 || fields[0].Field != "field" || fields[1].I18nKey != "key2" {
		t.Fatalf(tplFieldsState, fields)
	}
	buf, err = reo.JsonBytes()
	if err != nil {
		t.Fatalf(tplJsonBytesError, err)
	}
	if !strings.Contains(buf.String(), `"field2"`) {
		t.Fatalf(tplJsonBytesResult, buf.String())
	}
	recorder = httptest.NewRecorder()
	reo.Json(recorder)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"message2"`) {
		t.Fatalf(tplJsonResult, recorder.Code, recorder.Body.String())
	}
	err = ans.NewRestError(dic.Status().BadRequest, errors.New(errTextAsError)).AsError()
	if err == nil || err.Error() != errTextAsError {
		t.Fatalf(tplAsErrorByMessage, err)
	}
	err = ans.NewRestError(dic.Status().BadRequest, nil).CodeSet(0).Add("field", "value", "message").AsError()
	if err == nil || err.Error() != "ошибки значений в полях структуры данных" {
		t.Fatalf(tplAsErrorByFields, err)
	}
	err = ans.NewRestError(dic.Status().BadRequest, nil).CodeSet(7).AsError()
	if err == nil || err.Error() != "ошибка с кодом 7" {
		t.Fatalf(tplAsErrorByCode, err)
	}
	if err = ans.NewRestError(dic.Status().BadRequest, nil).CodeSet(0).AsError(); err != nil {
		t.Fatalf(tplAsErrorEmpty, err)
	}
}

// TestImpl_RestErrorJsonEncodeError Ошибка сериализации REST ошибки в JSON.
func TestImpl_RestErrorJsonEncodeError(t *testing.T) {
	const (
		errTextBad         = "bad"
		errTextEncodeError = "encode error"
		tplJsonBytesError  = "JsonBytes() с ошибкой кодирования вернул %v."
		tplJsonResult      = "Json() с ошибкой кодирования вернул code=%d body=%q."
	)
	var (
		ans      Interface
		reo      RestErrorInterface
		old      func(wr io.Writer, obj any) error
		recorder *httptest.ResponseRecorder
		err      error
	)

	ans = New(nil)
	reo = ans.NewRestError(dic.Status().BadRequest, errors.New(errTextBad))
	old = jsonEncode
	jsonEncode = func(wr io.Writer, obj any) error {
		if obj == reo {
			return errors.New(errTextEncodeError)
		}

		return old(wr, obj)
	}
	t.Cleanup(func() { jsonEncode = old })
	if _, err = reo.JsonBytes(); err == nil || !strings.Contains(err.Error(), "сериализация в JSON") {
		t.Fatalf(tplJsonBytesError, err)
	}
	recorder = httptest.NewRecorder()
	reo.Json(recorder)
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"code":-1`) {
		t.Fatalf(tplJsonResult, recorder.Code, recorder.Body.String())
	}
}

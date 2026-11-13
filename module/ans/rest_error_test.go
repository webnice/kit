package ans

import (
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

// TestImpl_RestErrorJsonEncodeError Ошибка сериализации REST ошибки в JSON.
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

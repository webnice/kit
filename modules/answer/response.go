package answer // import "gopkg.in/webnice/kit.v1/modules/answer"

//import "gopkg.in/webnice/debug.v1"
import "gopkg.in/webnice/log.v2"
import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"

	"gopkg.in/webnice/kit.v1/modules/verify"
	"gopkg.in/webnice/web.v1/header"
	"gopkg.in/webnice/web.v1/mime"
	"gopkg.in/webnice/web.v1/status"

	"github.com/wI2L/jettison"
)

// Response Ответ на запрос с проверкой передачи данных
func Response(wr http.ResponseWriter, statusCode int, buf []byte) {
	var err error

	wr.WriteHeader(statusCode)
	if len(buf) > 0 {
		if _, err = wr.Write(buf); err != nil {
			log.Errorf("http response error: %s", err)
		}
	}
}

// ResponseBadRequest Ответ на запрос с передачей ошибки запроса и структуры описывающей найденную ошибку
func ResponseBadRequest(wr http.ResponseWriter, statusCode int, data verify.Interface) {
	var (
		err error
		buf []byte
	)

	if buf = data.Json(); len(buf) == 0 {
		// TODO: Добавить проверку ошибки кодирования JSON в библиотеке gopkg.in/webnice/kit.v1/modules/verify
		err = fmt.Errorf("gopkg.in/webnice/kit.v1/modules/verify JSON encode error")
		InternalServerError(wr, err)
		return
	}
	wr.Header().Set(header.ContentType, mime.ApplicationJSONCharsetUTF8)
	Response(wr, statusCode, buf)
}

// InternalServerError Ответ на запрос с кодом ошибки 500 и структурой описывающей ошибку
func InternalServerError(wr http.ResponseWriter, err error) {
	var data verify.Interface

	log.Error(err.Error())
	data = verify.E5xx().Code(-1).Message(err.Error())
	ResponseBadRequest(wr, status.InternalServerError, data)
}

// JSON Ответ на запрос с передачей результата в JSON
func JSON(wr http.ResponseWriter, statusCode int, obj interface{}) { // nolint: unparam
	const sliceEmpty = `[]`
	var (
		err     error
		buf     *bytes.Buffer
		rvo     reflect.Value
		length  int
		isSlice bool
		enc     []byte
	)

	// Для среза получаем длинну
	switch reflect.TypeOf(obj).Kind() { // nolint: gocritic
	case reflect.Slice:
		isSlice, rvo = true, reflect.ValueOf(obj)
		length = rvo.Len()
	}
	// Если срез пустой, отвечаем константой
	if isSlice && length == 0 {
		buf = bytes.NewBufferString(sliceEmpty)
	} else {
		if enc, err = jettison.Marshal(obj); err != nil {
			err = fmt.Errorf("json encode error: %s", err)
			InternalServerError(wr, err)
			return
		}
		buf = bytes.NewBuffer(enc)
	}
	wr.Header().Set(header.ContentType, mime.ApplicationJSONCharsetUTF8)
	Response(wr, statusCode, buf.Bytes())
}

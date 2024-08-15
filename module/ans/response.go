package ans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/webnice/dic"
)

// NoContent Ответ кодом 204 "No Content" без передачи тела сообщения.
func (ans *impl) NoContent(wr http.ResponseWriter) Interface {
	wr.WriteHeader(dic.Status().NoContent.Code())
	return ans
}

// BadRequest Ответ на запрос с передачей ошибки запроса и структуры описывающей найденную ошибку
func (ans *impl) BadRequest(wr http.ResponseWriter, status dic.IStatus, data any) {

	// data verify.Interface

	//var (
	//	err error
	//	buf []byte
	//)
	//
	//if buf = data.Json(); len(buf) == 0 {
	//	// TODO: Добавить проверку ошибки кодирования JSON в библиотеке github.com/webnice/kit/v2/module/verify
	//	err = fmt.Errorf("github.com/webnice/kit/v2/module/verify JSON encode error")
	//	InternalServerError(wr, err)
	//	return
	//}
	//wr.Header().Set(header.ContentType, mime.ApplicationJSONCharsetUTF8)
	//Response(wr, statusCode, buf)
}

// InternalServerError Ответ на запрос с кодом ошибки 500 и структурой описывающей ошибку.
func (ans *impl) InternalServerError(wr http.ResponseWriter, err error) Interface {
	//var data verify.Interface
	//
	//log.Error(err.Error())
	//data = verify.E5xx().Code(-1).Message(err.Error())
	//ResponseBadRequest(wr, status.InternalServerError, data)

	return ans
}

// ContentType Установка типа контента передаваемых данных.
func (ans *impl) ContentType(wr http.ResponseWriter, mime dic.IMime) Interface {
	wr.Header().Set(dic.Header().ContentType.String(), mime.String())
	return ans
}

// ResponseBytes Ответ с проверкой передачи данных.
func (ans *impl) ResponseBytes(wr http.ResponseWriter, status dic.IStatus, data []byte) Interface {
	return ans.Response(wr, status, bytes.NewBuffer(data))
}

// Response Ответ с проверкой передачи данных.
func (ans *impl) Response(wr http.ResponseWriter, status dic.IStatus, buf *bytes.Buffer) Interface {
	var err error

	if buf == nil {
		wr.WriteHeader(status.Code())
		return ans
	}
	if buf.Len() > 0 {
		wr.Header().Set(dic.Header().ContentLength.String(), strconv.FormatInt(int64(buf.Len()), 10))
	}
	wr.WriteHeader(status.Code())
	if buf.Len() > 0 {
		if _, err = buf.WriteTo(wr); err != nil {
			ans.logErrorf("передача HTTP ответа прервана ошибкой: %s", err)
			return ans
		}
	}

	return ans
}

// Json Ответ на запрос с сериализацией результата в JSON формат.
func (ans *impl) Json(wr http.ResponseWriter, status dic.IStatus, obj any) Interface {
	const sliceEmpty = "[]"
	var (
		err     error
		buf     *bytes.Buffer
		enc     *json.Encoder
		rvo     reflect.Value
		length  int
		isSlice bool
	)

	// Для среза получаем длину.
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		isSlice, rvo = true, reflect.ValueOf(obj)
		length = rvo.Len()
	default:
		length = 0
	}
	// Если срез пустой, отвечаем константой.
	switch {
	case isSlice && length == 0:
		buf = bytes.NewBufferString(sliceEmpty)
	default:
		buf = &bytes.Buffer{}
		enc = json.NewEncoder(buf)
		if err = enc.Encode(obj); err != nil {
			err = fmt.Errorf(errEncode, err)
			ans.InternalServerError(wr, err)
			return ans
		}
	}
	ans.ContentType(wr, dic.Mime().ApplicationJson)
	ans.Response(wr, status, buf)

	return ans
}

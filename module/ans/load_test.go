package ans

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"github.com/webnice/dic"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestImpl_RqLoadVerify(t *testing.T) {
	var (
		err    error
		server *httptest.Server
	)

	tests := map[string]struct {
		err              error             // Ошибка загрузчика.
		route            *chi.Mux          // Роутер ВЕБ сервера.
		variable         any               // Объект загружаемой структуры.
		urn              string            // URN роутинга сервера.
		urnHandleFunc    http.HandlerFunc  // Функция-обработчик URN.
		serverUri        *url.URL          // URI запроса к ВЕБ серверу.
		requestMethod    string            // Метод запроса.
		requestBody      *bytes.Buffer     // Тело запроса.
		request          *http.Request     // Объект запроса.
		requestHeaders   map[string]string // Заголовки запроса.
		response         *http.Response    // Объект ответа.
		responseData     []byte            // Тело ответа.
		responseExpected string            // Ожидаемое тело ответа.
	}{
		"Тест 01": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &struct {
				Uint  uint64 `json:"uint"  validate:"required"`
				Email string `json:"email" validate:"required,email"`
				Age   uint16 `json:"age"   validate:"gte=21,lte=130"`
			}{},
			urn:              "/",
			requestMethod:    dic.Method().Post.String(),
			requestBody:      bytes.NewBufferString(``),
			requestHeaders:   map[string]string{},
			serverUri:        nil,
			request:          nil,
			response:         nil,
			responseData:     nil,
			responseExpected: `{"error":{"code":400,"message":"тело запроса пустое","i18nKey":"","errors":null}}`,
		},
		"Тест 02": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &struct {
				Uint  uint64 `json:"uint"  validate:"required"`
				Email string `json:"email" validate:"required,email"`
				Age   uint16 `json:"age"   validate:"gte=21,lte=130"`
			}{},
			urn:            "/",
			requestMethod:  dic.Method().Post.String(),
			requestBody:    bytes.NewBufferString(`{}`),
			requestHeaders: map[string]string{},
			serverUri:      nil,
			request:        nil,
			response:       nil,
			responseData:   nil,
			responseExpected: `{"error":{"code":400,"message":"заголовок Content-Type не передан, либо неизвестен ` +
				`тип контента: \"\"","i18nKey":"","errors":null}}`,
		},
		"Тест 03": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &struct {
				Uint  uint64 `json:"uint"  validate:"required"`
				Email string `json:"email" validate:"required,email"`
				Age   uint16 `json:"age"   validate:"gte=21,lte=130"`
			}{},
			urn:           "/",
			requestMethod: dic.Method().Post.String(),
			requestBody:   bytes.NewBufferString(`{}`),
			requestHeaders: map[string]string{
				dic.Header().ContentType.String(): dic.Mime().ApplicationMsgpack.String(),
			},
			serverUri:    nil,
			request:      nil,
			response:     nil,
			responseData: nil,
			responseExpected: `{"error":{"code":400,"message":"заголовок Content-Type не передан, либо неизвестен ` +
				`тип контента: \"application/msgpack\"","i18nKey":"","errors":null}}`,
		},
		"Тест 04": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &struct {
				Uint  uint64 `json:"uint"  validate:"required"`
				Email string `json:"email" validate:"required,email"`
				Age   uint16 `json:"age"   validate:"gte=21,lte=130"`
			}{},
			urn:           "/",
			requestMethod: dic.Method().Post.String(),
			requestBody:   bytes.NewBufferString(`{}`),
			requestHeaders: map[string]string{
				dic.Header().ContentType.String(): dic.Mime().ApplicationJson.String(),
			},
			serverUri:    nil,
			request:      nil,
			response:     nil,
			responseData: nil,
			responseExpected: `{"error":{"code":400,"message":"Key: 'Uint' Error:Field validation for 'Uint' failed ` +
				`on the 'required' tag\nKey: 'Email' Error:Field validation for 'Email' failed on the 'required' tag` +
				`\nKey: 'Age' Error:Field validation for 'Age' failed on the 'gte' tag","i18nKey":"","errors":[{"fie` +
				`ld":"Uint","fieldValue":"0","message":"required","i18nKey":""},{"field":"Email","fieldValue":"","me` +
				`ssage":"required","i18nKey":""},{"field":"Age","fieldValue":"0","message":"gte","i18nKey":""}]}}`,
		},
		"Тест 05": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &struct {
				Uint  uint64 `json:"uint"  validate:"required"`
				Email string `json:"email" validate:"required,email"`
				Age   uint16 `json:"age"   validate:"gte=21,lte=130"`
			}{},
			urn:           "/",
			requestMethod: dic.Method().Post.String(),
			requestBody:   bytes.NewBufferString(`{"uint":1, "email":"a@mail.ru"}`),
			requestHeaders: map[string]string{
				dic.Header().ContentType.String(): dic.Mime().ApplicationJson.String(),
			},
			serverUri:    nil,
			request:      nil,
			response:     nil,
			responseData: nil,
			responseExpected: `{"error":{"code":400,"message":"Key: 'Age' Error:Field validation for 'Age' failed ` +
				`on the 'gte' tag","i18nKey":"","errors":[{"field":"Age","fieldValue":"0","message":"gte","i18nKey` +
				`":""}]}}`,
		},
		"Тест 06": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &struct {
				Uint  uint64 `json:"uint"  validate:"required"`
				Email string `json:"email" validate:"required,email"`
				Age   uint16 `json:"age"   validate:"gte=21,lte=130"`
			}{},
			urn:           "/",
			requestMethod: dic.Method().Post.String(),
			requestBody:   bytes.NewBufferString(`{"uint":1, "email":"this is not e-mail"}`),
			requestHeaders: map[string]string{
				dic.Header().ContentType.String(): dic.Mime().ApplicationJson.String(),
			},
			serverUri:    nil,
			request:      nil,
			response:     nil,
			responseData: nil,
			responseExpected: `{"error":{"code":400,"message":"Key: 'Email' Error:Field validation for 'Email' ` +
				`failed on the 'email' tag\nKey: 'Age' Error:Field validation for 'Age' failed on the 'gte' tag` +
				`","i18nKey":"","errors":[{"field":"Email","fieldValue":"this is not e-mail","message":"email",` +
				`"i18nKey":""},{"field":"Age","fieldValue":"0","message":"gte","i18nKey":""}]}}`,
		},
		"Тест 07": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &struct {
				Uint  uint64 `json:"uint"  validate:"required"`
				Email string `json:"email" validate:"required,email"`
				Age   uint16 `json:"age"   validate:"gte=21,lte=130"`
			}{},
			urn:           "/",
			requestMethod: dic.Method().Post.String(),
			requestBody:   bytes.NewBufferString(`{"uint":1, "email":"a@mail.ru", "age": 20}`),
			requestHeaders: map[string]string{
				dic.Header().ContentType.String(): dic.Mime().ApplicationJson.String(),
			},
			serverUri:    nil,
			request:      nil,
			response:     nil,
			responseData: nil,
			responseExpected: `{"error":{"code":400,"message":"Key: 'Age' Error:Field validation for 'Age' failed ` +
				`on the 'gte' tag","i18nKey":"","errors":[{"field":"Age","fieldValue":"20","message":"gte","i18nKe` +
				`y":""}]}}`,
		},
		"Тест 08": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &struct {
				Uint  uint64 `json:"uint"  validate:"required"`
				Email string `json:"email" validate:"required,email"`
				Age   uint16 `json:"age"   validate:"gte=21,lte=130"`
			}{},
			urn:           "/",
			requestMethod: dic.Method().Post.String(),
			requestBody:   bytes.NewBufferString(`{"uint":1, "email":"a@mail.ru", "age": 21}`),
			requestHeaders: map[string]string{
				dic.Header().ContentType.String(): dic.Mime().ApplicationJson.String(),
			},
			serverUri:        nil,
			request:          nil,
			response:         nil,
			responseData:     nil,
			responseExpected: `{"uint":1,"email":"a@mail.ru","age":21}`,
		},
		"Тест 09": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &[]struct {
				ID    uint64 `json:"id"    validate:"required"`
				Email string `json:"email" validate:"required,email"`
			}{},
			urn:           "/",
			requestMethod: dic.Method().Post.String(),
			requestBody:   bytes.NewBufferString(`[{"id":1, "email":"a@mail.ru"},{"uint":2, "email":"b@mail.ru"}]`),
			requestHeaders: map[string]string{
				dic.Header().ContentType.String(): dic.Mime().ApplicationJson.String(),
			},
			serverUri:    nil,
			request:      nil,
			response:     nil,
			responseData: nil,
			responseExpected: `{"error":{"code":400,"message":"Key: 'ID' Error:Field validation for 'ID' failed on ` +
				`the 'required' tag","i18nKey":"","errors":[{"field":"ID","fieldValue":"0","message":"required","i1` +
				`8nKey":""}]}}`,
		},
		"Тест 10": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				return ret
			}(),
			variable: &[]struct {
				ID    uint64 `json:"id"    validate:"required"`
				Email string `json:"email" validate:"required,email"`
			}{},
			urn:           "/",
			requestMethod: dic.Method().Post.String(),
			requestBody:   bytes.NewBufferString(`[{"id":1, "email":"a@mail.ru"},{"id":2, "email":"b@mail.ru"}]`),
			requestHeaders: map[string]string{
				dic.Header().ContentType.String(): dic.Mime().ApplicationJson.String(),
			},
			serverUri:        nil,
			request:          nil,
			response:         nil,
			responseData:     nil,
			responseExpected: `[{"id":1,"email":"a@mail.ru"},{"id":2,"email":"b@mail.ru"}]`,
		},
	}

	for key := range tests {
		// Подготовка теста.
		test := tests[key]
		test.urnHandleFunc = func() http.HandlerFunc {
			return func(wr http.ResponseWriter, rq *http.Request) {
				var ans = New(nil)
				if err = ans.RqLoadVerify(wr, rq, test.variable); err != nil {
					return
				}
				ans.Json(wr, dic.Status().Ok, test.variable)
			}
		}()
		test.route.HandleFunc(test.urn, test.urnHandleFunc)
		server = httptest.NewServer(test.route)
		if test.serverUri, err = url.Parse(server.URL); err != nil {
			log.Fatal(err)
		}
		test.request, err = http.NewRequest(test.requestMethod, test.serverUri.String(), test.requestBody)
		if err != nil {
			log.Fatal(err)
		}
		for k := range test.requestHeaders {
			test.request.Header.Set(k, test.requestHeaders[k])
		}
		// Выполнение запроса к тестовому серверу.
		if test.response, err = http.DefaultClient.Do(test.request); err != nil {
			log.Fatal(err)
		}
		if test.responseData, err = io.ReadAll(test.response.Body); err != nil {
			log.Fatal(err)
		}
		_ = test.response.Body.Close()
		server.Close()
		// Проверка результата.
		if strings.TrimSpace(test.responseExpected) != strings.TrimSpace(string(test.responseData)) {
			t.Log(test.responseExpected)
			t.Log(string(test.responseData))
			t.Fatalf("Tест %q провален.", key)
		}
	}
}

package ans

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/webnice/dic"

	"github.com/go-chi/chi/v5"
)

// newChiRequest Формирование тестового запроса с параметрами маршрута chi.
func newChiRequest(params map[string]string) *http.Request {
	var (
		req  *http.Request
		rctx *chi.Context
		key  string
	)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rctx = chi.NewRouteContext()
	for key = range params {
		rctx.URLParams.Add(key, params[key])
	}

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// TestImpl_RqIdsAndRqId Идентификаторы из параметров маршрута.
func TestImpl_RqIdsAndRqId(t *testing.T) {
	const (
		tplRqIdsError       = "RqIds() вернул ошибку: %v."
		tplRqIdsUnexpected  = "RqIds() = %v, ожидалось [1 2 3]."
		errRqIdsNilError    = "RqIds() с невалидным значением вернул nil ошибку."
		tplRqIdsWrongAnswer = "RqIds() сформировал неверный ответ: code=%d body=%s."
		tplRqIdUnexpected   = "RqId() = %d, %v."
		tplRqIdEmpty        = "RqId() для пустого параметра = %d, %v."
		tplRqIdWrongValue   = "RqId() с невалидным значением вернул ошибку %v и код %d."
	)
	var (
		ans      Interface
		ids      []uint64
		id       uint64
		recorder *httptest.ResponseRecorder
		err      error
	)

	ans = New(nil)
	ids, err = ans.RqIds(httptest.NewRecorder(), newChiRequest(map[string]string{"ids": "1, 2,, 3"}), "ids")
	if err != nil {
		t.Fatalf(tplRqIdsError, err)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Fatalf(tplRqIdsUnexpected, ids)
	}
	recorder = httptest.NewRecorder()
	_, err = ans.RqIds(recorder, newChiRequest(map[string]string{"ids": "1,x"}), "ids")
	if err == nil {
		t.Fatal(errRqIdsNilError)
	}
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"field":"ids"`) {
		t.Fatalf(tplRqIdsWrongAnswer, recorder.Code, recorder.Body.String())
	}
	id, err = ans.RqId(httptest.NewRecorder(), newChiRequest(map[string]string{"id": "42"}), "id")
	if err != nil || id != 42 {
		t.Fatalf(tplRqIdUnexpected, id, err)
	}
	id, err = ans.RqId(httptest.NewRecorder(), newChiRequest(map[string]string{}), "id")
	if err != nil || id != 0 {
		t.Fatalf(tplRqIdEmpty, id, err)
	}
	recorder = httptest.NewRecorder()
	_, err = ans.RqId(recorder, newChiRequest(map[string]string{"id": "x"}), "id")
	if err == nil || recorder.Code != http.StatusBadRequest {
		t.Fatalf(tplRqIdWrongValue, err, recorder.Code)
	}
}

// TestImpl_RqVar Загрузчик переменных запроса.
func TestImpl_RqVar(t *testing.T) {
	const errRqVarNil = "RqVar() вернул nil."

	if New(nil).RqVar() == nil {
		t.Fatal(errRqVarNil)
	}
}

func TestImpl_RqLoadVerify(t *testing.T) {
	const (
		errTestFailed = "Тест %q провален."
	)
	// Описание одного теста загрузки и верификации данных.
	type testItem struct {
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
	}
	var (
		err    error
		server *httptest.Server
		tests  map[string]testItem
		key    string
		test   testItem
		k      string
	)

	tests = map[string]testItem{
		"Тест 01": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()

				return
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
			responseExpected: `{"error":{"code":-1,"message":"тело запроса пустое","i18nKey":"","errors":null}}`,
		},
		"Тест 02": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()

				return
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
			responseExpected: `{"error":{"code":-1,"message":"заголовок Content-Type не передан, либо неизвестен ` +
				`тип контента: \"\"","i18nKey":"","errors":null}}`,
		},
		"Тест 03": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()

				return
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
			responseExpected: `{"error":{"code":-1,"message":"заголовок Content-Type не передан, либо неизвестен ` +
				`тип контента: \"application/msgpack\"","i18nKey":"","errors":null}}`,
		},
		"Тест 04": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()

				return
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
			responseExpected: `{"error":{"code":-1,"message":"Key: 'Uint' Error:Field validation for 'Uint' failed ` +
				`on the 'required' tag\nKey: 'Email' Error:Field validation for 'Email' failed on the 'required' tag` +
				`\nKey: 'Age' Error:Field validation for 'Age' failed on the 'gte' tag","i18nKey":"","errors":[{"fie` +
				`ld":"Uint","fieldValue":"0","message":"required","i18nKey":""},{"field":"Email","fieldValue":"","me` +
				`ssage":"required","i18nKey":""},{"field":"Age","fieldValue":"0","message":"gte","i18nKey":""}]}}`,
		},
		"Тест 05": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()

				return
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
			responseExpected: `{"error":{"code":-1,"message":"Key: 'Age' Error:Field validation for 'Age' failed ` +
				`on the 'gte' tag","i18nKey":"","errors":[{"field":"Age","fieldValue":"0","message":"gte","i18nKey` +
				`":""}]}}`,
		},
		"Тест 06": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()

				return
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
			responseExpected: `{"error":{"code":-1,"message":"Key: 'Email' Error:Field validation for 'Email' ` +
				`failed on the 'email' tag\nKey: 'Age' Error:Field validation for 'Age' failed on the 'gte' tag` +
				`","i18nKey":"","errors":[{"field":"Email","fieldValue":"this is not e-mail","message":"email",` +
				`"i18nKey":""},{"field":"Age","fieldValue":"0","message":"gte","i18nKey":""}]}}`,
		},
		"Тест 07": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()

				return
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
			responseExpected: `{"error":{"code":-1,"message":"Key: 'Age' Error:Field validation for 'Age' failed ` +
				`on the 'gte' tag","i18nKey":"","errors":[{"field":"Age","fieldValue":"20","message":"gte","i18nKe` +
				`y":""}]}}`,
		},
		"Тест 08": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()

				return
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

				return
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
			responseExpected: `{"error":{"code":-1,"message":"Key: 'ID' Error:Field validation for 'ID' failed on ` +
				`the 'required' tag","i18nKey":"","errors":[{"field":"ID","fieldValue":"0","message":"required","i1` +
				`8nKey":""}]}}`,
		},
		"Тест 10": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()

				return
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

	for key = range tests {
		// Подготовка теста.
		test = tests[key]
		test.urnHandleFunc = func() http.HandlerFunc {
			return func(wr http.ResponseWriter, rq *http.Request) {
				var ans Interface

				ans = New(nil)
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
		for k = range test.requestHeaders {
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
			t.Fatalf(errTestFailed, key)
		}
	}
}

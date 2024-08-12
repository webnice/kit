package rqvar

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/webnice/dic"

	"github.com/go-chi/chi/v5"
)

// Извлечение значения поля через функцию, назначает значение библиотека с конвертацией.
type testBase64FieldDecoderToString struct {
	ID int64 `call-func:"Base64FieldDecoder"`
}

func (bfds *testBase64FieldDecoderToString) Base64FieldDecoder(rq *http.Request) (ret string) {
	s := struct {
		ID int `json:"id"`
		Dt int `json:"data"`
	}{}
	if tmp := rq.URL.Query().Get("id"); tmp != "" {
		buf, _ := base64.StdEncoding.DecodeString(tmp)
		_ = json.Unmarshal(buf, &s)
		ret = strconv.FormatInt(int64(s.ID*s.Dt), 10)
	}
	return
}

// Извлечение значения поля через функцию, назначает значение функция, при этом функция должна вернуть пустую строку.
type testBase64FieldDecoderToObject struct {
	GoodAdvice []rune `call-func:"ExtractRunes"`
}

func (bfdo *testBase64FieldDecoderToObject) ExtractRunes(rq *http.Request) (ret string) {
	if tmp := rq.URL.Query().Get("id"); tmp != "" {
		buf, _ := base64.StdEncoding.DecodeString(tmp)
		ret = strings.TrimSpace(string(buf))
	}
	return
}

type testSessionUser struct {
	UserID          uint64
	AuthType        string
	ExpiresDuration time.Duration
	ExpiresAt       time.Time
}

func TestGet(t *testing.T) {
	var rqv Interface

	singleton = nil
	rqv = Get()
	if singleton == nil {
		t.Errorf("Не верное начальное значение singleton")
	}
	if rqv.(*impl) != singleton {
		t.Errorf("Ошибка в функции Get()")
	}
}

func TestImpl_Load_NotAddressable(t *testing.T) {
	type NotAddressable struct{ ID uint64 }
	var (
		err       error
		testErr   error
		routePath *chi.Mux
		srv       *httptest.Server
		uri       *url.URL
		rsp       *http.Response
	)

	routePath = chi.NewMux()
	routePath.Get("/article/{ArticleID}/", func(wr http.ResponseWriter, rq *http.Request) {
		testErr = Get().Load(rq, NotAddressable{})
	})
	srv = httptest.NewServer(routePath)
	if uri, err = url.Parse(srv.URL); err != nil {
		log.Fatal(err)
	}
	uri.Path = "/article/813/"
	uri.RawQuery = "page_id=4321&number_id=-123456789012345"
	if rsp, err = http.Get(uri.String()); err != nil {
		log.Fatal(err)
	}
	_, err = io.ReadAll(rsp.Body)
	_ = rsp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	// Ожидается ошибка: "передана копия объекта, требуется ссылка на объект".
	if testErr == nil {
		t.Errorf("Выполнение функции Load() завершилось без ошибки, ожидалась ошибка")
		return
	}
}

func TestImpl_Load(t *testing.T) {
	var (
		err    error
		server *httptest.Server
	)

	tests := map[string]struct {
		err               error             // Ошибка загрузчика.
		route             *chi.Mux          // Роутер ВЕБ сервера.
		variable          any               // Объект загружаемой структуры.
		urn               string            // URN роутинга сервера.
		urnHandleFunc     http.HandlerFunc  // Функция-обработчик URN.
		serverUri         *url.URL          // URI запроса к ВЕБ серверу.
		serverUriPath     string            // Путь запроса.
		serverUriRawQuery string            // Параметры запроса.
		requestMethod     string            // Метод запроса.
		request           *http.Request     // Объект запроса.
		requestHeaders    map[string]string // Заголовки запроса.
		requestCookie     []*http.Cookie    // Печеньки запроса.
		response          *http.Response    // Объект ответа.
		responseData      []byte            // Тело ответа.
		responseExpected  string            // Ожидаемое тело ответа.
	}{
		"Тест 01": {
			err:   nil,
			route: chi.NewMux(),
			variable: &struct {
				Authorization string `header:"Authorization"`
				Auth          string `header:"Access-Token"`
				AuthCookie    string `cookie:"Access-Token"`
				AuthParams    string `urn-param:"accessToken,accesstoken"`
			}{},
			urn:               "/",
			serverUriPath:     "/",
			serverUriRawQuery: "accesstoken=C96C6A6E-095C-4F69-BB6B-C61921E7C0A0",
			requestMethod:     dic.Method().Get.String(),
			requestHeaders: map[string]string{
				dic.Header().Authorization.String(): "Bearer 1C236FFE-41D4-4D45-8718-BC2A49102F33",
				"Access-Token":                      "Bearer 764B24B2-087C-4AD4-BEBB-4B70CDC67B55",
			},
			requestCookie: []*http.Cookie{
				{Name: "Access-Token", Value: "5607A66D-559A-4B1F-98B1-D0033D3EAA39",
					Path: "/", MaxAge: 3600, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode},
			},
			serverUri:        nil,
			request:          nil,
			response:         nil,
			responseData:     nil,
			responseExpected: `{"Authorization":"1C236FFE-41D4-4D45-8718-BC2A49102F33","Auth":"Bearer 764B24B2-087C-4AD4-BEBB-4B70CDC67B55","AuthCookie":"5607A66D-559A-4B1F-98B1-D0033D3EAA39","AuthParams":"C96C6A6E-095C-4F69-BB6B-C61921E7C0A0"}`,
		},
		"Тест 02": {
			err:   nil,
			route: chi.NewMux(),
			variable: &struct {
				ProjectID int32  `cookie:"Project-Id"`
				ArticleID uint64 `path-param:"ArticleID"`
				Position  int    `urn-param:"pos"`
				Limit     int64  `urn-param:"limit"`
			}{},
			urn:               "/article/{ArticleID}",
			serverUriPath:     "/article/810",
			serverUriRawQuery: "pos=100&limit=-1",
			requestMethod:     dic.Method().Get.String(),
			requestHeaders:    map[string]string{},
			requestCookie: []*http.Cookie{
				{Name: "Project-Id", Value: "5607",
					Path: "/", MaxAge: 3600, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode},
			},
			serverUri:        nil,
			request:          nil,
			response:         nil,
			responseData:     nil,
			responseExpected: `{"ProjectID":5607,"ArticleID":810,"Position":100,"Limit":-1}`,
		},
		"Тест 03": {
			err:   nil,
			route: chi.NewMux(),
			variable: &struct {
				ID         uint64 `urn-param:"SectionId,Id,Ids,id"`
				ApiVersion string `path-param:"ApiVersion"`
				ArticleID  uint64 `path-param:"ArticleID"`
				Page       uint16 `path-param:"page-id"`
				IsOpen     *bool  `urn-param:"open"`
			}{},
			urn:               "/api/{ApiVersion}/article/{ArticleID}/page/{page-id}",
			serverUriPath:     "/api/v1.0/article/810/page/200",
			serverUriRawQuery: "id=18&open=1",
			requestMethod:     dic.Method().Get.String(),
			requestHeaders:    map[string]string{},
			requestCookie:     []*http.Cookie{},
			serverUri:         nil,
			request:           nil,
			response:          nil,
			responseData:      nil,
			responseExpected:  `{"ID":18,"ApiVersion":"v1.0","ArticleID":810,"Page":200,"IsOpen":true}`,
		},
		"Тест 04": {
			err:   nil,
			route: chi.NewMux(),
			variable: &struct {
				Authorization string  `header:"   Authorization    "  cookie:",  Access-Token,  "  urn-param:" accessToken, accesstoken"`
				IP            net.IP  `header:"X-Client-Forwarded-For,    X-Real-IP   ,,,,X-Forwarded-For"  context:""`
				Pi            float64 `cookie:"piko"                  context:"-"`
			}{},
			urn:               "/api/{ApiVersion}/article/{ArticleID}/page/{page-id}",
			serverUriPath:     "/api/v1.0/article/810/page/200",
			serverUriRawQuery: "accesstoken=C96C6A6E-095C-4F69-BB6B-C61921E7C0A0",
			requestMethod:     dic.Method().Get.String(),
			requestHeaders: map[string]string{
				dic.Header().Authorization.String(): "Bearer 1C236FFE-41D4-4D45-8718-BC2A49102F33",
				"X-Real-IP":                         "4.3.2.1",
				"X-Forwarded-For":                   "8.7.6.5",
			},
			requestCookie: []*http.Cookie{
				{Name: "Access-Token", Value: "5607A66D-559A-4B1F-98B1-D0033D3EAA39",
					Path: "/", MaxAge: 3600, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode},
				{Name: "piko", Value: strconv.FormatFloat(math.Pi, 'g', -1, 64),
					Path: "/", MaxAge: 3600, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode},
			},
			serverUri:        nil,
			request:          nil,
			response:         nil,
			responseData:     nil,
			responseExpected: `{"Authorization":"1C236FFE-41D4-4D45-8718-BC2A49102F33","IP":"4.3.2.1","Pi":3.141592653589793}`,
		},
		"Тест 05": {
			err:           nil,
			route:         chi.NewMux(),
			variable:      new(testBase64FieldDecoderToString),
			urn:           "/",
			serverUriPath: "/",
			serverUriRawQuery: func() string {
				v := url.Values{}
				v.Set("id", "eyJpZCI6LTMxOCwiZGF0YSI6MTh9")
				return v.Encode()
			}(),
			requestMethod:    dic.Method().Get.String(),
			requestHeaders:   map[string]string{},
			requestCookie:    []*http.Cookie{},
			serverUri:        nil,
			request:          nil,
			response:         nil,
			responseData:     nil,
			responseExpected: `{"ID":-5724}`,
		},
		"Тест 06": {
			err:           nil,
			route:         chi.NewMux(),
			variable:      new(testBase64FieldDecoderToObject),
			urn:           "/",
			serverUriPath: "/",
			serverUriRawQuery: func() string {
				v := url.Values{}
				v.Set("id", "5a2m5L+E6K+t77yB")
				return v.Encode()
			}(),
			requestMethod:    dic.Method().Get.String(),
			requestHeaders:   map[string]string{},
			requestCookie:    []*http.Cookie{},
			serverUri:        nil,
			request:          nil,
			response:         nil,
			responseData:     nil,
			responseExpected: `{"GoodAdvice":[23398,20420,35821,65281]}`,
		},
		"Тест 07": {
			err: nil,
			route: func() (ret *chi.Mux) {
				ret = chi.NewMux()
				ret.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(wr http.ResponseWriter, rq *http.Request) {
						rq = rq.WithContext(context.WithValue(rq.Context(), "context_string_value", "GjE6p3Z7OcliODc"))
						rq = rq.WithContext(context.WithValue(rq.Context(), "AuthorizationUserID", "87514"))
						rq = rq.WithContext(context.WithValue(rq.Context(), "session-user",
							&testSessionUser{
								UserID:          87514,
								AuthType:        "login/password",
								ExpiresDuration: time.Hour,
								ExpiresAt:       time.Unix(1723426620, 0).UTC(),
							},
						))
						rq = rq.WithContext(context.WithValue(rq.Context(), "project-id", 819))
						next.ServeHTTP(wr, rq)
					})
				})
				return ret
			}(),
			variable: &struct {
				ContextString string           `context:"Context-String-Value,myStringValue,context_string_value,,"`
				UserID        uint64           `context:"AuthorizationUserID"`
				Session       *testSessionUser `context:"session-user"`
				ProjectID     int64            `context:"project-id"`
			}{},
			urn:               "/",
			serverUriPath:     "/",
			serverUriRawQuery: "",
			requestMethod:     dic.Method().Get.String(),
			requestHeaders:    map[string]string{},
			requestCookie:     []*http.Cookie{},
			serverUri:         nil,
			request:           nil,
			response:          nil,
			responseData:      nil,
			responseExpected: `{"ContextString":"GjE6p3Z7OcliODc","UserID":87514,"Session":{"UserID":87514,"AuthType"` +
				`:"login/password","ExpiresDuration":3600000000000,"ExpiresAt":"2024-08-12T01:37:00Z"},"ProjectID":819}`,
		},
	}
	for key := range tests {
		// Подготовка теста.
		test := tests[key]
		test.urnHandleFunc = func() http.HandlerFunc {
			return func(wr http.ResponseWriter, rq *http.Request) {
				test.err = Get().Load(rq, test.variable)
				buf, _ := json.Marshal(test.variable)
				_, _ = wr.Write(buf)
			}
		}()
		test.route.HandleFunc(test.urn, test.urnHandleFunc)
		server = httptest.NewServer(test.route)
		if test.serverUri, err = url.Parse(server.URL); err != nil {
			log.Fatal(err)
		}
		test.serverUri.Path, test.serverUri.RawQuery = test.serverUriPath, test.serverUriRawQuery
		if test.request, err = http.NewRequest(test.requestMethod, test.serverUri.String(), nil); err != nil {
			log.Fatal(err)
		}
		for k := range test.requestHeaders {
			test.request.Header.Set(k, test.requestHeaders[k])
		}
		for n := range test.requestCookie {
			test.request.AddCookie(test.requestCookie[n])
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
		if test.responseExpected != string(test.responseData) {
			t.Log(test.responseExpected)
			t.Log(string(test.responseData))
			t.Fatalf("Tест %q провален.", key)
		}
	}
}

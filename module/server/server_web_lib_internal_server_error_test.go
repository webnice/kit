package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInternalServerErrorGetSet(t *testing.T) {
	const (
		errGetNilDefault = "функция InternalServerErrorGet(), вернулся nil, ожидалась функция по умолчанию."
		errSetNil        = "функция InternalServerErrorSet(), вернулся nil, ожидался интерфейс."
		errGetNilCustom  = "функция InternalServerErrorGet(), вернулся nil, ожидалась пользовательская функция."
		errNilChanged    = "функция InternalServerErrorSet(nil), значение не должно изменяться."
	)
	var (
		srv Interface
		iwl *implWebLib
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	// Функция обработки ошибки по умолчанию.
	if fn := iwl.InternalServerErrorGet(); fn == nil {
		t.Error(errGetNilDefault)
	}
	// Установка пользовательской функции.
	custom := func(http.ResponseWriter, *http.Request) {}
	if ret := iwl.InternalServerErrorSet(custom); ret == nil {
		t.Error(errSetNil)
	}
	if fn := iwl.InternalServerErrorGet(); fn == nil {
		t.Error(errGetNilCustom)
	}
	// Установка nil функции не изменяет значение.
	iwl.InternalServerErrorSet(nil)
	if fn := iwl.InternalServerErrorGet(); fn == nil {
		t.Error(errNilChanged)
	}
}

func TestDefaultInternalServerError(t *testing.T) {
	const (
		errTextServerError = "тестовая ошибка сервера."
		tplCodeWrong       = "функция defaultInternalServerError(), код %d, ожидался код 500."
		tplBodyNoStatus    = "функция defaultInternalServerError(), тело %q не содержит код статуса."
		tplBodyNoError     = "функция defaultInternalServerError(), тело %q не содержит текст ошибки."
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
		rsp *httptest.ResponseRecorder
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rsp = httptest.NewRecorder()
	// Ответ без установленной ошибки InternalServerError, в ответ передаётся код статуса.
	iwl.defaultInternalServerError(rsp, rq)
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf(tplCodeWrong, rsp.Code)
	}
	if body := rsp.Body.String(); !strings.Contains(body, "Internal Server Error") {
		t.Errorf(tplBodyNoStatus, body)
	}
	// Ответ с установленной ошибкой InternalServerError, в ответ передаётся текст ошибки.
	rsp = httptest.NewRecorder()
	srv.Web().Error().InternalServerError(errors.New(errTextServerError))
	iwl.defaultInternalServerError(rsp, rq)
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf(tplCodeWrong, rsp.Code)
	}
	if body := rsp.Body.String(); !strings.Contains(body, errTextServerError) {
		t.Errorf(tplBodyNoError, body)
	}
}

package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInternalServerErrorGetSet(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
	)

	// Функция обработки ошибки по умолчанию.
	if fn := iwl.InternalServerErrorGet(); fn == nil {
		t.Error("функция InternalServerErrorGet(), вернулся nil, ожидалась функция по умолчанию")
	}
	// Установка пользовательской функции.
	custom := func(http.ResponseWriter, *http.Request) {}
	if ret := iwl.InternalServerErrorSet(custom); ret == nil {
		t.Error("функция InternalServerErrorSet(), вернулся nil, ожидался интерфейс")
	}
	if fn := iwl.InternalServerErrorGet(); fn == nil {
		t.Error("функция InternalServerErrorGet(), вернулся nil, ожидалась пользовательская функция")
	}
	// Установка nil функции не изменяет значение.
	iwl.InternalServerErrorSet(nil)
	if fn := iwl.InternalServerErrorGet(); fn == nil {
		t.Error("функция InternalServerErrorSet(nil), значение не должно изменяться")
	}
}

func TestDefaultInternalServerError(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rsp = httptest.NewRecorder()
	)

	// Ответ без установленной ошибки InternalServerError, в ответ передаётся код статуса.
	iwl.defaultInternalServerError(rsp, rq)
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf("функция defaultInternalServerError(), код %d, ожидался код 500", rsp.Code)
	}
	if body := rsp.Body.String(); !strings.Contains(body, "Internal Server Error") {
		t.Errorf("функция defaultInternalServerError(), тело %q не содержит код статуса", body)
	}
	// Ответ с установленной ошибкой InternalServerError, в ответ передаётся текст ошибки.
	rsp = httptest.NewRecorder()
	srv.Web().Error().InternalServerError(errors.New("тестовая ошибка сервера"))
	iwl.defaultInternalServerError(rsp, rq)
	if rsp.Code != http.StatusInternalServerError {
		t.Errorf("функция defaultInternalServerError(), код %d, ожидался код 500", rsp.Code)
	}
	if body := rsp.Body.String(); !strings.Contains(body, "тестовая ошибка сервера") {
		t.Errorf("функция defaultInternalServerError(), тело %q не содержит текст ошибки", body)
	}
}

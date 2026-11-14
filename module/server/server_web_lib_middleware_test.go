package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetContextWrapper(t *testing.T) {
	const (
		tplNoDataNotNil    = "функция getContextWrapper() без данных в контексте, вернула %v, ожидался nil."
		tplWrongTypeNotNil = "функция getContextWrapper() с не верным типом данных, вернула %v, ожидался nil."
		tplWrongNameNotNil = "функция getContextWrapper() с не верным именем, вернула %v, ожидался nil."
		tplNilValueNotNil  = "функция getContextWrapper() с nil значением, вернула %v, ожидался nil."
		tplCorrectNil      = "функция getContextWrapper() с корректными данными, вернула %v, ожидался объект."
	)
	var (
		srv Interface
		iwl *implWebLib
		rq  *http.Request
		key *contextKey
	)

	srv = New(&testLogger{})
	iwl = srv.(*impl).serverWeb.lib
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	key = &contextKey{"test-key"}
	// Данные в контексте отсутствуют.
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret != nil {
		t.Errorf(tplNoDataNotNil, ret)
	}
	// В контексте данные не *contextWrapper.
	rq = rq.WithContext(context.WithValue(rq.Context(), key, "wrong value"))
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret != nil {
		t.Errorf(tplWrongTypeNotNil, ret)
	}
	// В контексте *contextWrapper с не верным именем.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(context.WithValue(rq.Context(), key, &contextWrapper{Name: "wrong-name", Value: 1}))
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret != nil {
		t.Errorf(tplWrongNameNotNil, ret)
	}
	// В контексте *contextWrapper с nil значением.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(context.WithValue(rq.Context(), key, &contextWrapper{Name: "test-name"}))
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret != nil {
		t.Errorf(tplNilValueNotNil, ret)
	}
	// В контексте корректные данные.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(context.WithValue(rq.Context(), key, &contextWrapper{Name: "test-name", Value: 1}))
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret == nil || ret.Value != 1 {
		t.Errorf(tplCorrectNil, ret)
	}
}

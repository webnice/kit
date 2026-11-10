package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetContextWrapper(t *testing.T) {
	var (
		srv = New(&testLogger{})
		iwl = srv.(*impl).serverWeb.lib
		rq  = httptest.NewRequest(http.MethodGet, "/", nil)
		key = &contextKey{"test-key"}
	)

	// Данные в контексте отсутствуют.
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret != nil {
		t.Errorf("функция getContextWrapper() без данных в контексте, вернула %v, ожидался nil", ret)
	}
	// В контексте данные не *contextWrapper.
	rq = rq.WithContext(context.WithValue(rq.Context(), key, "wrong value"))
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret != nil {
		t.Errorf("функция getContextWrapper() с не верным типом данных, вернула %v, ожидался nil", ret)
	}
	// В контексте *contextWrapper с не верным именем.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(context.WithValue(rq.Context(), key, &contextWrapper{Name: "wrong-name", Value: 1}))
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret != nil {
		t.Errorf("функция getContextWrapper() с не верным именем, вернула %v, ожидался nil", ret)
	}
	// В контексте *contextWrapper с nil значением.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(context.WithValue(rq.Context(), key, &contextWrapper{Name: "test-name"}))
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret != nil {
		t.Errorf("функция getContextWrapper() с nil значением, вернула %v, ожидался nil", ret)
	}
	// В контексте корректные данные.
	rq = httptest.NewRequest(http.MethodGet, "/", nil)
	rq = rq.WithContext(context.WithValue(rq.Context(), key, &contextWrapper{Name: "test-name", Value: 1}))
	if ret := iwl.getContextWrapper(rq, key, "test-name"); ret == nil || ret.Value != 1 {
		t.Errorf("функция getContextWrapper() с корректными данными, вернула %v, ожидался объект", ret)
	}
}

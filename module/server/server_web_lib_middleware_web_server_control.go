package server

import (
	"context"
	"fmt"
	"net/http"

	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// WebServerControlHandler Обработчик установки в контекст запросов ВЕБ сервера, объекта контроля за ВЕБ сервером.
func (iwl *implWebLib) WebServerControlHandler(ctl *kitTypesServer.WebServerControl) (
	ret func(http.Handler) http.Handler,
) {
	ret = func(next http.Handler) (fn http.Handler) {
		fn = http.HandlerFunc(func(wr http.ResponseWriter, rq *http.Request) {
			rq = rq.WithContext(
				context.WithValue(rq.Context(), contextKeyMiddlewareWebServerControl, &contextWrapper{
					Name:  contextNameMiddlewareWebServerControl,
					Value: ctl,
				}))
			next.ServeHTTP(wr, rq)
		})
		return
	}

	return
}

// WebServerControlGetFromContext Функция извлечения объекта контроля за ВЕБ сервером из контекста ВЕБ сервера.
func (iwl *implWebLib) WebServerControlGetFromContext(rq *http.Request) (
	ret *kitTypesServer.WebServerControl,
	err error,
) {
	const (
		tplNotFound   = "данные по ключу %q в контексте не найдены"
		tplWrongType  = "не верный тип данных полученных по ключу %q из контекста"
		tplWrongValue = "по ключу %q получена обёртка контекста содержащая не верные данные"
	)
	var (
		ok      bool
		value   any
		wrapper *contextWrapper
	)

	if value = rq.Context().Value(contextKeyMiddlewareWebServerControl); value == nil {
		err = fmt.Errorf(tplNotFound, contextNameMiddlewareWebServerControl)
		return
	}
	if wrapper, ok = value.(*contextWrapper); !ok {
		err = fmt.Errorf(tplWrongType, contextNameMiddlewareWebServerControl)
		return
	}
	if wrapper == nil || wrapper.Value == nil || wrapper.Name != contextNameMiddlewareWebServerControl {
		err = fmt.Errorf(tplWrongValue, contextNameMiddlewareWebServerControl)
		return
	}
	if ret, ok = wrapper.Value.(*kitTypesServer.WebServerControl); !ok {
		err = fmt.Errorf(tplWrongValue, contextNameMiddlewareWebServerControl)
		return
	}

	return
}

package server

import (
	"net/http"

	"github.com/webnice/dic"
)

// InternalServerErrorGet Получение функции для обработки внутренней ошибки ВЕБ сервера.
func (iwl *implWebLib) InternalServerErrorGet() (ret http.HandlerFunc) {
	if ret = iwl.funcs.fnInternalServerError; ret == nil {
		ret = iwl.defaultInternalServerError
	}

	return
}

// InternalServerErrorSet Установка пользовательской функции для обработки внутренней ошибки ВЕБ сервера.
func (iwl *implWebLib) InternalServerErrorSet(fn http.HandlerFunc) InterfaceHandlerFunc {
	if fn != nil {
		iwl.funcs.fnInternalServerError = fn
	}

	return iwl
}

func (iwl *implWebLib) defaultInternalServerError(wr http.ResponseWriter, _ *http.Request) {
	var err error

	wr.Header().Set(dic.Header().ContentType.String(), dic.Mime().TextPlain.String())
	wr.WriteHeader(dic.Status().InternalServerError.Code())
	if err = iwl.parent.serverWeb.error.InternalServerError(nil); err == nil {
		_, _ = wr.Write(dic.Status().InternalServerError.Bytes())
		return
	}
	_, _ = wr.Write([]byte(err.Error()))
}

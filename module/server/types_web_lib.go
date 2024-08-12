package server

import (
	"net/http"

	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// InterfaceWebLib Интерфейс библиотеки функций ВЕБ сервера.
type InterfaceWebLib interface {
	// БИБЛИОТЕКИ

	// Handler Интерфейс библиотеки ВЕБ функций Handler и HandlerFunc.
	Handler() InterfaceHandlerFunc

	// Middleware Интерфейс библиотеки ВЕБ функций "промежуточного слоя" для роутинга на базе библиотеки chi.
	Middleware() InterfaceMiddleware
}

// InterfaceHandlerFunc Интерфейс библиотеки ВЕБ функций Handler и HandlerFunc.
type InterfaceHandlerFunc interface {
	// ФУНКЦИИ

	// InternalServerErrorGet Получение функции для обработки внутренней ошибки ВЕБ сервера.
	InternalServerErrorGet() (ret http.HandlerFunc)

	// InternalServerErrorSet Установка пользовательской функции для обработки внутренней ошибки ВЕБ сервера.
	InternalServerErrorSet(fn http.HandlerFunc) InterfaceHandlerFunc
}

// InterfaceMiddleware Интерфейс библиотеки ВЕБ функций "промежуточного слоя".
type InterfaceMiddleware interface {
	// ПРОМЕЖУТОЧНЫЙ СЛОЙ

	// RecoverHandler Обработчик восстановления после паники в ВЕБ сервере.
	RecoverHandler() (ret func(http.Handler) http.Handler)

	// WebServerControlHandler Обработчик установки в контекст запросов ВЕБ сервера, объекта контроля за ВЕБ сервером.
	WebServerControlHandler(ctl *kitTypesServer.WebServerControl) (ret func(http.Handler) http.Handler)

	// WebServerControlGetFromContext Функция извлечения объекта контроля за ВЕБ сервером из контекста ВЕБ сервера.
	WebServerControlGetFromContext(rq *http.Request) (ret *kitTypesServer.WebServerControl, err error)

	// LogHandler Запись в журнал запросов к ВЕБ серверу.
	LogHandler() (ret func(http.Handler) http.Handler)
}

// Объект сущности, реализующий интерфейс InterfaceWebLib.
type implWebLib struct {
	parent *impl      // Адрес объекта родительской сущности (parent), интерфейс Interface.
	funcs  *fnHandler // Функции HandlerFunc.
}

// Функции Handler и HandlerFunc.
type fnHandler struct {
	fnInternalServerError http.HandlerFunc // Функция обработки ошибки InternalServerError.
}

// Обёртка для формирования ключа данных контекста ВЕБ запросов.
type contextKey struct {
	name string
}

// Обёртка над данными размещаемыми в контексте ВЕБ запросов.
type contextWrapper struct {
	Name  string
	Value any
}

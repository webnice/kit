package server

import kitTypes "github.com/webnice/kit/v4/types"

// New Конструктор объекта сущности пакета, возвращается интерфейс пакета.
func New(l kitTypes.Logger) Interface {
	var sri = &impl{
		logger: l,
		web:    new(implWeb),
		grpc:   new(implGrpc),
	}
	sri.gist = newEssence(sri)
	sri.web.parent, sri.grpc.parent = sri, sri

	return sri
}

// Ссылка на менеджер логирования.
func (sri *impl) log() kitTypes.Logger { return sri.logger }

// Errors Справочник ошибок.
func (sri *impl) Errors() *Error { return Errors() }

// Gist Интерфейс к служебным методам.
func (sri *impl) Gist() Essence { return sri.gist }

// Web Интерфейс веб сервера.
func (sri *impl) Web() InterfaceWeb { return sri.web }

// Grpc Интерфейс GRPC сервера.
func (sri *impl) Grpc() InterfaceGrpc { return sri.grpc }

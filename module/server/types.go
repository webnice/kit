package server

import (
	kitModuleUuid "github.com/webnice/kit/v4/module/uuid"
	kitTypes "github.com/webnice/kit/v4/types"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// Interface Интерфейс пакета.
type Interface interface {
	// Gist Интерфейс к служебным методам.
	Gist() Essence

	// Web Интерфейс веб сервера.
	Web() InterfaceWeb

	// Grpc Интерфейс GRPC сервера.
	Grpc() InterfaceGrpc

	// Start Запуск сервера.
	// Если не переданы идентификаторы, запускаются все добавленные серверы.
	// Если идентификаторы переданы, тогда будут запущены только серверы с указанными идентификаторами.
	// Если переданы не корректные идентификаторы, будет возвращена ошибка.
	Start(IDs ...string) (err error)

	// Stop Остановка сервера.
	// Если не переданы идентификаторы, останавливаются все серверы.
	// Если идентификаторы переданы, тогда будут остановлены только серверы с указанными идентификаторами.
	// Если переданы не корректные идентификаторы, будет возвращена ошибка.
	Stop(IDs ...string) (err error)
}

// Essence Служебный публичный интерфейс.
type Essence interface {
	// Debug Присвоение нового значения режима отладки.
	Debug(debug bool) Essence

	// ConfigurationWeb Добавление конфигурации веб сервера. Функцию можно вызывать многократно, для добавления
	// нескольких веб серверов. Серверы не должны пересекаться по занимаемому IP и порту или другим монопольно
	// занимаемым ресурсам.
	// Возвращается UUID идентификатор добавленного веб сервера.
	ConfigurationWeb(cfg *kitTypesServer.WebServerConfiguration) (ret string)
}

// Объект сущности, реализующий интерфейс Interface.
type impl struct {
	debug   bool            // Флаг режима отладки.
	logger  kitTypes.Logger // Интерфейс менеджера логирования.
	gist    *gist           // Объект служебного интерфейса Essence.
	web     *implWeb        // Объект WEB сервера.
	grpc    *implGrpc       // Объект GRPC сервера.
	servers []*server       // Добавленные конфигурации серверов.
}

// Объект сути сущности, интерфейс Essence.
type gist struct {
	parent *impl // Адрес объекта основной сущности, интерфейс Interface.
}

// Описание сервера.
type server struct {
	ID   kitModuleUuid.UUID
	Cfg  *kitTypesServer.WebServerConfiguration
	Type serverType
}

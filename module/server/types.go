/**

	Сервер реализует обработку доступа к ресурсам с использованием:
	1. ВЕБ, REST, GRPC запросов;
	2. TCP/IP соединений и UDP пакето;

	Особенности:
	Если в конфигурации приложения не определён GRPC сервер, запускаемый на выделенном сокете или IP/порт, но в
    приложении зарегистрированы GRPC контроллеры, тогда обслуживание GRPC запросов берёт на себя ВЕБ сервер,
    используя мультиплексирование WEB/REST и GRPC запросов, разделяя их по заголовкам типа контента.

**/

package server

import (
	"sync"

	kitTypes "github.com/webnice/kit/v4/types"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// Interface Интерфейс пакета.
type Interface interface {
	// Gist Интерфейс к служебным методам.
	Gist() Essence

	// Web Интерфейс веб сервера.
	Web() kitTypesServer.IServer[kitTypesServer.Web]

	// Grpc Интерфейс GRPC сервера.
	//Grpc() kitTypesServer.IServer[kitTypesServer.Grpc]

	// Tcp Интерфейс TCP/IP сервера.
	//Tcp() kitTypesServer.IServer[kitTypesServer.Tcp]

	// Udp Интерфейс UDP сервера.
	//Udp() kitTypesServer.IServer[kitTypesServer.Udp]

	// Start Запуск всех зарегистрированных серверов.
	// Если не зарегистрирован ни один сервер, функция возвращает ошибку.
	//Start() (err error)

	// Stop Остановка всех зарегистрированных серверов.
	// Если не зарегистрирован ни один сервер, функция возвращает ошибку.
	//Stop() (err error)

	// Errors Справочник ошибок.
	Errors() *Error
}

// Essence Служебный публичный интерфейс.
type Essence interface {
	// Debug Присвоение нового значения режима отладки.
	Debug(debug bool) Essence

	// WebAdd Добавление конфигурации веб сервера. Функцию можно вызывать многократно, для добавления
	// нескольких веб серверов. Серверы не должны пересекаться по занимаемому IP и порту или другим монопольно
	// выделяемым ресурсам.
	// Возвращается объект добавленного веб сервера.
	WebAdd(cfg *kitTypesServer.WebConfiguration) (ret *kitTypesServer.Server)

	// WebDel Удаление конфигурации сервера. В качестве идентификатора, передаётся UUID сервера,
	// полученный при добавлении конфигурации (Свойство ID структуры сервера).
	// Нельзя удалять запущенный сервер, функция вернёт ошибку.
	WebDel(IDs ...string) (err error)
}

// Объект сущности, реализующий интерфейс Interface.
type impl[
	T master | kitTypesServer.Web | kitTypesServer.Grpc | kitTypesServer.Tcp | kitTypesServer.Udp,
] struct {
	debug      bool                                 // Флаг режима отладки.
	logger     kitTypes.Logger                      // Интерфейс менеджера логирования.
	gist       *gist[T]                             // Объект служебного интерфейса Essence.
	sWeb       *implIServer[kitTypesServer.Web]     // Объект интерфейса WEB сервера.
	sGrpc      *implIServer[kitTypesServer.Grpc]    // Объект интерфейса GRPC сервера.
	sTcp       *implIServer[kitTypesServer.Tcp]     // Объект интерфейса TCP/IP сервера.
	sUdp       *implIServer[kitTypesServer.Udp]     // Объект интерфейса TCP/IP сервера.
	serverLock *sync.RWMutex                        // Защита карты от конкурентного доступа.
	server     map[kitTypesServer.Type][]*server[T] // Карта добавленных конфигураций серверов разбитых по типу.
}

// Объект сути сущности, интерфейс Essence.
type gist[
	T master | kitTypesServer.Web | kitTypesServer.Grpc | kitTypesServer.Tcp | kitTypesServer.Udp,
] struct {
	p *impl[master] // Адрес объекта родительской сущности (parent), интерфейс Interface.
}

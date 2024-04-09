package server

import (
	"sync"

	kitTypes "github.com/webnice/kit/v4/types"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// New Конструктор объекта сущности пакета, возвращается интерфейс пакета.
func New(l kitTypes.Logger) Interface {
	var sri = &impl[master]{
		logger:     l,
		sWeb:       new(implIServer[kitTypesServer.Web]),
		sGrpc:      new(implIServer[kitTypesServer.Grpc]),
		sTcp:       new(implIServer[kitTypesServer.Tcp]),
		sUdp:       new(implIServer[kitTypesServer.Udp]),
		serverLock: new(sync.RWMutex),
		server:     make(map[kitTypesServer.Type][]*server[master]),
	}
	sri.gist = newEssence(sri)
	sri.sWeb.p, sri.sGrpc.p, sri.sTcp.p = sri, sri, sri

	return sri
}

// Ссылка на менеджер логирования.
func (sio *impl[T]) log() kitTypes.Logger { return sio.logger }

// Errors Справочник ошибок.
func (sio *impl[T]) Errors() *Error { return Errors() }

// Gist Интерфейс к служебным методам.
func (sio *impl[T]) Gist() Essence { return sio.gist }

// Web Интерфейс веб сервера.
func (sio *impl[T]) Web() kitTypesServer.IServer[kitTypesServer.Web] { return sio.sWeb }

// Grpc Интерфейс GRPC сервера.
func (sio *impl[T]) Grpc() kitTypesServer.IServer[kitTypesServer.Grpc] { return sio.sGrpc }

// Tcp Интерфейс TCP/IP сервера.
func (sio *impl[T]) Tcp() kitTypesServer.IServer[kitTypesServer.Tcp] { return sio.sTcp }

// Udp Интерфейс UDP сервера.
func (sio *impl[T]) Udp() kitTypesServer.IServer[kitTypesServer.Udp] { return sio.sUdp }

// Start Запуск всех зарегистрированных серверов.
// Если не зарегистрирован ни один сервер, функция возвращает ошибку.
func (sio *impl[T]) Start() (err error) {
	// Сервер GRPC соединений.
	if err = sio.sGrpc.
		Start(); err != nil {
		return
	}
	// Сервер ВЕБ или REST соединений.
	if err = sio.sWeb.
		Start(); err != nil {
		return
	}
	// Сервер TCP/IP соединений.
	if err = sio.sTcp.
		Start(); err != nil {
		return
	}
	// Сервер UDP соединений.
	if err = sio.sUdp.
		Start(); err != nil {
		return
	}

	return
}

// Stop Остановка всех зарегистрированных серверов.
// Если не зарегистрирован ни один сервер, функция возвращает ошибку.
func (sio *impl[T]) Stop() (err error) {
	if err = sio.sUdp.
		Stop(); err != nil {
		return
	}
	if err = sio.sTcp.
		Stop(); err != nil {
		return
	}
	if err = sio.sWeb.
		Stop(); err != nil {
		return
	}
	if err = sio.sGrpc.
		Stop(); err != nil {
		return
	}

	return
}

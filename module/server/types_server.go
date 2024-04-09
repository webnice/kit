package server

import (
	"net"

	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// Вспомогательный тип данных мастер-объекта.
type master struct{}

// Объект процесса веб сервера обслуживающего один сокет.
type server[
	T master | kitTypesServer.Web | kitTypesServer.Grpc | kitTypesServer.Tcp | kitTypesServer.Udp,
] struct {
	p          *impl[master]          // Адрес объекта родительской сущности (parent), интерфейс Interface.
	r          []*T                   // Ресурсы сервера.
	s          *kitTypesServer.Server // Конфигурация сервера.
	t          interface{}            // Тип сервера *T, для более удобного определения типа через switch (type).
	l          net.Listener           // Слушатель соединений.
	isStarted  bool                   // Состояние выполнения сервера.
	isShutdown bool                   // Флаг нормального завершения работы.
	err        error                  // Последняя ошибка.
}

// Объект сущности, реализующий интерфейс IServer.
type implIServer[
	T kitTypesServer.Web | kitTypesServer.Grpc | kitTypesServer.Tcp | kitTypesServer.Udp,
] struct {
	p       *impl[master] // Адрес объекта родительской сущности (parent).
	r       []*T          // Зарегистрированные ресурсы веб сервера.
	servers []*server[T]  // Серверы.
}

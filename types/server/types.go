package server

// IResource Интерфейс регистрации контроллера ресурса сервера.
type IResource[T Web | Grpc | Tcp | Udp] interface {
	// Resource Описание ресурса.
	Resource() (ret *T)
}

// IServer Интерфейс регистрации контроллеров сервера.
type IServer[T Web | Grpc | Tcp | Udp] interface {
	// ResourceCount Количество зарегистрированных контроллеров ресурсов сервера.
	ResourceCount() (ret int)

	// ResourceRegistration Регистрация контроллеров ресурсов сервера.
	ResourceRegistration(s IResource[T])

	// ServerGetById Загрузка конфигурации сервера по идентификатору сервера.
	ServerGetById(id string) (ret *Server, err error)

	// Start Запуск сервера с указанным идентификатором или запуск всех серверов, если идентификаторы не переданы.
	// Если переданы не корректные идентификаторы, будет возвращена ошибка.
	Start(IDs ...string) (err error)

	// Stop Остановка сервера с указанным идентификатором или остановка всех серверов, если идентификаторы не переданы.
	// Если переданы не корректные идентификаторы, будет возвращена ошибка.
	Stop(IDs ...string) (err error)
}

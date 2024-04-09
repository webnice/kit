package server

import (
	kitModuleUuid "github.com/webnice/kit/v4/module/uuid"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// ResourceCount Количество зарегистрированных контроллеров ресурсов сервера.
func (iis *implIServer[T]) ResourceCount() (ret int) { return len(iis.r) }

// ResourceRegistration Регистрация контроллеров ресурсов сервера.
func (iis *implIServer[T]) ResourceRegistration(s kitTypesServer.IResource[T]) {
	var ifc T

	defer func() { _ = recover() }()
	ifc = interface{}(*s.Resource()).(T)
	iis.r = append(iis.r, &ifc)
}

// ServerGetById Загрузка конфигурации сервера по идентификатору сервера.
func (iis *implIServer[T]) ServerGetById(id string) (ret *kitTypesServer.Server, err error) {
	var servers []*server[master]

	if servers, err = iis.p.
		serverGetByTypeById(new(T), id); err != nil {
		return
	}
	if len(servers) <= 0 || id == "" {
		err = iis.p.Errors().ServerWithUuidNotFound(id)
		return
	}
	ret = servers[0].s

	return
}

// Start Запуск сервера с указанным идентификатором или запуск всех серверов, если идентификаторы не переданы.
// Если переданы не корректные идентификаторы, будет возвращена ошибка.
func (iis *implIServer[T]) Start(IDs ...string) (err error) {
	var (
		servers []*server[master]
		srv     *server[T]
		n, r    int
	)

	// Загрузка среза конфигураций серверов текущего типа.
	if servers, err = iis.p.
		serverGetByTypeById(new(T), IDs...); err != nil {
		return
	}
	// Копирование среза конфигураций серверов в тип процесса сервера.
	iis.servers = make([]*server[T], 0, len(servers))
	for n = range servers {
		srv = &server[T]{
			p: servers[n].p,
			r: make([]*T, 0, len(servers[n].r)),
			s: servers[n].s,
			t: new(T),
		}
		iis.servers = append(iis.servers, srv)
	}
	servers = servers[:0]
	// Обход всех конфигураций серверов, наполнение данными и запуск потоков.
	for n = range iis.servers {
		// Добавление ресурсов.
		for r = range iis.r {
			// Конвертирование ресурсов "master" в "T".
			resource := interface{}(*iis.r[r]).(T)
			iis.servers[n].r = append(iis.servers[n].r, &resource)
		}
		// Запуск сервера с ожиданием старта потока.
		if err = goWaitError(iis.servers[n].do); err != nil {
			return
		}
	}

	return
}

// Stop Остановка сервера с указанным идентификатором или остановка всех серверов, если идентификаторы не переданы.
// Если переданы не корректные идентификаторы, будет возвращена ошибка.
func (iis *implIServer[T]) Stop(IDs ...string) (err error) {
	var n int

	for n = range iis.servers {
		if !iis.servers[n].isStarted {
			continue
		}
		if err = iis.servers[n].
			doStop(); err != nil {
			return
		}
	}

	return
}

// Функция возвращает ошибку, если сервер с указанным UUID выполняется.
func (iis *implIServer[T]) isRun(uus []kitModuleUuid.UUID) (err error) {
	var (
		n, s int
		ids  []string
	)

	ids = make([]string, 0, len(uus))
	for n = range uus {
		ids = append(ids, uus[n].String())
	}
	// Проверка выполняются ли процессы серверов.
	for n = range uus {
		for s = range iis.servers {
			if iis.servers[s].s.ID.Equal(uus[n]) && iis.servers[s].IsStarted() {
				err = iis.p.Errors().ServerIsStarted(ids...)
				return
			}
		}
	}

	return
}

package server

import (
	"github.com/webnice/debug"
	kitModuleUuid "github.com/webnice/kit/v4/module/uuid"
)

// Поиск зарегистрированных серверов по UUID, если список IDs пустой, возврат всех серверов.
func (sri *impl) findServersByUuid(IDs ...string) (ret []*server, err error) {
	var (
		uuid kitModuleUuid.UUID
		n, s int
		srv  *server
	)

	ret = make([]*server, 0, len(sri.servers))
	if len(IDs) == 0 {
		ret = append(ret, sri.servers...)
		return
	}
	for n = range IDs {
		if uuid = kitModuleUuid.Get().FromString(IDs[n]); uuid.Equal(kitModuleUuid.NULL) {
			err = sri.Errors().UuidError(IDs[n])
			return
		}
		for s = range sri.servers {
			if sri.servers[s].ID.Equal(uuid) {
				srv = sri.servers[s]
				break
			}
		}
		if srv == nil {
			err = sri.Errors().ServerWithUuidNotFound(IDs[n])
			return
		}
		ret = append(ret, srv)
	}

	return
}

// Start Запуск WEB сервера.
// Если не переданы идентификаторы, запускаются все добавленные серверы.
// Если идентификаторы переданы, тогда будут запущены только сервера с указанными идентификаторами.
// Если переданы не корректные идентификаторы, будет возвращена ошибка.
func (sri *impl) Start(IDs ...string) (err error) {
	var start []*server

	if start, err = sri.findServersByUuid(IDs...); err != nil {
		return
	}

	sri.log().Trace(debug.DumperString(IDs, start))

	return
}

// Stop Остановка сервера.
// Если не переданы идентификаторы, останавливаются все серверы.
// Если идентификаторы переданы, тогда будут остановлены только серверы с указанными идентификаторами.
// Если переданы не корректные идентификаторы, будет возвращена ошибка.
func (sri *impl) Stop(IDs ...string) (err error) {
	var stop []*server

	if stop, err = sri.findServersByUuid(IDs...); err != nil {
		return
	}

	sri.log().Trace(debug.DumperString(IDs, stop))

	return
}

package server

import (
	"reflect"

	kitModuleUuid "github.com/webnice/kit/v4/module/uuid"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// Создание объекта и возвращение интерфейса Essence.
func newEssence(parent *impl[master]) *gist[master] {
	var ece = &gist[master]{p: parent}

	return ece
}

// Debug Присвоение нового значения режима отладки.
func (ece *gist[T]) Debug(debug bool) Essence { ece.p.debug = debug; return ece }

// Удаление конфигурации сервера. В качестве идентификатора, передаётся UUID сервера,
// полученный при добавлении конфигурации (Свойство ID структуры сервера).
// Нельзя удалять запущенный сервер, функция вернёт ошибку.
func (ece *gist[T]) deleteById(t kitTypesServer.Type, IDs ...string) (err error) {
	var (
		uus     []kitModuleUuid.UUID
		n, s    int
		found   bool
		servers []*server[master]
	)

	// Парсинг UUID идентификаторов.
	if uus, err = ece.p.parseIDsToObject(IDs...); err != nil {
		return
	}
	// Проверка выполняются ли процессы серверов с указанными UUID идентификаторами.
	switch t {
	case kitTypesServer.TWeb:
		err = ece.p.sWeb.isRun(uus)
	case kitTypesServer.TGrpc:
		err = ece.p.sGrpc.isRun(uus)
	case kitTypesServer.TTcp:
		err = ece.p.sTcp.isRun(uus)
	default:
		err = ece.p.Errors().TypeNotImplemented(reflect.TypeOf(t).String())
	}
	if err != nil {
		return
	}
	ece.p.serverLock.Lock()
	defer ece.p.serverLock.Unlock()
	// Проверка наличия конфигураций серверов для удаления.
	for n = range uus {
		found = false
		for s = range ece.p.server[t] {
			if ece.p.server[t][s].s.ID.Equal(uus[n]) {
				found = true
				break
			}
		}
		if !found {
			err = ece.p.Errors().ServerWithUuidNotFound(IDs[n])
			return
		}
	}
	// Удаление конфигураций.
	servers = make([]*server[master], 0, len(ece.p.server[t]))
	for s = range ece.p.server[t] {
		found = false
		for n = range uus {
			if ece.p.server[t][s].s.ID.Equal(uus[n]) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		servers = append(servers, ece.p.server[t][s])
	}
	// Обновление среза.
	ece.p.server[t] = make([]*server[master], 0, len(servers))
	ece.p.server[t] = append(ece.p.server[t], servers...)

	return
}

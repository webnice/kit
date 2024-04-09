package server

import (
	kitModuleUuid "github.com/webnice/kit/v4/module/uuid"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// WebAdd Добавление конфигурации веб сервера. Функцию можно вызывать многократно, для добавления
// нескольких веб серверов. Серверы не должны пересекаться по занимаемому IP и порту или другим монопольно
// выделяемым ресурсам.
// Возвращается объект добавленного веб сервера.
func (ece *gist[T]) WebAdd(cfg *kitTypesServer.WebConfiguration) (ret *kitTypesServer.Server) {
	var ok bool

	ret = &kitTypesServer.Server{
		ID:   kitModuleUuid.Get().V4(),
		Type: kitTypesServer.TWeb,
		Web:  cfg,
	}
	ece.p.serverLock.Lock()
	if _, ok = ece.p.server[kitTypesServer.TWeb]; !ok {
		ece.p.server[kitTypesServer.TWeb] = make([]*server[master], 0, 1)
	}
	ece.p.server[kitTypesServer.TWeb] = append(
		ece.p.server[kitTypesServer.TWeb],
		&server[master]{
			p: ece.p,
			s: ret,
		},
	)
	ece.p.serverLock.Unlock()

	return
}

// WebDel Удаление конфигурации сервера. В качестве идентификатора, передаётся UUID сервера,
// полученный при добавлении конфигурации (Свойство ID структуры сервера).
// Нельзя удалять запущенный сервер, функция вернёт ошибку.
func (ece *gist[T]) WebDel(IDs ...string) (err error) {
	switch interface{}(new(T)).(type) {
	case kitTypesServer.Web:
		err = ece.deleteById(kitTypesServer.TWeb, IDs...)
	case kitTypesServer.Grpc:
		err = ece.deleteById(kitTypesServer.TGrpc, IDs...)
	case kitTypesServer.Tcp:
		err = ece.deleteById(kitTypesServer.TTcp, IDs...)
	default:
		return
	}

	return
}

package server

import (
	kitModuleUuid "github.com/webnice/kit/v4/module/uuid"
	kitTypesServer "github.com/webnice/kit/v4/types/server"
)

// Создание объекта и возвращение интерфейса Essence.
func newEssence(parent *impl) *gist {
	var essence = &gist{parent: parent}
	return essence
}

// Debug Присвоение нового значения режима отладки.
func (essence *gist) Debug(debug bool) Essence { essence.parent.debug = debug; return essence }

// ConfigurationWeb Добавление конфигурации веб сервера. Функцию можно вызывать многократно, для добавления
// нескольких веб серверов. Серверы не должны пересекаться по занимаемому IP и порту или другим монопольно
// занимаемым ресурсам.
// Возвращается UUID идентификатор добавленного веб сервера.
func (essence *gist) ConfigurationWeb(cfg *kitTypesServer.WebServerConfiguration) (ret string) {
	var (
		item *server
		uuid kitModuleUuid.Interface
	)

	uuid = kitModuleUuid.Get()
	item = &server{
		ID:   uuid.V4(),
		Cfg:  cfg,
		Type: serverWeb,
	}
	essence.parent.servers = append(essence.parent.servers, item)
	ret = item.ID.String()

	return
}

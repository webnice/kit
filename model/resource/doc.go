package resource

/*
Для создания модели с ресурсами, необходимо создать пакет с синонимом интерфейса и тип со встраиванием.
Пример пакета:

//НАЧАЛО ПРИМЕРА-->
package resource

import kitModelResource "github.com/webnice/kit/v4/model/resource"

// Генератор кода встраиваемых ресурсов. Аргументы:
// --package - Название пакета для генерации .go файлов.
// --path    - Путь создания .go файлов с ресурсами. Если пусто, файлы создаются в текущей папке.
// Так же ожидаются переменные окружения:
// EMBEDDER_STATIC_BASE_DIR  - Корневая директория статических ресурсов.
// EMBEDDER_STATIC_RESOURCES - Группы ресурсов и папками самих ресурсов.
//                             Формат: group_name1:path/to/folder1,group_name2:path/to/folder2
//go:generate go run "github.com/webnice/kit/v4/model/resource/embedder" --package "resource" --path "."

var singleton *impl

// Объект сущности пакета.
type impl struct {
	kitModelResource.Implementation
}

func init() { singleton = &impl{Implementation: *kitModelResource.Constructor()} }

// Get Получение объекта сущности пакета, возвращается интерфейс пакета.
//
//goland:noinspection GoUnusedExportedFunction
func Get() kitModelResource.Interface { return singleton }

//<--ОКОНЧАНИЕ ПРИМЕРА.
*/

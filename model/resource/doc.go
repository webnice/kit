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
//go:generate go run "github.com/webnice/kit/v4/model/resource/embedder/embedder.go" --package "resource" --path "."

// Создаются .go файлы по шаблону имени файла:
// resource_content_{{ groupName }}_{{ resourceNumber }}.go
// Где:
// groupName      - строка приведённая к нижнему регистру не содержащая пробелы.
// resourceNumber - порядковый номер ресурса в формате %020d.
// Пример: resource_content_test_group_00000000000000000001.go

var singleton *impl

// Объект сущности пакета.
type impl struct {
	kitModelResource.Implementation
}

func init() { singleton = new(impl) }

// Get Получение объекта сущности пакета, возвращается интерфейс пакета.
//
//goland:noinspection GoUnusedExportedFunction
func Get() kitModelResource.Interface { return singleton }

//<--ОКОНЧАНИЕ ПРИМЕРА.
*/

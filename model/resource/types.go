package resource

import (
	"bytes"
	"sync"
	"time"
)

// Interface Интерфейс пакета.
type Interface interface {
	// Add Добавление ресурса в группу ресурсов.
	Add(group string, name string, resource Resource) error

	// Group Возвращается список групп ресурсов.
	Group() (ret []string)

	// ResourceByGroup Возвращает список ресурсов в указанной группе.
	ResourceByGroup(group string) (ret []string)

	// ResourceData Получение ресурса по имени группы и ресурсу. Если ресурса нет, возвращается nil.
	ResourceData(group string, resource string) (ret *Resource)

	// ResourceByGroupTarReader В памяти создаётся tar контейнер со всеми ресурсами группы и
	// возвращается *bytes.Reader к tar контейнеру.
	ResourceByGroupTarReader(group string) (ret *bytes.Reader, err error)
}

// Implementation Встраиваемая структура в модель ресурсов.
type Implementation struct {
	Res     map[string]map[string]*Resource // Карта ресурсов, map[название группы]map[название ресурса]*resource
	ResLock sync.RWMutex                    // Защита карты от конкурентного доступа на запись.
}

// Resource Описание ресурса.
type Resource struct {
	Size        uint64    // Размер ресурса в байтах.
	Time        time.Time // Дата и время создания ресурса.
	ContentType string    // Определённый по расширению имени файла тип контента ресурса.
	Content     []byte    // Контент ресурса.
}

// Package cfg
package cfg

import (
	"time"

	kitTypes "github.com/webnice/kit/types"
)

// Recorder Интерфейс логирования
type Recorder kitTypes.Logger

// Структура объекта сущности логирования, интерфейс Logger
type recorder struct {
	parent *impl // Адрес объекта основной сущности, интерфейс Interface
}

// Структура объекта переноса данных
type record struct {
	recorder  *recorder              // Родительский объект
	timestamp time.Time              // Время записи в UTC
	traceInfo *kitTypes.TraceInfo    // Данные о вызывающей функции и стеке
	stackBack int                    // Дополнительный откат записей стека
	keys      map[string]interface{} // Ключи логирования
	fatality  *bool                  // Флаг фатальности записи лога
}

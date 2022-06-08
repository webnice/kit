// Package pdw
package pdw

/*

Бассейн обёрток передачи данных.

*/

import (
	"context"
	"reflect"
	"sync"
)

// Interface Интерфейс пакета.
type Interface interface {
	// Get Получение объекта из бассейна.
	Get() Data

	// Put Возвращение объекта в бассейн.
	Put(wdo Data)
}

// Data Интерфейс обёртки данных.
type Data interface {
	// DataPut Загрузка в объект данных, флага и контекста.
	DataPut(data interface{}, isSync bool, ctx context.Context)

	// DataGet Возвращение оборачиваемых данных.
	DataGet() interface{}

	// Type Тип обёрнутых данных.
	Type() reflect.Type

	// IsSync Флаг режима передачи данных.
	// - истина - данные передаются в синхронном режиме, издатель ожидает ответ.
	// - ложь   - данные передаются в асинхронном режиме.
	IsSync() bool

	// DoneSet Установка флага "обработка данных завершена".
	DoneSet()

	// Done Канал сигнала окончания обработки данных и агрегации результата.
	Done() <-chan struct{}

	// Result Возвращение результатов обработки данных.
	Result() Result

	// Context Контекст.
	Context() context.Context

	// Reset Очистка всех данных объекта.
	Reset()
}

// Result Интерфейс результата обработки данных.
type Result interface {
	// DataPut Добавление данных с результатом.
	DataPut(d ...interface{})

	// DataGet Возвращение данных с результатом.
	DataGet() []interface{}

	// ErrPut Добавление ошибок в результат.
	ErrPut(e ...error)

	// ErrGet Возвращение ошибок результата.
	ErrGet(e ...error) []error
}

// Объект сущности, реализующий интерфейс Interface, бассейн объектов.
type impl struct {
	sync.Pool
}

// Обёртка над передаваемыми данными.
type data struct {
	data     interface{}     // Данные.
	dataType reflect.Type    // Тип данных.
	sync     bool            // Флаг синхронной передачи.
	done     chan struct{}   // Канал обратной связи для синхронного вызова.
	result   *result         // Результат обработки данных.
	ctx      context.Context // Контекст.
}

// Обёртка над результатом передачи данных потребителю.
type result struct {
	data []interface{}
	errs []error
}

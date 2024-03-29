// Package bus
package bus

import (
	"context"
	"time"

	kitModulePdw "github.com/webnice/kit/v3/module/pdw"
	kitTypes "github.com/webnice/kit/v3/types"
)

// Interface Интерфейс объекта сущности.
type Interface interface {
	// Subscribe Регистрация потребителя данных.
	// Вернётся ошибка, если:
	// - потребитель данных является nil.
	// - функция регистрации типов данных вернула недопустимые значения.
	Subscribe(databuser kitTypes.Databuser) (err error)

	// Unsubscribe Удаление потребителя данных.
	// Вернётся ошибка, если:
	// - потребитель данных является nil.
	// - потребитель данных не регистрировался или подписка потребителя была уже удалена.
	Unsubscribe(databuser kitTypes.Databuser) (err error)

	// PublishSync Передача в шину данных объекта данных в синхронном режиме, функция блокируется до окончания передачи
	// данных всем зарегистрированным потребителям, подписанным на получение передаваемого типа данных.
	// Функция вернёт ошибку, если:
	// - тип переданных данных не зарегистрирован ни одним потребителем данных, то есть некому передать данные.
	// - тип данных является пустым интерфейсом или nil.
	// - ошибку вернул потребитель данных.
	PublishSync(data interface{}) (ret []interface{}, errs []error)

	// PublishSyncWithContext Передача в шину данных объекта данных в синхронном режиме с контекстом,
	// функция блокируется до окончания передачи данных всем зарегистрированным потребителям, подписанным на получение
	// передаваемого типа данных.
	// Прервать ожидание ответа можно через контекст.
	// Функция вернёт ошибку, если:
	// - тип переданных данных не зарегистрирован ни одним потребителем данных, то есть некому передать данные.
	// - тип данных является пустым интерфейсом или nil.
	// - ошибку вернул потребитель данных.
	// - произошло прерывание ожидания ответа через контекст.
	PublishSyncWithContext(ctx context.Context, data interface{}) (ret []interface{}, errs []error)

	// PublishSyncWithTimeout Передача в шину данных объекта данных в синхронном режиме с таймаутом,
	// функция блокируется до окончания передачи данных всем зарегистрированным потребителям, подписанным на получение
	// передаваемого типа данных.
	// Ожидание автоматически прервётся через время указанное в timeout.
	// Функция вернёт ошибку, если:
	// - тип переданных данных не зарегистрирован ни одним потребителем данных, то есть некому передать данные.
	// - тип данных является пустым интерфейсом или nil.
	// - ошибку вернул потребитель данных.
	// - произошло прерывание ожидания ответа по таймауту.
	PublishSyncWithTimeout(timeout time.Duration, data interface{}) (ret []interface{}, errs []error)

	// PublishAsync Передача в шину данных объекта данных в асинхронном режиме.
	// Функция вернёт ошибку, если:
	// - тип переданных данных не зарегистрирован ни одним потребителем данных, то есть некому передать данные.
	// - тип данных является пустым интерфейсом или nil.
	PublishAsync(data interface{}) (err error)

	// Gist Интерфейс к публичным служебным методам.
	Gist() Essence

	// ОШИБКИ

	// Errors Все ошибки известного состояния, которые может вернуть приложение или функция.
	Errors() *Error
}

// Essence Служебный публичный интерфейс.
type Essence interface {
	// Debug Присвоение нового значения режима отладки.
	Debug(debug bool) Essence

	// WorkerStart Запуск обработчика шины данных.
	WorkerStart(workerCount int) Essence

	// WorkerStop Остановка обработчика шины данных с подтверждением остановки.
	// Функция блокируется до подтверждения завершения потока обработчика.
	WorkerStop() Essence

	// Statistic Статистика работы бассейна шины данных.
	// Статистика ведётся только если шина данных создана с флагом отладки New(..., isDebug=true).
	// Если шина данных создана без флага отладки, статистика вернёт nil.
	Statistic() *kitModulePdw.Statistic
}

// Package bus
package bus

import (
	"context"
	"reflect"
	"runtime"
	"time"

	kitModulePdw "github.com/webnice/kit/module/pdw"
	kitModuleTrace "github.com/webnice/kit/module/trace"
	kitTypes "github.com/webnice/kit/types"
)

// New Конструктор объекта пакета, возвращает интерфейс Interface.
// bufferLength - Размер буфера шины данных, если указано 0 или отрицательное число, используется размер по умолчанию,
//                равный 1000000 сообщений.
// workerCount  - Количество запускаемых обработчиков сообщений шины данных. Если указано 0 или отрицательное число,
//                используется значение по умолчанию, равное 1000 обработчиков.
// isDebug      - Флаг отладки.
func New(bufferLength int, workerCount int, isDebug bool) Interface {
	var (
		data *databus
		bus  *impl
	)

	if bufferLength <= 0 {
		bufferLength = defaultDatabusBufferLength
	}
	bus = &impl{
		debug: isDebug,
	}
	data = &databus{
		Wrappers:    kitModulePdw.New(bus.debug),
		Subscribers: newSubscribers(),
		Bus:         make(chan kitModulePdw.Data, bufferLength),
	}
	bus.databus = data
	runtime.SetFinalizer(bus, destructorBus)
	bus.essence = newEssence(bus).WorkerStart(workerCount)

	return bus
}

// Деструктор объекта *impl.
func destructorBus(bus *impl) {
	defer func() { _ = recover() }()
	bus.essence = nil
	destructorSafeCloseDatabusBus(bus.databus.Bus)
	safeCloseSignalChannel(bus.workerOnDone)
}

func destructorSafeCloseDatabusBus(c chan kitModulePdw.Data) {
	defer func() { _ = recover() }()
	close(c)
}

// Gist Интерфейс к публичным служебным методам.
func (bus *impl) Gist() Essence { return bus.essence }

// Errors Все ошибки известного состояния, которые может вернуть приложение или функция.
func (bus *impl) Errors() *Error { return Errors() }

// Subscribe Регистрация потребителя данных.
// Вернётся ошибка, если:
// - потребитель данных является nil.
// - функция регистрации типов данных вернула недопустимые значения.
func (bus *impl) Subscribe(databuser kitTypes.Databuser) (err error) {
	var (
		types []interface{}
		rt    reflect.Type
		n     int
		sti   *subscriberTypeInfo
		item  *subscriber
	)

	defer func() {
		if e := recover(); e != nil {
			err = bus.Errors().DatabusPanicException(0, e, kitModuleTrace.StackShort())
		}
	}()
	if databuser == nil {
		err = bus.Errors().DatabusObjectIsNil(0)
		return
	}
	types = databuser.KnownType()
	item = &subscriber{
		Name:  getFuncFullName(databuser),
		Item:  databuser,
		Types: make([]*subscriberTypeInfo, 0, len(types)),
	}
	for n = range types {
		rt = reflect.TypeOf(types[n])
		if sti, err = makeSubscriberTypeInfo(rt); err != nil {
			return
		}
		sti.Original = types[n]
		item.Types = append(item.Types, sti)
	}
	bus.databus.Subscribers.Store(item)

	return
}

// Unsubscribe Удаление потребителя данных.
// Вернётся ошибка, если:
// - потребитель данных является nil.
// - потребитель данных не регистрировался или подписка потребителя была уже удалена.
func (bus *impl) Unsubscribe(databuser kitTypes.Databuser) (err error) {
	var databuserName string

	defer func() {
		if e := recover(); e != nil {
			err = bus.Errors().DatabusPanicException(0, e, kitModuleTrace.StackShort())
		}
	}()
	if databuser == nil {
		err = bus.Errors().DatabusObjectIsNil(0)
		return
	}
	databuserName = getFuncFullName(databuser)
	err = bus.databus.Subscribers.Delete(databuserName)

	return
}

// PublishSync Передача в шину данных объекта данных в синхронном режиме, функция блокируется до окончания передачи
// данных всем зарегистрированным потребителям, подписанным на получение передаваемого типа данных.
// Функция вернёт ошибку, если:
// - тип переданных данных не зарегистрирован ни одним потребителем данных, то есть некому передать данные.
// - тип данных является пустым интерфейсом или nil.
// - ошибку вернул потребитель данных.
func (bus *impl) PublishSync(data interface{}) (ret []interface{}, errs []error) {
	return bus.publishSync(context.Background(), data)
}

// PublishSyncWithContext Передача в шину данных объекта данных в синхронном режиме с контекстом,
// функция блокируется до окончания передачи данных всем зарегистрированным потребителям, подписанным на получение
// передаваемого типа данных.
// Прервать ожидание ответа можно через контекст.
// Функция вернёт ошибку, если:
// - тип переданных данных не зарегистрирован ни одним потребителем данных, то есть некому передать данные.
// - тип данных является пустым интерфейсом или nil.
// - ошибку вернул потребитель данных.
// - произошло прерывание ожидания ответа через контекст.
func (bus *impl) PublishSyncWithContext(ctx context.Context, data interface{}) (ret []interface{}, errs []error) {
	return bus.publishSync(ctx, data)
}

// PublishSyncWithTimeout Передача в шину данных объекта данных в синхронном режиме с таймаутом,
// функция блокируется до окончания передачи данных всем зарегистрированным потребителям, подписанным на получение
// передаваемого типа данных.
// Ожидание автоматически прервётся через время указанное в timeout.
// Функция вернёт ошибку, если:
// - тип переданных данных не зарегистрирован ни одним потребителем данных, то есть некому передать данные.
// - тип данных является пустым интерфейсом или nil.
// - ошибку вернул потребитель данных.
// - произошло прерывание ожидания ответа по таймауту.
func (bus *impl) PublishSyncWithTimeout(timeout time.Duration, data interface{}) (ret []interface{}, errs []error) {
	var (
		ctx context.Context
		cfn context.CancelFunc
	)
	ctx, cfn = context.WithTimeout(context.Background(), timeout)
	defer cfn()
	return bus.publishSync(ctx, data)
}

func (bus *impl) publishSync(ctx context.Context, data interface{}) (ret []interface{}, errs []error) {
	var (
		err error
		wdi kitModulePdw.Data
		n   int
	)

	if err = bus.publishDataCheck(data); err != nil {
		errs = append(errs, err)
		return
	}
	if wdi = bus.databus.Wrappers.Get(); wdi == nil {
		err = bus.Errors().DatabusPoolInternalError(0)
		errs = append(errs, err)
		return
	}
	wdi.DataPut(data, true, ctx)
	bus.publish(wdi)
	if n = len(wdi.Result().DataGet()); n > 0 {
		ret = make([]interface{}, 0, n)
		ret = append(ret, wdi.Result().DataGet()...)
	}
	if n = len(wdi.Result().ErrGet()); n > 0 {
		errs = make([]error, 0, n)
		errs = append(errs, wdi.Result().ErrGet()...)
	}
	// Возвращение объртки обратно в бассейн.
	bus.databus.Wrappers.Put(wdi)

	return
}

// PublishAsync Передача в шину данных объекта данных в асинхронном режиме.
// Функция вернёт ошибку, если:
// - тип переданных данных не зарегистрирован ни одним потребителем данных, то есть некому передать данные.
// - тип данных является пустым интерфейсом или nil.
func (bus *impl) PublishAsync(data interface{}) (err error) {
	var wdi kitModulePdw.Data

	if err = bus.publishDataCheck(data); err != nil {
		return
	}
	if wdi = bus.databus.Wrappers.Get(); wdi == nil {
		err = bus.Errors().DatabusPoolInternalError(0)
		return
	}
	wdi.DataPut(data, false, nil)
	bus.publish(wdi)

	return
}

// Проверка передаваемых данных.
func (bus *impl) publishDataCheck(data interface{}) (err error) {
	var rdt reflect.Type

	if data == nil {
		err = bus.Errors().DatabusObjectIsNil(0)
		return
	}
	if rdt = reflect.TypeOf(data); !bus.databus.Subscribers.IsExistSubscriber(rdt) {
		err = bus.Errors().DatabusNotSubscribersForType(0, rdt.String())
		return
	}

	return
}

func (bus *impl) publish(wdi kitModulePdw.Data) {
	// Передача данных в шину данных.
	bus.databus.Bus <- wdi
	// Для синхронных данных, ожидание готовности результата.
	if wdi.IsSync() {
		<-wdi.Done()
	}

	return
}

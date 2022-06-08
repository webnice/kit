// Package pdw
package pdw

import (
	"context"
	"runtime"
)

// New Конструктор бассейна с объектами для обёртки данных.
func New() Interface {
	var pdw = new(impl)

	pdw.New = constructorWrapperData

	return pdw
}

// Конструктор объектов пула.
func constructorWrapperData() interface{} {
	var wdo = &data{
		data:     nil,
		sync:     false,
		dataType: nil,
		done:     make(chan struct{}),
		result:   new(result),
		ctx:      context.Background(),
	}

	runtime.SetFinalizer(wdo, destructorWrapperData)

	return wdo
}

// Деструктор объектов бассейна.
func destructorWrapperData(wdo *data) {
	safeCloseSignalChannel(wdo.done)
}

// Закрытие канала с защитой от паники (закрытие закрытого канала).
func safeCloseSignalChannel(c chan struct{}) {
	defer func() { _ = recover() }()
	close(c)
}

// Get Получение объекта из бассейна.
func (wdp *impl) Get() (ret Data) {
	var (
		wdo *data
		ok  bool
	)

	if wdo, ok = wdp.Pool.Get().(*data); ok {
		ret = wdo
	}

	return
}

// Put Возвращение объекта в бассейн.
func (wdp *impl) Put(wdi Data) {
	var wdo = wdi.(*data)

	wdo.Reset()
	wdp.Pool.Put(wdo)
}

package app

import (
	"context"

	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	kitTypes "github.com/webnice/kit/v4/types"
)

// Функция вызова функции Initiate() у компоненты с контролем длительности выполнения и прерыванием по таймауту
func (app *impl) initiateFn(component *kitTypes.ComponentInfo) (err kitTypes.ErrorWithCode) {
	var (
		ctx  context.Context
		ctf  context.CancelFunc
		call chan kitTypes.ErrorWithCode
	)

	// Функция защиты от паники
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentPanicException(0, component.ComponentName, e, kitModuleTrace.StackShort())
		}
	}()
	// Создание контекста контроля таймаута
	ctx, ctf = context.WithTimeout(context.Background(), component.InitiateTimeout)
	defer ctf()
	// Запуск функции Initiate() у компоненты с защитой от паники
	call = app.initiateCallFn(component.ComponentName, component.Component)
	defer func() { close(call) }()
	// Ожидание, либо таймаута, либо завершения функции Initiate()
	select {
	case <-ctx.Done():
		err = app.cfg.Errors().
			ComponentInitiateTimeout(0, component.ComponentName)
	case err = <-call:
		component.IsInitiate = err == nil
	}

	return
}

// Запуск горутины с каналом обратной связи для получения ошибки из вызываемой функции Initiate()
func (app *impl) initiateCallFn(componentName string, cpt kitTypes.Component) (ret chan kitTypes.ErrorWithCode) {
	ret = make(chan kitTypes.ErrorWithCode)
	go func() { ret <- app.initiateSafeCall(componentName, cpt) }()

	return
}

// Запуск функции Initiate() в компоненте с защитой от паники
func (app *impl) initiateSafeCall(componentName string, cpt kitTypes.Component) (err kitTypes.ErrorWithCode) {
	var e error

	// Функция защиты от паники
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentInitiatePanicException(0, componentName, e, kitModuleTrace.StackShort())
		}
	}()
	if e = cpt.Initiate(); e != nil {
		switch eto := e.(type) {
		case kitTypes.ErrorWithCode:
			err = eto
		default:
			err = app.cfg.Errors().ComponentInitiateExecution(0, componentName, eto)
		}
	}

	return
}

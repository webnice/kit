package app

import (
	"context"

	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	kitTypes "github.com/webnice/kit/v4/types"
)

// Функция вызова функции Initiate() у компоненты с контролем длительности выполнения и прерыванием по таймауту.
func (app *impl) initiateFn(component *kitTypes.ComponentInfo) (err error) {
	var (
		ctx  context.Context
		ctf  context.CancelFunc
		call chan error
	)

	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentPanicException.
				Bind(component.ComponentName, e, kitModuleTrace.StackShort())
		}
	}()
	// Создание контекста контроля таймаута.
	ctx, ctf = context.WithTimeout(context.Background(), component.InitiateTimeout)
	defer ctf()
	// Запуск функции Initiate() у компоненты с защитой от паники.
	call = app.initiateCallFn(component.ComponentName, component.Component)
	defer func() { close(call) }()
	// Ожидание, либо таймаута, либо завершения функции Initiate().
	select {
	case <-ctx.Done():
		err = app.cfg.Errors().ComponentInitiateTimeout.
			Bind(component.ComponentName)
	case err = <-call:
		component.IsInitiate = err == nil
	}

	return
}

// Запуск горутины с каналом обратной связи для получения ошибки из вызываемой функции Initiate().
func (app *impl) initiateCallFn(componentName string, cpt kitTypes.Component) (ret chan error) {
	ret = make(chan error)
	go func() { ret <- app.initiateSafeCall(componentName, cpt) }()

	return
}

// Запуск функции Initiate() в компоненте с защитой от паники.
func (app *impl) initiateSafeCall(componentName string, cpt kitTypes.Component) (err error) {
	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentInitiatePanicException.Bind(componentName, e, kitModuleTrace.StackShort())
		}
	}()
	if err = cpt.Initiate(); err != nil {
		if app.cfg.Errors().Unbind(err) == nil {
			err = app.cfg.Errors().ComponentInitiateExecution.Bind(componentName, err)
		}
	}

	return
}

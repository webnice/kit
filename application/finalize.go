// Package application
package application

import (
	"context"
	"time"

	kitModuleTrace "github.com/webnice/kit/module/trace"
	kitTypes "github.com/webnice/kit/types"
)

// Функция вызова функции Finalize() у компоненты
func (app *impl) finalizeFn(component *kitTypes.ComponentInfo) (err kitTypes.ErrorWithCode) {
	var (
		ctx context.Context
		cfn context.CancelFunc
	)

	// Создание контекста
	ctx, cfn = context.WithCancel(context.Background())
	defer cfn()
	// Запуск вспомогательной горутины, которая по таймауту выведет в лог сообщение о долгой работе функции Finalize()
	go func(cx context.Context, name string, tout time.Duration) {
		select {
		case <-cx.Done():
			return
		case <-time.After(tout):
			app.cfg.Log().Warning(app.cfg.Errors().ComponentFinalizeWarning(0, name, tout).Error())
		}
	}(ctx, component.ComponentName, app.cfg.Gist().ComponentFinalizeWarningTimeout())
	err = app.finalizeSafeCall(component.ComponentName, component.Component)

	return
}

// Запуск функции Finalize() в компоненте с защитой от паники
func (app *impl) finalizeSafeCall(componentName string, cpt kitTypes.Component) (err kitTypes.ErrorWithCode) {
	var e error

	// Функция защиты от паники
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentFinalizePanicException(0, componentName, e, kitModuleTrace.StackShort())
		}
	}()
	if e = cpt.Finalize(); e != nil {
		switch eto := e.(type) {
		case kitTypes.ErrorWithCode:
			err = eto
		default:
			err = app.cfg.Errors().ComponentFinalizeExecution(0, componentName, eto)
		}
	}

	return
}

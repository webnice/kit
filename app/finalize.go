package app

import (
	"context"
	"errors"
	"os"
	"time"

	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	kitTypes "github.com/webnice/kit/v4/types"
)

// Функция вызова функции Finalize() у компоненты.
func (app *impl) finalizeFn(component *kitTypes.ComponentInfo) (err kitTypes.ErrorWithCode) {
	var (
		ctx context.Context
		cfn context.CancelFunc
	)

	// Создание контекста.
	ctx, cfn = context.WithCancel(context.Background())
	defer cfn()
	// Запуск вспомогательной горутины, которая по таймауту выведет в лог сообщение о долгой работе функции Finalize().
	go func(cx context.Context, name string, tout time.Duration) {
		const maxCount = 5
		var count int8

		warning := app.cfg.Errors().ComponentFinalizeWarning(0, name, tout)
		for {
			count++
			select {
			case <-cx.Done():
				return
			case <-time.After(tout):
				app.cfg.Log().Warning(warning.Error())
			}
			if count >= maxCount {
				os.Exit(int(warning.Code()))
			}
		}
	}(ctx, component.ComponentName, app.cfg.Gist().ComponentFinalizeWarningTimeout())
	err = app.finalizeSafeCall(component.ComponentName, component.Component)

	return
}

// Запуск функции Finalize() в компоненте с защитой от паники.
func (app *impl) finalizeSafeCall(componentName string, cpt kitTypes.Component) (err kitTypes.ErrorWithCode) {
	var e error

	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentFinalizePanicException(0, componentName, e, kitModuleTrace.StackShort())
		}
	}()
	if e = cpt.Finalize(); e != nil {
		var eto kitTypes.ErrorWithCode
		switch {
		case errors.As(e, &eto):
			err = eto
		default:
			err = app.cfg.Errors().ComponentFinalizeExecution(0, componentName, e)
		}
	}

	return
}

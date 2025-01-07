package app

import (
	"context"
	"os"
	"time"

	"github.com/webnice/dic"
	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	kitTypes "github.com/webnice/kit/v4/types"
)

// Функция вызова функции Finalize() у компоненты.
func (app *impl) finalizeFn(component *kitTypes.ComponentInfo) (err error) {
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
		var (
			warning error
			errcode int
			count   int8
		)

		errcode = app.cfg.Errors().ComponentFinalizeWarning.CodeI().Get()
		warning = app.cfg.Errors().ComponentFinalizeWarning.Bind(name, tout)
		for {
			count++
			select {
			case <-cx.Done():
				return
			case <-time.After(tout):
				app.cfg.Log().Warning(warning.Error())
			}
			if count >= maxCount {
				os.Exit(errcode)
			}
		}
	}(ctx, component.ComponentName, app.cfg.Gist().ComponentFinalizeWarningTimeout())
	err = app.finalizeSafeCall(component.ComponentName, component.Component)

	return
}

// Запуск функции Finalize() в компоненте с защитой от паники.
func (app *impl) finalizeSafeCall(componentName string, cpt kitTypes.Component) (err error) {
	var ierr dic.IError

	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentFinalizePanicException.Bind(componentName, e, kitModuleTrace.StackShort())
		}
	}()
	if err = cpt.Finalize(); err != nil {
		if ierr = app.Cfg().Errors().Unbind(err); ierr == nil {
			err = app.cfg.Errors().ComponentFinalizeExecution.Bind(componentName, err)
		}
	}

	return
}

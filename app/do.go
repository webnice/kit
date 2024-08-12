package app

import (
	"errors"

	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	kitTypes "github.com/webnice/kit/v4/types"
)

// Функция вызова функции Do() у компоненты.
func (app *impl) doFn(component *kitTypes.ComponentInfo) (err kitTypes.ErrorWithCode) {
	var (
		levelDone bool
		levelExit bool
	)

	levelDone, levelExit, err = app.doSafeCall(component.ComponentName, component.Component)
	if levelDone {
		// Отключение автоматического переключения уровня работы приложения.
		app.cfg.Gist().RunlevelAutoincrementStop()
	}
	if levelExit {
		// Асинхронное переключение уровня работы приложения на завершение работы.
		app.cfg.Gist().RunlevelExitAsync()
	}

	return
}

// Запуск функции Do() в компоненте с защитой от паники.
func (app *impl) doSafeCall(componentName string, cpt kitTypes.Component) (
	levelDone bool, // ______________ Отключение автоматического переключения уровня работы приложения.
	levelExit bool, // ______________ Переключение работы приложения на уровень завершения работы.
	err kitTypes.ErrorWithCode, // __ Ошибка.
) {
	var e error

	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentDoPanicException(0, componentName, e, kitModuleTrace.StackShort())
		}
	}()
	if levelDone, levelExit, e = cpt.Do(); e != nil {
		var eto kitTypes.ErrorWithCode
		switch {
		case errors.As(e, &eto):
			err = eto
		default:
			err = app.cfg.Errors().ComponentDoExecution(0, componentName, e)
		}
	}

	return
}

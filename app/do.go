package app

import (
	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	kitTypes "github.com/webnice/kit/v4/types"
)

// Функция вызова функции Do() у компоненты.
func (app *impl) doFn(component *kitTypes.ComponentInfo) (err error) {
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
	levelDone bool, // Отключение автоматического переключения уровня работы приложения.
	levelExit bool, // Переключение работы приложения на уровень завершения работы.
	err error,      // ____ Ошибка.
) {
	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentDoPanicException.Bind(componentName, e, kitModuleTrace.StackShort())
		}
	}()
	if levelDone, levelExit, err = cpt.Do(); err != nil {
		if app.cfg.Errors().Unbind(err) == nil {
			err = app.cfg.Errors().ComponentDoExecution.Bind(componentName, err)
		}
	}

	return
}

package app

import "math"

// Выполнение действий приложения на уровнях от 10 до 65534.
// [10] выполнение функций Do() компонентов приложения.
func (app *impl) mainRunlevelDo(_ uint16, new uint16) {
	const runlevelEnd = math.MaxUint16 - 1
	var (
		err  error
		code uint8
	)

	// Пропускаем уведомления об изменении уровня работы приложения, для не обрабатываемых уровней.
	if new < app.cfg.Gist().RunlevelDefault() || new >= runlevelEnd {
		return
	}
	// Выполнение компонентов приложения, функций Do().
	// Фильтрация компонентов по runlevel и команде.
	if code, err = app.cfg.Gist().
		ComponentDo(new, app.doFn); code != 0 || err != nil {
		if app.cfg.Errors().Unbind(err) == nil { // Ошибка не dic.IError, приведение к стандартизированной ошибке.
			err = app.cfg.Errors().ComponentDoUnknownError.Bind(err)
		}
		app.cfg.Gist().ErrorAppend(err)
	}
	// Остановка автоматического переключения уровня работы приложения при достижении целевого уровня.
	if app.cfg.Targetlevel() == new {
		app.cfg.Gist().RunlevelAutoincrementStop()
	}
	// Если включено автоматическое переключение уровня, увеличение уровня работы приложения асинхронно,
	// иначе произойдёт deadlock.
	if app.cfg.RunlevelAutoincrement() {
		app.cfg.Gist().RunlevelNextAsync()
	}
}

// Выполнение действий приложения на уровне 65535.
// [65535] выполнение функций завершения работы компонентов.
func (app *impl) mainRunlevelFinalize(_ uint16, new uint16) {
	const runlevel = math.MaxUint16
	var (
		err  error
		code uint8
	)

	// Пропускаем уведомления об изменении уровня работы приложения, для не обрабатываемых уровней.
	if new != runlevel {
		return
	}
	if app.cfg.Debug() {
		app.cfg.Log().Info(tplApplicationFinalize)
	}
	// После завершения функции, завершение приложения.
	defer func() { app.finalize <- struct{}{} }()
	if code, err = app.cfg.Gist().
		ComponentFinalize(app.finalizeFn); code != 0 || err != nil {
		if app.cfg.Errors().Unbind(err) == nil { // Ошибка не dic.IError, приведение к стандартизированной ошибке.
			err = app.cfg.Errors().ComponentFinalizeUnknownError.Bind(err)
		}
		app.cfg.Gist().ErrorAppend(err)
	}
}

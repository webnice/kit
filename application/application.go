// Package application
package application

import (
	"bytes"
	"errors"
	"runtime"

	kitModuleCfg "github.com/webnice/kit/v3/module/cfg"
	kitTypes "github.com/webnice/kit/v3/types"
)

func init() { singleton = newApp().(*impl) }

// Создание нового экземпляра объекта приложения
func newApp() Interface {
	return &impl{
		cfg:      kitModuleCfg.Get(),
		finalize: make(chan struct{}),
	}
}

// Get Функция возвращает интерфейс объекта пакета
func Get() Interface { return singleton }

// Cfg Возвращает интерфейс конфигурации приложения
func (app *impl) Cfg() kitModuleCfg.Interface { return app.cfg }

// Проверка накопленных ошибок приложения, вернётся истина, если ошибки есть
func (app *impl) mainCheckingAccumulatedErrors(log func(string, ...interface{})) (ret bool) {
	if ret = app.cfg.Gist().IsError(); ret && app.cfg.Debug() {
		log(tplIsAccumulatedError, app.cfg.Runlevel())
	}

	return
}

// Main Точка запуска, выполнения и завершения приложения. Дирижёр логики работы с компонентами.
// Функция возвращает код ошибки, который передаётся в операционную систему и может быть считан запускающим
// приложением, скриптом или операционной системой.
func (app *impl) Main() (code uint8, err error) {
	var (
		e       kitTypes.ErrorWithCode
		ok, end bool
	)

	for {
		if end {
			break
		}
		// Проверка накопленных ошибок приложения, если ошибки есть, выход
		if app.mainCheckingAccumulatedErrors(app.cfg.Log().Errorf) {
			return
		}
		switch app.cfg.Gist().RunlevelNext(); app.cfg.Runlevel() {
		case 1: // [1] Закрытие регистрации компонентов и объектов конфигурации
			// Настройка менеджера логирования
			if err = app.Cfg().Gist().
				Logger().Debug(app.cfg.Debug()).Initialization(); err != nil {
				app.cfg.Gist().ErrorAppend(app.cfg.Errors().InitLogging(0, err))
				return
			}
			if app.cfg.Debug() {
				app.cfg.Log().Info(tplApplicationBegin)
			}
		case 2: // [2] Опрос компонентов, получение настроек компонентов (preferences)
			if code, err = app.cfg.Gist().
				ComponentPreferences(app.preferencesFn); code != 0 || err != nil {
				return
			}
		case 3: // [3] Проверка конфликтов компонентов (conflict)
			if code, err = app.cfg.Gist().
				ComponentCheckConflict(app.conflictFn); code != 0 || err != nil {
				return
			}
		case 4: // [4] Проверка зависимостей компонентов (requires)
			if code, err = app.cfg.Gist().
				ComponentRequiresCheck(app.requiresFn); code != 0 || err != nil {
				return
			}
		case 5: // [5] Сортировка компонентов в соответствии с настройками (before) и (after)
			if code, err = app.cfg.Gist().
				ComponentSort(app.sortFn); code != 0 || err != nil {
				return
			}
		case 6: // [6] инициализация параметров командного интерфейса и переменных окружения
			if code, err = app.mainRunCliInitWithParseError(); code != 0 || err != nil {
				return
			}
			// Обновление флага отладки в менеджере логирования.
			app.Cfg().Gist().Logger().Debug(app.cfg.Debug())
		case 7: // [7] Построение шагов переключения уровня выполнения приложения (runlevel)
			if code, err = app.cfg.Gist().
				ComponentMapRunlevel(app.cfg.Targetlevel()); code != 0 || err != nil {
				return
			}
		case 8: // [8] Выполнение функций инициализации компонентов (initiate)
			if code, err = app.cfg.Gist().
				ComponentInitiate(app.initiateFn); code != 0 || err != nil {
				return
			}
		case 9: // [9] Регистрация функций слежения за уровнем приложения
			// Функция mainRunlevelDo - слежение за переключением уровней с 10 по 65535
			if err = app.cfg.RunlevelSubscribe(app.mainRunlevelDo); err != nil {
				if e, ok = err.(kitTypes.ErrorWithCode); ok {
					code, err = e.Code(), errors.New(e.Error())
				}
				return
			}
			// Функция mainRunlevelFinalize - слежение за переключением на уровень 65535, запуск функций Finalize()
			if err = app.cfg.RunlevelSubscribe(app.mainRunlevelFinalize); err != nil {
				if e, ok = err.(kitTypes.ErrorWithCode); ok {
					code, err = e.Code(), errors.New(e.Error())
				}
				return
			}
		default: // [10] дальнейшее переключение уровней выполняется из функции mainRunlevelDo()
			if app.mainCheckingAccumulatedErrors(app.cfg.Log().Warningf) {
				app.cfg.Gist().RunlevelExitAsync()
			}
			end = true
		}
	}
	defer func() {
		runtime.Gosched()
		// Остановка шины данных приложения.
		app.Cfg().Bus().Gist().WorkerStop()
		runtime.Gosched()
		// Остановка менеджера логирования.
		app.Cfg().Gist().Logger().FlushAndClose()
	}()
	// Ожидание сигнала об окончании выполнения функций Finalize() всех зарегистрированных компонентов приложения
	<-app.finalize
	// Удаление регистрации функции mainRunlevelDo
	if err = app.cfg.RunlevelUnsubscribe(app.mainRunlevelDo); err != nil {
		if e, ok = err.(kitTypes.ErrorWithCode); ok {
			code, err = e.Code(), errors.New(e.Error())
		}
		return
	}
	// Удаление регистрации функции mainRunlevelFinalize
	if err = app.cfg.RunlevelUnsubscribe(app.mainRunlevelFinalize); err != nil {
		if e, ok = err.(kitTypes.ErrorWithCode); ok {
			code, err = e.Code(), errors.New(e.Error())
		}
		return
	}
	// Завершение работы приложения, код ошибки будет взят из массива ошибок и возвращён в операционную систему
	if app.cfg.Debug() {
		app.cfg.Log().Info(tplApplicationEnd)
	}

	return
}

// Выполнение инициализации библиотеки командного интерфейса и разбор ошибок
func (app *impl) mainRunCliInitWithParseError() (code uint8, err error) {
	var (
		help *bytes.Buffer
		desc string
		ewc  kitModuleCfg.Err
		bcfw *kitTypes.BootstrapConfigurationForkWorker
	)

	// Инициализация CLI, загрузка команды приложения, аргументов, флагов и переменных окружения
	if help, desc, err = app.cfg.Gist().
		CLI().Init(); err != nil {
		switch err {
		// Отображение помощи по командам, аргументам и флагам приложения
		case app.cfg.Gist().CLI().Errors().HelpDisplayed():
			ewc = app.cfg.Errors().ApplicationHelpDisplayed(0, help)
			code, err = ewc.Code(), errors.New(ewc.Error())
		// Требуется указать команду, аргумент или флаг командной строки
		case app.cfg.Gist().CLI().Errors().RequiredCommand():
			ewc = app.cfg.Errors().CommandLineArgumentRequired(0, desc)
			code, err = ewc.Code(), errors.New(ewc.Error())
		// Указана не известная команда
		case app.cfg.Gist().CLI().Errors().UnknownCommand():
			ewc = app.cfg.Errors().CommandLineArgumentUnknown(0, desc)
			code, err = ewc.Code(), errors.New(ewc.Error())
		// Не верное значение, тип значения, битность значения, аргумента, флага или параметра (проверка значений)
		case app.cfg.Gist().CLI().Errors().NotCorrectArgument():
			ewc = app.cfg.Errors().CommandLineArgumentNotCorrect(0, desc)
			code, err = ewc.Code(), errors.New(ewc.Error())
		// Неизвестный аргумент
		case app.cfg.Gist().CLI().Errors().UnknownArgument():
			ewc = app.cfg.Errors().CommandLineArgumentUnknown(0, desc)
			code, err = ewc.Code(), errors.New(ewc.Error())
		// Не указан один или несколько обязательных флагов
		case app.cfg.Gist().CLI().Errors().RequiredFlag():
			ewc = app.cfg.Errors().CommandLineRequiredFlag(0, desc)
			code, err = ewc.Code(), errors.New(ewc.Error())
		// Любая иная не предвиденная ошибка
		default:
			ewc = app.cfg.Errors().CommandLineUnexpectedError(0, desc, err)
			code, err = ewc.Code(), errors.New(ewc.Error())
		}

		return
	}
	// Коррекция значений загруженных через CLI и присвоение загруженной команды приложения
	app.cfg.Gist().
		UpdateBootstrapConfigurationPathValue().
		CommandFull(app.cfg.Gist().CLI().CommandFull())
	// Переключение флага режима forkWorker
	if bcfw = app.cfg.ForkWorker(); bcfw.Master != "" || bcfw.Component != "" {
		app.cfg.Gist().ForkWorkerMode(true)
	}

	return
}

// Package application
package application

import (
	"math"
	"regexp"

	kitModuleTrace "github.com/webnice/kit/v3/module/trace"
	kitTypes "github.com/webnice/kit/v3/types"
)

// Опрос зарегистрированных компонентов, получение настроек компонентов
// Функция загрузки настроек компонента
func (app *impl) preferencesFn(c *kitTypes.ComponentInfo) (ret *kitTypes.ComponentInfo, err kitTypes.ErrorWithCode) {
	const (
		cBefore   = `before`
		cAfter    = `after`
		cRequire  = `require`
		cConflict = `conflict`
	)
	var (
		preferences kitTypes.ComponentPreferences
		n           int
	)

	// Функция защиты от паники, в связи с тем что вызывается код внешней компоненты, в нём может быть любая ошибка
	defer func() {
		if e := recover(); e != nil {
			err = app.cfg.Errors().ComponentPanicException(0, c.ComponentName, e, kitModuleTrace.StackShort())
		}
	}()
	// Копирование настроек из регистрации компоненты
	ret = &kitTypes.ComponentInfo{
		InitiateTimeout: c.InitiateTimeout,
		Component:       c.Component,
		ComponentName:   c.ComponentName,
	}
	// Запрос настроек компоненты
	preferences = ret.
		Component.
		Preferences()
	// Присвоение настройки максимального время ожидания инициализации компоненты
	if preferences.InitiateTimeout > 0 {
		ret.InitiateTimeout = preferences.InitiateTimeout
	}
	// Проверка и присвоение настройки уровня запуска компоненты
	if preferences.Runlevel > 0 && (preferences.Runlevel < 10 || preferences.Runlevel > math.MaxUint16-1) {
		ret, err = nil, app.cfg.Errors().ComponentRunlevelError(0, ret.ComponentName, preferences.Runlevel)
		return
	}
	ret.Runlevel = preferences.Runlevel
	// Присвоение настройки активности компоненты
	ret.IsDisable = preferences.IsDisable
	// Компиляция regexp правил с контролем ошибок
	// BEFORE
	if ret.Before, err = app.preferencesRegexpMake(preferences.Before, cBefore, ret.ComponentName); err != nil {
		return
	}
	// AFTER
	if ret.After, err = app.preferencesRegexpMake(preferences.After, cAfter, ret.ComponentName); err != nil {
		return
	}
	// REQUIRE
	if ret.Require, err = app.preferencesRegexpMake(preferences.Require, cRequire, ret.ComponentName); err != nil {
		return
	}
	// CONFLICT
	if ret.Conflict, err = app.preferencesRegexpMake(preferences.Conflict, cConflict, ret.ComponentName); err != nil {
		return
	}
	// КОМАНДЫ Динамические команды компоненты приложения
	ret.Command = make([]string, 0, len(preferences.Command))
	for n = range preferences.Command {
		// Регистрация команды и группы команд компоненты для отображения в CLI
		app.cfg.Gist().ComponentCommandRegister(preferences.Command[n])
		// Добавление команды
		ret.Command = append(ret.Command, preferences.Command[n].Command)
	}
	// ФЛАГИ Глобальные флаги компоненты приложения
	for n = range preferences.Flag {
		if preferences.Flag[n].Flag == "" {
			continue
		}
		// Регистрация флага компоненты для отображения в CLI
		app.cfg.Gist().ComponentFlagRegister(preferences.Flag[n])
	}

	return
}

// Преобразование массива строк в массив regexp правил
func (app *impl) preferencesRegexpMake(rules []string, key string, componentName string) (
	ret []*regexp.Regexp,
	err kitTypes.ErrorWithCode,
) {
	var (
		e   error
		n   int
		rex *regexp.Regexp
	)

	ret = make([]*regexp.Regexp, 0, len(rules))
	for n = range rules {
		if rex, e = regexp.Compile(rules[n]); e != nil {
			err = app.cfg.Errors().ComponentRulesError(0, key, componentName, e)
			return
		}
		ret = append(ret, rex)
	}

	return
}

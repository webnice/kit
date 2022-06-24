// Package migrations
package migrations

import (
	kitModuleCfgReg "github.com/webnice/kit/v3/module/cfg/reg"
	kitTypes "github.com/webnice/kit/v3/types"
	//
	//kitWorkflow "git.webdesk.ru/wd/kit/workflow"
	//
	//log "github.com/webnice/lv2"
)

// Структура объекта компоненты.
type impl struct {
	//Pid pidfile.Interface
}

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component { return new(impl) }

// Preferences Функция возвращает настройки компоненты.
func (migrations *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cConfiguration = `(?mi)application/component/configuration$`
		cLogging       = `(?mi)application/component/logging$`
		cPidfile       = `(?mi)application/component/pidfile$`
	)
	return kitTypes.ComponentPreferences{
		After:   []string{cConfiguration, cLogging, cPidfile},
		Require: []string{cPidfile},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (migrations *impl) Initiate() (err error) { return }

// Do Выполнение компонента приложения.
func (migrations *impl) Do() (levelDone bool, levelExit bool, err error) {
	//const commandVersion = `version`
	//var command string
	//
	//exitCode = kitWorkflow.ErrNone
	//if cmd == commandVersion {
	//	return
	//}
	//log.Info(`Application database migrations apply started`)
	//// Поиск утилиты применения миграций
	//if command = mrs.migrationsUtility(); command == "" {
	//	return
	//}
	//if err = mrs.migrationsSQL(command); err != nil {
	//	log.Errorf("SQL migration error: %s", err)
	//}
	////if err = cgn.migrationsClickhouse(command); err != nil {
	////	log.Errorf("ClickHouse migration error: %s", err)
	////}
	//log.Info(`Application database migrations apply completed`)

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (migrations *impl) Finalize() (err error) { return }

package bootstrap

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
// Компонента не выполняет никакой полезной нагрузки, предназначена для определения опорной точки,
// после которой приложение полностью инициализировано и готово к работе.
//
// В последующих компонентах приложения, достаточно указать то, что они выполняются после компоненты bootstrap,
// вместо перечисления всех необходимых для работы компонентов.
//
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/v4/module/cfg/reg"
	kitTypes "github.com/webnice/kit/v4/types"
)

const tplApplicationFullStart = "Приложение полностью инициализировано."

// Структура объекта компоненты.
type impl struct {
	cfg kitModuleCfg.Interface
}

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component { return &impl{cfg: kitModuleCfg.Get()} }

// Ссылка на функцию получения значения режима отладки, для удобного использования внутри компоненты.
func (brp *impl) debug() bool { return brp.cfg.Debug() }

// Ссылка на менеджер логирования, для удобного использования внутри компоненты.
func (brp *impl) log() kitTypes.Logger { return brp.cfg.Log() }

// Preferences Функция возвращает настройки компоненты.
func (brp *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cEnvironment   = "(?mi)/component/environment$"
		cInterrupt     = "(?mi)/component/interrupt$"
		cConfiguration = "(?mi)/component/configuration$"
		cLogging       = "(?mi)/component/logg.*"
		cLoggerConsole = "(?mi)/component/logger_console$"
		cPidfile       = "(?mi)/component/pidfile$"
		cMigration     = "(?mi)/component/migration.*$"
	)
	return kitTypes.ComponentPreferences{
		After: []string{
			cEnvironment,
			cConfiguration,
			cLogging,
			cLoggerConsole,
			cInterrupt,
			cPidfile,
			cMigration,
		},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (brp *impl) Initiate() (err error) { return }

// Do Выполнение компонента приложения.
func (brp *impl) Do() (levelDone bool, levelExit bool, err error) {
	if brp.debug() {
		brp.log().Info(tplApplicationFullStart)
	}

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (brp *impl) Finalize() (err error) { return }

package migration_sql

import (
	"strings"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/v4/module/cfg/reg"
	kitModuleDbSql "github.com/webnice/kit/v4/module/db/sql"
	kitTypes "github.com/webnice/kit/v4/types"
	kitTypesDb "github.com/webnice/kit/v4/types/db"
)

// Структура объекта компоненты.
type impl struct {
	cfg         kitModuleCfg.Interface
	databaseSql *kitTypesDb.DatabaseSqlConfiguration
}

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component {
	var m8s = &impl{
		cfg:         kitModuleCfg.Get(),
		databaseSql: new(kitTypesDb.DatabaseSqlConfiguration),
	}

	// Регистрация конфигураций.
	if !m8s.cfg.Gist().ConfigurationRegistration(m8s.databaseSql) {
		m8s.log().Error("ошибка регистрации конфигурации")
	}

	return m8s
}

// Ссылка на менеджер логирования, для удобного использования внутри компоненты или модуля.
func (m8s *impl) log() kitTypes.Logger { return m8s.cfg.Log() }

// Preferences Функция возвращает настройки компоненты.
func (m8s *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cEnvironment   = "(?mi)/component/environment$"
		cInterrupt     = "(?mi)/component/interrupt$"
		cConfiguration = "(?mi)/component/configuration$"
		cLogging       = "(?mi)/component/logg.*"
		cLoggerConsole = "(?mi)/component/logger_console$"
		cPidfile       = "(?mi)/component/pidfile$"
		cBootstrap     = "(?mi)/component/bootstrap$"
	)
	return kitTypes.ComponentPreferences{
		After:   []string{cConfiguration, cLoggerConsole, cLogging, cPidfile, cInterrupt, cEnvironment},
		Require: []string{cPidfile},
		Before:  []string{cBootstrap},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (m8s *impl) Initiate() (err error) {
	var (
		elm any
		ok  bool
		c   *kitTypesDb.DatabaseSqlConfiguration
	)

	if m8s.isSkip() {
		return
	}
	// Загрузка конфигурации базы данных, сохранённой в конфигурации приложения.
	if elm, err = m8s.cfg.ConfigurationByObject(m8s.databaseSql); err != nil {
		return
	}
	// Приведение пустого интерфейса к типу данных.
	if c, ok = elm.(*kitTypesDb.DatabaseSqlConfiguration); ok {
		// Исправление пути к миграции на абсолютный путь, исправление по адресу, поэтому все кто запросят
		// конфигурацию базы данных, получат исправленный вариант.
		m8s.cfg.Gist().AbsolutePathAndUpdate(&c.SqlDB.Migration)
		// Обновление локальной копии конфигурации, так как после работы yaml библиотеки может слетать адрес.
		m8s.databaseSql = c
	}

	return
}

// Do Выполнение компонента приложения.
func (m8s *impl) Do() (levelDone bool, levelExit bool, err error) {
	if m8s.isSkip() {
		return
	}
	if err = kitModuleDbSql.Get().MigrationUp(); err != nil {
		levelDone, levelExit = true, true
	}

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (m8s *impl) Finalize() (err error) { return }

func (m8s *impl) isSkip() (ret bool) {
	const cmdVersion, cmdConfig = `version`, `config`

	// Для стандартной команды версии приложения миграцию не запускаем.
	switch {
	case m8s.cfg.Command() == cmdVersion:
		ret = true
	case strings.HasPrefix(m8s.cfg.Command(), cmdConfig):
		ret = true
	}

	return
}

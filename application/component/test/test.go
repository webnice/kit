// Package test
package test

import (
	"context"
	"github.com/webnice/debug"
	kitModuleCfg "github.com/webnice/kit/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/module/cfg/reg"
	kitTypes "github.com/webnice/kit/types"
)

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component {
	var tst = &impl{
		cfg:    kitModuleCfg.Get(),
		opt:    &options{},
		Dumper: debug.Dumper,
	}
	tst.ctx, tst.cfn = context.WithCancel(context.Background())
	tst.registrationConfiguration()
	return tst
}

// Регистрация объекта конфигурации приложения
func (tst *impl) registrationConfiguration() {
	var err error

	tst.cobj = new(Configuration)
	if err = tst.cfg.Gist().ConfigurationRegistration(tst.cobj); err != nil {
		switch eto := err.(type) {
		case kitModuleCfg.Err:
			tst.cfg.Gist().ErrorAppend(eto)
		default:
			tst.cfg.Gist().ErrorAppend(tst.cfg.Errors().ConfigurationApplicationObject(0, eto))
		}
		return
	}
}

// Ссылка на функцию получения значения режима отладки, для удобного использования внутри компоненты.
func (tst *impl) debug() bool { return tst.cfg.Debug() }

// Ссылка на менеджер логирования, для удобного использования внутри компоненты.
func (tst *impl) log() kitTypes.Logger { return tst.cfg.Log() }

// Preferences Функция возвращает настройки компоненты.
func (tst *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cConfiguration = `(?mi)application/component/configuration$`
		cLogging       = `(?mi)application/component/logging$`
		cInterrupt     = `(?mi)application/component/interrupt$`
		cBootstrap     = `(?mi)application/component/bootstrap$`
		cDaemon        = `(?mi)application/component/daemon$`
		cUnknown       = `(?mi)application/component/unknown$`
	)

	return kitTypes.ComponentPreferences{
		After:    []string{cBootstrap},
		Before:   []string{cDaemon},
		Conflict: []string{cUnknown},
		Require:  []string{cInterrupt, cConfiguration},
		Runlevel: 200,
		Command: []kitTypes.ComponentCommand{
			{
				GroupKey:         "test",
				GroupTitle:       "Тестовая группа команд:",
				GroupDescription: "Временные тестовые команды приложения.",
				Command:          "test",
				Description:      "Просто тестовая команда.",
				Value:            &tst.opt,
				//IsDefault:        true,
				//IsHidden:         false,
			},
		},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (tst *impl) Initiate() (err error) {

	//tst.logPrint()

	return
}

// Do Выполнение компонента приложения.
func (tst *impl) Do() (levelDone bool, levelExit bool, err error) {

	//levelDone = true
	//tst.subscribeDatabuser()
	//go tst.publisher()

	//levelDone = true
	tst.logPrint()

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (tst *impl) Finalize() (err error) {

	//tst.log().Debug("Запущена функция Finalize().")
	//defer tst.unsubscribeDatabuser()

	tst.cfn()

	//panic("finalize")
	//err = errors.New("finalize")

	//if cpn.Pid == nil {
	//	return
	//}
	//// Удаление PID файла при завершении приложения
	//if err = cpn.Pid.Unlock(); err != nil {
	//	exitCode, err = kitWorkflow.ErrPidFileDelete, fmt.Errorf("delete PID file error: %s", err)
	//	return
	//}
	//if cpn.Cfg.Debug() {
	//	log.Info(`Application PID file deleted successfully`)
	//}

	return
}

// Package logger
package logger

import (
	kitModuleCfgReg "github.com/webnice/kit/module/cfg/reg"
	kitTypes "github.com/webnice/kit/types"
	//
	//kitWorkflow "git.webdesk.ru/wd/kit/workflow"
	//
	//r "github.com/webnice/lv2/receiver"
	//s "github.com/webnice/lv2/sender"
)

// Структура объекта компоненты.
type impl struct {
}

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component { return new(impl) }

// Preferences Функция возвращает настройки компоненты.
func (logging *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cConfiguration = `(?mi)application/component/configuration$`
		cLoggerLogrus  = `(?mi)application/component/logger_logrus$`
	)
	return kitTypes.ComponentPreferences{
		After:    []string{cConfiguration},
		Conflict: []string{cLoggerLogrus},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (logging *impl) Initiate() (err error) {
	//var receiver s.Receiver
	//
	//cpn.Cfg = configuration.Get()
	//if cpn.Cfg.Log() == nil {
	//	return
	//}
	//if cpn.Cfg.Log().GraylogEnable {
	//	receiver = r.GelfReceiver.
	//		SetAddress(cpn.Cfg.Log().GraylogProto, cpn.Cfg.Log().GraylogAddress, cpn.Cfg.Log().GraylogPort).
	//		SetCompression("gzip").
	//		Receiver
	//	s.Gist().AddSender(receiver)
	//}
	//if cpn.Cfg.Debug() {
	//	s.Gist().AddSender(r.Default.Receiver)
	//}
	//log.Gist().StandardLogSet()
	//if cpn.Cfg.Debug() {
	//	log.Info(`Logging system has initialized successfully`)
	//}

	//l := cfg.Get().Loglevel()
	//cfg.Get().Log().Infof("@ уровень логирования: %d (%s)", l.Int(), l)

	return
}

// Do Выполнение компонента приложения.
func (logging *impl) Do() (levelDone bool, levelExit bool, err error) { return }

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (logging *impl) Finalize() (err error) {
	//if cpn.Cfg.Debug() {
	//	log.Info(`Logging system has been shutdown`)
	//}
	//log.Gist().StandardLogUnset()
	//log.Done()

	return
}

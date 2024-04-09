package loggerconsole

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/webnice/kit/v4/app/component/logger_console/tpl"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/v4/module/cfg/reg"
	kitModuleLogLevel "github.com/webnice/kit/v4/module/log/level"
	kitTypes "github.com/webnice/kit/v4/types"
)

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() (ret kitTypes.Component) {
	var lgc = &impl{
		cfg:            kitModuleCfg.Get(),
		handlerMux:     new(sync.Mutex),
		handlerControl: true,
	}

	lgc.handlerMux.Lock()
	lgc.tpl = tpl.New(lgc.cfg.RawWriter(), defaultTpl)
	lgc.err = lgc.cfg.Gist().Logger().
		HandlerSubscribe(lgc.Handler, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace)

	return lgc
}

// Ссылка на функцию получения значения режима отладки, для удобного использования внутри компоненты.
func (lgc *impl) debug() bool { return lgc.cfg.Debug() }

// Ссылка на менеджер логирования, для удобного использования внутри компоненты.
func (lgc *impl) log() kitTypes.Logger { return lgc.cfg.Log() }

// Ссылка на интерфейс вывода потоковых сообщений в консоль.
func (lgc *impl) wr() kitTypes.SyncWriter { return lgc.cfg.RawWriter() }

// Preferences Функция возвращает настройки компоненты.
func (lgc *impl) Preferences() kitTypes.ComponentPreferences {
	const cConfiguration = `(?mi)app/component/configuration$`
	return kitTypes.ComponentPreferences{
		After: []string{cConfiguration},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (lgc *impl) Initiate() (err error) {
	if err = lgc.err; err != nil {
		return
	}
	if err = lgc.tpl.Parse(); err != nil {
		return
	}
	lgc.handlerControl = false
	lgc.handlerMux.Unlock()
	runtime.Gosched()
	if lgc.cfg.Debug() {
		lgc.log().Info(tplOnInitiate)
		runtime.Gosched()
	}

	return
}

// Do Выполнение компонента приложения.
func (lgc *impl) Do() (levelDone bool, levelExit bool, err error) { return }

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (lgc *impl) Finalize() (err error) {
	if lgc.cfg.Debug() {
		lgc.log().Info(tplOnFinalize)
		runtime.Gosched()
	}
	if err = lgc.cfg.Gist().Logger().HandlerUnsubscribe(lgc.Handler); err != nil {
		_, _ = fmt.Fprintln(lgc.wr(), err.Error())
		return
	}

	return
}

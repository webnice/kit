package interrupt

import (
	"os"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/v4/module/cfg/reg"
	kitModuleInterrupt "github.com/webnice/kit/v4/module/interrupt"
	kitTypes "github.com/webnice/kit/v4/types"
)

const (
	newLine              = "\n"
	tplInterruptNew      = "Получен сигнал прерывания %q."
	tplInterruptEnabled  = "Включён перехват системных прерываний."
	tplInterruptDisabled = "Отключён перехват системных прерываний."
	sigInterrupt         = "interrupt"
)

// Объект сущности пакета.
type impl struct {
	i7t kitModuleInterrupt.Interface
}

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component { return new(impl) }

// Ссылка на функцию получения значения режима отладки, для удобного использования внутри компоненты.
func (ipt *impl) debug() bool { return kitModuleCfg.Get().Debug() }

// Ссылка на менеджер логирования, для удобного использования внутри компоненты.
func (ipt *impl) log() kitTypes.Logger { return kitModuleCfg.Get().Log() }

// Preferences Функция возвращает настройки компоненты.
func (ipt *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cConfiguration = "(?mi)/component/configuration$"
		cLogging       = "(?mi)/component/logging$"
	)
	return kitTypes.ComponentPreferences{
		After: []string{cConfiguration, cLogging},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (ipt *impl) Initiate() (err error) { return }

// Do Выполнение компонента приложения.
func (ipt *impl) Do() (levelDone bool, levelExit bool, err error) {
	ipt.i7t = kitModuleInterrupt.New(func(sig os.Signal) {
		// Печать новой строки, если был нажат Ctrl-C
		if sig.String() == sigInterrupt {
			_, _ = os.Stdout.WriteString(newLine)
		}
		ipt.log().Alertf(tplInterruptNew, sig.String())
		kitModuleCfg.Get().Gist().RunlevelExitAsync()
	})
	if ipt.i7t.Start(); ipt.debug() {
		ipt.log().Infof(tplInterruptEnabled)
	}

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом
func (ipt *impl) Finalize() (err error) {
	if ipt.i7t == nil {
		return
	}
	if ipt.i7t.Stop(); ipt.debug() {
		ipt.log().Infof(tplInterruptDisabled)
	}

	return
}

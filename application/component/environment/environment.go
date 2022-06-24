// Package environment
package environment

import (
	"context"
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"

	kitModuleCfg "github.com/webnice/kit/v3/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/v3/module/cfg/reg"
	kitTypes "github.com/webnice/kit/v3/types"
)

// Структура объекта компоненты.
type impl struct {
	Ctx       context.Context    // Контекст для прерывания работы горутины.
	Cfn       context.CancelFunc // Функция контекста для прерывания работы.
	Done      chan struct{}      // Канал обратной связи.
	IsRunning atomic.Value       // Флаг состояния запуска горутины.
	MaxCpu    int32              // Значение флага командной строки с количеством используемых приложением процессоров.
}

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component {
	var ret = &impl{
		Done:   make(chan struct{}),
		MaxCpu: int32(runtime.NumCPU()),
	}
	ret.IsRunning.Store(false)

	return ret
}

// Ссылка на функцию получения значения режима отладки, для удобного использования внутри компоненты.
func (env *impl) debug() bool { return kitModuleCfg.Get().Debug() }

// Ссылка на менеджер логирования, для удобного использования внутри компоненты.
func (env *impl) log() kitTypes.Logger { return kitModuleCfg.Get().Log() }

// Preferences Функция возвращает настройки компоненты.
func (env *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cmdFlag        = "max-cpu-usage"
		cmdDescription = "Максимальное количество используемых CPU, если указано 0, тогда все доступные CPU."
		cmdEnvironment = "APPLICATION_MAX_CPU_USAGE"
		cmdPlaceholder = "0"
		cConfiguration = `(?mi)application/component/configuration$`
	)
	return kitTypes.ComponentPreferences{
		Before: []string{cConfiguration},
		Flag: []kitTypes.ComponentFlag{
			{
				Flag:        cmdFlag,
				Description: cmdDescription,
				Environment: cmdEnvironment,
				Placeholder: cmdPlaceholder,
				Value:       &env.MaxCpu,
			},
		},
	}
}

// Initiate Функция инициализации компоненты и подготовки к запуску.
func (env *impl) Initiate() (err error) {
	var maxCpu int

	// Установка значения runtime.NumCPU() по данным переданным в --max-cpu-usage.
	if maxCpu = int(env.MaxCpu); maxCpu == 0 || maxCpu > runtime.NumCPU() {
		maxCpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(maxCpu)
	// Инициализация генератора случайных чисел.
	rand.Seed(time.Now().UnixNano())
	// Инициализация контекста.
	env.Ctx, env.Cfn = context.WithCancel(context.Background())

	return
}

// Do Выполнение компоненты приложения.
func (env *impl) Do() (levelDone bool, levelExit bool, err error) {
	if !env.IsRunning.Load().(bool) && env.debug() {
		env.IsRunning.Store(true)
		go env.gcScheduler()
	}

	return
}

// Finalize Функция вызывается перед завершением компоненты и приложения в целом.
func (env *impl) Finalize() (err error) {
	env.Cfn()
	if env.IsRunning.Load().(bool) {
		_ = <-env.Done
	}
	close(env.Done)

	return
}

// Горутина постоянного запуска GC для очистки памяти. Выполняется только в режиме отладки приложения.
// Так как GC по умолчанию оптимизирован и запускается при накоплении 100% переменных от общей кучи
// в режиме отладки запускаем GC раз в секунду, для отладки "финализаторов", отладки и проверки утечки памяти.
func (env *impl) gcScheduler() {
	const (
		tickerDuration = time.Second * 3
		tplInfoBegin   = "Запущен таймер принудительного выполнения сборщика мусора с периодичностью: %s."
		tplInfoEnd     = "Остановлен таймер принудительного выполнения сборщика мусора."
	)
	var (
		tic    *time.Ticker
		isDone bool
	)

	tic = time.NewTicker(tickerDuration)
	defer tic.Stop()
	env.log().Infof(tplInfoBegin, tickerDuration)
	defer env.log().Infof(tplInfoEnd)
	for {
		if isDone {
			break
		}
		select {
		case <-tic.C:
			runtime.GC()
		case <-env.Ctx.Done():
			isDone = true
		}
	}
	env.Done <- struct{}{}
}

// Package daemon
package daemon

import (
	kitModuleCfgReg "github.com/webnice/kit/module/cfg/reg"
	kitTypes "github.com/webnice/kit/types"
)

// Структура объекта компоненты.
type impl struct {
	//Cfg configuration.Interface
	//Jbo job.Interface
}

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component { return new(impl) }

// Preferences Функция возвращает настройки компоненты.
func (daemon *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cBootstrap = `(?mi)application/component/bootstrap$`
		cProjectID = `(?mi)application/component/projectid$`
	)
	return kitTypes.ComponentPreferences{
		After:    []string{cBootstrap, cProjectID},
		Runlevel: 100,
		Command: []kitTypes.ComponentCommand{
			{
				Command:          "daemon",
				Description:      "Запуск приложения в режиме службы.",
				GroupKey:         "main",
				GroupTitle:       "Основные режимы работы:",
				GroupDescription: "Команды основных режимов работы приложения.",
			},
		},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (daemon *impl) Initiate() (err error) {
	//cpn.Cfg = configuration.Get()
	//if cpn.Jbo, err = workers.Init(); err != nil {
	//	exitCode = kitWorkflow.ErrCantCreateWorkers
	//	return
	//}

	return
}

// Do Выполнение компонента приложения.
func (daemon *impl) Do() (levelDone bool, levelExit bool, err error) {
	//const commandDaemon = `daemon`
	//var fatality bool
	//
	//if cmd != commandDaemon {
	//	return
	//}
	//done = true
	//cpn.Jbo.
	//	// Регистрация функции получения ошибок
	//	RegisterErrorFunc(func(id string, err error) {
	//		if fatality {
	//			return
	//		}
	//		log.Errorf(" - worker process %q error: %s", id, err)
	//	}).
	//	// Регистрация функции получения изменений состояни процессов
	//	RegisterChangeStateFunc(func(id string, running bool) {
	//		if !cpn.Cfg.Debug() || fatality {
	//			return
	//		}
	//		if running {
	//			log.Noticef(" - worker process %q started", id)
	//		} else {
	//			log.Noticef(" - worker process %q stopped", id)
	//		}
	//	})
	//if cpn.Cfg.Debug() {
	//	log.Info(`Application workers has initialized`)
	//}
	//// Запуск, запускаются все воркеры с флагом Autostart=true
	//if err = cpn.Jbo.Do(); err != nil {
	//	defer log.Done()
	//	fatality = true
	//	err = fmt.Errorf("workers error: %s", err)
	//	return
	//}
	//// На этом этапе все воркеры приложения запущены
	//if cpn.Cfg.Debug() {
	//	log.Info(`Application workers started successfully`)
	//}
	//// Ожидание завершения всех воркеров
	//cpn.Jbo.Wait()

	//debug.Dumper(os.Getwd())

	levelDone = true
	//cfg.Get().Log().Fatal("Daemon ok")

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (daemon *impl) Finalize() (err error) { return }

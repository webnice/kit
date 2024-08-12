package version

import (
	"fmt"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitModuleCfgReg "github.com/webnice/kit/v4/module/cfg/reg"
	kitTypes "github.com/webnice/kit/v4/types"
)

// Структура объекта компоненты.
type impl struct {
}

// Регистрация компоненты в приложении.
func init() { kitModuleCfgReg.Registration(newComponent()) }

// Конструктор объекта компоненты.
func newComponent() kitTypes.Component { return new(impl) }

// Preferences Функция возвращает настройки компоненты.
func (ver *impl) Preferences() kitTypes.ComponentPreferences {
	const (
		cBootstrap = "(?mi)/component/bootstrap$"
	)
	return kitTypes.ComponentPreferences{
		After:    []string{cBootstrap},
		Runlevel: 100,
		Command: []kitTypes.ComponentCommand{
			{
				Command:          "version",
				Description:      "Отображение версии приложения и завершение работы.",
				GroupKey:         "main",
				GroupTitle:       "Основные режимы работы:",
				GroupDescription: "Команды основных режимов работы приложения.",
			},
		},
	}
}

// Initiate Функция инициализации компонента и подготовки компонента к запуску.
func (ver *impl) Initiate() (err error) { return }

// Do Выполнение компонента приложения.
func (ver *impl) Do() (levelDone bool, levelExit bool, err error) {
	_, err = fmt.Println(kitModuleCfg.Get().Version().String())
	levelDone, levelExit = true, true

	return
}

// Finalize Функция вызывается перед завершением компонента и приложения в целом.
func (ver *impl) Finalize() (err error) { return }

// Package reg
package reg

import (
	kitModuleCfg "github.com/webnice/kit/v3/module/cfg"
	kitTypes "github.com/webnice/kit/v3/types"
)

// Registration Регистрация компонента приложения.
// Ошибки регистрации накапливаются и отобразятся на следующих уровнях работы приложения.
func Registration(obj kitTypes.Component) {
	var (
		err error
		cf  kitModuleCfg.Interface
		cn  string
	)

	cf = kitModuleCfg.Get()
	// Проверка на ошибку разработчика
	if obj == nil {
		cf.Gist().ErrorAppend(cf.Errors().ComponentIsNull(0))
		return
	}
	// Проверка уровня работы приложения, компоненты регистрируются только на уровне 0
	cn = cf.Gist().ComponentName(obj)
	if cf.Runlevel() > 0 {
		cf.Gist().ErrorAppend(cf.Errors().ComponentRegistrationProhibited(0, cn))
		return
	}
	// Регистрация компоненты
	if err = cf.Gist().Registration(cn, obj); err != nil {
		cf.Gist().ErrorAppend(cf.Errors().ComponentRegistrationError(0, cn, err))
		return
	}
}

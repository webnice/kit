package reg

import (
	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitTypes "github.com/webnice/kit/v4/types"
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
		cf.Gist().ErrorAppend(cf.Errors().ComponentIsNull.Bind())
		return
	}
	// Проверка уровня работы приложения, компоненты регистрируются только на уровне 0
	cn = cf.Gist().ComponentName(obj)
	if cf.Runlevel() > 0 {
		cf.Gist().ErrorAppend(cf.Errors().ComponentRegistrationProhibited.Bind(cn))
		return
	}
	// Регистрация компоненты
	if err = cf.Gist().Registration(cn, obj); err != nil {
		cf.Gist().ErrorAppend(cf.Errors().ComponentRegistrationError.Bind(cn, err))
		return
	}
}

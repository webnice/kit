package server

import (
	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	kitTypesServer "github.com/webnice/kit/v4/types/server"

	"github.com/go-chi/chi/v5"
)

// Prepare Подготовка зарегистрированных ресурсов веб сервера и создание роутинга.
func (iweb *implWeb) Prepare() (err error) {
	// Вызов функции Before.
	iweb.info("Подготовка зарегистрированных ресурсов веб сервера.")
	if err = iweb.callBefore(); err != nil {
		return
	}
	iweb.info("Подготовка зарегистрированных ресурсов веб сервера выполнено.")
	// Создание роутинга ВЕБ сервера.
	iweb.info("Создание роутинга ВЕБ сервера.")
	if err = iweb.makeRouting(); err != nil {
		return
	}
	iweb.info("Создание роутинга ВЕБ сервера выполнено.")

	return
}

// Вызов функции Before.
func (iweb *implWeb) callBefore() (err error) {
	var n, s int

	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = iweb.parent.Errors().
				ModulePanicException(moduleName, e, kitModuleTrace.StackShort())
		}
	}()
	// Запуск функции Before.
	for n = range iweb.res {
		if iweb.res[n].Resource() == nil {
			continue
		}
		if iweb.res[n].Resource().Before == nil {
			continue
		}
		for s = range iweb.parent.server {
			if iweb.parent.server[s].T != kitTypesServer.TWeb {
				continue
			}
			if err = iweb.res[n].
				Resource().
				Before(iweb.parent.server[s]); err != nil {
				err = iweb.parent.Errors().BeforeExitWithError(err)
				return
			}
		}
	}

	return
}

// Создание роутинга ВЕБ сервера.
func (iweb *implWeb) makeRouting() (err error) {
	var (
		grp     map[string][]*kitTypesServer.Web
		res     *kitTypesServer.Web
		pattern string
		n, m    int
	)

	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = iweb.parent.Errors().
				ModulePanicException(moduleName, e, kitModuleTrace.StackShort())
		}
	}()
	iweb.Error().Reset()
	iweb.router = chi.NewRouter()
	// Базовые промежуточные.
	iweb.router.
		Use(iweb.Lib().Middleware().RecoverHandler())
	// Группировка ресурсов по базовому пути URN.
	grp = make(map[string][]*kitTypesServer.Web)
	for n = range iweb.res {
		res = iweb.res[n].Resource()
		grp[res.Path] = append(grp[res.Path], res)
	}
	// Промежуточные без пути роутинга, подключаемые из плагинов.
	for pattern = range grp {
		if pattern == "" {
			for n = range grp[pattern] {
				for m = range grp[pattern][n].Middleware {
					iweb.router.Use(grp[pattern][n].Middleware[m])
				}
			}
		}
	}
	// Создание роутинга для зарегистрированных ресурсов.
	for pattern = range grp {
		if pattern == "" {
			continue
		}
		if err = iweb.addResources(pattern, grp[pattern]); err != nil {
			return
		}
	}

	return
}

// Создание роутинга для зарегистрированных ресурсов.
func (iweb *implWeb) addResources(pattern string, res []*kitTypesServer.Web) (err error) {
	var n, m, c int

	// Группа ресурсов привязанная к одному паттерну пути.
	iweb.router.Route(pattern, func(router chi.Router) {
		for n = range res {
			// Группа для множества контроллеров с общими "промежуточными слоями".
			router.Group(func(rr chi.Router) {
				for m = range res[n].Middleware {
					rr.Use(res[n].Middleware[m])
				}
				for c = range res[n].Controller {
					// Один контроллер со своими "промежуточными слоями".
					rr.Group(func(ctl chi.Router) {
						iweb.makeControllerMiddlewareGroup(ctl, res[n].Controller[c])
					})
				}
			})
		}
	})

	return
}

// Создание группы одного контроллер со своими "промежуточными слоями".
func (iweb *implWeb) makeControllerMiddlewareGroup(router chi.Router, wc kitTypesServer.WebController) {
	var n int

	// Обработчики "промежуточного слоя".
	for n = range wc.Middleware {
		router.Use(wc.Middleware[n])
	}
	// Перечисленные методы запросов.
	for n = range wc.Method {
		router.Method(wc.Method[n].String(), wc.Path, wc.Controller)
	}
	// Для пустого списка методов запросов - все методы.
	if len(wc.Method) == 0 {
		router.HandleFunc(wc.Path, wc.Controller)
	}

	return
}

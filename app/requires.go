package app

import kitTypes "github.com/webnice/kit/v4/types"

// Функция проверки зависимости между компонентами
func (app *impl) requiresFn(components []*kitTypes.ComponentInfo) (err kitTypes.ErrorWithCode) {
	var (
		n, i, j int
		reqOk   bool
	)

	// Просмотр всех зарегистрированных компонентов
	for n = range components {
		if len(components[n].Require) == 0 {
			continue
		}
		// Если есть зависимости
		for i = range components[n].Require {
			// Поиск зависимости
			reqOk = false
			for j = range components {
				if components[n].Require[i].MatchString(components[j].ComponentName) {
					reqOk = true
					break
				}
			}
			// Ошибка, если зависимость не удовлетворена
			if !reqOk {
				err = app.cfg.Errors().
					ComponentRequires(0, components[n].ComponentName, components[n].Require[i].String())
				return
			}
		}
	}

	return
}

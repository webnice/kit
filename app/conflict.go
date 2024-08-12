package app

import kitTypes "github.com/webnice/kit/v4/types"

// Функция проверки конфликтов между компонентами.
func (app *impl) conflictFn(components []*kitTypes.ComponentInfo) (err kitTypes.ErrorWithCode) {
	var n, c, i int

	// Обход всех компонентов.
	for n = range components {
		// Если у компоненты есть список конфликта.
		if len(components[n].Conflict) == 0 {
			continue
		}
		// Тогда проверка по списку.
		for c = range components[n].Conflict {
			// Всех других компонент.
			for i = range components {
				// На вхождение в список конфликта.
				if components[n].Conflict[c].MatchString(components[i].ComponentName) {
					err = app.cfg.Errors().
						ComponentConflict(0, components[n].ComponentName, components[i].ComponentName)
					return
				}
			}
		}
	}

	return
}

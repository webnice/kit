// Package application
package application

import (
	kitTypes "github.com/webnice/kit/v3/types"
)

// Функция проверки конфликтов между компонентами
func (app *impl) conflictFn(components []*kitTypes.ComponentInfo) (err kitTypes.ErrorWithCode) {
	var n, c, i int

	// Обход всех компонент,
	for n = range components {
		// если у компоненты есть список конфликта,
		if len(components[n].Conflict) == 0 {
			continue
		}
		// тогда проверка по списку,
		for c = range components[n].Conflict {
			// всех других компонент,
			for i = range components {
				// на вхождение в список конфликта
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

package app

import (
	"regexp"
	"sort"

	kitTypes "github.com/webnice/kit/v4/types"
)

// Функция сортировки компонентов в соответствии с настройками (before) и (after).
func (app *impl) sortFn(components []*kitTypes.ComponentInfo) (err kitTypes.ErrorWithCode) {
	var (
		fnGetIdx func(*regexp.Regexp) int64
		mp       map[string]int64
		mpLen    int64
		mkIdx    int64
		n, j     int
	)

	// Создание карты индекса сортировки для всех зарегистрированных компонентов.
	mp = make(map[string]int64)
	for n = range components {
		mp[components[n].ComponentName] = 0
	}
	mpLen = int64(len(mp))
	// Функция поиска текущего индекса сортировки компонента.
	fnGetIdx = func(rex *regexp.Regexp) (ret int64) {
		for i := range components {
			if rex.MatchString(components[i].ComponentName) {
				ret = mp[components[i].ComponentName]
			}
		}

		return
	}
	// Увеличение индекса сортировки в соответствии с требованием (after).
	for n = range components {
		for j = range components[n].After {
			mp[components[n].ComponentName] += mpLen + fnGetIdx(components[n].After[j])
		}
	}
	// Уменьшение индекса сортировки в соответствии с требованием (before).
	for n = range components {
		for j = range components[n].Before {
			// Поиск текущего индекса сортировки зависимого компонента.
			mkIdx = fnGetIdx(components[n].Before[j])
			// Уменьшение индекса сортировки зависящего компонента.
			if mp[components[n].ComponentName] >= mkIdx {
				mp[components[n].ComponentName] = mkIdx - 1
			}
		}
	}
	// Сортировка среза зарегистрированных компонентов в соответствии с получившейся картой индекса сортировки.
	sort.Slice(components, func(i, j int) bool {
		return mp[components[i].ComponentName] < mp[components[j].ComponentName]
	})

	return
}

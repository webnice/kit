package app

import (
	"regexp"

	kitTypes "github.com/webnice/kit/v4/types"
)

// Функция сортировки компонентов в соответствии с настройками (before) и (after).
func (app *impl) sortFn(components []*kitTypes.ComponentInfo) (err error) {
	var (
		cnt    int
		n      int
		ok     bool
		before map[int]map[int]bool
		after  map[int]map[int]bool
		order  []int
		sorted []*kitTypes.ComponentInfo
	)

	cnt = len(components)
	// Построение графа зависимостей запуска компонент по правилам (before) и (after).
	before, after = sortEdgesMake(components)
	// Топологическая сортировка с приоритетом правил (after) над правилами (before).
	order, ok = sortKahn(cnt, after, before)
	if !ok {
		// Разрешение конфликта правил: повторная сортировка только по правилам (after), имеющим приоритет.
		order, ok = sortKahn(cnt, after, nil)
	}
	if !ok {
		// Циклическая зависимость в правилах (after) — сортировка невозможна.
		err = app.cfg.Errors().ComponentSortCyclicDependency.
			Bind(components[order[0]].ComponentName, components[order[1]].ComponentName)

		return
	}
	// Перестановка компонент в соответствии с полученным порядком запуска.
	sorted = make([]*kitTypes.ComponentInfo, cnt)
	for n = range order {
		sorted[n] = components[order[n]]
	}
	copy(components, sorted)

	return
}

// Построение графа зависимостей запуска компонент по правилам (before) и (after).
// Ключ первого уровня - индекс компоненты, которая должна запуститься раньше, значение - множество индексов
// компонент, запускаемых после неё.
func sortEdgesMake(components []*kitTypes.ComponentInfo) (before, after map[int]map[int]bool) {
	var (
		n, i, j int
		rex     *regexp.Regexp
	)

	before, after = make(map[int]map[int]bool), make(map[int]map[int]bool)
	for n = range components {
		// Ребро от компоненты к компонентам, которые должны запуститься после неё по правилу (before).
		for j = range components[n].Before {
			rex = components[n].Before[j]
			for i = range components {
				// Пропуск самой компоненты, правило не создаёт зависимости компоненты от самой себя.
				if i == n || !rex.MatchString(components[i].ComponentName) {
					continue
				}
				if before[n] == nil {
					before[n] = make(map[int]bool)
				}
				before[n][i] = true
			}
		}
		// Ребро от компонент, которые должны запуститься раньше по правилу (after), к компоненте.
		for j = range components[n].After {
			rex = components[n].After[j]
			for i = range components {
				if i == n || !rex.MatchString(components[i].ComponentName) {
					continue
				}
				if after[i] == nil {
					after[i] = make(map[int]bool)
				}
				after[i][n] = true
			}
		}
	}

	return
}

// Топологическая сортировка методом Кана. primary и secondary - множества рёбер графа зависимостей,
// рёбра из secondary учитываются только при отсутствии циклической зависимости с рёбрами из primary.
// Возвращает порядок индексов компонент и истину; если сортировка невозможна из-за циклической зависимости,
// возвращает ложь и список индексов компонент, оставшихся в цикле.
func sortKahn(cnt int, primary, secondary map[int]map[int]bool) (ret []int, ok bool) {
	var (
		indegree []int
		done     []bool
		n, i     int
		next     int
	)

	// Подсчёт количества входящих зависимостей для каждой компоненты.
	indegree = make([]int, cnt)
	for n = range primary {
		for i = range primary[n] {
			indegree[i]++
		}
	}
	if secondary != nil {
		for n = range secondary {
			for i = range secondary[n] {
				indegree[i]++
			}
		}
	}
	done = make([]bool, cnt)
	ret = make([]int, 0, cnt)
	// Обход компонент: на каждом шаге выбирается первая по порядку регистрации компонента,
	// у которой отсутствуют неразрешённые входящие зависимости.
	for {
		next = -1
		for n = 0; n < cnt; n++ {
			if !done[n] && indegree[n] == 0 {
				next = n

				break
			}
		}
		// Отсутствие компонент без входящих зависимостей означает циклическую зависимость.
		if next < 0 {
			break
		}
		done[next] = true
		ret = append(ret, next)
		// Уменьшение количества входящих зависимостей у компонент, следующих за текущей.
		for i = range primary[next] {
			indegree[i]--
		}
		if secondary != nil {
			for i = range secondary[next] {
				indegree[i]--
			}
		}
	}
	if len(ret) == cnt {
		ok = true

		return
	}
	// Формирование списка индексов компонент, оставшихся в циклической зависимости.
	ret = ret[:0]
	for n = 0; n < cnt; n++ {
		if !done[n] {
			ret = append(ret, n)
		}
	}

	return
}

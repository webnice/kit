package cfg

import (
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/webnice/dic"
	kitModuleCfgCli "github.com/webnice/kit/v5/module/cfg/cli"
	kitTypes "github.com/webnice/kit/v5/types"
)

// ComponentName Получение уникального имени пакета компоненты.
func (essence *gist) ComponentName(obj any) (ret string) {
	var rt reflect.Type

	if rt = reflect.TypeOf(obj); rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	ret = rt.PkgPath()

	return
}

// ComponentNames Возвращает список зарегистрированных компонентов.
func (essence *gist) ComponentNames() (ret []string) {
	var c *kitTypes.ComponentInfo

	if !essence.parent.comp.IsClose {
		return
	}
	essence.parent.comp.ComponentMutex.Lock()
	for _, c = range essence.parent.comp.Component {
		ret = append(ret, c.ComponentName)
	}
	essence.parent.comp.ComponentMutex.Unlock()

	return
}

// ComponentPreferences Функция-менеджер загрузки и обработки настроек компонентов.
//
// Функция переносит временный список зарегистрированных компонентов Registered в основной список Component.
// Список Registered копируется под ComponentMutex, после чего пользовательская функция fn вызывается вне блокировки.
// Это позволяет fn безопасно обращаться к ComponentNames() и другим gist-методам. Результаты фиксируются
// в Component только после успешной обработки всех элементов снапшота; обработанные элементы удаляются из Registered
// по указателю, чтобы не потерять возможные новые регистрации, появившиеся во время работы fn.
func (essence *gist) ComponentPreferences(fn kitTypes.ComponentPreferencesFn) (code uint8, err error) {
	var (
		ierr       dic.IError
		registered []*kitTypes.ComponentInfo
		results    []*kitTypes.ComponentInfo
		src, res   *kitTypes.ComponentInfo
		n, i       int
	)

	essence.parent.comp.ComponentMutex.Lock()
	if !essence.parent.comp.IsClose {
		essence.parent.comp.ComponentMutex.Unlock()
		code = essence.parent.Errors().ComponentPreferencesCallBeforeCompleting.CodeU8().Get()
		essence.ErrorAppend(essence.parent.Errors().ComponentPreferencesCallBeforeCompleting.Bind())
		return
	}
	registered = append(registered, essence.parent.comp.Registered...)
	essence.parent.comp.ComponentMutex.Unlock()

	for _, src = range registered {
		if res, err = fn(src); err != nil {
			if ierr = Errors().Unbind(err); ierr != nil {
				code = ierr.CodeU8().Get()
			}
			return
		}
		results = append(results, res)
	}

	essence.parent.comp.ComponentMutex.Lock()
	defer essence.parent.comp.ComponentMutex.Unlock()
	essence.parent.comp.Component = append(essence.parent.comp.Component, results...)
	// Удаляются только те компоненты, которые были обработаны в этом проходе. Если во время обработки
	// пользовательским кодом появились новые регистрации, они сохраняются.
	for n = range registered {
		for i = range essence.parent.comp.Registered {
			if essence.parent.comp.Registered[i] == registered[n] {
				essence.parent.comp.Registered = append(essence.parent.comp.Registered[:i], essence.parent.comp.Registered[i+1:]...)
				break
			}
		}
	}

	return
}

// ComponentCheckConflict Проверка конфликтов между всеми зарегистрированными компонентами.
//
// Передаётся снапшот списка Component, поэтому пользовательская функция fn не вызывается с удержанием ComponentMutex.
// Снапшот защищает от self-deadlock и от паники container/list при параллельном чтении компонентов.
func (essence *gist) ComponentCheckConflict(fn kitTypes.ComponentConflictFn) (code uint8, err error) {
	var (
		component = essence.componentSnapshot()
		ierr      dic.IError
	)

	if err = fn(component); err != nil {
		if ierr = Errors().Unbind(err); ierr != nil {
			code = ierr.CodeU8().Get()
		}
		return
	}

	return
}

// ComponentRequiresCheck Проверка зависимостей между всеми зарегистрированными компонентами.
//
// Передаётся снапшот списка Component, поэтому пользовательская функция fn не вызывается с удержанием ComponentMutex.
// Проверка должна вернуть ошибку с кодом из справочника cfg.Errors, чтобы дирижёр мог извлечь код возврата.
func (essence *gist) ComponentRequiresCheck(fn kitTypes.ComponentRequiresFn) (code uint8, err error) {
	var (
		component = essence.componentSnapshot()
		ierr      dic.IError
	)

	if err = fn(component); err != nil {
		if ierr = Errors().Unbind(err); ierr != nil {
			code = ierr.CodeU8().Get()
		}
		return
	}

	return
}

// ComponentSort Сортировка зарегистрированных компонентов в соответствии с настройками (before) и (after).
//
// Функция сортирует снапшот указателей на компоненты вне ComponentMutex и фиксирует новый порядок в Component
// только после успешного завершения fn. Если fn вернула ошибку, исходный порядок компонентов не изменяется.
func (essence *gist) ComponentSort(fn kitTypes.ComponentSortFn) (code uint8, err error) {
	var (
		component = essence.componentSnapshot()
		ierr      dic.IError
	)

	if err = fn(component); err != nil {
		if ierr = Errors().Unbind(err); ierr != nil {
			code = ierr.CodeU8().Get()
		}
		return
	}
	essence.parent.comp.ComponentMutex.Lock()
	essence.parent.comp.Component = component
	essence.parent.comp.ComponentMutex.Unlock()

	return
}

// ComponentMapRunlevel Построение шагов переключения уровня выполнения приложения (runlevel).
func (essence *gist) ComponentMapRunlevel(targetlevel uint16) (code uint8, err error) {
	var (
		component []*kitTypes.ComponentInfo
		n         int
		l         uint16
		tmp       map[uint16]bool
		mapLevel  []uint16
	)

	component = essence.componentSnapshot()
	tmp = make(map[uint16]bool)
	tmp[defaultRunlevel], tmp[targetlevel], tmp[math.MaxUint16] = true, true, true
	// Обход всех зарегистрированных компонентов, получение уровня отличного от нуля
	for n = range component {
		if component[n].Runlevel != 0 {
			tmp[component[n].Runlevel] = true
		}
	}
	mapLevel = make([]uint16, 0, len(tmp))
	// Копирование получившейся карты из map в массив.
	for l = range tmp {
		mapLevel = append(mapLevel, l)
	}
	// Сортировка по значению уровня выполнения приложения.
	sort.Slice(mapLevel, func(i, j int) bool { return mapLevel[i] < mapLevel[j] })
	essence.parent.runLevelMutex.Lock()
	essence.parent.mapRunLevel = mapLevel
	essence.parent.runLevelMutex.Unlock()

	return
}

// возвращает копию списка компонентов для безопасного вызова пользовательского кода вне блокировки.
//
// Блокировка удерживается только на время копирования среза указателей. Сами объекты ComponentInfo не копируются:
// пользовательская функция может изменять их поля состояния (IsInitiate, IsDo, IsFinalize, Runlevel) по указателю
func (essence *gist) componentSnapshot() (ret []*kitTypes.ComponentInfo) {
	essence.parent.comp.ComponentMutex.Lock()
	ret = append(ret, essence.parent.comp.Component...)
	essence.parent.comp.ComponentMutex.Unlock()

	return
}

// ComponentInitiate Вызов функции Initiate у всех зарегистрированных компонентов в прямом порядке.
//
// Список компонентов копируется под ComponentMutex, после чего fn вызывается вне блокировки. Пропускаются отключённые
// компоненты (IsDisable) и уже инициализированные компоненты (IsInitiate). При ошибке выполнение останавливается,
// а код ошибки извлекается через cfg.Errors().Unbind.
func (essence *gist) ComponentInitiate(fn kitTypes.ComponentInitiateFn) (code uint8, err error) {
	var (
		component []*kitTypes.ComponentInfo
		ierr      dic.IError
		n         int
	)

	component = essence.componentSnapshot()
	for n = range component {
		// Пропускаем отключённые компоненты.
		if component[n].IsDisable {
			continue
		}
		// Пропускаем уже проинициализированные компоненты.
		if component[n].IsInitiate {
			continue
		}
		if err = fn(component[n]); err != nil {
			if ierr = Errors().Unbind(err); ierr != nil {
				code = ierr.CodeU8().Get()
			}
			break
		}
	}

	return
}

// ComponentDoFindByCommand Функция поиска компоненты по команде.
func (essence *gist) ComponentDoFindByCommand(n int) (ret bool) {
	var i int

	for i = range essence.parent.comp.Component[n].Command {
		// Компонента подходит если у неё полностью совпадает команда, а так же, если у компоненты есть
		// пустая команда, тогда она тоже запускается как для своей команды, так и для всех других команд.
		if strings.EqualFold(essence.parent.comp.Component[n].Command[i], essence.parent.Command()) ||
			essence.parent.comp.Component[n].Command[i] == "" {
			ret = true
		}
	}
	// Если у компоненты вообще нет команд, то она подходит для всех команд.
	if len(essence.parent.comp.Component[n].Command) == 0 {
		ret = true
	}

	return
}

// ComponentDo Вызов функции Do у всех зарегистрированных компонентов в прямом порядке для указанного уровня приложения.
//
// Список компонентов копируется под ComponentMutex, пользовательская функция fn вызывается вне блокировки.
// Фильтрация выполняется по флагам IsDisable/IsInitiate/IsDo, по runlevel компоненты, по команде приложения и режиму
// forkWorker. Компоненты с Runlevel=0 получают текущий уровень перед вызовом fn и помечаются IsDo=true до вызова fn.
func (essence *gist) ComponentDo(runlevel uint16, fn kitTypes.ComponentDoFn) (code uint8, err error) {
	var (
		bcfw      *kitTypes.BootstrapConfigurationForkWorker
		component []*kitTypes.ComponentInfo
		ierr      dic.IError
		n         int
	)

	bcfw = essence.parent.ForkWorker()
	component = essence.componentSnapshot()
	for n = range component {
		// Пропускаем отключённые компоненты.
		if component[n].IsDisable {
			continue
		}
		// Пропускаем не инициализированные компоненты.
		if !component[n].IsInitiate {
			continue
		}
		// Пропускаем ранее запущенные компоненты.
		if component[n].IsDo {
			continue
		}
		// Пропускаем компоненты с уровнем запуска не равному текущему уровню работы приложения.
		if component[n].Runlevel > 0 && component[n].Runlevel != runlevel {
			continue
		}
		// Выбор на основе флага isForkWorker.
		switch essence.parent.isForkWorker {
		case true:
			switch {
			// Режим forkWorker и указанная компонента, имеет приоритет над командой приложения.
			case bcfw.Component != "":
				// Запускается только компонента указанная для запуска.
				if !strings.EqualFold(bcfw.Component, component[n].ComponentName) {
					continue
				}
			default:
				// Если приложение запущено с командой, поиск компоненты с указанной командой, пропуск остальных.
				if essence.parent.Command() != "" && !essence.ComponentDoFindByCommand(n) {
					continue
				}
			}
		default:
			// Если приложение запущено с командой, поиск компоненты с указанной командой, пропуск остальных.
			if essence.parent.Command() != "" && !essence.ComponentDoFindByCommand(n) {
				continue
			}
		}
		// Если у компоненты уровень по умолчанию (=0), присваиваем уровень на котором был выполнен запуск компоненты.
		if component[n].Runlevel == 0 {
			component[n].Runlevel = runlevel
		}
		// Помечаем то что компонента была запущена.
		component[n].IsDo = true
		// Выполнение запуска компоненты с использованием внешней функции запуска.
		if err = fn(component[n]); err != nil {
			if ierr = Errors().Unbind(err); ierr != nil {
				code = ierr.CodeU8().Get()
			}
			break
		}
	}

	return
}

// ComponentFinalizeWarningTimeout Возвращает время отводимое на выполнение функции Finalize(), до печати в лог
// сообщения о долгой работе функции.
func (essence *gist) ComponentFinalizeWarningTimeout() (ret time.Duration) {
	ret = essence.parent.finalizeWarningTimeout
	return
}

// ComponentFinalize Вызов функции Finalize у всех зарегистрированных компонентов в обратном порядке.
//
// Список компонентов копируется под ComponentMutex, после чего fn вызывается вне блокировки. Обход выполняется
// от последнего компонента к первому; пропускаются отключённые, не инициализированные, не запущенные и уже
// финализированные компоненты. При ошибке выполнение останавливается, код ошибки извлекается через cfg.Errors().Unbind.
func (essence *gist) ComponentFinalize(fn kitTypes.ComponentFinalizeFn) (code uint8, err error) {
	var (
		component []*kitTypes.ComponentInfo
		ierr      dic.IError
		n         int
	)

	component = essence.componentSnapshot()
	for n = len(component) - 1; n >= 0; n-- {
		// Пропускаем отключённые компоненты.
		if component[n].IsDisable {
			continue
		}
		// Пропускаем не инициализированные компоненты.
		if !component[n].IsInitiate {
			continue
		}
		// Пропускаем не запущенные компоненты.
		if !component[n].IsDo {
			continue
		}
		// Пропускаем ранее "финализированные" компоненты.
		if component[n].IsFinalize {
			continue
		}
		if err = fn(component[n]); err != nil {
			if ierr = Errors().Unbind(err); ierr != nil {
				code = ierr.CodeU8().Get()
			}
			break
		}
	}

	return
}

// ComponentCommandRegister Регистрация команды и группы команд компоненты.
func (essence *gist) ComponentCommandRegister(cc kitTypes.ComponentCommand) {
	var (
		groupKey string
		foundPos int
		n        int
	)

	groupKey, foundPos = cc.GroupKey, -1
	for n = range essence.parent.comp.Groups {
		if strings.EqualFold(essence.parent.comp.Groups[n].Key, groupKey) {
			foundPos = n
			break
		}
	}
	if foundPos >= 0 {
		if essence.parent.comp.Groups[foundPos].Title == "" {
			essence.parent.comp.Groups[foundPos].Title = cc.GroupTitle
		}
		if essence.parent.comp.Groups[foundPos].Description == "" {
			essence.parent.comp.Groups[foundPos].Description = cc.GroupDescription
		}
	} else {
		essence.parent.comp.Groups = append(essence.parent.comp.Groups, &kitTypes.CommandGroup{
			Key:         cc.GroupKey,
			Title:       cc.GroupTitle,
			Description: cc.GroupDescription,
		})
	}
	essence.parent.cli.RegisterCommand(&kitModuleCfgCli.Command{
		Command:     cc.Command,
		Description: cc.Description,
		GroupKey:    cc.GroupKey,
		IsDefault:   cc.IsDefault,
		IsHidden:    cc.IsHidden,
		Value:       cc.Value,
	})
}

// ComponentFlagRegister Регистрация глобального флага компоненты приложения.
func (essence *gist) ComponentFlagRegister(cf kitTypes.ComponentFlag) {
	essence.parent.cli.RegisterFlag(&kitModuleCfgCli.Flag{
		ShortKey:    cf.ShortKey,
		Flag:        cf.Flag,
		Description: cf.Description,
		Environment: cf.Environment,
		Placeholder: cf.Placeholder,
		IsRequired:  cf.IsRequired,
		IsHidden:    cf.IsHidden,
		Value:       cf.Value,
	})
}

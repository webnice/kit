package cfg

import (
	"container/list"

	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
)

// Функция создаёт канал изменения уровня работы приложения.
// Запускает поток, выполняющий изменение уровня работы приложения и вызов функций, подписанных на события изменения
// уровня работы приложения.
func runlevelChangeFuncNew() (ret chan *runLevelUp) {
	ret = make(chan *runLevelUp)
	go singleton.runlevelChangeFunc(ret)

	return
}

// Поток изменения уровня работы приложения и вызов функций подписчиков на события изменения уровня работы приложения.
func (cfg *impl) runlevelChangeFunc(ch chan *runLevelUp) {
	var (
		msg                *runLevelUp
		oldLevel, newLevel uint16
		isEqual            bool
	)

	// Цикл чтения канала, цикл завершится с закрытием канала.
	for msg = range ch {
		newLevel = msg.newLever
		cfg.runLevelMutex.RLock()
		isEqual = newLevel == cfg.runLevel
		cfg.runLevelMutex.RUnlock()
		if isEqual {
			msg.done <- struct{}{}
			continue
		}
		// Закрытие регистрации компонентов.
		if !cfg.comp.IsClose && newLevel > 0 {
			cfg.comp.IsClose = true
		}
		// Применение изменения уровня работы приложения.
		cfg.runLevelMutex.Lock()
		oldLevel, cfg.runLevel = cfg.runLevel, newLevel
		cfg.runLevelMutex.Unlock()
		// Для режима отладки, печать в лог информации о переключении режима работы приложения в состояние завершения.
		if cfg.Debug() && oldLevel != defaultRunlevelExit && newLevel == defaultRunlevelExit {
			cfg.Log().Info(tplRunlevelExit)
		}
		// Вызов функций подписчиков на изменение уровня работы приложения.
		//
		// Список подписчиков копируется под runLevelSubscribersMutex, после чего функции вызываются вне блокировки.
		// Это допускает подписку/отписку из callback-ов и защищает container/list от параллельной модификации.
		cfg.runlevelCallSubscribers(oldLevel, newLevel)
		// Подтверждение изменения уровня работы приложения.
		msg.done <- struct{}{}
	}
}

// Вызов функций подписчиков на изменение уровня работы приложения.
func (cfg *impl) runlevelCallSubscribers(oldLevel, newLevel uint16) {
	var (
		err        error
		item       RunlevelSubscriberFn
		subscriber []RunlevelSubscriberFn
	)

	cfg.runLevelSubscribersMutex.RLock()
	for element := cfg.runLevelSubscribers.Front(); element != nil; element = element.Next() {
		subscriber = append(subscriber, element.Value.(RunlevelSubscriberFn))
	}
	cfg.runLevelSubscribersMutex.RUnlock()

	for n := range subscriber {
		item = subscriber[n]
		if err = cfg.runlevelSafeCallSubscriber(item, oldLevel, newLevel); err != nil {
			cfg.Gist().ErrorAppend(err)
		}
	}
}

// Безопасный вызов функции обратного вызова подписчика на изменения уровня работы приложения.
func (cfg *impl) runlevelSafeCallSubscriber(fn RunlevelSubscriberFn, oldLevel uint16, newLevel uint16) (
	err error,
) {
	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = cfg.Errors().RunlevelSubscriptionPanicException.Bind(e, kitModuleTrace.StackShort())
		}
	}()
	// Вызов функции подписчика.
	fn(oldLevel, newLevel)

	return
}

// Runlevel Возвращает текущее значение уровня работы приложения.
func (cfg *impl) Runlevel() (ret uint16) {
	cfg.runLevelMutex.RLock()
	ret = cfg.runLevel
	cfg.runLevelMutex.RUnlock()

	return
}

// RunlevelMap Возвращает карту, описывающую план переключения уровней работы приложения.
func (cfg *impl) RunlevelMap() (ret []uint16) {
	cfg.runLevelMutex.RLock()
	ret = append(ret, cfg.mapRunLevel...)
	cfg.runLevelMutex.RUnlock()

	return
}

// RunlevelSubscribe Подписка на события изменения уровня работы приложения.
//
// Подписчик идентифицируется полным именем функции. Повторная подписка той же функции запрещена и возвращает ошибку
// RunlevelAlreadySubscribedFunction. Список подписчиков защищён отдельным mutex.
func (cfg *impl) RunlevelSubscribe(fn RunlevelSubscriberFn) (err error) {
	var (
		element      *list.Element
		item         RunlevelSubscriberFn
		found        bool
		funcFullName string
	)

	// Проверка на ошибку разработчика
	if fn == nil {
		err = cfg.Errors().RunlevelSubscribeUnsubscribeNilFunction.Bind()
		return
	}
	funcFullName = getFuncFullName(fn)
	cfg.runLevelSubscribersMutex.Lock()
	defer cfg.runLevelSubscribersMutex.Unlock()
	// Поиск подписчика среди существующих подписчиков
	for element = cfg.runLevelSubscribers.Front(); element != nil; element = element.Next() {
		item = element.Value.(RunlevelSubscriberFn)
		if getFuncFullName(item) == funcFullName {
			found = true
			break
		}
	}
	// Если подписчик уже существует, возвращение ошибки
	if found {
		err = cfg.Errors().RunlevelAlreadySubscribedFunction.Bind(funcFullName)
		return
	}
	cfg.runLevelSubscribers.PushBack(fn)

	return
}

// RunlevelUnsubscribe Отписка от событий изменения уровня работы приложения.
//
// Удаляются все элементы списка подписчиков, чьё полное имя функции совпадает с именем fn. Если подписчик не найден,
// возвращается ошибка RunlevelSubscriptionNotFound. Операция потокобезопасна.
func (cfg *impl) RunlevelUnsubscribe(fn RunlevelSubscriberFn) (err error) {
	var (
		element      *list.Element
		item         RunlevelSubscriberFn
		forDelete    []*list.Element
		funcFullName string
		n            int
	)

	// Проверка на ошибку разработчика
	if fn == nil {
		err = cfg.Errors().RunlevelSubscribeUnsubscribeNilFunction.Bind()
		return
	}
	funcFullName = getFuncFullName(fn)
	cfg.runLevelSubscribersMutex.Lock()
	defer cfg.runLevelSubscribersMutex.Unlock()
	// Поиск подписчика среди существующих подписчиков
	for element = cfg.runLevelSubscribers.Front(); element != nil; element = element.Next() {
		item = element.Value.(RunlevelSubscriberFn)
		if getFuncFullName(item) == funcFullName {
			forDelete = append(forDelete, element)
		}
	}
	// Подписчик не нашёлся
	if len(forDelete) == 0 {
		err = cfg.Errors().RunlevelSubscriptionNotFound.Bind(funcFullName)
		return
	}
	// Удаление подписчика
	for n = range forDelete {
		cfg.runLevelSubscribers.Remove(forDelete[n])
	}

	return
}

// RunlevelAutoincrement Режим работы автоматического увеличения уровня работы приложения.
func (cfg *impl) RunlevelAutoincrement() (ret bool) {
	cfg.runLevelMutex.RLock()
	ret = !cfg.runLevelStopAutoincrement
	cfg.runLevelMutex.RUnlock()

	return
}

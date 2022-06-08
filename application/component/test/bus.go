// Package test
package test

import (
	"fmt"
	"runtime"
	"time"

	kitModuleCfg "github.com/webnice/kit/module/cfg"
	kitModuleTrace "github.com/webnice/kit/module/trace"
)

func (tst *impl) subscribeDatabuser() {
	var err error

	if err = tst.cfg.Bus().Subscribe(tst); err != nil {
		switch eto := err.(type) {
		case kitModuleCfg.Err:
			tst.cfg.Gist().ErrorAppend(eto)
		default:
			tst.cfg.Gist().ErrorAppend(tst.cfg.Errors().ApplicationUnknownError(0, eto))
		}
		return
	}
}

func (tst *impl) unsubscribeDatabuser() {
	var err error

	if err = tst.cfg.Bus().Unsubscribe(tst); err != nil {
		switch eto := err.(type) {
		case kitModuleCfg.Err:
			tst.cfg.Gist().ErrorAppend(eto)
		default:
			tst.cfg.Gist().ErrorAppend(tst.cfg.Errors().ApplicationUnknownError(0, eto))
		}
	}
}

// Consumer Функция получения данных из шины данных.
// В функцию передаются объекты, типы структур которых были получены через вызов функции KnownType()
// Для синхронного вызова, функция должна вернуть ответа, он будет передан издателю.
// Для асинхронного вызова, функция не должна возвращать никакие данные, ничего не будет передаваться издателю.
func (tst *impl) Consumer(isSync bool, data interface{}) (ret []interface{}, errs []error) {
	//switch isSync {
	//case true:
	//	ret = append(ret, isSync)
	//	ret = append(ret, data)
	//	ret = append(ret, time.Now().String())
	//default:
	//	ret = append(ret, time.Now().String())
	//}

	return
}

// KnownType Функция вызывается один раз, при регистрации подписчика в шине данных и должна вернуть срез структур
// данных, которые готова получать. Для получения данных любого типа, необходимо вернуть срез нулевой длинны.
func (tst *impl) KnownType() (ret []interface{}) {

	t1 := new(Configuration)
	t2 := new(Psql)
	ret = append(ret, t1, t2)

	return
}

// Тестирование издателя.
func (tst *impl) publisher() {
	const tplPanic = `Работа с подпиской потребителя, в шине данных, прервана паникой:` + "\n%v\n%s."
	var (
		err     error
		errs    []error
		rto     time.Time
		eto     time.Duration
		average time.Duration
		n       int
	)
	var maxCount = 2000000

	defer func() {
		if e := recover(); e != nil {
			tst.log().Fatalf(tplPanic, e, kitModuleTrace.StackShort())
		}
	}()

	_ = maxCount
	_, _ = err, errs
	_, _, _, _ = rto, eto, average, n

	<-time.After(time.Second * 3)
	tst.log().Noticef("Запущен обмен данными через шину данных.")

	tst.cfg.Bus().PublishSync(&Configuration{TestExotic: "Тест синхронный"})
	if err = tst.cfg.Bus().PublishAsync(&Configuration{TestExotic: "Тест асинхронный"}); err != nil {
		tst.log().Errorf("ошибка: %s", err)
	}
	// Небольшое издевательство, для тестирования стабильности.
	tst.log().Noticef("Остановка шины данных.")
	tst.cfg.Bus().Gist().WorkerStop()
	tst.log().Noticef("Запуск шины данных.")
	tst.cfg.Bus().Gist().WorkerStart(0)
	tst.log().Noticef("Шина данных запущена.")

	for loop := 0; loop < 10; loop++ {
		average = 0

		// Выполняется 1 000 000 асинхронных отправок данных в шину данных с ожиданием результата.
		// Замеряется среднее время запросов.
		average = 0
		for n = 0; n < maxCount; n++ {
			rto = time.Now()
			cfg := &Configuration{
				TestExotic: fmt.Sprintf("Проверка шины данных %d.", n),
			}
			if err := tst.cfg.Bus().PublishAsync(cfg); err != nil {
				tst.log().Errorf("Ошибка передачи данных в шину данных: %s", err)
				return
			}
			eto = time.Since(rto)
			average += eto
		}

		average = average / time.Duration(maxCount)
		tst.log().Debugf("Среднее время одного асинхронного запроса %q, для %d запросов.", average, maxCount)
		//runtime.Gosched()
		//<-time.After(time.Second * 4)
		//fmt.Printf("Создано: %d, Уничтожено: %d\n", bus.ConstructorWrapperData, bus.DestructorWrapperData)
		//fmt.Printf("Получено: %d, Отдано: %d\n", bus.WrapperDataGet, bus.WrapperDataPut)

		// Выполняется 1 000 000 синхронных отправок данных в шину данных с ожиданием результата.
		// Замеряется среднее время запросов.
		for n = 0; n < maxCount; n++ {
			rto = time.Now()
			cfg := &Configuration{
				TestExotic: fmt.Sprintf("Проверка шины данных %d.", n),
			}
			if _, errs = tst.cfg.Bus().PublishSyncWithTimeout(time.Second*5, cfg); len(errs) > 0 {
				tst.log().Errorf("Ошибка передачи данных в шину данных: %s", errs[0])
				return
			}
			eto = time.Since(rto)
			average += eto
		}

		average = average / time.Duration(maxCount)
		tst.log().Debugf("Среднее время одного синхронного запроса %q, для %d запросов.", average, maxCount)
		//runtime.Gosched()
		//<-time.After(time.Second * 4)
		//fmt.Printf("Создано: %d, Уничтожено: %d\n", bus.ConstructorWrapperData, bus.DestructorWrapperData)
		//fmt.Printf("Получено: %d, Отдано: %d\n", bus.WrapperDataGet, bus.WrapperDataPut)

		// Небольшое издевательство, для тестирования стабильности.
		//tst.log().Noticef("Остановка шины данных.")
		//tst.cfg.Bus().Gist().WorkerStop()
		//tst.log().Noticef("Запуск шины данных.")
		//tst.cfg.Bus().Gist().WorkerStart(0)
		//tst.log().Noticef("Шина данных запущена.")
	}

	runtime.Gosched()
	<-time.After(time.Second * 4)
	//fmt.Printf("Создано: %d, Уничтожено: %d\n", bus.ConstructorWrapperData, bus.DestructorWrapperData)
	//fmt.Printf("Получено: %d, Отдано: %d\n", bus.WrapperDataGet, bus.WrapperDataPut)
	//tst.log().Debugf("Создано: %d, Уничтожено: %d", bus.ConstructorWrapperData, bus.DestructorWrapperData)
	//tst.log().Debugf("Получено: %d, Отдано: %d", bus.WrapperDataGet, bus.WrapperDataPut)

	//runtime.Gosched()
	//<-time.After(time.Second * 4)
	//fmt.Printf("Создано: %d, Уничтожено: %d\n", bus.ConstructorWrapperData, bus.DestructorWrapperData)
	//fmt.Printf("Получено: %d, Отдано: %d\n", bus.WrapperDataGet, bus.WrapperDataPut)
	////tst.log().Debugf("Создано: %d, Уничтожено: %d", bus.ConstructorWrapperData, bus.DestructorWrapperData)
	////tst.log().Debugf("Получено: %d, Отдано: %d", bus.WrapperDataGet, bus.WrapperDataPut)

	//tst.log().Noticef("Остановлен обмен данными через шину данных.")
	//<-time.After(time.Hour * 24)

	// Выход.
	tst.cfg.Gist().RunlevelExitAsync()
	tst.log().Noticef("Процесс обмена данными через шину данных завершён.")
}

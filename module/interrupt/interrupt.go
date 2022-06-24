// Package interrupt
package interrupt

import (
	"os"
	"os/signal"
	"runtime"
	runtimeDebug "runtime/debug"
	"syscall"

	kitModuleCfg "github.com/webnice/kit/v3/module/cfg"
	kitTypes "github.com/webnice/kit/v3/types"
)

// New Создание объекта и возвращение интерфейса.
func New(fn CatchFn) Interface {
	var itp = &impl{
		doExitUp:   make(chan struct{}, 1),
		doExitDone: make(chan struct{}, 1),
		catch:      fn,
		signal:     make(chan os.Signal, 100),
	}
	itp.isRun.Store(false)
	runtime.SetFinalizer(itp, destructor)

	return itp
}

func destructor(itp *impl) {
	defer func() { _ = recover() }()
	close(itp.signal)
	close(itp.doExitDone)
	close(itp.doExitUp)
}

// Log Ссылка на менеджер логирования для использования внутри компоненты.
func (i7t *impl) Log() kitTypes.Logger { return kitModuleCfg.Get().Log() }

// Start Запуск перехвата сигналов прерывания.
func (i7t *impl) Start() Interface {
	defer func() {
		if e := recover(); e != nil {
			i7t.Log().Criticalf("выполнение прервано паникой:\n%v\n%s", e.(error), string(runtimeDebug.Stack()))
		}
	}()
	// Если горутина перехвата работает, тогда второй раз старт не вызывается.
	if i7t.isRun.Load().(bool) {
		return i7t
	}
	i7t.isRun.Store(true)
	signal.Notify(i7t.signal,
		syscall.SIGABRT,
		syscall.SIGALRM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	go i7t.Catcher()

	return i7t
}

// Stop Остановка перехвата сигналов прерывания.
func (i7t *impl) Stop() Interface {
	// Если горутина перехвата не работает, тогда останавливать нечего.
	if !i7t.isRun.Load().(bool) {
		return i7t
	}
	signal.Stop(i7t.signal)
	// Выход после того как убедимся что горутина остановилась.
	i7t.doExitUp <- struct{}{}
	<-i7t.doExitDone

	return i7t
}

// Catcher Получаем сигналы, вызываем функцию.
func (i7t *impl) Catcher() {
	var (
		end bool
		sig os.Signal
	)

	for {
		if end {
			break
		}
		select {
		case <-i7t.doExitUp:
			end = true
		case sig = <-i7t.signal:
			if i7t.catch != nil {
				i7t.catch(sig)
			}
		}
	}
	i7t.doExitDone <- struct{}{}
}

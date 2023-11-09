package interrupt

import (
	"os"
	"sync/atomic"
)

// Interface is an interface of package.
type Interface interface {
	// Start Запуск перехвата внешних прерываний.
	Start() Interface

	// Stop Остановка перехвата внешних прерываний.
	Stop() Interface
}

// impl is an implementation of package.
type impl struct {
	doExitUp   chan struct{}  // Канал сигнала о завершении горутины.
	doExitDone chan struct{}  // Канал сигнала "горутина завершена".
	catch      CatchFn        // Функция вызывается при перехвате сигнала os.
	signal     chan os.Signal // Канал поступления сигналов от os.
	isRun      atomic.Value   // Флаг - горутина работает.
}

// CatchFn Описание функции вызываемой при поступлении прерывания от OS.
type CatchFn func(os.Signal)

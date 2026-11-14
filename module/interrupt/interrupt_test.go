package interrupt

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestImpl_StopResetState(t *testing.T) {
	const errStopHang = "повторный вызов Stop() или перезапуск после Stop() привели к зависанию."
	var (
		i7t  Interface
		done chan struct{}
	)

	i7t = New(nil)
	done = make(chan struct{})
	go func() {
		i7t.Start().Stop().Start().Stop().Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal(errStopHang)
	}
}

func TestImpl_Destructor(t *testing.T) {
	var res cleanupResource

	res = cleanupResource{
		doExitUp:   make(chan struct{}),
		doExitDone: make(chan struct{}),
		signal:     make(chan os.Signal),
	}
	destructor(res)
	destructor(res)
}

func TestImpl_Log(t *testing.T) {
	const errLogNil = "Log() вернул nil."
	var i7t Interface

	i7t = New(nil)
	if i7t.(*impl).Log() == nil {
		t.Fatal(errLogNil)
	}
}

func TestImpl_StartAlreadyRun(t *testing.T) {
	const (
		errFirstNotSelf  = "первый вызов Start() вернул не исходный объект."
		errSecondNotSelf = "повторный вызов Start() при работающем catcher вернул не исходный объект."
	)
	var i7t Interface

	i7t = New(nil)
	if i7t.Start() != i7t {
		t.Fatal(errFirstNotSelf)
	}
	if i7t.Start() != i7t {
		t.Fatal(errSecondNotSelf)
	}
	i7t.Stop()
}

func TestImpl_StartPanicRecovered(t *testing.T) {
	const (
		tplPanicNotRecovered = "Start() не восстановился после паники: %v."
		errNotNilInterface   = "после паники Start() должен вернуть nil Interface."
	)
	var i7t *impl

	i7t = &impl{signal: make(chan os.Signal, 1)}
	defer func() {
		if e := recover(); e != nil {
			t.Fatalf(tplPanicNotRecovered, e)
		}
	}()
	if i7t.Start() != nil {
		t.Fatal(errNotNilInterface)
	}
}

func TestImpl_CatcherWithoutCatch(t *testing.T) {
	const (
		errNoSignal = "catcher не получил сигнал."
		errNoDone   = "catcher не завершился после сигнала остановки."
	)
	var (
		i7t  *impl
		sent chan struct{}
	)

	i7t = &impl{
		doExitUp:   make(chan struct{}),
		doExitDone: make(chan struct{}),
		signal:     make(chan os.Signal),
	}
	sent = make(chan struct{})
	go i7t.Catcher()
	go func() {
		i7t.signal <- syscall.SIGINT
		close(sent)
	}()
	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal(errNoSignal)
	}
	i7t.doExitUp <- struct{}{}
	select {
	case <-i7t.doExitDone:
	case <-time.After(time.Second):
		t.Fatal(errNoDone)
	}
}

func TestImpl_CatcherCatch(t *testing.T) {
	const (
		tplSignalMismatch = "catcher передал сигнал %v, ожидался %v."
		errCatchNotCalled = "catcher не вызвал функцию обработки сигнала."
		errNoDone         = "catcher не завершился после сигнала остановки."
	)
	var (
		called chan os.Signal
		i7t    *impl
	)

	called = make(chan os.Signal, 1)
	i7t = &impl{
		doExitUp:   make(chan struct{}),
		doExitDone: make(chan struct{}),
		signal:     make(chan os.Signal),
		catch: func(sig os.Signal) {
			called <- sig
		},
	}
	go i7t.Catcher()
	go func() {
		i7t.signal <- syscall.SIGTERM
	}()
	select {
	case sig := <-called:
		if sig != syscall.SIGTERM {
			t.Fatalf(tplSignalMismatch, sig, syscall.SIGTERM)
		}
	case <-time.After(time.Second):
		t.Fatal(errCatchNotCalled)
	}
	i7t.doExitUp <- struct{}{}
	select {
	case <-i7t.doExitDone:
	case <-time.After(time.Second):
		t.Fatal(errNoDone)
	}
}

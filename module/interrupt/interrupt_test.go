package interrupt

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestImpl_StopResetState(t *testing.T) {
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
		t.Fatal("повторный вызов Stop() или перезапуск после Stop() привели к зависанию")
	}
}

func TestImpl_Destructor(t *testing.T) {
	var res = cleanupResource{
		doExitUp:   make(chan struct{}),
		doExitDone: make(chan struct{}),
		signal:     make(chan os.Signal),
	}

	destructor(res)
	destructor(res)
}

func TestImpl_Log(t *testing.T) {
	var i7t = New(nil)

	if i7t.(*impl).Log() == nil {
		t.Fatal("Log() вернул nil")
	}
}

func TestImpl_StartAlreadyRun(t *testing.T) {
	var i7t = New(nil)

	if i7t.Start() != i7t {
		t.Fatal("первый вызов Start() вернул не исходный объект")
	}
	if i7t.Start() != i7t {
		t.Fatal("повторный вызов Start() при работающем catcher вернул не исходный объект")
	}
	i7t.Stop()
}

func TestImpl_StartPanicRecovered(t *testing.T) {
	var i7t = &impl{signal: make(chan os.Signal, 1)}

	defer func() {
		if e := recover(); e != nil {
			t.Fatalf("Start() не восстановился после паники: %v", e)
		}
	}()
	if i7t.Start() != nil {
		t.Fatal("после паники Start() должен вернуть nil Interface")
	}
}

func TestImpl_CatcherWithoutCatch(t *testing.T) {
	var (
		i7t = &impl{
			doExitUp:   make(chan struct{}),
			doExitDone: make(chan struct{}),
			signal:     make(chan os.Signal),
		}
		sent = make(chan struct{})
	)

	go i7t.Catcher()
	go func() {
		i7t.signal <- syscall.SIGINT
		close(sent)
	}()

	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal("catcher не получил сигнал")
	}
	i7t.doExitUp <- struct{}{}
	select {
	case <-i7t.doExitDone:
	case <-time.After(time.Second):
		t.Fatal("catcher не завершился после сигнала остановки")
	}
}

func TestImpl_CatcherCatch(t *testing.T) {
	var (
		called = make(chan os.Signal, 1)
		i7t    = &impl{
			doExitUp:   make(chan struct{}),
			doExitDone: make(chan struct{}),
			signal:     make(chan os.Signal),
			catch: func(sig os.Signal) {
				called <- sig
			},
		}
	)

	go i7t.Catcher()
	go func() {
		i7t.signal <- syscall.SIGTERM
	}()

	select {
	case sig := <-called:
		if sig != syscall.SIGTERM {
			t.Fatalf("catcher передал сигнал %v, ожидался %v", sig, syscall.SIGTERM)
		}
	case <-time.After(time.Second):
		t.Fatal("catcher не вызвал функцию обработки сигнала")
	}
	i7t.doExitUp <- struct{}{}
	select {
	case <-i7t.doExitDone:
	case <-time.After(time.Second):
		t.Fatal("catcher не завершился после сигнала остановки")
	}
}

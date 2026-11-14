package cfg

import (
	"testing"
	"time"

	kitTypes "github.com/webnice/kit/v4/types"
)

// TestComponentInitiateDoesNotSelfDeadlock ComponentInitiate не приводит к самоблокировке.
func TestComponentInitiateDoesNotSelfDeadlock(t *testing.T) {
	const (
		errSelfDeadlock = "ComponentInitiate() привёл к self-deadlock."
		errNotInitiated = "ComponentInitiate() не выполнил Initiate компоненты."
	)
	var (
		oldComponent = singleton.comp.Component
		oldIsClose   = singleton.comp.IsClose
		done         chan struct{}
	)

	t.Cleanup(func() {
		singleton.comp.Component = oldComponent
		singleton.comp.IsClose = oldIsClose
	})
	singleton.comp.Component = []*kitTypes.ComponentInfo{{ComponentName: "self-deadlock-test"}}
	singleton.comp.IsClose = true
	done = make(chan struct{})
	go func() {
		Get().Gist().ComponentInitiate(func(ci *kitTypes.ComponentInfo) error {
			_ = Get().Gist().ComponentNames()
			ci.IsInitiate = true

			return nil
		})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal(errSelfDeadlock)
	}
	if !singleton.comp.Component[0].IsInitiate {
		t.Fatal(errNotInitiated)
	}
}

// TestComponentPreferencesDoesNotHoldMutexDuringCallback ComponentPreferences не удерживает мьютекс во время обратного вызова.
func TestComponentPreferencesDoesNotHoldMutexDuringCallback(t *testing.T) {
	const (
		tplPreferencesResult = "ComponentPreferences() = %d, %v."
		tplNotMoved          = "ComponentPreferences() не перенёс компоненты: component=%d registered=%d."
	)
	var (
		oldComponent  = singleton.comp.Component
		oldRegistered = singleton.comp.Registered
		oldIsClose    = singleton.comp.IsClose
		code          uint8
		err           error
	)

	t.Cleanup(func() {
		singleton.comp.Component = oldComponent
		singleton.comp.Registered = oldRegistered
		singleton.comp.IsClose = oldIsClose
	})
	singleton.comp.Component = nil
	singleton.comp.Registered = []*kitTypes.ComponentInfo{
		{ComponentName: "preferences-one"},
		{ComponentName: "preferences-two"},
	}
	singleton.comp.IsClose = true
	code, err = Get().Gist().ComponentPreferences(func(ci *kitTypes.ComponentInfo) (*kitTypes.ComponentInfo, error) {
		_ = Get().Gist().ComponentNames()
		ci.IsInitiate = false

		return ci, nil
	})
	if err != nil || code != 0 {
		t.Fatalf(tplPreferencesResult, code, err)
	}
	if len(singleton.comp.Component) != 2 || len(singleton.comp.Registered) != 0 {
		t.Fatalf(tplNotMoved, len(singleton.comp.Component), len(singleton.comp.Registered))
	}
}

// TestComponentChecksAndSortDoNotHoldMutexDuringCallback Проверки и сортировка компонент не удерживают мьютекс во время обратного вызова.
func TestComponentChecksAndSortDoNotHoldMutexDuringCallback(t *testing.T) {
	const (
		tplConflictResult = "ComponentCheckConflict() = %d, %v."
		tplRequiresResult = "ComponentRequiresCheck() = %d, %v."
		tplSortResult     = "ComponentSort() = %d, %v."
		tplSortNoOrder    = "ComponentSort() не сохранил порядок: %#v."
	)
	var (
		oldComponent = singleton.comp.Component
		oldIsClose   = singleton.comp.IsClose
		code         uint8
		err          error
	)

	t.Cleanup(func() {
		singleton.comp.Component = oldComponent
		singleton.comp.IsClose = oldIsClose
	})
	singleton.comp.IsClose = true
	singleton.comp.Component = []*kitTypes.ComponentInfo{
		{ComponentName: "zeta"},
		{ComponentName: "alpha"},
	}
	if code, err = Get().Gist().ComponentCheckConflict(func([]*kitTypes.ComponentInfo) error {
		_ = Get().Gist().ComponentNames()

		return nil
	}); err != nil || code != 0 {
		t.Fatalf(tplConflictResult, code, err)
	}
	if code, err = Get().Gist().ComponentRequiresCheck(func([]*kitTypes.ComponentInfo) error {
		_ = Get().Gist().ComponentNames()

		return nil
	}); err != nil || code != 0 {
		t.Fatalf(tplRequiresResult, code, err)
	}
	if code, err = Get().Gist().ComponentSort(func(component []*kitTypes.ComponentInfo) error {
		_ = Get().Gist().ComponentNames()
		if len(component) == 2 && component[0].ComponentName == "zeta" {
			component[0], component[1] = component[1], component[0]
		}

		return nil
	}); err != nil || code != 0 {
		t.Fatalf(tplSortResult, code, err)
	}
	if singleton.comp.Component[0].ComponentName != "alpha" || singleton.comp.Component[1].ComponentName != "zeta" {
		t.Fatalf(tplSortNoOrder, singleton.comp.Component)
	}
}

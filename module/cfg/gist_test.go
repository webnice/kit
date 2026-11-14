package cfg

import (
	"errors"
	"sync"
	"testing"
)

// TestErrorAppendConcurrentSafe Конкурентное добавление ошибок объекта сущности.
func TestErrorAppendConcurrentSafe(t *testing.T) {
	const (
		errTextTestError = "test error"
		errIsErrorFalse  = "IsError() = false после ErrorAppend."
	)
	var (
		obj *impl
		ess *gist
		wg  sync.WaitGroup
		n   int
	)

	obj = &impl{}
	ess = &gist{parent: obj}
	for n = 0; n < 64; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ess.ErrorAppend(errors.New(errTextTestError))
		}()
	}
	wg.Wait()
	if !ess.IsError() {
		t.Fatal(errIsErrorFalse)
	}
}

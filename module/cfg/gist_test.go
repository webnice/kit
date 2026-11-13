package cfg

import (
	"errors"
	"sync"
	"testing"
)

// TestErrorAppendConcurrentSafe Конкурентное добавление ошибок объекта сущности.
func TestErrorAppendConcurrentSafe(t *testing.T) {
	var (
		obj = &impl{}
		ess = &gist{parent: obj}
		wg  sync.WaitGroup
	)
	for n := 0; n < 64; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ess.ErrorAppend(errors.New("test error"))
		}()
	}
	wg.Wait()
	if !ess.IsError() {
		t.Fatal("IsError() = false после ErrorAppend")
	}
}

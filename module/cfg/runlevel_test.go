package cfg

import (
	"container/list"
	"sync"
	"testing"
)

// testRunlevelSubscriberA Тестовый подписчик изменения уровня.
func testRunlevelSubscriberA(uint16, uint16) {}

// testRunlevelSubscriberB Тестовый подписчик изменения уровня.
func testRunlevelSubscriberB(uint16, uint16) {}

// testRunlevelSubscriberC Тестовый подписчик изменения уровня.
func testRunlevelSubscriberC(uint16, uint16) {}

// TestRunlevelConcurrentSubscribeAndDispatch Конкурентная подписка и диспетчеризация уровней.
func TestRunlevelConcurrentSubscribeAndDispatch(t *testing.T) {
	var wg sync.WaitGroup
	cfgObj := &impl{runLevelSubscribers: list.New()}

	if err := cfgObj.RunlevelSubscribe(testRunlevelSubscriberA); err != nil {
		t.Fatalf("RunlevelSubscribe(A) вернул ошибку: %v", err)
	}
	if err := cfgObj.RunlevelSubscribe(testRunlevelSubscriberB); err != nil {
		t.Fatalf("RunlevelSubscribe(B) вернул ошибку: %v", err)
	}
	for n := 0; n < 64; n++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = cfgObj.RunlevelSubscribe(testRunlevelSubscriberC)
		}()
		go func() {
			defer wg.Done()
			_ = cfgObj.RunlevelUnsubscribe(testRunlevelSubscriberB)
		}()
		go func() {
			defer wg.Done()
			cfgObj.runlevelCallSubscribers(1, 2)
		}()
	}
	wg.Wait()
}

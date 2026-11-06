package ambry

import (
	"fmt"
	"sync"
	"testing"
)

func TestImpl_SetOverwrite(t *testing.T) {
	var prm = New()

	prm.Add(k, v1)
	prm.Add(k, v2)
	prm.Set(k, v3)
	if all := prm.GetAll(k); len(all) != 1 || all[0] != v3 {
		t.Fatalf("Set() не перезаписал значения: %+v", all)
	}
}

func TestImpl_DelMultiple(t *testing.T) {
	var prm = New()

	prm.Add(k, v1)
	prm.Add(k, v2)
	if got := prm.Del(k); got != v1 {
		t.Fatalf("Del() = %v, ожидался первый элемент %v", got, v1)
	}
	if prm.Has(k) {
		t.Fatal("ключ не удалён после Del()")
	}
	if all := prm.GetAll(k); len(all) != 0 {
		t.Fatalf("GetAll() после Del() вернул %+v", all)
	}
}

func TestImpl_EmptySlices(t *testing.T) {
	var prm = New()

	if keys := prm.Keys(); keys == nil || len(keys) != 0 {
		t.Fatalf("Keys() для пустого объекта вернул %#v", keys)
	}
	if all := prm.GetAll(k); all == nil || len(all) != 0 {
		t.Fatalf("GetAll() для отсутствующего ключа вернул %#v", all)
	}
}

func TestImpl_DifferentKeyTypes(t *testing.T) {
	type customKey struct {
		ID   int
		Name string
	}

	var (
		prm = New()
		ck  = customKey{ID: 1, Name: "one"}
	)

	prm.Set(1, "int")
	prm.Set("str", "string")
	prm.Set(ck, "struct")
	if got := prm.Get(1); got != "int" {
		t.Fatalf("Get(int) = %v", got)
	}
	if got := prm.Get("str"); got != "string" {
		t.Fatalf("Get(string) = %v", got)
	}
	if got := prm.Get(ck); got != "struct" {
		t.Fatalf("Get(struct) = %v", got)
	}
	if keys := prm.Keys(); len(keys) != 3 {
		t.Fatalf("Keys() = %v, ожидалось 3 ключа", keys)
	}
}

func TestImpl_ConcurrentAdd(t *testing.T) {
	const total = 128
	var (
		n     int
		ok    bool
		value any
		all   []any
		prm   Interface
		wg    sync.WaitGroup
		seen  map[any]struct{}
	)

	prm = New()
	for n = range total {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			prm.Add(k, fmt.Sprintf("value-%d", n))
		}(n)
	}
	wg.Wait()
	all = prm.GetAll(k)
	if len(all) != total {
		t.Fatalf("после конкурентного Add() значений %d, ожидалось %d", len(all), total)
	}
	seen = make(map[any]struct{}, len(all))
	for _, value = range all {
		seen[value] = struct{}{}
	}
	for n = range total {
		if _, ok = seen[fmt.Sprintf("value-%d", n)]; !ok {
			t.Fatalf("после конкурентного Add() потеряно значение value-%d", n)
		}
	}
}

func TestImpl_ConcurrentMixedAccess(t *testing.T) {
	var (
		n, m int
		prm  Interface
		wg   sync.WaitGroup
		key  string
	)

	prm = New()
	for n = range 64 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key = fmt.Sprintf("key-%d", n)
			for m = range 50 {
				switch m % 4 {
				case 0:
					prm.Set(key, m)
				case 1:
					prm.Add(key, m)
				case 2:
					prm.Has(key)
					prm.Get(key)
					prm.GetAll(key)
				case 3:
					prm.Del(key)
				}
			}
		}(n)
	}
	wg.Wait()
	prm.Keys()
}

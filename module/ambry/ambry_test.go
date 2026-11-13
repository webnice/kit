package ambry

import (
	"fmt"
	"sync"
	"testing"
)

var (
	k  = "A69F2F4D-0B8F-43D3-8E1B-1EA90766FEB5"
	v1 = "ac5110c701777822d999d0be2fd44a945dde58b46398d2c12ada878c5271e167770305532a6d1aa05cf16c575edf3ae99fd3f55db8143"
	v2 = "205c23c53fc4cf5ea6bf7a352fc4c7cf91ac001dcbfbebbe97b0c59cdf053096ceaf8e41e392c7953b379d0c6f6094c56471aa253f2b3"
	v3 = "5f4f7401e818ccc9af6d6c6c19eff4ce0696925bfb2a839c7bea211dda029816f48bc2417a9420b401450b6d6092dd4273f2a2d6a9364"
)

func TestImpl_Set(t *testing.T) {
	var prm = New()

	if prm.Has(k) {
		t.Errorf("key %s exist", k)
	}
	prm.Set(k, v1)
	if !prm.Has(k) {
		t.Errorf("key is %s exist", k)
	}
	if prm.Get(k) != v1 {
		t.Errorf("value is not equal")
	}
}

func TestImpl_Get(t *testing.T) {
	var prm = New()

	if prm.Get(k) != nil {
		t.Errorf("Get return incorrect value")
	}
	prm.Set(k, v1)
	if prm.Get(k) != v1 {
		t.Errorf("value is not equal")
	}
}

func TestImpl_Add(t *testing.T) {
	var (
		prm Interface
		all []any
	)

	prm = New()

	all = prm.GetAll(k)
	if len(all) != 0 {
		t.Errorf("Error in GetAll")
	}
	prm.Add(k, v1)
	prm.Add(k, v2)
	prm.Add(k, v3)
	if prm.Get(k) != v1 {
		t.Errorf("value is not equal")
	}
	all = prm.GetAll(k)
	if len(all) != 3 {
		t.Errorf("incorrect length")
	}
	if all[0] != v1 || all[1] != v2 || all[2] != v3 {
		t.Errorf("incorrect value")
	}
}

func TestImpl_Keys(t *testing.T) {
	var (
		prm Interface
		arr []any
		oks []bool
	)

	prm = New()
	arr = prm.Keys()
	if len(arr) != 0 {
		t.Errorf("Error in Keys")
	}

	prm.Add("k1", v1)
	prm.Add("k2", v2)
	prm.Set("k3", v3)
	arr = prm.Keys()
	if len(arr) != 3 {
		t.Errorf("Error in Keys")
	}
	for i := range arr {
		if arr[i] == "k1" {
			oks = append(oks, true)
		}
		if arr[i] == "k2" {
			oks = append(oks, true)
		}
		if arr[i] == "k3" {
			oks = append(oks, true)
		}
	}
	if len(oks) != 3 {
		t.Errorf("Error in Keys")
	}
}

func TestImpl_Del(t *testing.T) {
	var prm = New()

	if prm.Del(k) != nil {
		t.Errorf("Error in Del")
	}
	prm.Set(k, v1)
	if prm.Get(k) != v1 {
		t.Errorf("Error in Get")
	}
	if prm.Del(k) != v1 {
		t.Errorf("Error in Del")
	}
	if prm.Get(k) != nil {
		t.Errorf("Error in Get")
	}
}

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

const (
	KEY = "foo"
	VAL = "bar"
)

var res any

func BenchmarkImpl_Get(b *testing.B) {
	var (
		str any
		prm = New()
	)

	prm.Set(KEY, VAL)
	for n := 0; n < b.N; n++ {
		if x := prm.Get(KEY); x != "" {
			str = x
		}
	}
	res = str
}

func BenchmarkImpl_Set(b *testing.B) {
	var prm = New()

	for n := 0; n < b.N; n++ {
		prm.Set(KEY, VAL)
	}
}

func BenchmarkImpl_Has(b *testing.B) {
	var (
		has bool
		prm = New()
	)

	for n := 0; n < b.N; n++ {
		res = prm.Has(KEY)
	}
	res = has
}

func BenchmarkImpl_Del(b *testing.B) {
	var prm = New()

	prm.Set(KEY, VAL)
	for n := 0; n < b.N; n++ {
		prm.Del(KEY)
	}
}

package ambry

import (
	"fmt"
	"sync"
	"testing"
)

// Тестовые ключи и значения.
var (
	k  = "A69F2F4D-0B8F-43D3-8E1B-1EA90766FEB5"
	v1 = "ac5110c701777822d999d0be2fd44a945dde58b46398d2c12ada878c5271e167770305532a6d1aa05cf16c575edf3ae99fd3f55db8143"
	v2 = "205c23c53fc4cf5ea6bf7a352fc4c7cf91ac001dcbfbebbe97b0c59cdf053096ceaf8e41e392c7953b379d0c6f6094c56471aa253f2b3"
	v3 = "5f4f7401e818ccc9af6d6c6c19eff4ce0696925bfb2a839c7bea211dda029816f48bc2417a9420b401450b6d6092dd4273f2a2d6a9364"
)

func TestImpl_Set(t *testing.T) {
	const (
		errKeyExists    = "ключ %s существует до установки."
		errKeyNotExists = "ключ %s не существует после установки."
		errValueNotEq   = "значение не равно ожидаемому."
	)
	var prm Interface

	prm = New()
	if prm.Has(k) {
		t.Errorf(errKeyExists, k)
	}
	prm.Set(k, v1)
	if !prm.Has(k) {
		t.Errorf(errKeyNotExists, k)
	}
	if prm.Get(k) != v1 {
		t.Error(errValueNotEq)
	}
}

func TestImpl_Get(t *testing.T) {
	const (
		errGetNotNil  = "Get() для отсутствующего ключа вернул не nil."
		errValueNotEq = "значение не равно ожидаемому."
	)
	var prm Interface

	prm = New()
	if prm.Get(k) != nil {
		t.Error(errGetNotNil)
	}
	prm.Set(k, v1)
	if prm.Get(k) != v1 {
		t.Error(errValueNotEq)
	}
}

func TestImpl_Add(t *testing.T) {
	const (
		errGetAllNotEmpty = "GetAll() для отсутствующего ключа вернул не пустой срез."
		errValueNotEq     = "значение не равно ожидаемому."
		errLengthWrong    = "неверная длина среза значений."
		errValuesWrong    = "неверные значения в срезе."
	)
	var (
		prm Interface
		all []any
	)

	prm = New()
	all = prm.GetAll(k)
	if len(all) != 0 {
		t.Error(errGetAllNotEmpty)
	}
	prm.Add(k, v1)
	prm.Add(k, v2)
	prm.Add(k, v3)
	if prm.Get(k) != v1 {
		t.Error(errValueNotEq)
	}
	all = prm.GetAll(k)
	if len(all) != 3 {
		t.Error(errLengthWrong)
	}
	if all[0] != v1 || all[1] != v2 || all[2] != v3 {
		t.Error(errValuesWrong)
	}
}

func TestImpl_Keys(t *testing.T) {
	const (
		errKeysNotEmpty = "Keys() для пустого объекта вернул не пустой срез."
		errKeysCount    = "Keys() вернул неверное количество ключей."
		errKeysMissing  = "Keys() не вернул все добавленные ключи."
	)
	var (
		prm Interface
		arr []any
		oks []bool
		n   int
	)

	prm = New()
	arr = prm.Keys()
	if len(arr) != 0 {
		t.Error(errKeysNotEmpty)
	}
	prm.Add("k1", v1)
	prm.Add("k2", v2)
	prm.Set("k3", v3)
	arr = prm.Keys()
	if len(arr) != 3 {
		t.Error(errKeysCount)
	}
	for n = range arr {
		if arr[n] == "k1" || arr[n] == "k2" || arr[n] == "k3" {
			oks = append(oks, true)
		}
	}
	if len(oks) != 3 {
		t.Error(errKeysMissing)
	}
}

func TestImpl_Del(t *testing.T) {
	const (
		errDelNotNil  = "Del() для отсутствующего ключа вернул не nil."
		errGetNotEq   = "Get() вернул неверное значение."
		errDelNotEq   = "Del() вернул неверное значение."
		errGetNotNil2 = "Get() после Del() вернул не nil."
	)
	var prm Interface

	prm = New()
	if prm.Del(k) != nil {
		t.Error(errDelNotNil)
	}
	prm.Set(k, v1)
	if prm.Get(k) != v1 {
		t.Error(errGetNotEq)
	}
	if prm.Del(k) != v1 {
		t.Error(errDelNotEq)
	}
	if prm.Get(k) != nil {
		t.Error(errGetNotNil2)
	}
}

func TestImpl_SetOverwrite(t *testing.T) {
	const (
		tplSetNotOverwrite = "Set() не перезаписал значения: %+v."
	)
	var (
		prm Interface
		all []any
	)

	prm = New()
	prm.Add(k, v1)
	prm.Add(k, v2)
	prm.Set(k, v3)
	all = prm.GetAll(k)
	if len(all) != 1 || all[0] != v3 {
		t.Fatalf(tplSetNotOverwrite, all)
	}
}

func TestImpl_DelMultiple(t *testing.T) {
	const (
		tplDelWrong       = "Del() = %v, ожидался первый элемент %v."
		errKeyNotDeleted  = "ключ не удалён после Del()."
		tplGetAllNotEmpty = "GetAll() после Del() вернул %+v."
	)
	var (
		prm Interface
		got any
		all []any
	)

	prm = New()
	prm.Add(k, v1)
	prm.Add(k, v2)
	got = prm.Del(k)
	if got != v1 {
		t.Fatalf(tplDelWrong, got, v1)
	}
	if prm.Has(k) {
		t.Fatal(errKeyNotDeleted)
	}
	all = prm.GetAll(k)
	if len(all) != 0 {
		t.Fatalf(tplGetAllNotEmpty, all)
	}
}

func TestImpl_EmptySlices(t *testing.T) {
	const (
		tplKeysNotEmpty = "Keys() для пустого объекта вернул %#v."
		tplAllNotEmpty  = "GetAll() для отсутствующего ключа вернул %#v."
	)
	var (
		prm  Interface
		keys []any
		all  []any
	)

	prm = New()
	keys = prm.Keys()
	if keys == nil || len(keys) != 0 {
		t.Fatalf(tplKeysNotEmpty, keys)
	}
	all = prm.GetAll(k)
	if all == nil || len(all) != 0 {
		t.Fatalf(tplAllNotEmpty, all)
	}
}

func TestImpl_DifferentKeyTypes(t *testing.T) {
	const (
		tplGetIntWrong    = "Get(int) = %v."
		tplGetStringWrong = "Get(string) = %v."
		tplGetStructWrong = "Get(struct) = %v."
		tplKeysCount      = "Keys() = %v, ожидалось 3 ключа."
	)
	type customKey struct {
		ID   int
		Name string
	}
	var (
		prm  Interface
		ck   customKey
		got  any
		keys []any
	)

	prm = New()
	ck = customKey{ID: 1, Name: "one"}
	prm.Set(1, "int")
	prm.Set("str", "string")
	prm.Set(ck, "struct")
	got = prm.Get(1)
	if got != "int" {
		t.Fatalf(tplGetIntWrong, got)
	}
	got = prm.Get("str")
	if got != "string" {
		t.Fatalf(tplGetStringWrong, got)
	}
	got = prm.Get(ck)
	if got != "struct" {
		t.Fatalf(tplGetStructWrong, got)
	}
	keys = prm.Keys()
	if len(keys) != 3 {
		t.Fatalf(tplKeysCount, keys)
	}
}

func TestImpl_ConcurrentAdd(t *testing.T) {
	const (
		total            = 128
		tplValue         = "value-%d"
		tplCountMismatch = "после конкурентного Add() значений %d, ожидалось %d."
		tplValueLost     = "после конкурентного Add() потеряно значение value-%d."
	)
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
			prm.Add(k, fmt.Sprintf(tplValue, n))
		}(n)
	}
	wg.Wait()
	all = prm.GetAll(k)
	if len(all) != total {
		t.Fatalf(tplCountMismatch, len(all), total)
	}
	seen = make(map[any]struct{}, len(all))
	for _, value = range all {
		seen[value] = struct{}{}
	}
	for n = range total {
		if _, ok = seen[fmt.Sprintf(tplValue, n)]; !ok {
			t.Fatalf(tplValueLost, n)
		}
	}
}

func TestImpl_ConcurrentMixedAccess(t *testing.T) {
	const (
		tplKey = "key-%d"
	)
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
			key = fmt.Sprintf(tplKey, n)
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

// Ключ и значение для бенчмарков.
const (
	keyBenchmark = "foo"
	valBenchmark = "bar"
)

// Результат бенчмарков, предотвращающий оптимизацию компилятора.
var res any

func BenchmarkImpl_Get(b *testing.B) {
	var (
		n   int
		str any
		prm Interface
	)

	prm = New()
	prm.Set(keyBenchmark, valBenchmark)
	for n = 0; n < b.N; n++ {
		if x := prm.Get(keyBenchmark); x != "" {
			str = x
		}
	}
	res = str
}

func BenchmarkImpl_Set(b *testing.B) {
	var (
		n   int
		prm Interface
	)

	prm = New()
	for n = 0; n < b.N; n++ {
		prm.Set(keyBenchmark, valBenchmark)
	}
}

func BenchmarkImpl_Has(b *testing.B) {
	var (
		n   int
		has bool
		prm Interface
	)

	prm = New()
	for n = 0; n < b.N; n++ {
		res = prm.Has(keyBenchmark)
	}
	res = has
}

func BenchmarkImpl_Del(b *testing.B) {
	var (
		n   int
		prm Interface
	)

	prm = New()
	prm.Set(keyBenchmark, valBenchmark)
	for n = 0; n < b.N; n++ {
		prm.Del(keyBenchmark)
	}
}

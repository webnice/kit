package ambry

import "testing"

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

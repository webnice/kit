package utility

import "testing"

func assertSameSet[T comparable](t *testing.T, got, expected []T) {
	const (
		tplLenWrong   = "len(got) = %d, ожидалось %d: %#v."
		tplUnexpected = "неожиданный элемент %v в %#v."
	)
	var (
		items map[T]struct{}
		item  T
	)

	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf(tplLenWrong, len(got), len(expected), got)
	}
	items = make(map[T]struct{}, len(expected))
	for _, item = range expected {
		items[item] = struct{}{}
	}
	for _, item = range got {
		if _, ok := items[item]; !ok {
			t.Fatalf(tplUnexpected, item, got)
		}
	}
}

func TestSliceUniqueInts(t *testing.T) {
	var got []int

	got = SliceUnique([]int{4, 2, 4, 1, 2, 3, 1})
	assertSameSet(t, got, []int{1, 2, 3, 4})
}

func TestSliceUniqueStrings(t *testing.T) {
	var got []string

	got = SliceUnique([]string{"b", "a", "b", "c", "a"})
	assertSameSet(t, got, []string{"a", "b", "c"})
}

func TestSliceUniqueEmpty(t *testing.T) {
	const (
		tplEmptyWrong = "SliceUnique(empty) = %#v."
	)
	var got []float64

	got = SliceUnique([]float64{})
	if len(got) != 0 || got == nil {
		t.Fatalf(tplEmptyWrong, got)
	}
}

func TestSliceUniqueComplex(t *testing.T) {
	var got []complex128

	got = SliceUnique([]complex128{1 + 2i, 3 + 4i, 1 + 2i})
	assertSameSet(t, got, []complex128{1 + 2i, 3 + 4i})
}

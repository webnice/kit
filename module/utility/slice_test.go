package utility

import "testing"

func assertSameSet[T comparable](t *testing.T, got, expected []T) {
	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf("len(got) = %d, ожидалось %d: %#v", len(got), len(expected), got)
	}
	items := make(map[T]struct{}, len(expected))
	for _, item := range expected {
		items[item] = struct{}{}
	}
	for _, item := range got {
		if _, ok := items[item]; !ok {
			t.Fatalf("неожиданный элемент %v в %#v", item, got)
		}
	}
}

func TestSliceUniqueInts(t *testing.T) {
	got := SliceUnique([]int{4, 2, 4, 1, 2, 3, 1})
	assertSameSet(t, got, []int{1, 2, 3, 4})
}

func TestSliceUniqueStrings(t *testing.T) {
	got := SliceUnique([]string{"b", "a", "b", "c", "a"})
	assertSameSet(t, got, []string{"a", "b", "c"})
}

func TestSliceUniqueEmpty(t *testing.T) {
	got := SliceUnique([]float64{})
	if len(got) != 0 || got == nil {
		t.Fatalf("SliceUnique(empty) = %#v", got)
	}
}

func TestSliceUniqueComplex(t *testing.T) {
	got := SliceUnique([]complex128{1 + 2i, 3 + 4i, 1 + 2i})
	assertSameSet(t, got, []complex128{1 + 2i, 3 + 4i})
}

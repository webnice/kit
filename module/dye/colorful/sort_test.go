package colorful

import "testing"

// TestSortSimple tests the sorting of a small set of colors.
func TestSortSimple(t *testing.T) {
	// Sort a list of reds and blues.
	in := make([]Color, 0, 6)
	for i := 0; i < 3; i++ {
		in = append(in, Color{1.0 - float64(i+1)*0.25, 0.0, 0.0}) // Reds
		in = append(in, Color{0.0, 0.0, 1.0 - float64(i+1)*0.25}) // Blues
	}
	out := Sorted(in)

	// Ensure the output matches what we expected.
	exp := []Color{
		Color{R: 0.25, G: 0.0, B: 0},
		Color{R: 0.50, G: 0.0, B: 0},
		Color{R: 0.75, G: 0.0, B: 0},
		Color{R: 0.0, G: 0.0, B: 0.25},
		Color{R: 0.0, G: 0.0, B: 0.50},
		Color{R: 0.0, G: 0.0, B: 0.75},
	}
	for i, e := range exp {
		if out[i] != e {
			t.Fatalf("Expected %v but saw %v", e, out[i])
		}
	}
}

// TestUnion Объединение непересекающихся множеств.
func TestUnion(t *testing.T) {
	// Объединение элемента с самим собой, без изменений.
	e := newElement()
	union(e, e)
	if e.find() != e {
		t.Error("union(): элемент не является корнем собственного множества")
	}
	// Объединение двух множеств одинакового ранга.
	e1, e2 := newElement(), newElement()
	union(e1, e2)
	if e1.find() != e2.find() {
		t.Error("union(): множества не объединены")
	}
	if e1.find().rank != 1 {
		t.Errorf("union(): ожидался ранг 1, получено %d", e1.find().rank)
	}
	// Объединение множеств разного ранга, в обе стороны.
	short := newElement()
	tall := newElement()
	union(tall, newElement()) // rank(tall) == 1
	union(short, tall)        // rank(short) == 0 < rank(tall)
	if short.find() != tall.find() {
		t.Error("union(): множества не объединены")
	}
	short2, tall2 := newElement(), newElement()
	union(tall2, newElement())
	union(tall2, short2) // rank(tall2) > rank(short2)
	if short2.find() != tall2.find() {
		t.Error("union(): множества не объединены")
	}
}

// TestSortedTrivial Сортировка пустого списка и списка из одного цвета.
func TestSortedTrivial(t *testing.T) {
	if got := Sorted(nil); len(got) != 0 {
		t.Errorf("Sorted(nil): ожидался пустой список, получено %v", got)
	}
	one := []Color{{1, 0, 0}}
	if got := Sorted(one); len(got) != 1 || got[0] != one[0] {
		t.Errorf("Sorted(один цвет): ожидалось %v, получено %v", one, got)
	}
}

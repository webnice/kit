package colorful

import "testing"

// TestFastHappyPalette Быстрая генерация палитры в HSV пространстве.
func TestFastHappyPalette(t *testing.T) {
	const n = 5
	colors := FastHappyPalette(n)
	if len(colors) != n {
		t.Errorf("FastHappyPalette(): ожидалось %d цветов, получено %d", n, len(colors))
	}
	for _, c := range colors {
		if !c.IsValid() {
			t.Errorf("FastHappyPalette(): невалидный цвет %v", c)
		}
	}
}

// TestHappyPalette Генерация палитры через L*a*b* пространство.
func TestHappyPalette(t *testing.T) {
	const n = 3
	colors, err := HappyPalette(n)
	if err != nil {
		t.Fatalf("HappyPalette(): неожиданная ошибка: %v", err)
	}
	if len(colors) != n {
		t.Errorf("HappyPalette(): ожидалось %d цветов, получено %d", n, len(colors))
	}
	for _, c := range colors {
		if !c.IsValid() {
			t.Errorf("HappyPalette(): невалидный цвет %v", c)
		}
	}
}

package colorful

import "testing"

// TestFastWarmPalette Быстрая генерация палитры в HSV пространстве.
func TestFastWarmPalette(t *testing.T) {
	const n = 7
	colors := FastWarmPalette(n)
	if len(colors) != n {
		t.Errorf("FastWarmPalette(): ожидалось %d цветов, получено %d", n, len(colors))
	}
	for _, c := range colors {
		if !c.IsValid() {
			t.Errorf("FastWarmPalette(): невалидный цвет %v", c)
		}
	}
}

// TestWarmPalette Генерация палитры через L*a*b* пространство.
func TestWarmPalette(t *testing.T) {
	const n = 3
	colors, err := WarmPalette(n)
	if err != nil {
		t.Fatalf("WarmPalette(): неожиданная ошибка: %v", err)
	}
	if len(colors) != n {
		t.Errorf("WarmPalette(): ожидалось %d цветов, получено %d", n, len(colors))
	}
	for _, c := range colors {
		if !c.IsValid() {
			t.Errorf("WarmPalette(): невалидный цвет %v", c)
		}
	}
}

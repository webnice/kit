package colorful

import (
	"fmt"
	"math"
	"testing"
)

// This is really difficult to test, if you've got a good idea, pull request!

// Check if it returns all valid and enough colors.
func TestColorCount(t *testing.T) {
	fmt.Printf("Testing up to %v palettes may take a while...\n", 100)
	for i := 0; i < 100; i++ {
		//pal, err := SoftPaletteEx(i, SoftPaletteGenSettings{nil, 50, true})
		pal, err := SoftPalette(i)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Check the color count of the palette
		if len(pal) != i {
			t.Errorf("Requested %v colors but got %v", i, len(pal))
		}

		// Also check whether all colors exist in RGB space.
		for icol, col := range pal {
			if !col.IsValid() {
				t.Errorf("Color %v in palette of %v is invalid: %v", icol, len(pal), col)
			}
		}
	}
	fmt.Println("Done with that, but more tests to run.")
}

// Check if it errors-out on an impossible constraint
func TestImpossibleConstraint(t *testing.T) {
	never := func(l, a, b float64) bool { return false }

	pal, err := SoftPaletteEx(10, SoftPaletteSettings{never, 50, true})
	if err == nil || pal != nil {
		t.Error("Should error-out on impossible constraint!")
	}
}

// Check whether the constraint is respected
func TestConstraint(t *testing.T) {
	octant := func(l, a, b float64) bool { return l <= 0.5 && a <= 0.0 && b <= 0.0 }

	pal, err := SoftPaletteEx(100, SoftPaletteSettings{octant, 50, true})
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Check ALL the colors!
	for icol, col := range pal {
		if !col.IsValid() {
			t.Errorf("Color %v in constrained palette is invalid: %v", icol, col)
		}

		l, a, b := col.Lab()
		if l > 0.5 || a > 0.0 || b > 0.0 {
			t.Errorf("Color %v in constrained palette violates the constraint: %v (lab: %v)", icol, col, [3]float64{l, a, b})
		}
	}
}

// TestSoftPaletteExactCount Запрошено ровно столько цветов, сколько доступно образцов.
func TestSoftPaletteExactCount(t *testing.T) {
	// Ограничение выбирает ровно 2 тёмно-серых образца.
	onlyGray := func(l, a, b float64) bool {
		return l > 0.15 && l < 0.25 && math.Abs(a) < 0.01 && math.Abs(b) < 0.01
	}
	palette, err := SoftPaletteEx(2, SoftPaletteSettings{onlyGray, 1, false})
	if err != nil {
		t.Fatalf("SoftPaletteEx(): неожиданная ошибка: %v", err)
	}
	if len(palette) != 2 {
		t.Fatalf("SoftPaletteEx(): ожидалось 2 цвета, получено %d", len(palette))
	}
}

// TestSoftPaletteMedoid Среднее кластера выходит за пределы допустимого пространства, режим medoid.
func TestSoftPaletteMedoid(t *testing.T) {
	// Разрешаем только тёмные и светлые цвета, среднее всех образцов попадает в запрещённую середину.
	noMiddleL := func(l, a, b float64) bool { return l < 0.3 || l > 0.7 }
	palette, err := SoftPaletteEx(1, SoftPaletteSettings{noMiddleL, 2, false})
	if err != nil {
		t.Fatalf("SoftPaletteEx(): неожиданная ошибка: %v", err)
	}
	if len(palette) != 1 {
		t.Fatalf("SoftPaletteEx(): ожидался 1 цвет, получено %d", len(palette))
	}
	if !palette[0].IsValid() {
		t.Fatalf("SoftPaletteEx(): невалидный цвет %v", palette[0])
	}
}

package dye

import (
	"image/color"
	"testing"
)

// customColor Произвольный тип цвета, для проверки ветки неизвестного типа в Convert().
type customColor struct{}

// Sequence Реализация интерфейса Color.
func (customColor) Sequence(bool) string { return "" }

// String Реализация интерфейса Color.
func (customColor) String() string { return "" }

// TestTerminalProfileConvert Конвертация цветов под профиль терминала.
func TestTerminalProfileConvert(t *testing.T) {
	const (
		tplTypeNone    = "Convert(): ожидался тип colorNone, получено %T."
		tplTypeAnsi    = "Convert(): ожидался тип colorAnsi(%d), получено %T(%v)."
		tplTypeAnsi256 = "Convert(): ожидался тип colorAnsi256(196), получено %T(%v)."
		tplTypeRgb     = "Convert(): ожидался тип colorRgb(#ff8000), получено %T(%v)."
		tplNilExpected = "Convert(): ожидалось nil, получено %T(%v)."
		tplTypeCustom  = "Convert(): ожидался тип customColor, получено %T."
		hexRed         = "#ff0000"
		hexOrange      = "#ff8000"
	)
	t.Run("Ascii", func(t *testing.T) {
		var c Color

		c = terminalAscii.Convert(colorAnsi(colorANSIRed))
		if _, ok := c.(colorNone); !ok {
			t.Fatalf(tplTypeNone, c)
		}
	})
	t.Run("AnsiPassthrough", func(t *testing.T) {
		var c Color

		c = terminalANSI.Convert(colorAnsi(colorANSIRed))
		if v, ok := c.(colorAnsi); !ok || v != colorANSIRed {
			t.Fatalf(tplTypeAnsi, colorANSIRed, c, c)
		}
	})
	t.Run("Ansi256ToAnsi", func(t *testing.T) {
		var c Color

		c = terminalANSI.Convert(colorAnsi256(196))
		if v, ok := c.(colorAnsi); !ok || v != colorANSIBrightRed {
			t.Fatalf(tplTypeAnsi, colorANSIBrightRed, c, c)
		}
	})
	t.Run("Ansi256Passthrough", func(t *testing.T) {
		var c Color

		c = terminalANSI256.Convert(colorAnsi256(196))
		if v, ok := c.(colorAnsi256); !ok || v != 196 {
			t.Fatalf(tplTypeAnsi256, c, c)
		}
		c = terminalTrueColor.Convert(colorAnsi256(196))
		if v, ok := c.(colorAnsi256); !ok || v != 196 {
			t.Fatalf(tplTypeAnsi256, c, c)
		}
	})
	t.Run("RgbToAnsi256", func(t *testing.T) {
		var c Color

		c = terminalANSI256.Convert(colorRgb(hexRed))
		if v, ok := c.(colorAnsi256); !ok || v != 196 {
			t.Fatalf(tplTypeAnsi256, c, c)
		}
	})
	t.Run("RgbToAnsi", func(t *testing.T) {
		var c Color

		c = terminalANSI.Convert(colorRgb(hexRed))
		if v, ok := c.(colorAnsi); !ok || v != colorANSIBrightRed {
			t.Fatalf(tplTypeAnsi, colorANSIBrightRed, c, c)
		}
	})
	t.Run("RgbTrueColor", func(t *testing.T) {
		var c Color

		c = terminalTrueColor.Convert(colorRgb(hexOrange))
		if v, ok := c.(colorRgb); !ok || v != hexOrange {
			t.Fatalf(tplTypeRgb, c, c)
		}
	})
	t.Run("RgbInvalid", func(t *testing.T) {
		var c Color

		if c = terminalANSI256.Convert(colorRgb("невалидный цвет")); c != nil {
			t.Fatalf(tplNilExpected, c, c)
		}
		if c = terminalTrueColor.Convert(colorRgb("невалидный цвет")); c != nil {
			t.Fatalf(tplNilExpected, c, c)
		}
	})
	t.Run("UnknownType", func(t *testing.T) {
		var c Color

		c = terminalANSI256.Convert(customColor{})
		if _, ok := c.(customColor); !ok {
			t.Fatalf(tplTypeCustom, c)
		}
	})
}

// TestTerminalProfileColor Создание цвета из строки.
func TestTerminalProfileColor(t *testing.T) {
	const (
		tplNilExpected = "Color(): ожидалось nil, получено %T(%v)."
		tplTypeRgb     = "Color(): ожидался тип colorRgb(#ff8000), получено %T(%v)."
		tplTypeAnsi    = "Color(): ожидался тип colorAnsi(7), получено %T(%v)."
		tplTypeAnsi256 = "Color(): ожидался тип colorAnsi256(123), получено %T(%v)."
		tplTypeNone    = "Color(): ожидался тип colorNone, получено %T(%v)."
		hexOrange      = "#ff8000"
	)
	var c Color

	if c = terminalTrueColor.Color(""); c != nil {
		t.Fatalf(tplNilExpected, c, c)
	}
	if c = terminalTrueColor.Color("абв"); c != nil {
		t.Fatalf(tplNilExpected, c, c)
	}
	c = terminalTrueColor.Color(hexOrange)
	if v, ok := c.(colorRgb); !ok || v != hexOrange {
		t.Fatalf(tplTypeRgb, c, c)
	}
	c = terminalTrueColor.Color("7")
	if v, ok := c.(colorAnsi); !ok || v != colorAnsi(7) {
		t.Fatalf(tplTypeAnsi, c, c)
	}
	c = terminalTrueColor.Color("123")
	if v, ok := c.(colorAnsi256); !ok || v != 123 {
		t.Fatalf(tplTypeAnsi256, c, c)
	}
	c = terminalAscii.Color(hexOrange)
	if _, ok := c.(colorNone); !ok {
		t.Fatalf(tplTypeNone, c, c)
	}
}

// TestTerminalProfileFromColor Создание цвета из стандартного image/color.Color.
func TestTerminalProfileFromColor(t *testing.T) {
	const (
		tplTypeRgb  = "FromColor(): ожидался тип colorRgb(#ff8000), получено %T(%v)."
		tplTypeNone = "FromColor(): ожидался тип colorNone, получено %T(%v)."
		hexOrange   = "#ff8000"
	)
	var (
		src color.RGBA
		c   Color
	)

	src = color.RGBA{R: 255, G: 128, B: 0, A: 255}
	c = terminalTrueColor.FromColor(src)
	if v, ok := c.(colorRgb); !ok || v != hexOrange {
		t.Fatalf(tplTypeRgb, c, c)
	}
	c = terminalAscii.FromColor(src)
	if _, ok := c.(colorNone); !ok {
		t.Fatalf(tplTypeNone, c, c)
	}
}

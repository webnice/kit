package dye

import (
	"image/color"
	"testing"
)

// customColor Произвольный тип цвета, для проверки ветки неизвестного типа в Convert().
type customColor struct{}

func (customColor) Sequence(bool) string { return "" }
func (customColor) String() string       { return "" }

// TestTerminalProfileConvert Конвертация цветов под профиль терминала.
func TestTerminalProfileConvert(t *testing.T) {
	t.Run("Ascii", func(t *testing.T) {
		c := terminalAscii.Convert(colorAnsi(colorANSIRed))
		if _, ok := c.(colorNone); !ok {
			t.Fatalf("Convert(): ожидался тип colorNone, получено %T", c)
		}
	})
	t.Run("AnsiPassthrough", func(t *testing.T) {
		c := terminalANSI.Convert(colorAnsi(colorANSIRed))
		if v, ok := c.(colorAnsi); !ok || v != colorANSIRed {
			t.Fatalf("Convert(): ожидался тип colorAnsi(%d), получено %T(%v)", colorANSIRed, c, c)
		}
	})
	t.Run("Ansi256ToAnsi", func(t *testing.T) {
		c := terminalANSI.Convert(colorAnsi256(196))
		if v, ok := c.(colorAnsi); !ok || v != colorANSIBrightRed {
			t.Fatalf("Convert(): ожидался тип colorAnsi(%d), получено %T(%v)", colorANSIBrightRed, c, c)
		}
	})
	t.Run("Ansi256Passthrough", func(t *testing.T) {
		c := terminalANSI256.Convert(colorAnsi256(196))
		if v, ok := c.(colorAnsi256); !ok || v != 196 {
			t.Fatalf("Convert(): ожидался тип colorAnsi256(196), получено %T(%v)", c, c)
		}
		c = terminalTrueColor.Convert(colorAnsi256(196))
		if v, ok := c.(colorAnsi256); !ok || v != 196 {
			t.Fatalf("Convert(): ожидался тип colorAnsi256(196), получено %T(%v)", c, c)
		}
	})
	t.Run("RgbToAnsi256", func(t *testing.T) {
		c := terminalANSI256.Convert(colorRgb("#ff0000"))
		if v, ok := c.(colorAnsi256); !ok || v != 196 {
			t.Fatalf("Convert(): ожидался тип colorAnsi256(196), получено %T(%v)", c, c)
		}
	})
	t.Run("RgbToAnsi", func(t *testing.T) {
		c := terminalANSI.Convert(colorRgb("#ff0000"))
		if v, ok := c.(colorAnsi); !ok || v != colorANSIBrightRed {
			t.Fatalf("Convert(): ожидался тип colorAnsi(%d), получено %T(%v)", colorANSIBrightRed, c, c)
		}
	})
	t.Run("RgbTrueColor", func(t *testing.T) {
		c := terminalTrueColor.Convert(colorRgb("#ff8000"))
		if v, ok := c.(colorRgb); !ok || v != "#ff8000" {
			t.Fatalf("Convert(): ожидался тип colorRgb(#ff8000), получено %T(%v)", c, c)
		}
	})
	t.Run("RgbInvalid", func(t *testing.T) {
		if c := terminalANSI256.Convert(colorRgb("невалидный цвет")); c != nil {
			t.Fatalf("Convert(): ожидалось nil, получено %T(%v)", c, c)
		}
		if c := terminalTrueColor.Convert(colorRgb("невалидный цвет")); c != nil {
			t.Fatalf("Convert(): ожидалось nil, получено %T(%v)", c, c)
		}
	})
	t.Run("UnknownType", func(t *testing.T) {
		c := terminalANSI256.Convert(customColor{})
		if _, ok := c.(customColor); !ok {
			t.Fatalf("Convert(): ожидался тип customColor, получено %T", c)
		}
	})
}

// TestTerminalProfileColor Создание цвета из строки.
func TestTerminalProfileColor(t *testing.T) {
	if c := terminalTrueColor.Color(""); c != nil {
		t.Fatalf("Color(): ожидалось nil, получено %T(%v)", c, c)
	}
	if c := terminalTrueColor.Color("абв"); c != nil {
		t.Fatalf("Color(): ожидалось nil, получено %T(%v)", c, c)
	}
	c := terminalTrueColor.Color("#ff8000")
	if v, ok := c.(colorRgb); !ok || v != "#ff8000" {
		t.Fatalf("Color(): ожидался тип colorRgb(#ff8000), получено %T(%v)", c, c)
	}
	c = terminalTrueColor.Color("7")
	if v, ok := c.(colorAnsi); !ok || v != colorAnsi(7) {
		t.Fatalf("Color(): ожидался тип colorAnsi(7), получено %T(%v)", c, c)
	}
	c = terminalTrueColor.Color("123")
	if v, ok := c.(colorAnsi256); !ok || v != 123 {
		t.Fatalf("Color(): ожидался тип colorAnsi256(123), получено %T(%v)", c, c)
	}
	c = terminalAscii.Color("#ff8000")
	if _, ok := c.(colorNone); !ok {
		t.Fatalf("Color(): ожидался тип colorNone, получено %T(%v)", c, c)
	}
}

// TestTerminalProfileFromColor Создание цвета из стандартного image/color.Color.
func TestTerminalProfileFromColor(t *testing.T) {
	src := color.RGBA{R: 255, G: 128, B: 0, A: 255}
	c := terminalTrueColor.FromColor(src)
	if v, ok := c.(colorRgb); !ok || v != "#ff8000" {
		t.Fatalf("FromColor(): ожидался тип colorRgb(#ff8000), получено %T(%v)", c, c)
	}
	c = terminalAscii.FromColor(src)
	if _, ok := c.(colorNone); !ok {
		t.Fatalf("FromColor(): ожидался тип colorNone, получено %T(%v)", c, c)
	}
}

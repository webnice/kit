package dye

import "testing"

// TestColorNone Отсутствие раскраски, заглушка для терминалов без поддержки цвета.
func TestColorNone(t *testing.T) {
	var c colorNone
	if got := c.String(); got != "" {
		t.Fatalf("String(): ожидалась пустая строка, получено %q", got)
	}
	if got := c.Sequence(false); got != "" {
		t.Fatalf("Sequence(false): ожидалась пустая строка, получено %q", got)
	}
	if got := c.Sequence(true); got != "" {
		t.Fatalf("Sequence(true): ожидалась пустая строка, получено %q", got)
	}
}

// TestColorAnsi ANSI цвета, последовательности текста и фона обычной и повышенной яркости.
func TestColorAnsi(t *testing.T) {
	var cases = []struct {
		c   colorAnsi
		fg  string
		bg  string
		str string
	}{
		{colorANSIBlack, "30", "40", "#000000"},
		{colorANSIRed, "31", "41", "#800000"},
		{colorANSIWhite, "37", "47", "#c0c0c0"},
		{colorANSIBrightBlack, "90", "100", "#808080"},
		{colorANSIBrightRed, "91", "101", "#ff0000"},
		{colorANSIBrightWhite, "97", "107", "#ffffff"},
	}
	for _, tt := range cases {
		if got := tt.c.Sequence(false); got != tt.fg {
			t.Fatalf("Sequence(false): ожидалась последовательность %q, получено %q", tt.fg, got)
		}
		if got := tt.c.Sequence(true); got != tt.bg {
			t.Fatalf("Sequence(true): ожидалась последовательность %q, получено %q", tt.bg, got)
		}
		if got := tt.c.String(); got != tt.str {
			t.Fatalf("String(): ожидалось значение %q, получено %q", tt.str, got)
		}
	}
}

// TestColorAnsi256 Цвета ANSI256.
func TestColorAnsi256(t *testing.T) {
	var c colorAnsi256 = 196
	if got := c.Sequence(false); got != "38;5;196" {
		t.Fatalf("Sequence(false): ожидалась последовательность %q, получено %q", "38;5;196", got)
	}
	if got := c.Sequence(true); got != "48;5;196" {
		t.Fatalf("Sequence(true): ожидалась последовательность %q, получено %q", "48;5;196", got)
	}
	if got := c.String(); got != "#ff0000" {
		t.Fatalf("String(): ожидалось значение %q, получено %q", "#ff0000", got)
	}
	c = 16
	if got := c.Sequence(false); got != "38;5;16" {
		t.Fatalf("Sequence(false): ожидалась последовательность %q, получено %q", "38;5;16", got)
	}
}

// TestColorRgb RGB цвета.
func TestColorRgb(t *testing.T) {
	var c colorRgb = "#ff8000"
	if got := c.String(); got != "#ff8000" {
		t.Fatalf("String(): ожидалось значение %q, получено %q", "#ff8000", got)
	}
	if got := c.Sequence(false); got != "38;2;255;128;0" {
		t.Fatalf("Sequence(false): ожидалась последовательность %q, получено %q", "38;2;255;128;0", got)
	}
	if got := c.Sequence(true); got != "48;2;255;128;0" {
		t.Fatalf("Sequence(true): ожидалась последовательность %q, получено %q", "48;2;255;128;0", got)
	}
	c = "невалидный цвет"
	if got := c.Sequence(false); got != "" {
		t.Fatalf("Sequence(false): ожидалась пустая строка, получено %q", got)
	}
}

package dye

import "testing"

// TestColorNone Отсутствие раскраски, заглушка для терминалов без поддержки цвета.
func TestColorNone(t *testing.T) {
	const (
		tplStringNotEmpty   = "String(): ожидалась пустая строка, получено %q."
		tplSeqFalseNotEmpty = "Sequence(false): ожидалась пустая строка, получено %q."
		tplSeqTrueNotEmpty  = "Sequence(true): ожидалась пустая строка, получено %q."
	)
	var c colorNone

	if got := c.String(); got != "" {
		t.Fatalf(tplStringNotEmpty, got)
	}
	if got := c.Sequence(false); got != "" {
		t.Fatalf(tplSeqFalseNotEmpty, got)
	}
	if got := c.Sequence(true); got != "" {
		t.Fatalf(tplSeqTrueNotEmpty, got)
	}
}

// TestColorAnsi ANSI цвета, последовательности текста и фона обычной и повышенной яркости.
func TestColorAnsi(t *testing.T) {
	const (
		tplSeqFgMismatch  = "Sequence(false): ожидалась последовательность %q, получено %q."
		tplSeqBgMismatch  = "Sequence(true): ожидалась последовательность %q, получено %q."
		tplStringMismatch = "String(): ожидалось значение %q, получено %q."
	)
	// Описание одного теста ANSI цвета.
	type testCase struct {
		c   colorAnsi
		fg  string
		bg  string
		str string
	}
	var (
		cases []testCase
		tt    testCase
		n     int
	)

	cases = []testCase{
		{colorANSIBlack, "30", "40", "#000000"},
		{colorANSIRed, "31", "41", "#800000"},
		{colorANSIWhite, "37", "47", "#c0c0c0"},
		{colorANSIBrightBlack, "90", "100", "#808080"},
		{colorANSIBrightRed, "91", "101", "#ff0000"},
		{colorANSIBrightWhite, "97", "107", "#ffffff"},
	}
	for n = range cases {
		tt = cases[n]
		if got := tt.c.Sequence(false); got != tt.fg {
			t.Fatalf(tplSeqFgMismatch, tt.fg, got)
		}
		if got := tt.c.Sequence(true); got != tt.bg {
			t.Fatalf(tplSeqBgMismatch, tt.bg, got)
		}
		if got := tt.c.String(); got != tt.str {
			t.Fatalf(tplStringMismatch, tt.str, got)
		}
	}
}

// TestColorAnsi256 Цвета ANSI256.
func TestColorAnsi256(t *testing.T) {
	const (
		tplSeqFgMismatch  = "Sequence(false): ожидалась последовательность %q, получено %q."
		tplSeqBgMismatch  = "Sequence(true): ожидалась последовательность %q, получено %q."
		tplStringMismatch = "String(): ожидалось значение %q, получено %q."
	)
	var c colorAnsi256

	c = 196
	if got := c.Sequence(false); got != "38;5;196" {
		t.Fatalf(tplSeqFgMismatch, "38;5;196", got)
	}
	if got := c.Sequence(true); got != "48;5;196" {
		t.Fatalf(tplSeqBgMismatch, "48;5;196", got)
	}
	if got := c.String(); got != "#ff0000" {
		t.Fatalf(tplStringMismatch, "#ff0000", got)
	}
	c = 16
	if got := c.Sequence(false); got != "38;5;16" {
		t.Fatalf(tplSeqFgMismatch, "38;5;16", got)
	}
}

// TestColorRgb RGB цвета.
func TestColorRgb(t *testing.T) {
	const (
		tplStringMismatch = "String(): ожидалось значение %q, получено %q."
		tplSeqFgMismatch  = "Sequence(false): ожидалась последовательность %q, получено %q."
		tplSeqBgMismatch  = "Sequence(true): ожидалась последовательность %q, получено %q."
		tplSeqNotEmpty    = "Sequence(false): ожидалась пустая строка, получено %q."
	)
	var c colorRgb

	c = "#ff8000"
	if got := c.String(); got != "#ff8000" {
		t.Fatalf(tplStringMismatch, "#ff8000", got)
	}
	if got := c.Sequence(false); got != "38;2;255;128;0" {
		t.Fatalf(tplSeqFgMismatch, "38;2;255;128;0", got)
	}
	if got := c.Sequence(true); got != "48;2;255;128;0" {
		t.Fatalf(tplSeqBgMismatch, "48;2;255;128;0", got)
	}
	c = "невалидный цвет"
	if got := c.Sequence(false); got != "" {
		t.Fatalf(tplSeqNotEmpty, got)
	}
}

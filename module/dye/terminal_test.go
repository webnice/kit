package dye

import (
	"errors"
	"os"
	"testing"

	"github.com/webnice/kit/v5/module/dye/colorful"
)

// TestIsTTY Проверка TTY на файловом дескрипторе не являющимся терминалом.
func TestIsTTY(t *testing.T) {
	const (
		tplFatalOpen       = "Open(): неожиданная ошибка: %v."
		errDevNullNotFalse = "isTTY(): для /dev/null ожидалось false."
		errCiNotFalse      = "isTTY(): в режиме CI ожидалось false."
	)
	var (
		f   *os.File
		err error
	)

	f, err = os.Open(os.DevNull)
	if err != nil {
		t.Fatalf(tplFatalOpen, err)
	}
	defer func() { _ = f.Close() }()
	if isTTY(f.Fd()) {
		t.Fatal(errDevNullNotFalse)
	}
	t.Setenv(envCi, "true")
	if isTTY(f.Fd()) {
		t.Fatal(errCiNotFalse)
	}
}

// TestCliColorForced Проверка переменной окружения принудительного включения цвета.
func TestCliColorForced(t *testing.T) {
	const (
		errEmptyTrue  = "cliColorForced(): при пустой переменной окружения ожидалось false."
		errZeroTrue   = "cliColorForced(): при значении 0 ожидалось false."
		errOneFalse   = "cliColorForced(): при значении 1 ожидалось true."
		errValueFalse = "cliColorForced(): при непустом значении ожидалось true."
	)

	t.Setenv(envCliColorForce, "")
	if cliColorForced() {
		t.Fatal(errEmptyTrue)
	}
	t.Setenv(envCliColorForce, "0")
	if cliColorForced() {
		t.Fatal(errZeroTrue)
	}
	t.Setenv(envCliColorForce, "1")
	if !cliColorForced() {
		t.Fatal(errOneFalse)
	}
	t.Setenv(envCliColorForce, "true")
	if !cliColorForced() {
		t.Fatal(errValueFalse)
	}
}

// TestXTermColor Разбор ответа терминала на OSC запрос цвета.
func TestXTermColor(t *testing.T) {
	const (
		tplFatalXTermColor    = "xTermColor(%q): неожиданная ошибка: %v."
		tplXTermColorMismatch = "xTermColor(%q): ожидалось %q, получено %q."
		tplXTermColorInvalid  = "xTermColor(%q): ожидалась ошибка errInvalidColor, получено %v."
	)
	// Описание одного валидного теста разбора цвета.
	type validCase struct {
		in   string
		want colorRgb
	}
	var (
		valid   []validCase
		invalid []string
		tt      validCase
		got     colorRgb
		in      string
		err     error
		n       int
	)

	valid = []validCase{
		{"\x1b]11;rgb:1111/2222/3333\a", "#112233"},
		{"\x1b]11;rgb:1111/2222/3333\x1b", "#112233"},
		{"\x1b]11;rgb:1111/2222/3333\x1b\\", "#112233"},
		{"\x1b]10;rgb:aaaa/bbbb/cccc\a", "#aabbcc"},
	}
	for n = range valid {
		tt = valid[n]
		got, err = xTermColor(tt.in)
		if err != nil {
			t.Fatalf(tplFatalXTermColor, tt.in, err)
		}
		if got != tt.want {
			t.Fatalf(tplXTermColorMismatch, tt.in, tt.want, got)
		}
	}
	invalid = []string{
		"",
		"\x1b]11;rgb:1111/2222/3333",
		"\x1b]11;rgb:1111/2222/3333x",
		"\x1b]11;XXX:1111/2222/3333\a",
		"\x1b]11;rgb:11/22/33\a",
	}
	for n = range invalid {
		in = invalid[n]
		if _, err = xTermColor(in); !errors.Is(err, errInvalidColor) {
			t.Fatalf(tplXTermColorInvalid, in, err)
		}
	}
}

// TestAnsi256ToANSIColor Поиск ближайшего ANSI цвета для ANSI256 цвета.
func TestAnsi256ToANSIColor(t *testing.T) {
	const (
		tplAnsiMismatch   = "ansi256ToANSIColor(%d): ожидалось %d, получено %d."
		tplAnsiOutOfRange = "ansi256ToANSIColor(%d): результат %d вне диапазона 0-15."
	)
	// Описание одного теста конвертации ANSI256 в ANSI.
	type testCase struct {
		in   colorAnsi256
		want colorAnsi
	}
	var (
		cases []testCase
		tt    testCase
		n, i  int
	)

	cases = []testCase{
		{16, colorANSIBlack},
		{196, colorANSIBrightRed},
		{21, colorANSIBrightBlue},
		{46, colorANSIBrightGreen},
		{231, colorANSIBrightWhite},
		{232, colorANSIBlack},
	}
	for n = range cases {
		tt = cases[n]
		if got := ansi256ToANSIColor(tt.in); got != tt.want {
			t.Fatalf(tplAnsiMismatch, tt.in, tt.want, got)
		}
	}
	for i = 0; i < 256; i++ {
		if got := ansi256ToANSIColor(colorAnsi256(i)); got < 0 || got > 15 {
			t.Fatalf(tplAnsiOutOfRange, i, got)
		}
	}
}

// TestHexToANSI256Color Поиск ближайшего ANSI256 цвета для RGB цвета.
func TestHexToANSI256Color(t *testing.T) {
	const tplHexMismatch = "hexToANSI256Color(%v): ожидалось %d, получено %d."
	// Описание одного теста конвертации RGB в ANSI256.
	type testCase struct {
		in   colorful.Color
		want colorAnsi256
	}
	var (
		cases []testCase
		tt    testCase
		n     int
	)

	cases = []testCase{
		{colorful.Color{R: 0, G: 0, B: 0}, 16},
		{colorful.Color{R: 1, G: 1, B: 1}, 231},
		{colorful.Color{R: 8.0 / 255.0, G: 8.0 / 255.0, B: 8.0 / 255.0}, 232},
		{colorful.Color{R: 1, G: 0, B: 0}, 196},
		{colorful.Color{R: 64.0 / 255.0, G: 64.0 / 255.0, B: 64.0 / 255.0}, 59},
	}
	for n = range cases {
		tt = cases[n]
		if got := hexToANSI256Color(tt.in); got != tt.want {
			t.Fatalf(tplHexMismatch, tt.in, tt.want, got)
		}
	}
}

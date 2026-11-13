package dye

import (
	"errors"
	"os"
	"testing"

	"github.com/webnice/kit/v4/module/dye/colorful"
)

// TestIsTTY Проверка TTY на файловом дескрипторе не являющимся терминалом.
func TestIsTTY(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("Open(): неожиданная ошибка: %v", err)
	}
	defer func() { _ = f.Close() }()
	if isTTY(f.Fd()) {
		t.Fatal("isTTY(): для /dev/null ожидалось false")
	}
	t.Setenv("CI", "true")
	if isTTY(f.Fd()) {
		t.Fatal("isTTY(): в режиме CI ожидалось false")
	}
}

// TestCliColorForced Проверка переменной окружения принудительного включения цвета.
func TestCliColorForced(t *testing.T) {
	t.Setenv(envCliColorForce, "")
	if cliColorForced() {
		t.Fatal("cliColorForced(): при пустой переменной окружения ожидалось false")
	}
	t.Setenv(envCliColorForce, "0")
	if cliColorForced() {
		t.Fatal("cliColorForced(): при значении 0 ожидалось false")
	}
	t.Setenv(envCliColorForce, "1")
	if !cliColorForced() {
		t.Fatal("cliColorForced(): при значении 1 ожидалось true")
	}
	t.Setenv(envCliColorForce, "true")
	if !cliColorForced() {
		t.Fatal("cliColorForced(): при непустом значении ожидалось true")
	}
}

// TestXTermColor Разбор ответа терминала на OSC запрос цвета.
func TestXTermColor(t *testing.T) {
	var valid = []struct {
		in   string
		want colorRgb
	}{
		{"\x1b]11;rgb:1111/2222/3333\a", "#112233"},
		{"\x1b]11;rgb:1111/2222/3333\x1b", "#112233"},
		{"\x1b]11;rgb:1111/2222/3333\x1b\\", "#112233"},
		{"\x1b]10;rgb:aaaa/bbbb/cccc\a", "#aabbcc"},
	}
	for _, tt := range valid {
		got, err := xTermColor(tt.in)
		if err != nil {
			t.Fatalf("xTermColor(%q): неожиданная ошибка: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("xTermColor(%q): ожидалось %q, получено %q", tt.in, tt.want, got)
		}
	}
	var invalid = []string{
		"",
		"\x1b]11;rgb:1111/2222/3333",
		"\x1b]11;rgb:1111/2222/3333x",
		"\x1b]11;XXX:1111/2222/3333\a",
		"\x1b]11;rgb:11/22/33\a",
	}
	for _, in := range invalid {
		if _, err := xTermColor(in); !errors.Is(err, errInvalidColor) {
			t.Fatalf("xTermColor(%q): ожидалась ошибка errInvalidColor, получено %v", in, err)
		}
	}
}

// TestAnsi256ToANSIColor Поиск ближайшего ANSI цвета для ANSI256 цвета.
func TestAnsi256ToANSIColor(t *testing.T) {
	var cases = []struct {
		in   colorAnsi256
		want colorAnsi
	}{
		{16, colorANSIBlack},
		{196, colorANSIBrightRed},
		{21, colorANSIBrightBlue},
		{46, colorANSIBrightGreen},
		{231, colorANSIBrightWhite},
		{232, colorANSIBlack},
	}
	for _, tt := range cases {
		if got := ansi256ToANSIColor(tt.in); got != tt.want {
			t.Fatalf("ansi256ToANSIColor(%d): ожидалось %d, получено %d", tt.in, tt.want, got)
		}
	}
	for i := 0; i < 256; i++ {
		if got := ansi256ToANSIColor(colorAnsi256(i)); got < 0 || got > 15 {
			t.Fatalf("ansi256ToANSIColor(%d): результат %d вне диапазона 0-15", i, got)
		}
	}
}

// TestHexToANSI256Color Поиск ближайшего ANSI256 цвета для RGB цвета.
func TestHexToANSI256Color(t *testing.T) {
	var cases = []struct {
		in   colorful.Color
		want colorAnsi256
	}{
		{colorful.Color{R: 0, G: 0, B: 0}, 16},
		{colorful.Color{R: 1, G: 1, B: 1}, 231},
		{colorful.Color{R: 8.0 / 255.0, G: 8.0 / 255.0, B: 8.0 / 255.0}, 232},
		{colorful.Color{R: 1, G: 0, B: 0}, 196},
		{colorful.Color{R: 64.0 / 255.0, G: 64.0 / 255.0, B: 64.0 / 255.0}, 59},
	}
	for _, tt := range cases {
		if got := hexToANSI256Color(tt.in); got != tt.want {
			t.Fatalf("hexToANSI256Color(%v): ожидалось %d, получено %d", tt.in, tt.want, got)
		}
	}
}

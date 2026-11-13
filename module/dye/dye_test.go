package dye

import (
	"bytes"
	"strings"
	"testing"
)

// testImpl Конструктор объекта с фиксированным профилем терминала, для детерминированных тестов.
func testImpl(profile terminalProfile) *impl { return &impl{profile: profile} }

// TestNew Конструктор пакета без последовательностей.
func TestNew(t *testing.T) {
	d := New()
	if got := d.Source(); len(got) != 0 {
		t.Fatalf("New(): ожидалась пустая последовательность, получено %q", strings.Join(got, ";"))
	}
}

// TestNewWithSequences Конструктор пакета с начальными последовательностями.
func TestNewWithSequences(t *testing.T) {
	d := New("1", "31")
	if got := strings.Join(d.Source(), ";"); got != "1;31" {
		t.Fatalf("New(1, 31): ожидалась последовательность %q, получено %q", "1;31", got)
	}
	if got := d.Done().String(); got != "\x1b[1;31m" {
		t.Fatalf("Done(): ожидалась последовательность %q, получено %q", "\x1b[1;31m", got)
	}
}

// TestAdd Добавление последовательностей в конец текущей последовательности.
func TestAdd(t *testing.T) {
	d := New()
	if got := d.Add("1").Add("31", "32").Done().String(); got != "\x1b[1;31;32m" {
		t.Fatalf("Add(): ожидалась последовательность %q, получено %q", "\x1b[1;31;32m", got)
	}
	if got := len(d.Source()); got != 3 {
		t.Fatalf("Source(): ожидалось %d элемента, получено %d", 3, got)
	}
}

// TestDone Завершение последовательности, в том числе пустой.
func TestDone(t *testing.T) {
	if got := New().Done().String(); got != "\x1b[m" {
		t.Fatalf("Done() пустой: ожидалась последовательность %q, получено %q", "\x1b[m", got)
	}
	seq := New("0").Done()
	if !bytes.Equal(seq.Byte(), []byte("\x1b[0m")) {
		t.Fatalf("Byte(): ожидалась последовательность %q, получено %q", "\x1b[0m", seq.Byte())
	}
}

// TestSwitches Переключатели текста.
func TestSwitches(t *testing.T) {
	var cases = []struct {
		name string
		want string
		fn   func(d Interface) Interface
	}{
		{"Bold", "1", func(d Interface) Interface { return d.Bold() }},
		{"Faded", "2", func(d Interface) Interface { return d.Faded() }},
		{"Italic", "3", func(d Interface) Interface { return d.Italic() }},
		{"Underline", "4", func(d Interface) Interface { return d.Underline() }},
		{"Reverse", "7", func(d Interface) Interface { return d.Reverse() }},
		{"CrossOut", "9", func(d Interface) Interface { return d.CrossOut() }},
		{"ResetBold", "21", func(d Interface) Interface { return d.ResetBold() }},
		{"ResetFaded", "22", func(d Interface) Interface { return d.ResetFaded() }},
		{"ResetItalic", "23", func(d Interface) Interface { return d.ResetItalic() }},
		{"ResetUnderline", "24", func(d Interface) Interface { return d.ResetUnderline() }},
		{"ResetReverse", "27", func(d Interface) Interface { return d.ResetReverse() }},
		{"ResetCrossOut", "29", func(d Interface) Interface { return d.ResetCrossOut() }},
		{"Reset", "0", func(d Interface) Interface { return d.Reset() }},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(New()).Done().String(); got != "\x1b["+tt.want+"m" {
				t.Fatalf("%s: ожидалась последовательность %q, получено %q", tt.name, "\x1b["+tt.want+"m", got)
			}
		})
	}
}

// TestColors Цвета текста, варианты яркости и фона.
func TestColors(t *testing.T) {
	var cases = []struct {
		name string
		want string
		fn   func(d Interface) Interface
	}{
		{"Black", "30", func(d Interface) Interface { return d.Black() }},
		{"Red", "31", func(d Interface) Interface { return d.Red() }},
		{"Green", "32", func(d Interface) Interface { return d.Green() }},
		{"Yellow", "33", func(d Interface) Interface { return d.Yellow() }},
		{"Blue", "34", func(d Interface) Interface { return d.Blue() }},
		{"Magenta", "35", func(d Interface) Interface { return d.Magenta() }},
		{"Cyan", "36", func(d Interface) Interface { return d.Cyan() }},
		{"White", "37", func(d Interface) Interface { return d.White() }},
		{"BrightBlack", "90", func(d Interface) Interface { return d.Bright().Black() }},
		{"BrightRed", "91", func(d Interface) Interface { return d.Bright().Red() }},
		{"BrightGreen", "92", func(d Interface) Interface { return d.Bright().Green() }},
		{"BrightYellow", "93", func(d Interface) Interface { return d.Bright().Yellow() }},
		{"BrightBlue", "94", func(d Interface) Interface { return d.Bright().Blue() }},
		{"BrightMagenta", "95", func(d Interface) Interface { return d.Bright().Magenta() }},
		{"BrightCyan", "96", func(d Interface) Interface { return d.Bright().Cyan() }},
		{"BrightWhite", "97", func(d Interface) Interface { return d.Bright().White() }},
		{"BackgroundBlack", "40", func(d Interface) Interface { return d.Background().Black() }},
		{"BackgroundRed", "41", func(d Interface) Interface { return d.Background().Red() }},
		{"BackgroundGreen", "42", func(d Interface) Interface { return d.Background().Green() }},
		{"BackgroundYellow", "43", func(d Interface) Interface { return d.Background().Yellow() }},
		{"BackgroundBlue", "44", func(d Interface) Interface { return d.Background().Blue() }},
		{"BackgroundMagenta", "45", func(d Interface) Interface { return d.Background().Magenta() }},
		{"BackgroundCyan", "46", func(d Interface) Interface { return d.Background().Cyan() }},
		{"BackgroundWhite", "47", func(d Interface) Interface { return d.Background().White() }},
		{"BrightBackgroundBlack", "100", func(d Interface) Interface { return d.Bright().Background().Black() }},
		{"BrightBackgroundRed", "101", func(d Interface) Interface { return d.Bright().Background().Red() }},
		{"BrightBackgroundGreen", "102", func(d Interface) Interface { return d.Bright().Background().Green() }},
		{"BrightBackgroundYellow", "103", func(d Interface) Interface { return d.Bright().Background().Yellow() }},
		{"BrightBackgroundBlue", "104", func(d Interface) Interface { return d.Bright().Background().Blue() }},
		{"BrightBackgroundMagenta", "105", func(d Interface) Interface { return d.Bright().Background().Magenta() }},
		{"BrightBackgroundCyan", "106", func(d Interface) Interface { return d.Bright().Background().Cyan() }},
		{"BrightBackgroundWhite", "107", func(d Interface) Interface { return d.Bright().Background().White() }},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(New()).Done().String(); got != "\x1b["+tt.want+"m" {
				t.Fatalf("%s: ожидалась последовательность %q, получено %q", tt.name, "\x1b["+tt.want+"m", got)
			}
		})
	}
}

// TestColorFlagsReset Флаги яркости и фона сбрасываются после добавления цвета в последовательность.
func TestColorFlagsReset(t *testing.T) {
	var cases = []struct {
		name string
		want string
		fn   func(d Interface) Interface
	}{
		{"BackgroundReset", "\x1b[41;31m", func(d Interface) Interface { return d.Background().Red().Red() }},
		{"BrightReset", "\x1b[91;34m", func(d Interface) Interface { return d.Bright().Red().Blue() }},
		{"Normal", "\x1b[31m", func(d Interface) Interface { return d.Bright().Normal().Red() }},
		{"Foreground", "\x1b[31m", func(d Interface) Interface { return d.Background().Foreground().Red() }},
		{"BrightNormal", "\x1b[31m", func(d Interface) Interface { return d.Normal().Bright().Normal().Red() }},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(New()).Done().String(); got != tt.want {
				t.Fatalf("%s: ожидалась последовательность %q, получено %q", tt.name, tt.want, got)
			}
		})
	}
}

// TestHexRgb Конвертация HEX RGB цвета в цвет терминала.
func TestHexRgb(t *testing.T) {
	t.Run("TrueColor", func(t *testing.T) {
		d := testImpl(terminalTrueColor)
		if err := d.HexRgb("#ff8000"); err != nil {
			t.Fatalf("HexRgb(): неожиданная ошибка: %v", err)
		}
		if got := d.Done().String(); got != "\x1b[38;2;255;128;0m" {
			t.Fatalf("HexRgb(): ожидалась последовательность %q, получено %q", "\x1b[38;2;255;128;0m", got)
		}
	})
	t.Run("TrueColorUpperCase", func(t *testing.T) {
		d := testImpl(terminalTrueColor)
		if err := d.HexRgb("#FF8000"); err != nil {
			t.Fatalf("HexRgb(): неожиданная ошибка: %v", err)
		}
		if got := d.Done().String(); got != "\x1b[38;2;255;128;0m" {
			t.Fatalf("HexRgb(): ожидалась последовательность %q, получено %q", "\x1b[38;2;255;128;0m", got)
		}
	})
	t.Run("TrueColorBackground", func(t *testing.T) {
		d := testImpl(terminalTrueColor)
		if err := d.Background().HexRgb("#ff8000"); err != nil {
			t.Fatalf("HexRgb(): неожиданная ошибка: %v", err)
		}
		if got := d.Done().String(); got != "\x1b[48;2;255;128;0m" {
			t.Fatalf("HexRgb(): ожидалась последовательность %q, получено %q", "\x1b[48;2;255;128;0m", got)
		}
	})
	t.Run("BackgroundFlagReset", func(t *testing.T) {
		d := testImpl(terminalTrueColor)
		if err := d.Background().HexRgb("#ff8000"); err != nil {
			t.Fatalf("HexRgb(): неожиданная ошибка: %v", err)
		}
		// Флаг фона должен быть сброшен, следующий цвет - цвет текста.
		d.Red()
		if got := d.Done().String(); got != "\x1b[48;2;255;128;0;31m" {
			t.Fatalf("HexRgb(): ожидалась последовательность %q, получено %q", "\x1b[48;2;255;128;0;31m", got)
		}
	})
	t.Run("ANSI256", func(t *testing.T) {
		d := testImpl(terminalANSI256)
		if err := d.HexRgb("#ff0000"); err != nil {
			t.Fatalf("HexRgb(): неожиданная ошибка: %v", err)
		}
		if got := d.Done().String(); got != "\x1b[38;5;196m" {
			t.Fatalf("HexRgb(): ожидалась последовательность %q, получено %q", "\x1b[38;5;196m", got)
		}
	})
	t.Run("ANSI", func(t *testing.T) {
		d := testImpl(terminalANSI)
		if err := d.HexRgb("#ff0000"); err != nil {
			t.Fatalf("HexRgb(): неожиданная ошибка: %v", err)
		}
		if got := d.Done().String(); got != "\x1b[91m" {
			t.Fatalf("HexRgb(): ожидалась последовательность %q, получено %q", "\x1b[91m", got)
		}
	})
	t.Run("Ascii", func(t *testing.T) {
		d := testImpl(terminalAscii)
		if err := d.HexRgb("#ff0000"); err != nil {
			t.Fatalf("HexRgb(): неожиданная ошибка: %v", err)
		}
		if got := d.Done().String(); got != "\x1b[m" {
			t.Fatalf("HexRgb(): ожидалась последовательность %q, получено %q", "\x1b[m", got)
		}
	})
	t.Run("BadParams", func(t *testing.T) {
		for _, param := range []string{"", "ff0000", "#fff", "#fffff", "#fffffff", "#ff00000", "#ff000g", "#gg0000"} {
			d := testImpl(terminalTrueColor)
			err := d.HexRgb(param)
			if err == nil {
				t.Fatalf("HexRgb(%q): ожидалась ошибка, получено без ошибки", param)
			}
			if !strings.Contains(err.Error(), "не верный параметр") {
				t.Fatalf("HexRgb(%q): неожиданная ошибка: %v", param, err)
			}
		}
	})
}

// TestInterface Компиляционная проверка реализации интерфейсов пакета.
func TestInterface(t *testing.T) {
	var _ Interface = New()
	var _ Sequence = New().Done()
	var _ Color = colorNone{}
	var _ Color = colorAnsi(0)
	var _ Color = colorAnsi256(0)
	var _ Color = colorRgb("")
}

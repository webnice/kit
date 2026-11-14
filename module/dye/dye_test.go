package dye

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// Конструктор объекта с фиксированным профилем терминала, для детерминированных тестов.
func testImpl(profile terminalProfile) *impl { return &impl{profile: profile} }

// TestNew Конструктор пакета без последовательностей.
func TestNew(t *testing.T) {
	const tplNewNotEmpty = "New(): ожидалась пустая последовательность, получено %q."
	var d Interface

	d = New()
	if got := d.Source(); len(got) != 0 {
		t.Fatalf(tplNewNotEmpty, strings.Join(got, ";"))
	}
}

// TestNewWithSequences Конструктор пакета с начальными последовательностями.
func TestNewWithSequences(t *testing.T) {
	const (
		tplSeqMismatch   = "New(1, 31): ожидалась последовательность %q, получено %q."
		tplDoneMismatch  = "Done(): ожидалась последовательность %q, получено %q."
		sequenceExpected = "1;31"
		sequenceDone     = "\x1b[1;31m"
	)
	var (
		d   Interface
		got string
	)

	d = New("1", "31")
	got = strings.Join(d.Source(), ";")
	if got != sequenceExpected {
		t.Fatalf(tplSeqMismatch, sequenceExpected, got)
	}
	got = d.Done().String()
	if got != sequenceDone {
		t.Fatalf(tplDoneMismatch, sequenceDone, got)
	}
}

// TestAdd Добавление последовательностей в конец текущей последовательности.
func TestAdd(t *testing.T) {
	const (
		tplAddMismatch   = "Add(): ожидалась последовательность %q, получено %q."
		tplCountMismatch = "Source(): ожидалось %d элемента, получено %d."
		sequenceExpected = "\x1b[1;31;32m"
	)
	var (
		d   Interface
		got string
	)

	d = New()
	got = d.Add("1").Add("31", "32").Done().String()
	if got != sequenceExpected {
		t.Fatalf(tplAddMismatch, sequenceExpected, got)
	}
	if got = string(rune(len(d.Source()))); len(d.Source()) != 3 {
		t.Fatalf(tplCountMismatch, 3, len(d.Source()))
	}
	_ = got
}

// TestDone Завершение последовательности, в том числе пустой.
func TestDone(t *testing.T) {
	const (
		tplDoneEmpty    = "Done() пустой: ожидалась последовательность %q, получено %q."
		tplByteMismatch = "Byte(): ожидалась последовательность %q, получено %q."
		seqEmpty        = "\x1b[m"
		seqZero         = "\x1b[0m"
	)
	var (
		got string
		seq Sequence
	)

	got = New().Done().String()
	if got != seqEmpty {
		t.Fatalf(tplDoneEmpty, seqEmpty, got)
	}
	seq = New("0").Done()
	if !bytes.Equal(seq.Byte(), []byte(seqZero)) {
		t.Fatalf(tplByteMismatch, seqZero, seq.Byte())
	}
}

// Описание одного теста последовательности управления.
type testSequenceCase struct {
	name string
	want string
	fn   func(d Interface) Interface
}

// TestSwitches Переключатели текста.
func TestSwitches(t *testing.T) {
	const (
		tplSeqCaseMismatch = "%s: ожидалась последовательность %q, получено %q."
		seqTemplate        = "\x1b[%sm"
	)
	var (
		cases []testSequenceCase
		n     int
		got   string
	)

	cases = []testSequenceCase{
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
	for n = range cases {
		t.Run(cases[n].name, func(t *testing.T) {
			if got = cases[n].fn(New()).Done().String(); got != fmt.Sprintf(seqTemplate, cases[n].want) {
				t.Fatalf(tplSeqCaseMismatch, cases[n].name, fmt.Sprintf(seqTemplate, cases[n].want), got)
			}
		})
	}
}

// TestColors Цвета текста, варианты яркости и фона.
func TestColors(t *testing.T) {
	const (
		tplSeqCaseMismatch = "%s: ожидалась последовательность %q, получено %q."
		seqTemplate        = "\x1b[%sm"
	)
	var (
		cases []testSequenceCase
		n     int
		got   string
	)

	cases = []testSequenceCase{
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
	for n = range cases {
		t.Run(cases[n].name, func(t *testing.T) {
			if got = cases[n].fn(New()).Done().String(); got != fmt.Sprintf(seqTemplate, cases[n].want) {
				t.Fatalf(tplSeqCaseMismatch, cases[n].name, fmt.Sprintf(seqTemplate, cases[n].want), got)
			}
		})
	}
}

// TestColorFlagsReset Флаги яркости и фона сбрасываются после добавления цвета в последовательность.
func TestColorFlagsReset(t *testing.T) {
	const tplSeqCaseMismatch = "%s: ожидалась последовательность %q, получено %q."
	var (
		cases []testSequenceCase
		n     int
		got   string
	)

	cases = []testSequenceCase{
		{"BackgroundReset", "\x1b[41;31m", func(d Interface) Interface { return d.Background().Red().Red() }},
		{"BrightReset", "\x1b[91;34m", func(d Interface) Interface { return d.Bright().Red().Blue() }},
		{"Normal", "\x1b[31m", func(d Interface) Interface { return d.Bright().Normal().Red() }},
		{"Foreground", "\x1b[31m", func(d Interface) Interface { return d.Background().Foreground().Red() }},
		{"BrightNormal", "\x1b[31m", func(d Interface) Interface { return d.Normal().Bright().Normal().Red() }},
	}
	for n = range cases {
		t.Run(cases[n].name, func(t *testing.T) {
			if got = cases[n].fn(New()).Done().String(); got != cases[n].want {
				t.Fatalf(tplSeqCaseMismatch, cases[n].name, cases[n].want, got)
			}
		})
	}
}

// TestHexRgb Конвертация HEX RGB цвета в цвет терминала.
func TestHexRgb(t *testing.T) {
	const (
		tplFatalHexRgb      = "HexRgb(): неожиданная ошибка: %v."
		tplHexRgbMismatch   = "HexRgb(): ожидалась последовательность %q, получено %q."
		tplHexRgbNoError    = "HexRgb(%q): ожидалась ошибка, получено без ошибки."
		tplHexRgbUnexpected = "HexRgb(%q): неожиданная ошибка: %v."
		errParam            = "не верный параметр"
	)
	t.Run("TrueColor", func(t *testing.T) {
		const (
			hexColor   = "#ff8000"
			seqFgColor = "\x1b[38;2;255;128;0m"
		)
		var (
			d   *impl
			got string
		)

		d = testImpl(terminalTrueColor)
		if err := d.HexRgb(hexColor); err != nil {
			t.Fatalf(tplFatalHexRgb, err)
		}
		got = d.Done().String()
		if got != seqFgColor {
			t.Fatalf(tplHexRgbMismatch, seqFgColor, got)
		}
	})
	t.Run("TrueColorUpperCase", func(t *testing.T) {
		const (
			hexColor   = "#FF8000"
			seqFgColor = "\x1b[38;2;255;128;0m"
		)
		var (
			d   *impl
			got string
		)

		d = testImpl(terminalTrueColor)
		if err := d.HexRgb(hexColor); err != nil {
			t.Fatalf(tplFatalHexRgb, err)
		}
		got = d.Done().String()
		if got != seqFgColor {
			t.Fatalf(tplHexRgbMismatch, seqFgColor, got)
		}
	})
	t.Run("TrueColorBackground", func(t *testing.T) {
		const (
			hexColor   = "#ff8000"
			seqBgColor = "\x1b[48;2;255;128;0m"
		)
		var (
			d   *impl
			got string
		)

		d = testImpl(terminalTrueColor)
		if err := d.Background().HexRgb(hexColor); err != nil {
			t.Fatalf(tplFatalHexRgb, err)
		}
		got = d.Done().String()
		if got != seqBgColor {
			t.Fatalf(tplHexRgbMismatch, seqBgColor, got)
		}
	})
	t.Run("BackgroundFlagReset", func(t *testing.T) {
		const (
			hexColor       = "#ff8000"
			seqBgFlagReset = "\x1b[48;2;255;128;0;31m"
		)
		var (
			d   *impl
			got string
		)

		d = testImpl(terminalTrueColor)
		if err := d.Background().HexRgb(hexColor); err != nil {
			t.Fatalf(tplFatalHexRgb, err)
		}
		// Флаг фона должен быть сброшен, следующий цвет - цвет текста.
		d.Red()
		got = d.Done().String()
		if got != seqBgFlagReset {
			t.Fatalf(tplHexRgbMismatch, seqBgFlagReset, got)
		}
	})
	t.Run("ANSI256", func(t *testing.T) {
		const (
			hexColor = "#ff0000"
			seqColor = "\x1b[38;5;196m"
		)
		var (
			d   *impl
			got string
		)

		d = testImpl(terminalANSI256)
		if err := d.HexRgb(hexColor); err != nil {
			t.Fatalf(tplFatalHexRgb, err)
		}
		got = d.Done().String()
		if got != seqColor {
			t.Fatalf(tplHexRgbMismatch, seqColor, got)
		}
	})
	t.Run("ANSI", func(t *testing.T) {
		const (
			hexColor = "#ff0000"
			seqColor = "\x1b[91m"
		)
		var (
			d   *impl
			got string
		)

		d = testImpl(terminalANSI)
		if err := d.HexRgb(hexColor); err != nil {
			t.Fatalf(tplFatalHexRgb, err)
		}
		got = d.Done().String()
		if got != seqColor {
			t.Fatalf(tplHexRgbMismatch, seqColor, got)
		}
	})
	t.Run("Ascii", func(t *testing.T) {
		const (
			hexColor = "#ff0000"
			seqColor = "\x1b[m"
		)
		var (
			d   *impl
			got string
		)

		d = testImpl(terminalAscii)
		if err := d.HexRgb(hexColor); err != nil {
			t.Fatalf(tplFatalHexRgb, err)
		}
		got = d.Done().String()
		if got != seqColor {
			t.Fatalf(tplHexRgbMismatch, seqColor, got)
		}
	})
	t.Run("BadParams", func(t *testing.T) {
		var (
			d     *impl
			param string
			err   error
		)

		for _, param = range []string{"", "ff0000", "#fff", "#fffff", "#fffffff", "#ff00000", "#ff000g", "#gg0000"} {
			d = testImpl(terminalTrueColor)
			err = d.HexRgb(param)
			if err == nil {
				t.Fatalf(tplHexRgbNoError, param)
			}
			if !strings.Contains(err.Error(), errParam) {
				t.Fatalf(tplHexRgbUnexpected, param, err)
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

//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package dye

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// TestColorProfile Определение профиля терминала по переменным окружения.
func TestColorProfile(t *testing.T) {
	const (
		tplProfileMismatch = "colorProfile(): ожидался профиль %d, получено %d."
	)
	// Описание одного теста определения профиля терминала.
	type testCase struct {
		name      string
		term      string
		colorTerm string
		program   string
		want      terminalProfile
	}
	var (
		cases []testCase
		tt    testCase
		n     int
	)

	cases = []testCase{
		{"TrueColor24Bit", "xterm-256color", "24bit", "", terminalTrueColor},
		{"TrueColorTrueColor", "xterm-256color", "truecolor", "", terminalTrueColor},
		{"ScreenNoTmux", "screen-256color", "truecolor", "", terminalANSI256},
		{"ScreenTmux", "screen-256color", "truecolor", "tmux", terminalTrueColor},
		{"ColorTermYes", "xterm", "yes", "", terminalANSI256},
		{"ColorTermTrue", "xterm", "true", "", terminalANSI256},
		{"XtermKitty", "xterm-kitty", "", "", terminalTrueColor},
		{"Linux", "linux", "", "", terminalANSI},
		{"Term256Color", "xterm-256color", "", "", terminalANSI256},
		{"TermColor", "xterm-color", "", "", terminalANSI},
		{"TermAnsi", "ansi", "", "", terminalANSI},
		{"Dumb", "dumb", "", "", terminalAscii},
		{"Empty", "", "", "", terminalAscii},
	}
	for n = range cases {
		tt = cases[n]
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envTerm, tt.term)
			t.Setenv(envColorTerm, tt.colorTerm)
			t.Setenv(envTermProgram, tt.program)
			if got := colorProfile(); got != tt.want {
				t.Fatalf(tplProfileMismatch, tt.want, got)
			}
		})
	}
}

// TestForegroundColor Цвет текста терминала, ветки переменной окружения COLORFGBG.
func TestForegroundColor(t *testing.T) {
	const (
		tplColorMismatch = "foregroundColor(): ожидался цвет %v, получено %v."
	)

	t.Setenv(envTerm, "screen") // Отключение запроса к терминалу, ответ будет получен из окружения.
	t.Run("ColorFgbg", func(t *testing.T) {
		t.Setenv(envColorFgbg, "1;2")
		if got := foregroundColor(); got != colorAnsi(1) {
			t.Fatalf(tplColorMismatch, colorAnsi(1), got)
		}
	})
	t.Run("ColorFgbgBadValue", func(t *testing.T) {
		t.Setenv(envColorFgbg, "default;default")
		if got := foregroundColor(); got != colorAnsi(7) {
			t.Fatalf(tplColorMismatch, colorAnsi(7), got)
		}
	})
	t.Run("Default", func(t *testing.T) {
		t.Setenv(envColorFgbg, "")
		if got := foregroundColor(); got != colorAnsi(7) {
			t.Fatalf(tplColorMismatch, colorAnsi(7), got)
		}
	})
}

// TestBackgroundColor Цвет фона терминала, ветки переменной окружения COLORFGBG.
func TestBackgroundColor(t *testing.T) {
	const (
		tplColorMismatch = "backgroundColor(): ожидался цвет %v, получено %v."
	)

	t.Setenv(envTerm, "screen") // Отключение запроса к терминалу, ответ будет получен из окружения.
	t.Run("ColorFgbg", func(t *testing.T) {
		t.Setenv(envColorFgbg, "1;2")
		if got := backgroundColor(); got != colorAnsi(2) {
			t.Fatalf(tplColorMismatch, colorAnsi(2), got)
		}
	})
	t.Run("ColorFgbgBadValue", func(t *testing.T) {
		t.Setenv(envColorFgbg, "default;default")
		if got := backgroundColor(); got != colorAnsi(0) {
			t.Fatalf(tplColorMismatch, colorAnsi(0), got)
		}
	})
	t.Run("Default", func(t *testing.T) {
		t.Setenv(envColorFgbg, "")
		if got := backgroundColor(); got != colorAnsi(0) {
			t.Fatalf(tplColorMismatch, colorAnsi(0), got)
		}
	})
}

// TestTermStatusReport Запрос отчёта статуса терминала в среде без терминала.
func TestTermStatusReport(t *testing.T) {
	const (
		tplFatalStatusReport = "termStatusReport(): ожидалась ошибка errStatusReport, получено %v."
		skipStdoutTerminal   = "пропуск: stdout является терминалом."
	)

	t.Run("Screen", func(t *testing.T) {
		t.Setenv(envTerm, "screen-256color")
		if _, err := termStatusReport(11); !errors.Is(err, errStatusReport) {
			t.Fatalf(tplFatalStatusReport, err)
		}
	})
	t.Run("NoTerminal", func(t *testing.T) {
		if isTTY(os.Stdout.Fd()) {
			t.Skip(skipStdoutTerminal)
		}
		t.Setenv(envTerm, "xterm-256color")
		if _, err := termStatusReport(11); !errors.Is(err, errStatusReport) {
			t.Fatalf(tplFatalStatusReport, err)
		}
	})
}

// TestWaitForData Ожидание данных на файловом дескрипторе.
func TestWaitForData(t *testing.T) {
	const (
		tplFatalPipe      = "Pipe(): неожиданная ошибка: %v."
		tplFatalWait      = "waitForData(): ожидалась ошибка errTimeout, получено %v."
		tplFatalWaitError = "waitForData(): неожиданная ошибка: %v."
		tplFatalWrite     = "Write(): неожиданная ошибка: %v."
	)
	var (
		r   *os.File
		w   *os.File
		err error
	)

	r, w, err = os.Pipe()
	if err != nil {
		t.Fatalf(tplFatalPipe, err)
	}
	defer func() { _ = r.Close() }()
	defer func() { _ = w.Close() }()
	if err = waitForData(r.Fd(), 50*time.Millisecond); !errors.Is(err, errTimeout) {
		t.Fatalf(tplFatalWait, err)
	}
	if _, err = w.Write([]byte("x")); err != nil {
		t.Fatalf(tplFatalWrite, err)
	}
	if err = waitForData(r.Fd(), time.Second); err != nil {
		t.Fatalf(tplFatalWaitError, err)
	}
}

// TestReadNextByte Чтение одного байта из файлового дескриптора.
func TestReadNextByte(t *testing.T) {
	const (
		tplFatalPipe      = "Pipe(): неожиданная ошибка: %v."
		tplFatalWrite     = "Write(): неожиданная ошибка: %v."
		tplFatalRead      = "readNextByte(): неожиданная ошибка: %v."
		tplByteMismatch   = "readNextByte(): ожидался байт %q, получено %q."
		tplFatalOpen      = "Open(): неожиданная ошибка: %v."
		errReadDirNoError = "readNextByte(): при чтении каталога ожидалась ошибка."
	)
	var (
		r   *os.File
		w   *os.File
		d   *os.File
		b   byte
		err error
	)

	r, w, err = os.Pipe()
	if err != nil {
		t.Fatalf(tplFatalPipe, err)
	}
	defer func() { _ = r.Close() }()
	if _, err = w.Write([]byte("z")); err != nil {
		t.Fatalf(tplFatalWrite, err)
	}
	b, err = readNextByte(r)
	if err != nil {
		t.Fatalf(tplFatalRead, err)
	}
	if b != 'z' {
		t.Fatalf(tplByteMismatch, 'z', b)
	}
	// Чтение из дескриптора на котором чтение невозможно, ожидается ошибка.
	d, err = os.Open(t.TempDir())
	if err != nil {
		t.Fatalf(tplFatalOpen, err)
	}
	defer func() { _ = d.Close() }()
	if _, err = readNextByte(d); err == nil {
		t.Fatal(errReadDirNoError)
	}
}

// TestReadNextByteTimeout Истечение времени ожидания данных.
func TestReadNextByteTimeout(t *testing.T) {
	const (
		tplFatalPipe = "Pipe(): неожиданная ошибка: %v."
		tplFatalRead = "readNextByte(): ожидалась ошибка errTimeout, получено %v."
	)
	var (
		r   *os.File
		err error
	)

	r, _, err = os.Pipe()
	if err != nil {
		t.Fatalf(tplFatalPipe, err)
	}
	defer func() { _ = r.Close() }()
	if _, err = readNextByte(r); !errors.Is(err, errTimeout) {
		t.Fatalf(tplFatalRead, err)
	}
}

// TestReadNextResponse Чтение ответа терминала: OSC ответ и ответ позиции курсора.
func TestReadNextResponse(t *testing.T) {
	const (
		tplFatalPipe        = "Pipe(): неожиданная ошибка: %v."
		tplFatalWrite       = "Write(): неожиданная ошибка: %v."
		tplFatalRead        = "readNextResponse(): неожиданная ошибка: %v."
		tplResponseMismatch = "readNextResponse(): ожидалось (%q, %v), получено (%q, %v)."
	)
	// Описание одного теста чтения ответа терминала.
	type testCase struct {
		name    string
		in      string
		want    string
		wantOSC bool
	}
	var (
		cases    []testCase
		tt       testCase
		n        int
		r, w     *os.File
		response string
		isOSC    bool
		err      error
	)

	cases = []testCase{
		{"OscBel", "\x1b]11;rgb:1111/2222/3333\a", "\x1b]11;rgb:1111/2222/3333\a", true},
		{"OscEsc", "\x1b]10;rgb:aaaa/bbbb/cccc\x1b", "\x1b]10;rgb:aaaa/bbbb/cccc\x1b", true},
		{"CursorPosition", "мусор\x1b[42;1R", "\x1b[42;1R", false},
	}
	for n = range cases {
		tt = cases[n]
		t.Run(tt.name, func(t *testing.T) {
			r, w, err = os.Pipe()
			if err != nil {
				t.Fatalf(tplFatalPipe, err)
			}
			defer func() { _ = r.Close() }()
			defer func() { _ = w.Close() }()
			if _, err = w.Write([]byte(tt.in)); err != nil {
				t.Fatalf(tplFatalWrite, err)
			}
			response, isOSC, err = readNextResponse(r)
			if err != nil {
				t.Fatalf(tplFatalRead, err)
			}
			if response != tt.want || isOSC != tt.wantOSC {
				t.Fatalf(tplResponseMismatch, tt.want, tt.wantOSC, response, isOSC)
			}
		})
	}
}

// TestReadNextResponseErrors Ошибки чтения ответа терминала.
func TestReadNextResponseErrors(t *testing.T) {
	const (
		tplFatalPipe    = "Pipe(): неожиданная ошибка: %v."
		tplFatalWrite   = "Write(): неожиданная ошибка: %v."
		tplFatalStatus  = "readNextResponse(): ожидалась ошибка errStatusReport, получено %v."
		tplFatalOpen    = "Open(): неожиданная ошибка: %v."
		errBreakNoError = "readNextResponse(): при обрыве потока ожидалась ошибка."
		errDirNoError   = "readNextResponse(): при чтении каталога ожидалась ошибка."
	)
	var (
		r   *os.File
		w   *os.File
		err error
	)

	t.Run("BadType", func(t *testing.T) {
		r, w, err = os.Pipe()
		if err != nil {
			t.Fatalf(tplFatalPipe, err)
		}
		defer func() { _ = r.Close() }()
		defer func() { _ = w.Close() }()
		if _, err = w.Write([]byte("\x1bZ")); err != nil {
			t.Fatalf(tplFatalWrite, err)
		}
		if _, _, err = readNextResponse(r); !errors.Is(err, errStatusReport) {
			t.Fatalf(tplFatalStatus, err)
		}
	})
	t.Run("TooLong", func(t *testing.T) {
		r, w, err = os.Pipe()
		if err != nil {
			t.Fatalf(tplFatalPipe, err)
		}
		defer func() { _ = r.Close() }()
		defer func() { _ = w.Close() }()
		if _, err = w.Write([]byte("\x1b[" + strings.Repeat("9", 30))); err != nil {
			t.Fatalf(tplFatalWrite, err)
		}
		if _, _, err = readNextResponse(r); !errors.Is(err, errStatusReport) {
			t.Fatalf(tplFatalStatus, err)
		}
	})
	t.Run("SkipLoopError", func(t *testing.T) {
		// Первый байт не ESC, второе чтение обрывается ошибкой EOF.
		r, w, err = os.Pipe()
		if err != nil {
			t.Fatalf(tplFatalPipe, err)
		}
		defer func() { _ = r.Close() }()
		if _, err = w.Write([]byte("x")); err != nil {
			t.Fatalf(tplFatalWrite, err)
		}
		_ = w.Close()
		if _, _, err = readNextResponse(r); err == nil {
			t.Fatal(errBreakNoError)
		}
	})
	t.Run("TypeReadError", func(t *testing.T) {
		// После ESC обрыв потока перед чтением типа ответа.
		r, w, err = os.Pipe()
		if err != nil {
			t.Fatalf(tplFatalPipe, err)
		}
		defer func() { _ = r.Close() }()
		if _, err = w.Write([]byte("\x1b")); err != nil {
			t.Fatalf(tplFatalWrite, err)
		}
		_ = w.Close()
		if _, _, err = readNextResponse(r); err == nil {
			t.Fatal(errBreakNoError)
		}
	})
	t.Run("BodyReadError", func(t *testing.T) {
		// После заголовка ответа обрыв потока перед чтением тела ответа.
		r, w, err = os.Pipe()
		if err != nil {
			t.Fatalf(tplFatalPipe, err)
		}
		defer func() { _ = r.Close() }()
		if _, err = w.Write([]byte("\x1b[")); err != nil {
			t.Fatalf(tplFatalWrite, err)
		}
		_ = w.Close()
		if _, _, err = readNextResponse(r); err == nil {
			t.Fatal(errBreakNoError)
		}
	})
	t.Run("NoData", func(t *testing.T) {
		// Чтение из дескриптора на котором чтение невозможно, ожидается ошибка.
		r, err = os.Open(t.TempDir())
		if err != nil {
			t.Fatalf(tplFatalOpen, err)
		}
		defer func() { _ = r.Close() }()
		if _, _, err = readNextResponse(r); err == nil {
			t.Fatal(errDirNoError)
		}
	})
}

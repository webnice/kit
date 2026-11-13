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
	var cases = []struct {
		name      string
		term      string
		colorTerm string
		program   string
		want      terminalProfile
	}{
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
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envTerm, tt.term)
			t.Setenv(envColorTerm, tt.colorTerm)
			t.Setenv(envTermProgram, tt.program)
			if got := colorProfile(); got != tt.want {
				t.Fatalf("colorProfile(): ожидался профиль %d, получено %d", tt.want, got)
			}
		})
	}
}

// TestForegroundColor Цвет текста терминала, ветки переменной окружения COLORFGBG.
func TestForegroundColor(t *testing.T) {
	t.Setenv(envTerm, "screen") // Отключение запроса к терминалу, ответ будет получен из окружения.
	t.Run("ColorFgbg", func(t *testing.T) {
		t.Setenv(envColorFgbg, "1;2")
		if got := foregroundColor(); got != colorAnsi(1) {
			t.Fatalf("foregroundColor(): ожидался цвет %v, получено %v", colorAnsi(1), got)
		}
	})
	t.Run("ColorFgbgBadValue", func(t *testing.T) {
		t.Setenv(envColorFgbg, "default;default")
		if got := foregroundColor(); got != colorAnsi(7) {
			t.Fatalf("foregroundColor(): ожидался цвет %v, получено %v", colorAnsi(7), got)
		}
	})
	t.Run("Default", func(t *testing.T) {
		t.Setenv(envColorFgbg, "")
		if got := foregroundColor(); got != colorAnsi(7) {
			t.Fatalf("foregroundColor(): ожидался цвет %v, получено %v", colorAnsi(7), got)
		}
	})
}

// TestBackgroundColor Цвет фона терминала, ветки переменной окружения COLORFGBG.
func TestBackgroundColor(t *testing.T) {
	t.Setenv(envTerm, "screen") // Отключение запроса к терминалу, ответ будет получен из окружения.
	t.Run("ColorFgbg", func(t *testing.T) {
		t.Setenv(envColorFgbg, "1;2")
		if got := backgroundColor(); got != colorAnsi(2) {
			t.Fatalf("backgroundColor(): ожидался цвет %v, получено %v", colorAnsi(2), got)
		}
	})
	t.Run("ColorFgbgBadValue", func(t *testing.T) {
		t.Setenv(envColorFgbg, "default;default")
		if got := backgroundColor(); got != colorAnsi(0) {
			t.Fatalf("backgroundColor(): ожидался цвет %v, получено %v", colorAnsi(0), got)
		}
	})
	t.Run("Default", func(t *testing.T) {
		t.Setenv(envColorFgbg, "")
		if got := backgroundColor(); got != colorAnsi(0) {
			t.Fatalf("backgroundColor(): ожидался цвет %v, получено %v", colorAnsi(0), got)
		}
	})
}

// TestTermStatusReport Запрос отчёта статуса терминала в среде без терминала.
func TestTermStatusReport(t *testing.T) {
	t.Run("Screen", func(t *testing.T) {
		t.Setenv(envTerm, "screen-256color")
		if _, err := termStatusReport(11); !errors.Is(err, errStatusReport) {
			t.Fatalf("termStatusReport(): ожидалась ошибка errStatusReport, получено %v", err)
		}
	})
	t.Run("NoTerminal", func(t *testing.T) {
		if isTTY(os.Stdout.Fd()) {
			t.Skip("пропуск: stdout является терминалом")
		}
		t.Setenv(envTerm, "xterm-256color")
		if _, err := termStatusReport(11); !errors.Is(err, errStatusReport) {
			t.Fatalf("termStatusReport(): ожидалась ошибка errStatusReport, получено %v", err)
		}
	})
}

// TestWaitForData Ожидание данных на файловом дескрипторе.
func TestWaitForData(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
	}
	defer func() { _ = r.Close() }()
	defer func() { _ = w.Close() }()
	if err = waitForData(r.Fd(), 50*time.Millisecond); !errors.Is(err, errTimeout) {
		t.Fatalf("waitForData(): ожидалась ошибка errTimeout, получено %v", err)
	}
	if _, err = w.Write([]byte("x")); err != nil {
		t.Fatalf("Write(): неожиданная ошибка: %v", err)
	}
	if err = waitForData(r.Fd(), time.Second); err != nil {
		t.Fatalf("waitForData(): неожиданная ошибка: %v", err)
	}
}

// TestReadNextByte Чтение одного байта из файлового дескриптора.
func TestReadNextByte(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
	}
	defer func() { _ = r.Close() }()
	if _, err = w.Write([]byte("z")); err != nil {
		t.Fatalf("Write(): неожиданная ошибка: %v", err)
	}
	b, err := readNextByte(r)
	if err != nil {
		t.Fatalf("readNextByte(): неожиданная ошибка: %v", err)
	}
	if b != 'z' {
		t.Fatalf("readNextByte(): ожидался байт %q, получено %q", 'z', b)
	}
	// Чтение из дескриптора на котором чтение невозможно, ожидается ошибка.
	d, err := os.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open(): неожиданная ошибка: %v", err)
	}
	defer func() { _ = d.Close() }()
	if _, err = readNextByte(d); err == nil {
		t.Fatal("readNextByte(): при чтении каталога ожидалась ошибка")
	}
}

// TestReadNextByteTimeout Истечение времени ожидания данных.
func TestReadNextByteTimeout(t *testing.T) {
	r, _, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
	}
	defer func() { _ = r.Close() }()
	if _, err = readNextByte(r); !errors.Is(err, errTimeout) {
		t.Fatalf("readNextByte(): ожидалась ошибка errTimeout, получено %v", err)
	}
}

// TestReadNextResponse Чтение ответа терминала: OSC ответ и ответ позиции курсора.
func TestReadNextResponse(t *testing.T) {
	var cases = []struct {
		name    string
		in      string
		want    string
		wantOSC bool
	}{
		{"OscBel", "\x1b]11;rgb:1111/2222/3333\a", "\x1b]11;rgb:1111/2222/3333\a", true},
		{"OscEsc", "\x1b]10;rgb:aaaa/bbbb/cccc\x1b", "\x1b]10;rgb:aaaa/bbbb/cccc\x1b", true},
		{"CursorPosition", "мусор\x1b[42;1R", "\x1b[42;1R", false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
			}
			defer func() { _ = r.Close() }()
			defer func() { _ = w.Close() }()
			if _, err = w.Write([]byte(tt.in)); err != nil {
				t.Fatalf("Write(): неожиданная ошибка: %v", err)
			}
			response, isOSC, err := readNextResponse(r)
			if err != nil {
				t.Fatalf("readNextResponse(): неожиданная ошибка: %v", err)
			}
			if response != tt.want || isOSC != tt.wantOSC {
				t.Fatalf("readNextResponse(): ожидалось (%q, %v), получено (%q, %v)", tt.want, tt.wantOSC, response, isOSC)
			}
		})
	}
}

// TestReadNextResponseErrors Ошибки чтения ответа терминала.
func TestReadNextResponseErrors(t *testing.T) {
	t.Run("BadType", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
		}
		defer func() { _ = r.Close() }()
		defer func() { _ = w.Close() }()
		if _, err = w.Write([]byte("\x1bZ")); err != nil {
			t.Fatalf("Write(): неожиданная ошибка: %v", err)
		}
		if _, _, err = readNextResponse(r); !errors.Is(err, errStatusReport) {
			t.Fatalf("readNextResponse(): ожидалась ошибка errStatusReport, получено %v", err)
		}
	})
	t.Run("TooLong", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
		}
		defer func() { _ = r.Close() }()
		defer func() { _ = w.Close() }()
		if _, err = w.Write([]byte("\x1b[" + strings.Repeat("9", 30))); err != nil {
			t.Fatalf("Write(): неожиданная ошибка: %v", err)
		}
		if _, _, err = readNextResponse(r); !errors.Is(err, errStatusReport) {
			t.Fatalf("readNextResponse(): ожидалась ошибка errStatusReport, получено %v", err)
		}
	})
	t.Run("SkipLoopError", func(t *testing.T) {
		// Первый байт не ESC, второе чтение обрывается ошибкой EOF.
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
		}
		defer func() { _ = r.Close() }()
		if _, err = w.Write([]byte("x")); err != nil {
			t.Fatalf("Write(): неожиданная ошибка: %v", err)
		}
		_ = w.Close()
		if _, _, err = readNextResponse(r); err == nil {
			t.Fatal("readNextResponse(): при обрыве потока ожидалась ошибка")
		}
	})
	t.Run("TypeReadError", func(t *testing.T) {
		// После ESC обрыв потока перед чтением типа ответа.
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
		}
		defer func() { _ = r.Close() }()
		if _, err = w.Write([]byte("\x1b")); err != nil {
			t.Fatalf("Write(): неожиданная ошибка: %v", err)
		}
		_ = w.Close()
		if _, _, err = readNextResponse(r); err == nil {
			t.Fatal("readNextResponse(): при обрыве потока ожидалась ошибка")
		}
	})
	t.Run("BodyReadError", func(t *testing.T) {
		// После заголовка ответа обрыв потока перед чтением тела ответа.
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
		}
		defer func() { _ = r.Close() }()
		if _, err = w.Write([]byte("\x1b[")); err != nil {
			t.Fatalf("Write(): неожиданная ошибка: %v", err)
		}
		_ = w.Close()
		if _, _, err = readNextResponse(r); err == nil {
			t.Fatal("readNextResponse(): при обрыве потока ожидалась ошибка")
		}
	})
	t.Run("NoData", func(t *testing.T) {
		// Чтение из дескриптора на котором чтение невозможно, ожидается ошибка.
		r, err := os.Open(t.TempDir())
		if err != nil {
			t.Fatalf("Open(): неожиданная ошибка: %v", err)
		}
		defer func() { _ = r.Close() }()
		if _, _, err = readNextResponse(r); err == nil {
			t.Fatal("readNextResponse(): при чтении каталога ожидалась ошибка")
		}
	})
}

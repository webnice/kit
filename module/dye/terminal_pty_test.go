//go:build darwin || linux
// +build darwin linux

package dye

import (
	"os"
	"testing"
)

// TestIsTTYPty Проверка TTY на псевдотерминале.
func TestIsTTYPty(t *testing.T) {
	master, slave := openTestPty(t)
	defer func() { _ = master.Close() }()
	defer func() { _ = slave.Close() }()

	t.Setenv("CI", "")
	if !isTTY(slave.Fd()) {
		t.Fatal("isTTY(): для псевдотерминала ожидалось true")
	}
	if !isTTY(master.Fd()) {
		t.Fatal("isTTY(): для master псевдотерминала ожидалось true")
	}
	t.Setenv("CI", "true")
	if isTTY(slave.Fd()) {
		t.Fatal("isTTY(): в режиме CI ожидалось false")
	}
}

// TestGetColorProfileFd Определение профиля терминала по дескриптору файла.
func TestGetColorProfileFd(t *testing.T) {
	master, slave := openTestPty(t)
	defer func() { _ = master.Close() }()
	defer func() { _ = slave.Close() }()

	r, _, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
	}
	defer func() { _ = r.Close() }()
	if got := getColorProfileFd(r.Fd()); got != terminalAscii {
		t.Fatalf("getColorProfileFd(): для пайпа ожидался профиль %d, получено %d", terminalAscii, got)
	}

	t.Setenv("CI", "")
	t.Setenv(envColorTerm, "truecolor")
	t.Setenv(envTerm, "xterm-256color")
	if got := getColorProfileFd(slave.Fd()); got != terminalTrueColor {
		t.Fatalf("getColorProfileFd(): ожидался профиль %d, получено %d", terminalTrueColor, got)
	}

	t.Setenv("CI", "true")
	if got := getColorProfileFd(slave.Fd()); got != terminalAscii {
		t.Fatalf("getColorProfileFd(): в режиме CI ожидался профиль %d, получено %d", terminalAscii, got)
	}
}

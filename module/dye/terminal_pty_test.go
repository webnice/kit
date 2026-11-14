//go:build darwin || linux
// +build darwin linux

package dye

import (
	"os"
	"testing"
)

// TestIsTTYPty Проверка TTY на псевдотерминале.
func TestIsTTYPty(t *testing.T) {
	const (
		errSlaveNotTrue  = "isTTY(): для псевдотерминала ожидалось true."
		errMasterNotTrue = "isTTY(): для master псевдотерминала ожидалось true."
		errCiNotFalse    = "isTTY(): в режиме CI ожидалось false."
	)
	var master, slave *os.File

	master, slave = openTestPty(t)
	defer func() { _ = master.Close() }()
	defer func() { _ = slave.Close() }()
	t.Setenv(envCi, "")
	if !isTTY(slave.Fd()) {
		t.Fatal(errSlaveNotTrue)
	}
	if !isTTY(master.Fd()) {
		t.Fatal(errMasterNotTrue)
	}
	t.Setenv(envCi, "true")
	if isTTY(slave.Fd()) {
		t.Fatal(errCiNotFalse)
	}
}

// TestGetColorProfileFd Определение профиля терминала по дескриптору файла.
func TestGetColorProfileFd(t *testing.T) {
	const (
		tplFatalPipe       = "Pipe(): неожиданная ошибка: %v."
		tplProfileMismatch = "getColorProfileFd(): для пайпа ожидался профиль %d, получено %d."
		tplProfileExpected = "getColorProfileFd(): ожидался профиль %d, получено %d."
		tplProfileCi       = "getColorProfileFd(): в режиме CI ожидался профиль %d, получено %d."
	)
	var (
		master, slave *os.File
		r             *os.File
		err           error
	)

	master, slave = openTestPty(t)
	defer func() { _ = master.Close() }()
	defer func() { _ = slave.Close() }()
	r, _, err = os.Pipe()
	if err != nil {
		t.Fatalf(tplFatalPipe, err)
	}
	defer func() { _ = r.Close() }()
	if got := getColorProfileFd(r.Fd()); got != terminalAscii {
		t.Fatalf(tplProfileMismatch, terminalAscii, got)
	}
	t.Setenv(envCi, "")
	t.Setenv(envColorTerm, "truecolor")
	t.Setenv(envTerm, "xterm-256color")
	if got := getColorProfileFd(slave.Fd()); got != terminalTrueColor {
		t.Fatalf(tplProfileExpected, terminalTrueColor, got)
	}
	t.Setenv(envCi, "true")
	if got := getColorProfileFd(slave.Fd()); got != terminalAscii {
		t.Fatalf(tplProfileCi, terminalAscii, got)
	}
}

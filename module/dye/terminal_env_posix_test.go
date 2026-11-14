//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd
// +build darwin dragonfly freebsd linux netbsd openbsd

package dye

import (
	"os"
	"testing"
)

// TestIsForegroundNotTerminal Проверка ошибки ioctl для дескриптора не являющегося терминалом.
func TestIsForegroundNotTerminal(t *testing.T) {
	const (
		tplFatalPipe = "Pipe(): неожиданная ошибка: %v."
		errPipeTrue  = "isForeground(): для пайпа ожидалось false."
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
	if isForeground(int(r.Fd())) {
		t.Fatal(errPipeTrue)
	}
}

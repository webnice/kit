//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd
// +build darwin dragonfly freebsd linux netbsd openbsd

package dye

import (
	"os"
	"testing"
)

// TestIsForegroundNotTerminal Проверка ошибки ioctl для дескриптора не являющегося терминалом.
func TestIsForegroundNotTerminal(t *testing.T) {
	r, _, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe(): неожиданная ошибка: %v", err)
	}
	defer func() { _ = r.Close() }()
	if isForeground(int(r.Fd())) {
		t.Fatal("isForeground(): для пайпа ожидалось false")
	}
}

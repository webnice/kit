//go:build linux
// +build linux

package dye

import (
	"fmt"
	"testing"

	"golang.org/x/sys/unix"
)

// ptySlavePath Получение пути slave части псевдотерминала.
func ptySlavePath(t *testing.T, masterFd int) (path string) {
	t.Helper()
	n, err := unix.IoctlGetInt(masterFd, unix.TIOCGPTN)
	if err != nil {
		t.Skipf("пропуск: псевдотерминал недоступен: %v", err)
	}

	return fmt.Sprintf("/dev/pts/%d", n)
}

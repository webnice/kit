//go:build darwin
// +build darwin

package dye

import (
	"bytes"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

// ptySlavePath Получение пути slave части псевдотерминала.
func ptySlavePath(t *testing.T, masterFd int) (path string) {
	t.Helper()
	var name [128]byte
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(masterFd), uintptr(unix.TIOCPTYGNAME), uintptr(unsafe.Pointer(&name[0])))
	if errno != 0 {
		t.Skipf("пропуск: псевдотерминал недоступен: %v", errno)
	}

	return string(name[:bytes.IndexByte(name[:], 0)])
}

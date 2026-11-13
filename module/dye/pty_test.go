//go:build darwin || linux
// +build darwin linux

package dye

import (
	"os"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

// openTestPty Создание пары псевдотерминалов (master, slave) для тестирования.
func openTestPty(t *testing.T) (master, slave *os.File) {
	t.Helper()
	mfd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("пропуск: псевдотерминал недоступен: %v", err)
	}
	slavePath := ptySlavePath(t, mfd)
	sfd, err := unix.Open(slavePath, unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		_ = unix.Close(mfd)
		t.Skipf("пропуск: псевдотерминал недоступен: %v", err)
	}

	return os.NewFile(uintptr(mfd), "ptmx"), os.NewFile(uintptr(sfd), slavePath)
}

// ptySetForegroundPgrp Установка группы процессов активного терминала псевдотерминала.
func ptySetForegroundPgrp(t *testing.T, master *os.File) {
	t.Helper()
	pgrp := unix.Getpgrp()
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, master.Fd(), uintptr(unix.TIOCSPGRP), uintptr(unsafe.Pointer(&pgrp)))
	if errno != 0 {
		t.Skipf("пропуск: невозможно установить группу процессов терминала: %v", errno)
	}
}

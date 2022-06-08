//go:build windows
// +build windows

// Package lock
package lock

import (
	"syscall"
	"unsafe"
)

var (
	kernel32, _         = syscall.LoadLibrary("kernel32.dll")
	procLockFileEx, _   = syscall.GetProcAddress(kernel32, "LockFileEx")
	procUnlockFileEx, _ = syscall.GetProcAddress(kernel32, "UnlockFileEx")
)

const (
	winLockfileFailImmediately = 0x00000001
	winLockfileExclusiveLock   = 0x00000002
	winLockfileSharedLock      = 0x00000000
)

// Документация где-то там :): https://msdn.microsoft.com/en-us/library/windows/desktop/aa365203(v=vs.85).aspx
func lockFileEx(handle syscall.Handle, flags uint32, reserved uint32, numberOfBytesToLockLow uint32, numberOfBytesToLockHigh uint32, offset *syscall.Overlapped) (bool, syscall.Errno) {
	r1, _, errNo := syscall.Syscall6(
		uintptr(procLockFileEx),
		6,
		uintptr(handle),
		uintptr(flags),
		uintptr(reserved),
		uintptr(numberOfBytesToLockLow),
		uintptr(numberOfBytesToLockHigh),
		uintptr(unsafe.Pointer(offset)))
	if r1 != 1 {
		if errNo == 0 {
			return false, syscall.EINVAL
		}

		return false, errNo
	}

	return true, 0
}

func unlockFileEx(handle syscall.Handle, reserved uint32, numberOfBytesToLockLow uint32, numberOfBytesToLockHigh uint32, offset *syscall.Overlapped) (bool, syscall.Errno) {
	r1, _, errNo := syscall.Syscall6(
		uintptr(procUnlockFileEx),
		5,
		uintptr(handle),
		uintptr(reserved),
		uintptr(numberOfBytesToLockLow),
		uintptr(numberOfBytesToLockHigh),
		uintptr(unsafe.Pointer(offset)),
		0)
	if r1 != 1 {
		if errNo == 0 {
			return false, syscall.EINVAL
		}

		return false, errNo
	}

	return true, 0
}

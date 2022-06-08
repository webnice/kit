// Package lock
package lock

import "syscall"

const (
	ErrorLockViolation syscall.Errno = 0x21 // dec: 33
)

// Lock Эксклюзивная блокировка файла.
func (lko *Lock) Lock() error { return lko.lock(&lko.isLocked, winLockfileExclusiveLock) }

// RLock Блокировка файла с доступом на чтение.
func (lko *Lock) RLock() error { return lko.lock(&lko.isReadLocked, winLockfileSharedLock) }

func (lko *Lock) lock(locked *bool, flag uint32) (err error) {
	var errNo syscall.Errno

	lko.mux.Lock()
	defer lko.mux.Unlock()
	if *locked {
		return
	}
	if lko.fh == nil {
		if err = lko.setFh(); err != nil {
			return
		}
		defer func() { _ = lko.ensureFhState() }()
	}
	if _, errNo = lockFileEx(syscall.Handle(lko.fh.Fd()), flag, 0, 1, 0, &syscall.Overlapped{}); errNo > 0 {
		err = errNo
		return
	}
	*locked, err = true, nil

	return
}

// Unlock Снятие блокировки.
func (lko *Lock) Unlock() (err error) {
	var errNo syscall.Errno

	lko.mux.Lock()
	defer lko.mux.Unlock()
	if (!lko.isLocked && !lko.isReadLocked) || lko.fh == nil {
		return
	}
	if _, errNo = unlockFileEx(syscall.Handle(lko.fh.Fd()), 0, 1, 0, &syscall.Overlapped{}); errNo > 0 {
		err = errNo
		return
	}
	_ = lko.fh.Close()
	lko.isLocked, lko.isReadLocked, lko.fh = false, false, nil

	return
}

func (lko *Lock) TryLock() (bool, error) { return lko.try(&lko.isLocked, winLockfileExclusiveLock) }

func (lko *Lock) TryRLock() (bool, error) { return lko.try(&lko.isReadLocked, winLockfileSharedLock) }

func (lko *Lock) try(locked *bool, flag uint32) (done bool, err error) {
	var errNo syscall.Errno

	lko.mux.Lock()
	defer lko.mux.Unlock()
	if *locked {
		done = true
		return
	}
	if lko.fh == nil {
		if err = lko.setFh(); err != nil {
			return
		}
		defer func() { _ = lko.ensureFhState() }()
	}

	_, errNo = lockFileEx(syscall.Handle(lko.fh.Fd()), flag|winLockfileFailImmediately, 0, 1, 0, &syscall.Overlapped{})
	if errNo > 0 {
		if errNo == ErrorLockViolation || errNo == syscall.ERROR_IO_PENDING {
			return
		}
		err = errNo
		return
	}
	*locked, done = true, true

	return
}

//go:build !aix && !windows
// +build !aix,!windows

// Package lock
package lock

import (
	"os"
	"syscall"
)

// Lock Эксклюзивная блокировка файла.
func (lko *Lock) Lock() error { return lko.lock(&lko.isLocked, syscall.LOCK_EX) }

// RLock Блокировка файла с доступом на чтение.
func (lko *Lock) RLock() error { return lko.lock(&lko.isReadLocked, syscall.LOCK_SH) }

func (lko *Lock) lock(locked *bool, flag int) (err error) {
	var (
		shouldRetry bool
		reopenErr   error
	)

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
	if err = syscall.Flock(int(lko.fh.Fd()), flag); err != nil {
		if shouldRetry, reopenErr = lko.reopenFDOnError(err); reopenErr != nil {
			err = reopenErr
			return
		}
		if !shouldRetry {
			return
		}
		if err = syscall.Flock(int(lko.fh.Fd()), flag); err != nil {
			return
		}
	}
	*locked = true

	return
}

// Unlock Снятие блокировки.
func (lko *Lock) Unlock() (err error) {
	lko.mux.Lock()
	defer lko.mux.Unlock()
	if (!lko.isLocked && !lko.isReadLocked) || lko.fh == nil {
		return
	}
	if err = syscall.Flock(int(lko.fh.Fd()), syscall.LOCK_UN); err != nil {
		return
	}
	_ = lko.fh.Close()
	lko.isLocked, lko.isReadLocked, lko.fh = false, false, nil

	return
}

func (lko *Lock) TryLock() (bool, error) { return lko.try(&lko.isLocked, syscall.LOCK_EX) }

func (lko *Lock) TryRLock() (bool, error) { return lko.try(&lko.isReadLocked, syscall.LOCK_SH) }

func (lko *Lock) try(locked *bool, flag int) (done bool, err error) {
	var (
		retried     bool
		shouldRetry bool
		reopenErr   error
	)

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
retryTry:
	switch err = syscall.Flock(int(lko.fh.Fd()), flag|syscall.LOCK_NB); err {
	case syscall.EWOULDBLOCK:
		err = nil
		return
	case nil:
		*locked, done, err = true, true, nil
		return
	}
	if !retried {
		if shouldRetry, reopenErr = lko.reopenFDOnError(err); reopenErr != nil {
			err = reopenErr
			return
		} else if shouldRetry {
			retried = true
			goto retryTry
		}
	}

	return
}

func (lko *Lock) reopenFDOnError(e error) (done bool, err error) {
	var (
		fi os.FileInfo
		fh *os.File
	)

	if e != syscall.EIO && e != syscall.EBADF {
		return
	}
	if fi, err = lko.fh.Stat(); err == nil {
		if fi.Mode()&0600 == 0600 {
			_ = lko.fh.Close()
			lko.fh = nil
			if fh, err = os.OpenFile(lko.path, os.O_CREATE|os.O_RDWR, os.FileMode(0600)); err != nil {
				return
			}
			lko.fh, done = fh, true
			return
		}
	}

	return
}

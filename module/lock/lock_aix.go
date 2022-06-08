//go:build aix
// +build aix

// Package lock
package lock

import (
	"errors"
	"io"
	"os"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

type lockType int16

const (
	readLock  lockType = unix.F_RDLCK
	writeLock lockType = unix.F_WRLCK
)

type cmdType int

const (
	tryLock  cmdType = unix.F_SETLK
	waitLock cmdType = unix.F_SETLKW
)

type inode = uint64

type inodeLock struct {
	owner *Lock
	queue []<-chan *Lock
}

var (
	mu     sync.Mutex
	inodes = map[*Lock]inode{}
	locks  = map[inode]inodeLock{}
)

// Lock Эксклюзивная блокировка файла.
func (lko *Lock) Lock() error { return lko.lock(&lko.isLocked, writeLock) }

// RLock Блокировка файла с доступом на чтение.
func (lko *Lock) RLock() error { return lko.lock(&lko.isReadLocked, readLock) }

func (lko *Lock) lock(locked *bool, flag lockType) (err error) {
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
	if _, err = lko.doLock(waitLock, flag, true); err != nil {
		return
	}
	*locked = true

	return
}

func (lko *Lock) doLock(cmd cmdType, lt lockType, blocking bool) (done bool, err error) {
	var (
		fi   os.FileInfo
		ino  inode
		i    inode
		dup  bool
		wait chan *Lock
		il   inodeLock
	)

	if fi, err = lko.fh.Stat(); err != nil {
		return
	}
	ino = inode(fi.Sys().(*syscall.Stat_t).Ino)
	mu.Lock()
	if i, dup = inodes[lko]; dup && i != ino {
		mu.Unlock()
		err = &os.PathError{
			Path: lko.Path(),
			Err:  errors.New("inode for file changed since last Lock or RLock"),
		}
		return
	}
	inodes[lko] = ino
	if il = locks[ino]; il.owner == lko {
		// Этот файл уже заблокирован.
	} else if il.owner == nil {
		il.owner = lko
	} else if !blocking {
		mu.Unlock()
		return
	} else {
		wait = make(chan *Lock)
		il.queue = append(il.queue, wait)
	}
	locks[ino] = il
	mu.Unlock()
	if wait != nil {
		wait <- lko
	}
	if err = setlkw(lko.fh.Fd(), cmd, lt); err != nil {
		_ = lko.doUnlock()
		if cmd == tryLock && err == unix.EACCES {
			err = nil
			return
		}
		return
	}
	done = true

	return
}

// Unlock Снятие блокировки.
func (lko *Lock) Unlock() (err error) {
	lko.mux.Lock()
	defer lko.mux.Unlock()
	if (!lko.isLocked && !lko.isReadLocked) || lko.fh == nil {
		return
	}
	if err = lko.doUnlock(); err != nil {
		return
	}
	_ = lko.fh.Close()
	lko.isLocked, lko.isReadLocked, lko.fh = false, false, nil

	return
}

func (lko *Lock) doUnlock() (err error) {
	var (
		owner *Lock
		ino   inode
		ok    bool
		il    inodeLock
	)

	mu.Lock()
	if ino, ok = inodes[lko]; ok {
		owner = locks[ino].owner
	}
	mu.Unlock()
	if owner == lko {
		err = setlkw(lko.fh.Fd(), waitLock, unix.F_UNLCK)
	}
	mu.Lock()
	il = locks[ino]
	if len(il.queue) == 0 {
		delete(locks, ino)
	} else {
		il.owner = <-il.queue[0]
		il.queue = il.queue[1:]
		locks[ino] = il
	}
	delete(inodes, lko)
	mu.Unlock()

	return
}

func (lko *Lock) TryLock() (bool, error) { return lko.try(&lko.isLocked, writeLock) }

func (lko *Lock) TryRLock() (bool, error) { return lko.try(&lko.isReadLocked, readLock) }

func (lko *Lock) try(locked *bool, flag lockType) (done bool, err error) {
	var hasLock bool

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
	if hasLock, err = lko.doLock(tryLock, flag, false); err != nil {
		return
	}
	*locked, done = hasLock, hasLock

	return
}

func setlkw(fd uintptr, cmd cmdType, lt lockType) (err error) {
	for {
		if err = unix.FcntlFlock(fd, int(cmd), &unix.Flock_t{
			Type:   int16(lt),
			Whence: io.SeekStart,
			Start:  0,
			Len:    0, // Все байты
		}); err != unix.EINTR {
			return
		}
	}
}

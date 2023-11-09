package lock

import (
	"context"
	"os"
	"sync"
	"time"
)

// Lock Описание структуры блокировки файлов.
type Lock struct {
	path         string       // Путь и имя файла.
	mux          sync.RWMutex // rw mutex.
	fh           *os.File     // Файловый дескриптор.
	isLocked     bool         // Состояние блокировки файла.
	isReadLocked bool         // Состояние блокировки файла с доступом только для чтения.
}

// New Конструктор объекта Lock.
func New(path string) (lko *Lock) {
	lko = &Lock{
		path: path,
	}

	return
}

// Close Снятие блокировки, освобождение файлового дескриптора без удаления файла.
func (lko *Lock) Close() error {
	return lko.Unlock()
}

// Path Возвращает путь.
func (lko *Lock) Path() string {
	return lko.path
}

// IsLocked Возвращает состояние блокировки файла.
func (lko *Lock) IsLocked() bool {
	lko.mux.RLock()
	defer lko.mux.RUnlock()
	return lko.isLocked
}

// IsReadLocked Возвращает состояние блокировки файла с доступом только для чтения.
func (lko *Lock) IsReadLocked() bool {
	lko.mux.RLock()
	defer lko.mux.RUnlock()
	return lko.isReadLocked
}

// String Реализация интерфейса Stringer.
func (lko *Lock) String() string { return lko.path }

// TryLockContext Попытка блокировки файла через указанные промежутки времени с прерыванием через контекст.
func (lko *Lock) TryLockContext(ctx context.Context, retryDelay time.Duration) (bool, error) {
	return tryCtx(ctx, lko.TryLock, retryDelay)
}

// TryRLockContext Попытка блокировки файла через указанные промежутки времени с прерыванием через контекст.
func (lko *Lock) TryRLockContext(ctx context.Context, retryDelay time.Duration) (bool, error) {
	return tryCtx(ctx, lko.TryRLock, retryDelay)
}

// Попытка блокировки файла через указанные промежутки времени с прерыванием через контекст.
func tryCtx(ctx context.Context, fn func() (bool, error), retryDelay time.Duration) (done bool, err error) {
	var isEnd bool

	if err = ctx.Err(); err != nil {
		return
	}
	for {
		if isEnd {
			break
		}
		if done, err = fn(); done || err != nil {
			isEnd = true
			continue
		}
		select {
		case <-ctx.Done():
			done, err, isEnd = false, ctx.Err(), true
		case <-time.After(retryDelay):
		}
	}

	return
}

func (lko *Lock) setFh() (err error) {
	const defaultFileMode = 0644
	var (
		flags int
		fh    *os.File
	)

	flags = os.O_CREATE | os.O_RDONLY
	if fh, err = os.OpenFile(lko.path, flags, os.FileMode(defaultFileMode)); err != nil {
		return
	}
	lko.fh = fh

	return
}

// Проверка того что файловый дескриптор закрыт при отсутствии блокировки файла.
func (lko *Lock) ensureFhState() (err error) {
	if !lko.isLocked && !lko.isReadLocked && lko.fh != nil {
		err = lko.fh.Close()
		lko.fh = nil
	}

	return
}

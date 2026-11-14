package lock

import (
	"os"
	"testing"
)

func Test(t *testing.T) {
	const (
		tplFatalCreateTemp = "CreateTemp(): неожиданная ошибка: %v."
		tplTryLockWrong    = "функция TryLock() вернулa %t или прервана ошибкой: %s."
		errFhNotNil        = "дескриптор файла должен был быть освобожден и иметь значение nil."
	)
	var (
		err       error
		tmpFileFh *os.File
		tmpFile   string
		lock      *Lock
		locked    bool
		newLock   *Lock
	)

	if tmpFileFh, err = os.CreateTemp(os.TempDir(), "lock-"); err != nil {
		t.Fatalf(tplFatalCreateTemp, err)

		return
	}
	_ = tmpFileFh.Close()
	tmpFile = tmpFileFh.Name()
	_ = os.Remove(tmpFile)
	lock = New(tmpFile)
	locked, err = lock.TryLock()
	if locked == false || err != nil {
		t.Fatalf(tplTryLockWrong, locked, err)

		return
	}
	newLock = New(tmpFile)
	locked, err = newLock.TryLock()
	if locked != false || err != nil {
		t.Fatalf(tplTryLockWrong, locked, err)

		return
	}
	if newLock.fh != nil {
		t.Fatal(errFhNotNil)

		return
	}
}

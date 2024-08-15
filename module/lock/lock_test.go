package lock

import (
	"os"
	"testing"
)

func Test(t *testing.T) {
	var (
		err       error
		tmpFileFh *os.File
		tmpFile   string
		lock      *Lock
		locked    bool
		newLock   *Lock
	)

	if tmpFileFh, err = os.CreateTemp(os.TempDir(), "lock-"); err != nil {
		t.Fatal(err.Error())
		return
	}
	_ = tmpFileFh.Close()
	tmpFile = tmpFileFh.Name()
	_ = os.Remove(tmpFile)
	lock = New(tmpFile)
	locked, err = lock.TryLock()
	if locked == false || err != nil {
		t.Fatalf("функция TryLock() вернула %t или прервана ошибкой: %s", locked, err)
		return
	}
	newLock = New(tmpFile)
	locked, err = newLock.TryLock()
	if locked != false || err != nil {
		t.Fatalf("функция TryLock() вернула %t или прервана ошибкой: %s", locked, err)
		return
	}
	if newLock.fh != nil {
		t.Fatal("дескриптор файла должен был быть освобожден и иметь значение nil")
		return
	}
}

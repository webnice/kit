// Package lock
package lock

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test(t *testing.T) {
	tmpFileFh, err := ioutil.TempFile(os.TempDir(), "lock-")
	_ = tmpFileFh.Close()
	tmpFile := tmpFileFh.Name()
	_ = os.Remove(tmpFile)

	lock := New(tmpFile)
	locked, err := lock.TryLock()
	if locked == false || err != nil {
		t.Fatalf("failed to lock file: locked: %t, err: %v", locked, err)
	}

	newLock := New(tmpFile)
	locked, err = newLock.TryLock()
	if locked != false || err != nil {
		t.Fatalf("should have failed locking file: locked: %t, err: %v", locked, err)
	}

	if newLock.fh != nil {
		t.Fatal("file handle should have been released and be nil")
	}
}

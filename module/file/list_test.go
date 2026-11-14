package file

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

// TestRecursiveFileList Рекурсивный список файлов.
func TestRecursiveFileList(t *testing.T) {
	const (
		tplFatalUnexpected = "RecursiveFileList(): неожиданная ошибка: %v."
		tplCountMismatch   = "RecursiveFileList(): ожидалось %d файлов, получено %d: %v."
		tplItemMismatch    = "RecursiveFileList(): ожидалось %q, получено %q."
		tplOneFileExpected = "RecursiveFileList(): ожидался 1 файл, получено %d."
		errAbsentNoError   = "RecursiveFileList(): для несуществующей директории ожидалась ошибка."
		tplFatalMkdirAll   = "MkdirAll(): неожиданная ошибка: %v."
		tplFatalChmod      = "Chmod(): неожиданная ошибка: %v."
		errLockedNoError   = "RecursiveFileList(): для недоступной поддиректории ожидалась ошибка."
	)
	var (
		dir    string
		err    error
		ret    []string
		want   []string
		locked string
		n      int
	)

	dir = t.TempDir()
	createTestFile(t, filepath.Join(dir, "sub1", "sub2"), "a.txt", testContent)
	createTestFile(t, filepath.Join(dir, "sub1"), "b.txt", testContent)
	createTestFile(t, dir, "c.txt", testContent)
	if ret, err = New().RecursiveFileList(dir); err != nil {
		t.Fatalf(tplFatalUnexpected, err)
	}
	sort.Strings(ret)
	want = []string{
		filepath.Join("c.txt"),
		filepath.Join("sub1", "b.txt"),
		filepath.Join("sub1", "sub2", "a.txt"),
	}
	if len(ret) != len(want) {
		t.Fatalf(tplCountMismatch, len(want), len(ret), ret)
	}
	for n = range want {
		if ret[n] != want[n] {
			t.Errorf(tplItemMismatch, want[n], ret[n])
		}
	}
	// Пустая директория.
	if ret, err = New().RecursiveFileList(filepath.Join(dir, "sub1", "sub2")); err != nil {
		t.Fatalf(tplFatalUnexpected, err)
	}
	if len(ret) != 1 {
		t.Errorf(tplOneFileExpected, len(ret))
	}
	// Несуществующая директория.
	if _, err = New().RecursiveFileList(filepath.Join(dir, "absent")); err == nil {
		t.Error(errAbsentNoError)
	}
	// Недоступная поддиректория, ошибка чтения внутри рекурсии.
	if runtime.GOOS != "windows" && os.Geteuid() != 0 {
		locked = filepath.Join(dir, "locked")
		if err = os.MkdirAll(locked, 0755); err != nil {
			t.Fatalf(tplFatalMkdirAll, err)
		}
		if err = os.Chmod(locked, 0000); err != nil {
			t.Fatalf(tplFatalChmod, err)
		}
		defer func() { _ = os.Chmod(locked, 0755) }()
		if _, err = New().RecursiveFileList(dir); err == nil {
			t.Error(errLockedNoError)
		}
	}
}

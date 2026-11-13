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
	var (
		dir = t.TempDir()
		err error
		ret []string
	)

	createTestFile(t, filepath.Join(dir, "sub1", "sub2"), "a.txt", testContent)
	createTestFile(t, filepath.Join(dir, "sub1"), "b.txt", testContent)
	createTestFile(t, dir, "c.txt", testContent)

	if ret, err = New().RecursiveFileList(dir); err != nil {
		t.Fatalf("RecursiveFileList(): неожиданная ошибка: %v", err)
	}
	sort.Strings(ret)
	var want = []string{
		filepath.Join("c.txt"),
		filepath.Join("sub1", "b.txt"),
		filepath.Join("sub1", "sub2", "a.txt"),
	}
	if len(ret) != len(want) {
		t.Fatalf("RecursiveFileList(): ожидалось %d файлов, получено %d: %v", len(want), len(ret), ret)
	}
	for i := range want {
		if ret[i] != want[i] {
			t.Errorf("RecursiveFileList(): ожидалось %q, получено %q", want[i], ret[i])
		}
	}

	// Пустая директория.
	if ret, err = New().RecursiveFileList(filepath.Join(dir, "sub1", "sub2")); err != nil {
		t.Fatalf("RecursiveFileList(): неожиданная ошибка: %v", err)
	}
	if len(ret) != 1 {
		t.Errorf("RecursiveFileList(): ожидался 1 файл, получено %d", len(ret))
	}
	// Несуществующая директория.
	if _, err = New().RecursiveFileList(filepath.Join(dir, "absent")); err == nil {
		t.Error("RecursiveFileList(): для несуществующей директории ожидалась ошибка")
	}
	// Недоступная поддиректория, ошибка чтения внутри рекурсии.
	if runtime.GOOS != "windows" && os.Geteuid() != 0 {
		locked := filepath.Join(dir, "locked")
		if err = os.MkdirAll(locked, 0755); err != nil {
			t.Fatalf("MkdirAll(): неожиданная ошибка: %v", err)
		}
		if err = os.Chmod(locked, 0000); err != nil {
			t.Fatalf("Chmod(): неожиданная ошибка: %v", err)
		}
		defer func() { _ = os.Chmod(locked, 0755) }()
		if _, err = New().RecursiveFileList(dir); err == nil {
			t.Error("RecursiveFileList(): для недоступной поддиректории ожидалась ошибка")
		}
	}
}

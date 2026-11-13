package file

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// testContent Тестовое содержимое файла.
var testContent = []byte("содержимое тестового файла пакета file")

// createTestFile Создание тестового файла с содержимым.
func createTestFile(t *testing.T, dir, name string, content []byte) (filename string) {
	t.Helper()
	filename = filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		t.Fatalf("MkdirAll(): неожиданная ошибка: %v", err)
	}
	if err := os.WriteFile(filename, content, 0644); err != nil {
		t.Fatalf("WriteFile(): неожиданная ошибка: %v", err)
	}

	return
}

// TestCleanEmptyFolder Удаление пустых директорий.
func TestCleanEmptyFolder(t *testing.T) {
	var (
		dir = t.TempDir()
		err error
	)

	// Дерево: dir/{empty1, sub/{empty2, full}}, dir/file.txt
	createTestFile(t, filepath.Join(dir, "sub", "full"), "f.txt", testContent)
	if err = os.MkdirAll(filepath.Join(dir, "empty1"), 0755); err != nil {
		t.Fatalf("MkdirAll(): неожиданная ошибка: %v", err)
	}
	if err = os.MkdirAll(filepath.Join(dir, "sub", "empty2"), 0755); err != nil {
		t.Fatalf("MkdirAll(): неожиданная ошибка: %v", err)
	}

	if err = New().CleanEmptyFolder(dir); err != nil {
		t.Fatalf("CleanEmptyFolder(): неожиданная ошибка: %v", err)
	}
	// Пустые директории удалены.
	for _, p := range []string{"empty1", filepath.Join("sub", "empty2")} {
		if _, err = os.Stat(filepath.Join(dir, p)); !os.IsNotExist(err) {
			t.Errorf("CleanEmptyFolder(): директория %q не удалена", p)
		}
	}
	// Непустые остались.
	for _, p := range []string{filepath.Join("sub", "full"), "sub"} {
		if _, err = os.Stat(filepath.Join(dir, p)); err != nil {
			t.Errorf("CleanEmptyFolder(): директория %q удалена ошибочно", p)
		}
	}
	// Корневая директория не удаляется даже если пуста.
	empty := t.TempDir()
	if err = New().CleanEmptyFolder(empty); err != nil {
		t.Fatalf("CleanEmptyFolder(): неожиданная ошибка: %v", err)
	}
	if _, err = os.Stat(empty); err != nil {
		t.Error("CleanEmptyFolder(): корневая директория не должна удаляться")
	}
	// Несуществующая директория.
	if err = New().CleanEmptyFolder(filepath.Join(dir, "absent")); err == nil {
		t.Error("CleanEmptyFolder(): для несуществующей директории ожидалась ошибка")
	}
}

// TestCleanEmptyFolderRemoveError Ошибка удаления директории.
func TestCleanEmptyFolderRemoveError(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("пропуск: права доступа unix, без пользователя root")
	}
	var (
		dir    = t.TempDir()
		locked = filepath.Join(dir, "locked")
		empty  = filepath.Join(locked, "empty")
		err    error
	)
	if err = os.MkdirAll(empty, 0755); err != nil {
		t.Fatalf("MkdirAll(): неожиданная ошибка: %v", err)
	}
	// Запрет записи в родительскую директорию делает удаление невозможным.
	if err = os.Chmod(locked, 0555); err != nil {
		t.Fatalf("Chmod(): неожиданная ошибка: %v", err)
	}
	defer func() { _ = os.Chmod(locked, 0755) }()
	if err = New().CleanEmptyFolder(dir); err == nil {
		t.Error("CleanEmptyFolder(): при ошибке удаления ожидалась ошибка")
	}
}

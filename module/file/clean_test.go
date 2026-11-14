package file

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// testContent Тестовое содержимое файла.
var testContent = []byte("содержимое тестового файла пакета file")

// Создание тестового файла с содержимым.
func createTestFile(t *testing.T, dir, name string, content []byte) (filename string) {
	const (
		tplFatalMkdirAll  = "MkdirAll(): неожиданная ошибка: %v."
		tplFatalWriteFile = "WriteFile(): неожиданная ошибка: %v."
	)

	t.Helper()
	filename = filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		t.Fatalf(tplFatalMkdirAll, err)
	}
	if err := os.WriteFile(filename, content, 0644); err != nil {
		t.Fatalf(tplFatalWriteFile, err)
	}

	return
}

// TestCleanEmptyFolder Удаление пустых директорий.
func TestCleanEmptyFolder(t *testing.T) {
	const (
		tplFatalMkdirAll   = "MkdirAll(): неожиданная ошибка: %v."
		tplFatalCleanEmpty = "CleanEmptyFolder(): неожиданная ошибка: %v."
		tplDirNotRemoved   = "CleanEmptyFolder(): директория %q не удалена."
		tplDirRemovedWrong = "CleanEmptyFolder(): директория %q удалена ошибочно."
		errRootRemoved     = "CleanEmptyFolder(): корневая директория не должна удаляться."
		errAbsentNoError   = "CleanEmptyFolder(): для несуществующей директории ожидалась ошибка."
	)
	var (
		dir   string
		empty string
		err   error
		p     string
	)

	// Дерево: dir/{empty1, sub/{empty2, full}}, dir/file.txt
	dir = t.TempDir()
	createTestFile(t, filepath.Join(dir, "sub", "full"), "f.txt", testContent)
	if err = os.MkdirAll(filepath.Join(dir, "empty1"), 0755); err != nil {
		t.Fatalf(tplFatalMkdirAll, err)
	}
	if err = os.MkdirAll(filepath.Join(dir, "sub", "empty2"), 0755); err != nil {
		t.Fatalf(tplFatalMkdirAll, err)
	}
	if err = New().CleanEmptyFolder(dir); err != nil {
		t.Fatalf(tplFatalCleanEmpty, err)
	}
	// Пустые директории удалены.
	for _, p = range []string{"empty1", filepath.Join("sub", "empty2")} {
		if _, err = os.Stat(filepath.Join(dir, p)); !os.IsNotExist(err) {
			t.Errorf(tplDirNotRemoved, p)
		}
	}
	// Непустые остались.
	for _, p = range []string{filepath.Join("sub", "full"), "sub"} {
		if _, err = os.Stat(filepath.Join(dir, p)); err != nil {
			t.Errorf(tplDirRemovedWrong, p)
		}
	}
	// Корневая директория не удаляется даже если пуста.
	empty = t.TempDir()
	if err = New().CleanEmptyFolder(empty); err != nil {
		t.Fatalf(tplFatalCleanEmpty, err)
	}
	if _, err = os.Stat(empty); err != nil {
		t.Error(errRootRemoved)
	}
	// Несуществующая директория.
	if err = New().CleanEmptyFolder(filepath.Join(dir, "absent")); err == nil {
		t.Error(errAbsentNoError)
	}
}

// TestCleanEmptyFolderRemoveError Ошибка удаления директории.
func TestCleanEmptyFolderRemoveError(t *testing.T) {
	const (
		skipNonUnix      = "пропуск: права доступа unix, без пользователя root."
		tplFatalMkdirAll = "MkdirAll(): неожиданная ошибка: %v."
		tplFatalChmod    = "Chmod(): неожиданная ошибка: %v."
		errRemoveNoError = "CleanEmptyFolder(): при ошибке удаления ожидалась ошибка."
	)
	var (
		dir    string
		locked string
		empty  string
		err    error
	)

	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip(skipNonUnix)
	}
	dir = t.TempDir()
	locked = filepath.Join(dir, "locked")
	empty = filepath.Join(locked, "empty")
	if err = os.MkdirAll(empty, 0755); err != nil {
		t.Fatalf(tplFatalMkdirAll, err)
	}
	// Запрет записи в родительскую директорию делает удаление невозможным.
	if err = os.Chmod(locked, 0555); err != nil {
		t.Fatalf(tplFatalChmod, err)
	}
	defer func() { _ = os.Chmod(locked, 0755) }()
	if err = New().CleanEmptyFolder(dir); err == nil {
		t.Error(errRemoveNoError)
	}
}

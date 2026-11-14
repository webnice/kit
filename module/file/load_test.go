package file

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestLoadFile Загрузка файла в память.
func TestLoadFile(t *testing.T) {
	const (
		tplFatalUnexpected = "LoadFile(): неожиданная ошибка: %v."
		tplInfoUnexpected  = "LoadFile(): неожиданная информация о файле: %v."
		errContentMismatch = "LoadFile(): содержимое не совпадает."
		errAbsentNoError   = "LoadFile(): для несуществующего файла ожидалась ошибка."
		errDirNoError      = "LoadFile(): для каталога ожидалась ошибка."
	)
	var (
		dir  string
		src  string
		data *bytes.Buffer
		info os.FileInfo
		err  error
	)

	dir = t.TempDir()
	src = createTestFile(t, dir, "load.txt", testContent)
	if data, info, err = New().LoadFile(src); err != nil {
		t.Fatalf(tplFatalUnexpected, err)
	}
	if info == nil || info.Name() != "load.txt" {
		t.Errorf(tplInfoUnexpected, info)
	}
	if !bytes.Equal(data.Bytes(), testContent) {
		t.Error(errContentMismatch)
	}
	// Несуществующий файл.
	if _, _, err = New().LoadFile(filepath.Join(dir, "absent.txt")); err == nil {
		t.Error(errAbsentNoError)
	}
	// Каталог: чтение данных невозможно.
	if runtime.GOOS != "windows" {
		if _, _, err = New().LoadFile(dir); err == nil {
			t.Error(errDirNoError)
		}
	}
}

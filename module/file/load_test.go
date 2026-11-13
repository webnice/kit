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
	var (
		dir  = t.TempDir()
		src  = createTestFile(t, dir, "load.txt", testContent)
		data *bytes.Buffer
		info os.FileInfo
		err  error
	)

	if data, info, err = New().LoadFile(src); err != nil {
		t.Fatalf("LoadFile(): неожиданная ошибка: %v", err)
	}
	if info == nil || info.Name() != "load.txt" {
		t.Errorf("LoadFile(): неожиданная информация о файле: %v", info)
	}
	if !bytes.Equal(data.Bytes(), testContent) {
		t.Error("LoadFile(): содержимое не совпадает")
	}

	// Несуществующий файл.
	if _, _, err = New().LoadFile(filepath.Join(dir, "absent.txt")); err == nil {
		t.Error("LoadFile(): для несуществующего файла ожидалась ошибка")
	}
	// Каталог: чтение данных невозможно.
	if runtime.GOOS != "windows" {
		if _, _, err = New().LoadFile(dir); err == nil {
			t.Error("LoadFile(): для каталога ожидалась ошибка")
		}
	}
}

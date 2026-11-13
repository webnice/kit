package file

import (
	"crypto/sha512"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestGetInfoSha512 Информация о файле с контрольной суммой.
func TestGetInfoSha512(t *testing.T) {
	var (
		dir  = t.TempDir()
		src  = createTestFile(t, dir, "info.txt", testContent)
		sum  = sha512.Sum512(testContent)
		want = hex.EncodeToString(sum[:])
		inf  *InfoSha512
		err  error
	)

	if inf, err = New().GetInfoSha512(src); err != nil {
		t.Fatalf("GetInfoSha512(): неожиданная ошибка: %v", err)
	}
	if inf.Name != "info.txt" {
		t.Errorf("GetInfoSha512(): ожидалось имя %q, получено %q", "info.txt", inf.Name)
	}
	if inf.Size != int64(len(testContent)) {
		t.Errorf("GetInfoSha512(): ожидался размер %d, получено %d", len(testContent), inf.Size)
	}
	if inf.Sha512 != want {
		t.Errorf("GetInfoSha512(): ожидалась сумма %q, получено %q", want, inf.Sha512)
	}

	// Несуществующий файл.
	if _, err = New().GetInfoSha512(filepath.Join(dir, "absent.txt")); err == nil {
		t.Error("GetInfoSha512(): для несуществующего файла ожидалась ошибка")
	}
	// Каталог: чтение невозможно.
	if _, err = New().GetInfoSha512(dir); err == nil {
		t.Error("GetInfoSha512(): для каталога ожидалась ошибка")
	}
	// Расхождение размера: файлы /proc имеют размер 0, но читаются с данными.
	if _, err = os.Stat("/proc/self/cmdline"); err == nil {
		if _, err = New().GetInfoSha512("/proc/self/cmdline"); err == nil {
			t.Error("GetInfoSha512(): при расхождении размера ожидалась ошибка")
		}
	}
}

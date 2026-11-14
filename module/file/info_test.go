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
	const (
		tplFatalUnexpected   = "GetInfoSha512(): неожиданная ошибка: %v."
		tplNameMismatch      = "GetInfoSha512(): ожидалось имя %q, получено %q."
		tplSizeMismatch      = "GetInfoSha512(): ожидался размер %d, получено %d."
		tplSumMismatch       = "GetInfoSha512(): ожидалась сумма %q, получено %q."
		errAbsentNoError     = "GetInfoSha512(): для несуществующего файла ожидалась ошибка."
		errDirNoError        = "GetInfoSha512(): для каталога ожидалась ошибка."
		errSizeMismatchNoErr = "GetInfoSha512(): при расхождении размера ожидалась ошибка."
	)
	var (
		dir  string
		src  string
		sum  [64]byte
		want string
		inf  *InfoSha512
		err  error
	)

	dir = t.TempDir()
	src = createTestFile(t, dir, "info.txt", testContent)
	sum = sha512.Sum512(testContent)
	want = hex.EncodeToString(sum[:])
	if inf, err = New().GetInfoSha512(src); err != nil {
		t.Fatalf(tplFatalUnexpected, err)
	}
	if inf.Name != "info.txt" {
		t.Errorf(tplNameMismatch, "info.txt", inf.Name)
	}
	if inf.Size != int64(len(testContent)) {
		t.Errorf(tplSizeMismatch, len(testContent), inf.Size)
	}
	if inf.Sha512 != want {
		t.Errorf(tplSumMismatch, want, inf.Sha512)
	}
	// Несуществующий файл.
	if _, err = New().GetInfoSha512(filepath.Join(dir, "absent.txt")); err == nil {
		t.Error(errAbsentNoError)
	}
	// Каталог: чтение невозможно.
	if _, err = New().GetInfoSha512(dir); err == nil {
		t.Error(errDirNoError)
	}
	// Расхождение размера: файлы /proc имеют размер 0, но читаются с данными.
	if _, err = os.Stat("/proc/self/cmdline"); err == nil {
		if _, err = New().GetInfoSha512("/proc/self/cmdline"); err == nil {
			t.Error(errSizeMismatchNoErr)
		}
	}
}

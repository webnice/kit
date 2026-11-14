package file

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// failingWriter Объект io.Writer всегда возвращающий ошибку.
type failingWriter struct{ err error }

func (fw *failingWriter) Write([]byte) (int, error) { return 0, fw.err }

// shortWriter Объект io.Writer возвращающий короткую запись без ошибки.
type shortWriter struct{}

func (sw *shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }

// failingReader Объект io.Reader возвращающий ошибку.
type failingReader struct{ err error }

func (fr *failingReader) Read([]byte) (int, error) { return 0, fr.err }

// TestCopy Копирование файла.
func TestCopy(t *testing.T) {
	const (
		tplFatalCopy       = "Copy(): неожиданная ошибка: %v."
		tplSizeMismatch    = "Copy(): ожидался размер %d, получено %d."
		tplFatalReadFile   = "ReadFile(): неожиданная ошибка: %v."
		errContentMismatch = "Copy(): содержимое скопированного файла не совпадает."
		errDstDirNoError   = "Copy(): для несуществующей директории ожидалась ошибка."
		errSrcNoError      = "Copy(): для несуществующего источника ожидалась ошибка."
		errSelfNoError     = "Copy(): при копировании файла в самого себя ожидалась ошибка."
		errSrcChanged      = "Copy(): исходный файл изменён при копировании в самого себя."
		errReadDirNoError  = "Copy(): при чтении каталога ожидалась ошибка."
	)
	var (
		dir    string
		src    string
		dst    string
		srcDir string
		size   int64
		err    error
		data   []byte
	)

	dir = t.TempDir()
	src = createTestFile(t, dir, "src.txt", testContent)
	dst = filepath.Join(dir, "dst.txt")
	if size, err = New().Copy(dst, src); err != nil {
		t.Fatalf(tplFatalCopy, err)
	}
	if size != int64(len(testContent)) {
		t.Errorf(tplSizeMismatch, len(testContent), size)
	}
	if data, err = os.ReadFile(dst); err != nil {
		t.Fatalf(tplFatalReadFile, err)
	}
	if !bytes.Equal(data, testContent) {
		t.Error(errContentMismatch)
	}
	// Ошибки копирования.
	if _, err = New().Copy(filepath.Join(dir, "absent", "dst.txt"), src); err == nil {
		t.Error(errDstDirNoError)
	}
	if _, err = New().Copy(dst, filepath.Join(dir, "absent.txt")); err == nil {
		t.Error(errSrcNoError)
	}
	// Копирование файла в самого себя.
	if _, err = New().Copy(src, src); err == nil {
		t.Error(errSelfNoError)
	}
	// После неудачной попытки исходный файл не изменён.
	if data, err = os.ReadFile(src); err != nil {
		t.Fatalf(tplFatalReadFile, err)
	}
	if !bytes.Equal(data, testContent) {
		t.Error(errSrcChanged)
	}
	// Копирование из каталога, чтение данных невозможно.
	if runtime.GOOS != "windows" {
		srcDir = t.TempDir()
		if _, err = New().Copy(filepath.Join(dir, "fromdir.txt"), srcDir); err == nil {
			t.Error(errReadDirNoError)
		}
	}
}

// TestCopyWithSha512Sum Копирование с контрольной суммой SHA512.
func TestCopyWithSha512Sum(t *testing.T) {
	const (
		errTextWriteError     = "ошибка записи"
		tplFatalUnexpected    = "CopyWithSha512Sum(): неожиданная ошибка: %v."
		tplSizeMismatch       = "CopyWithSha512Sum(): ожидался размер %d, получено %d."
		tplSumMismatch        = "CopyWithSha512Sum(): ожидалась сумма %q, получено %q."
		errContentMismatch    = "CopyWithSha512Sum(): содержимое не совпадает."
		tplUnexpectedResult   = "CopyWithSha512Sum(): неожиданный результат (%d, %q)."
		tplWriteErrorMismatch = "CopyWithSha512Sum(): ожидалась ошибка %v, получено %v."
		tplShortWriteMismatch = "CopyWithSha512Sum(): ожидалась ошибка io.ErrShortWrite, получено %v."
		errReadNoError        = "CopyWithSha512Sum(): при ошибке чтения ожидалась ошибка."
	)
	var (
		sum       [64]byte
		sum10     [64]byte
		want      string
		dst       *bytes.Buffer
		limited   *io.LimitedReader
		written   int64
		sha512sum string
		testErr   error
		err       error
	)

	sum = sha512.Sum512(testContent)
	want = hex.EncodeToString(sum[:])
	dst = new(bytes.Buffer)
	// Обычное копирование.
	written, sha512sum, err = New().CopyWithSha512Sum(dst, bytes.NewReader(testContent))
	if err != nil {
		t.Fatalf(tplFatalUnexpected, err)
	}
	if written != int64(len(testContent)) {
		t.Errorf(tplSizeMismatch, len(testContent), written)
	}
	if sha512sum != want {
		t.Errorf(tplSumMismatch, want, sha512sum)
	}
	if !bytes.Equal(dst.Bytes(), testContent) {
		t.Error(errContentMismatch)
	}
	// Источник *io.LimitedReader, буфер меньше стандартного.
	limited = &io.LimitedReader{R: bytes.NewReader(testContent), N: 10}
	dst.Reset()
	written, sha512sum, err = New().CopyWithSha512Sum(dst, limited)
	if err != nil {
		t.Fatalf(tplFatalUnexpected, err)
	}
	sum10 = sha512.Sum512(testContent[:10])
	if written != 10 || sha512sum != hex.EncodeToString(sum10[:]) {
		t.Errorf(tplUnexpectedResult, written, sha512sum)
	}
	// Пустой источник *io.LimitedReader.
	dst.Reset()
	written, sha512sum, err = New().CopyWithSha512Sum(dst, &io.LimitedReader{R: bytes.NewReader(testContent), N: 0})
	if err != nil {
		t.Fatalf(tplFatalUnexpected, err)
	}
	if written != 0 || len(sha512sum) == 0 {
		t.Errorf(tplUnexpectedResult, written, sha512sum)
	}
	// Ошибка записи.
	testErr = errors.New(errTextWriteError)
	if _, _, err = New().CopyWithSha512Sum(&failingWriter{err: testErr}, bytes.NewReader(testContent)); err != testErr {
		t.Errorf(tplWriteErrorMismatch, testErr, err)
	}
	// Короткая запись.
	if _, _, err = New().CopyWithSha512Sum(&shortWriter{}, bytes.NewReader(testContent)); !errors.Is(err, io.ErrShortWrite) {
		t.Errorf(tplShortWriteMismatch, err)
	}
	// Ошибка чтения.
	if _, _, err = New().CopyWithSha512Sum(new(bytes.Buffer), &failingReader{err: testErr}); err == nil {
		t.Error(errReadNoError)
	}
}

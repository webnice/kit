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
	var (
		dir  = t.TempDir()
		src  = createTestFile(t, dir, "src.txt", testContent)
		dst  = filepath.Join(dir, "dst.txt")
		size int64
		err  error
		data []byte
	)

	if size, err = New().Copy(dst, src); err != nil {
		t.Fatalf("Copy(): неожиданная ошибка: %v", err)
	}
	if size != int64(len(testContent)) {
		t.Errorf("Copy(): ожидался размер %d, получено %d", len(testContent), size)
	}
	if data, err = os.ReadFile(dst); err != nil {
		t.Fatalf("ReadFile(): неожиданная ошибка: %v", err)
	}
	if !bytes.Equal(data, testContent) {
		t.Error("Copy(): содержимое скопированного файла не совпадает")
	}

	// Ошибки копирования.
	if _, err = New().Copy(filepath.Join(dir, "absent", "dst.txt"), src); err == nil {
		t.Error("Copy(): для несуществующей директории ожидалась ошибка")
	}
	if _, err = New().Copy(dst, filepath.Join(dir, "absent.txt")); err == nil {
		t.Error("Copy(): для несуществующего источника ожидалась ошибка")
	}
	// Копирование файла в самого себя.
	if _, err = New().Copy(src, src); err == nil {
		t.Error("Copy(): при копировании файла в самого себя ожидалась ошибка")
	}
	// После неудачной попытки исходный файл не изменён.
	if data, err = os.ReadFile(src); err != nil {
		t.Fatalf("ReadFile(): неожиданная ошибка: %v", err)
	}
	if !bytes.Equal(data, testContent) {
		t.Error("Copy(): исходный файл изменён при копировании в самого себя")
	}
	// Копирование из каталога, чтение данных невозможно.
	if runtime.GOOS != "windows" {
		srcDir := t.TempDir()
		if _, err = New().Copy(filepath.Join(dir, "fromdir.txt"), srcDir); err == nil {
			t.Error("Copy(): при чтении каталога ожидалась ошибка")
		}
	}
}

// TestCopyWithSha512Sum Копирование с контрольной суммой SHA512.
func TestCopyWithSha512Sum(t *testing.T) {
	var (
		sum  = sha512.Sum512(testContent)
		want = hex.EncodeToString(sum[:])
		dst  = new(bytes.Buffer)
	)

	// Обычное копирование.
	written, sha512sum, err := New().CopyWithSha512Sum(dst, bytes.NewReader(testContent))
	if err != nil {
		t.Fatalf("CopyWithSha512Sum(): неожиданная ошибка: %v", err)
	}
	if written != int64(len(testContent)) {
		t.Errorf("CopyWithSha512Sum(): ожидался размер %d, получено %d", len(testContent), written)
	}
	if sha512sum != want {
		t.Errorf("CopyWithSha512Sum(): ожидалась сумма %q, получено %q", want, sha512sum)
	}
	if !bytes.Equal(dst.Bytes(), testContent) {
		t.Error("CopyWithSha512Sum(): содержимое не совпадает")
	}

	// Источник *io.LimitedReader, буфер меньше стандартного.
	limited := &io.LimitedReader{R: bytes.NewReader(testContent), N: 10}
	dst.Reset()
	written, sha512sum, err = New().CopyWithSha512Sum(dst, limited)
	if err != nil {
		t.Fatalf("CopyWithSha512Sum(): неожиданная ошибка: %v", err)
	}
	sum10 := sha512.Sum512(testContent[:10])
	if written != 10 || sha512sum != hex.EncodeToString(sum10[:]) {
		t.Errorf("CopyWithSha512Sum(): неожиданный результат (%d, %q)", written, sha512sum)
	}

	// Пустой источник *io.LimitedReader.
	dst.Reset()
	written, sha512sum, err = New().CopyWithSha512Sum(dst, &io.LimitedReader{R: bytes.NewReader(testContent), N: 0})
	if err != nil {
		t.Fatalf("CopyWithSha512Sum(): неожиданная ошибка: %v", err)
	}
	if written != 0 || len(sha512sum) == 0 {
		t.Errorf("CopyWithSha512Sum(): неожиданный результат (%d, %q)", written, sha512sum)
	}

	// Ошибка записи.
	testErr := errors.New("ошибка записи")
	if _, _, err = New().CopyWithSha512Sum(&failingWriter{err: testErr}, bytes.NewReader(testContent)); err != testErr {
		t.Errorf("CopyWithSha512Sum(): ожидалась ошибка %v, получено %v", testErr, err)
	}

	// Короткая запись.
	if _, _, err = New().CopyWithSha512Sum(&shortWriter{}, bytes.NewReader(testContent)); !errors.Is(err, io.ErrShortWrite) {
		t.Errorf("CopyWithSha512Sum(): ожидалась ошибка io.ErrShortWrite, получено %v", err)
	}

	// Ошибка чтения.
	if _, _, err = New().CopyWithSha512Sum(new(bytes.Buffer), &failingReader{err: testErr}); err == nil {
		t.Error("CopyWithSha512Sum(): при ошибке чтения ожидалась ошибка")
	}
}

package fileinfo

import (
	"os"
	"testing"
	"time"
)

// Объект эмуляции os.FileInfo для тестирования.
type testFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     any
}

// Name Реализация интерфейса os.FileInfo.
func (tfi testFileInfo) Name() string { return tfi.name }

// Size Реализация интерфейса os.FileInfo.
func (tfi testFileInfo) Size() int64 { return tfi.size }

// Mode Реализация интерфейса os.FileInfo.
func (tfi testFileInfo) Mode() os.FileMode { return tfi.mode }

// ModTime Реализация интерфейса os.FileInfo.
func (tfi testFileInfo) ModTime() time.Time { return tfi.modTime }

// IsDir Реализация интерфейса os.FileInfo.
func (tfi testFileInfo) IsDir() bool { return tfi.isDir }

// Sys Реализация интерфейса os.FileInfo.
func (tfi testFileInfo) Sys() any { return tfi.sys }

func TestImpl_New(t *testing.T) {
	const (
		errNewNil         = "New() вернул nil."
		tplNameNotEmpty   = "Name() = %q, ожидалось пустое значение."
		tplSizeNotZero    = "Size() = %d, ожидался 0."
		tplModeNotZero    = "Mode() = %v, ожидался 0."
		tplModTimeNotZero = "ModTime() = %v, ожидалось нулевое время."
		errIsDirTrue      = "IsDir() = true, ожидался false."
		tplSysNotNil      = "Sys() = %v, ожидался nil."
	)
	var fif Interface

	fif = New()
	if fif == nil {
		t.Fatal(errNewNil)
	}
	if fif.Name() != "" {
		t.Fatalf(tplNameNotEmpty, fif.Name())
	}
	if fif.Size() != 0 {
		t.Fatalf(tplSizeNotZero, fif.Size())
	}
	if fif.Mode() != 0 {
		t.Fatalf(tplModeNotZero, fif.Mode())
	}
	if !fif.ModTime().IsZero() {
		t.Fatalf(tplModTimeNotZero, fif.ModTime())
	}
	if fif.IsDir() {
		t.Fatal(errIsDirTrue)
	}
	if fif.Sys() != nil {
		t.Fatalf(tplSysNotNil, fif.Sys())
	}
}

func TestImpl_SettersAndGetters(t *testing.T) {
	const (
		errSetNameNotChain   = "SetName() вернул не исходный объект."
		errSetSizeNotChain   = "SetSize() вернул не исходный объект."
		errSetModeNotChain   = "SetMode() вернул не исходный объект."
		tplSetModTimeChain   = "SetModTime() вернул не исходный объект."
		errSetIsDirNotChain  = "SetIsDir() вернул не исходный объект."
		errSetSysNotChain    = "SetSys() вернул не исходный объект."
		tplNameUnexpected    = "Name() = %q, ожидалось %q."
		tplSizeUnexpected    = "Size() = %d, ожидалось %d."
		tplModeUnexpected    = "Mode() = %v, ожидалось %v."
		tplModTimeUnexpected = "ModTime() = %v, ожидалось %v."
		errIsDirFalse        = "IsDir() = false, ожидался true."
		tplSysUnexpected     = "Sys() = %v, ожидалось %v."
	)
	var (
		modTime time.Time
		sys     *struct{ Inode uint64 }
		fif     Interface
	)

	modTime = time.Date(2026, 11, 6, 12, 34, 56, 0, time.UTC)
	sys = &struct{ Inode uint64 }{Inode: 12345}
	fif = New()
	if fif.SetName("/tmp/test-file.txt") != fif {
		t.Fatal(errSetNameNotChain)
	}
	if fif.SetSize(12345) != fif {
		t.Fatal(errSetSizeNotChain)
	}
	if fif.SetMode(os.ModePerm|os.ModeDir) != fif {
		t.Fatal(errSetModeNotChain)
	}
	if fif.SetModTime(modTime) != fif {
		t.Fatal(tplSetModTimeChain)
	}
	if fif.SetIsDir(true) != fif {
		t.Fatal(errSetIsDirNotChain)
	}
	if fif.SetSys(sys) != fif {
		t.Fatal(errSetSysNotChain)
	}
	if fif.Name() != "test-file.txt" {
		t.Fatalf(tplNameUnexpected, fif.Name(), "test-file.txt")
	}
	if fif.Size() != 12345 {
		t.Fatalf(tplSizeUnexpected, fif.Size(), 12345)
	}
	if fif.Mode() != os.ModePerm|os.ModeDir {
		t.Fatalf(tplModeUnexpected, fif.Mode(), os.ModePerm|os.ModeDir)
	}
	if !fif.ModTime().Equal(modTime) {
		t.Fatalf(tplModTimeUnexpected, fif.ModTime(), modTime)
	}
	if !fif.IsDir() {
		t.Fatal(errIsDirFalse)
	}
	if fif.Sys() != sys {
		t.Fatalf(tplSysUnexpected, fif.Sys(), sys)
	}
}

func TestImpl_CopyFrom(t *testing.T) {
	const (
		errCopyFromNil       = "CopyFrom() вернул nil."
		tplNameUnexpected    = "Name() = %q, ожидалось %q."
		tplSizeUnexpected    = "Size() = %d, ожидалось %d."
		tplModeUnexpected    = "Mode() = %v, ожидалось %v."
		tplModTimeUnexpected = "ModTime() = %v, ожидалось %v."
		tplIsDirUnexpected   = "IsDir() = %t, ожидалось %t."
		tplSysUnexpected     = "Sys() = %v, ожидалось %v."
	)
	var (
		modTime time.Time
		sys     *struct{ Inode uint64 }
		src     testFileInfo
		dst     Interface
	)

	modTime = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	sys = &struct{ Inode uint64 }{Inode: 54321}
	src = testFileInfo{
		name:    "source-file.bin",
		size:    98765,
		mode:    os.ModeTemporary | 0o644,
		modTime: modTime,
		isDir:   true,
		sys:     sys,
	}
	dst = New().CopyFrom(src)
	if dst == nil {
		t.Fatal(errCopyFromNil)
	}
	if dst.Name() != src.Name() {
		t.Fatalf(tplNameUnexpected, dst.Name(), src.Name())
	}
	if dst.Size() != src.Size() {
		t.Fatalf(tplSizeUnexpected, dst.Size(), src.Size())
	}
	if dst.Mode() != src.Mode() {
		t.Fatalf(tplModeUnexpected, dst.Mode(), src.Mode())
	}
	if !dst.ModTime().Equal(src.ModTime()) {
		t.Fatalf(tplModTimeUnexpected, dst.ModTime(), src.ModTime())
	}
	if dst.IsDir() != src.IsDir() {
		t.Fatalf(tplIsDirUnexpected, dst.IsDir(), src.IsDir())
	}
	if dst.Sys() != src.Sys() {
		t.Fatalf(tplSysUnexpected, dst.Sys(), src.Sys())
	}
}

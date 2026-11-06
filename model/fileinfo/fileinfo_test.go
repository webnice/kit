package fileinfo

import (
	"os"
	"testing"
	"time"
)

type testFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     any
}

func (tfi testFileInfo) Name() string       { return tfi.name }
func (tfi testFileInfo) Size() int64        { return tfi.size }
func (tfi testFileInfo) Mode() os.FileMode  { return tfi.mode }
func (tfi testFileInfo) ModTime() time.Time { return tfi.modTime }
func (tfi testFileInfo) IsDir() bool        { return tfi.isDir }
func (tfi testFileInfo) Sys() any           { return tfi.sys }

func TestImpl_New(t *testing.T) {
	var fif = New()

	if fif == nil {
		t.Fatal("New() вернул nil")
	}
	if fif.Name() != "" {
		t.Fatalf("Name() = %q, ожидалось пустое значение", fif.Name())
	}
	if fif.Size() != 0 {
		t.Fatalf("Size() = %d, ожидался 0", fif.Size())
	}
	if fif.Mode() != 0 {
		t.Fatalf("Mode() = %v, ожидался 0", fif.Mode())
	}
	if !fif.ModTime().IsZero() {
		t.Fatalf("ModTime() = %v, ожидалось нулевое время", fif.ModTime())
	}
	if fif.IsDir() {
		t.Fatal("IsDir() = true, ожидался false")
	}
	if fif.Sys() != nil {
		t.Fatalf("Sys() = %v, ожидался nil", fif.Sys())
	}
}

func TestImpl_SettersAndGetters(t *testing.T) {
	var (
		modTime = time.Date(2026, 11, 6, 12, 34, 56, 0, time.UTC)
		sys     = &struct{ Inode uint64 }{Inode: 12345}
		fif     = New()
	)

	if fif.SetName("/tmp/test-file.txt") != fif {
		t.Fatal("SetName() вернул не исходный объект")
	}
	if fif.SetSize(12345) != fif {
		t.Fatal("SetSize() вернул не исходный объект")
	}
	if fif.SetMode(os.ModePerm|os.ModeDir) != fif {
		t.Fatal("SetMode() вернул не исходный объект")
	}
	if fif.SetModTime(modTime) != fif {
		t.Fatal("SetModTime() вернул не исходный объект")
	}
	if fif.SetIsDir(true) != fif {
		t.Fatal("SetIsDir() вернул не исходный объект")
	}
	if fif.SetSys(sys) != fif {
		t.Fatal("SetSys() вернул не исходный объект")
	}

	if fif.Name() != "test-file.txt" {
		t.Fatalf("Name() = %q, ожидалось %q", fif.Name(), "test-file.txt")
	}
	if fif.Size() != 12345 {
		t.Fatalf("Size() = %d, ожидалось %d", fif.Size(), 12345)
	}
	if fif.Mode() != os.ModePerm|os.ModeDir {
		t.Fatalf("Mode() = %v, ожидалось %v", fif.Mode(), os.ModePerm|os.ModeDir)
	}
	if !fif.ModTime().Equal(modTime) {
		t.Fatalf("ModTime() = %v, ожидалось %v", fif.ModTime(), modTime)
	}
	if !fif.IsDir() {
		t.Fatal("IsDir() = false, ожидался true")
	}
	if fif.Sys() != sys {
		t.Fatalf("Sys() = %v, ожидалось %v", fif.Sys(), sys)
	}
}

func TestImpl_CopyFrom(t *testing.T) {
	var (
		modTime = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		sys     = &struct{ Inode uint64 }{Inode: 54321}
		src     = testFileInfo{
			name:    "source-file.bin",
			size:    98765,
			mode:    os.ModeTemporary | 0o644,
			modTime: modTime,
			isDir:   true,
			sys:     sys,
		}
		dst = New().CopyFrom(src)
	)

	if dst == nil {
		t.Fatal("CopyFrom() вернул nil")
	}
	if dst.Name() != src.Name() {
		t.Fatalf("Name() = %q, ожидалось %q", dst.Name(), src.Name())
	}
	if dst.Size() != src.Size() {
		t.Fatalf("Size() = %d, ожидалось %d", dst.Size(), src.Size())
	}
	if dst.Mode() != src.Mode() {
		t.Fatalf("Mode() = %v, ожидалось %v", dst.Mode(), src.Mode())
	}
	if !dst.ModTime().Equal(src.ModTime()) {
		t.Fatalf("ModTime() = %v, ожидалось %v", dst.ModTime(), src.ModTime())
	}
	if dst.IsDir() != src.IsDir() {
		t.Fatalf("IsDir() = %t, ожидалось %t", dst.IsDir(), src.IsDir())
	}
	if dst.Sys() != src.Sys() {
		t.Fatalf("Sys() = %v, ожидалось %v", dst.Sys(), src.Sys())
	}
}

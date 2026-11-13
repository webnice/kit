package file

import (
	"path/filepath"
	"testing"
)

// TestGetFilename Выделение имени файла из пути.
func TestGetFilename(t *testing.T) {
	var sep = string(filepath.Separator)
	var cases = []struct {
		in   string
		want string
	}{
		{"", ""},
		{sep, ""},
		{"file.txt", "file.txt"},
		{"file", "file"},
		{filepath.Join("a", "b", "c.txt"), "c.txt"},
		{sep + filepath.Join("a", "b", "c.txt"), "c.txt"},
		{"a" + sep, ""},
		{filepath.Join("папка", "файл.txt"), "файл.txt"},
	}
	for _, tt := range cases {
		if got := New().GetFilename(tt.in); got != tt.want {
			t.Errorf("GetFilename(%q): ожидалось %q, получено %q", tt.in, tt.want, got)
		}
	}
}

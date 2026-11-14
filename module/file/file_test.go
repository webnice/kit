package file

import (
	"path/filepath"
	"testing"
)

// TestGetFilename Выделение имени файла из пути.
func TestGetFilename(t *testing.T) {
	const (
		tplUnexpected = "GetFilename(%q): ожидалось %q, получено %q."
	)
	// Описание одного теста выделения имени файла.
	type testItem struct {
		in   string
		want string
	}
	var (
		sep   string
		cases []testItem
		n     int
		got   string
	)

	sep = string(filepath.Separator)
	cases = []testItem{
		{"", ""},
		{sep, ""},
		{"file.txt", "file.txt"},
		{"file", "file"},
		{filepath.Join("a", "b", "c.txt"), "c.txt"},
		{sep + filepath.Join("a", "b", "c.txt"), "c.txt"},
		{"a" + sep, ""},
		{filepath.Join("папка", "файл.txt"), "файл.txt"},
	}
	for n = range cases {
		if got = New().GetFilename(cases[n].in); got != cases[n].want {
			t.Errorf(tplUnexpected, cases[n].in, cases[n].want, got)
		}
	}
}

package emt

import (
	"os"
	"strings"
	"testing"
)

func TestFilenameParsingAndHelpers(t *testing.T) {
	var sep = string(os.PathSeparator)

	fn := parseFilename("dir/sub/file.tar.gz")
	if len(fn.Path) != 2 || fn.Path[0] != "dir" || fn.Path[1] != "sub" {
		t.Fatalf("parseFilename().Path = %#v", fn.Path)
	}
	if fn.Filename != "file.tar" || fn.Extension != "gz" {
		t.Fatalf("parseFilename() = %#v", fn)
	}
	if fn.String() != strings.Join([]string{"dir", "sub", "file.tar.gz"}, sep) {
		t.Fatalf("filename.String() = %q", fn.String())
	}
	if fn.IsPath() || fn.IsDotFirst() || !fn.IsPathPrefix("sub") || fn.IsPathMacos() {
		t.Fatalf("filename helpers вернули некорректный результат: %#v", fn)
	}

	fn = parseFilename("dir/__MACOSX/file.txt")
	if !fn.IsPathMacos() {
		t.Fatalf("IsPathMacos() = false для %#v", fn.Path)
	}
	fn = parseFilename(".hidden")
	if fn.Filename != ".hidden" || fn.Extension != "" || !fn.IsDotFirst() {
		t.Fatalf("parseFilename(.hidden) = %#v", fn)
	}
	fn = parseFilename("")
	if !fn.IsPath() || fn.String() != "" {
		t.Fatalf("parseFilename(пустая строка) = %#v", fn)
	}
	fn = parseFilename("dir/")
	if !fn.IsPath() {
		t.Fatalf("parseFilename(dir/) не определил директорию: %#v", fn)
	}
}

func TestFilenameContentTypeFallback(t *testing.T) {
	if got := (filename{}).ContentType(); got != "application/octet-stream" {
		t.Fatalf("ContentType() без расширения = %q", got)
	}
	if got := (filename{Extension: "unknownextension"}).ContentType(); got != "application/octet-stream" {
		t.Fatalf("ContentType() с неизвестным расширением = %q", got)
	}
	if got := (filename{Extension: "png"}).ContentType(); !strings.HasPrefix(got, "image/png") {
		t.Fatalf("ContentType(png) = %q", got)
	}
}

func TestFilenameEdgeBranches(t *testing.T) {
	if (filename{}).IsPathPrefix("x") {
		t.Fatal("IsPathPrefix() с пустым путём и непустым prefix вернул true")
	}
	if !(filename{Path: []string{".dir"}, Filename: "file.txt"}).IsDotFirst() {
		t.Fatal("IsDotFirst() не обнаружил точку в сегменте пути")
	}
	if got := (filename{Extension: "png"}).FilenameFull(); got != "" {
		t.Fatalf("FilenameFull() без имени файла = %q", got)
	}
	fn := parseFilename("dir/.hidden/file.txt")
	if !fn.IsDotFirst() || fn.Filename != "file" || fn.Extension != "txt" {
		t.Fatalf("parseFilename(dir/.hidden/file.txt) = %#v", fn)
	}
}

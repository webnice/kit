package emt

import (
	"os"
	"strings"
	"testing"
)

func TestFilenameParsingAndHelpers(t *testing.T) {
	const (
		tplPathMismatch     = "parseFilename().Path = %#v."
		tplFilenameMismatch = "parseFilename() = %#v."
		tplStringMismatch   = "filename.String() = %q."
		tplHelpersWrong     = "filename helpers вернули некорректный результат: %#v."
		tplMacosFalse       = "IsPathMacos() = false для %#v."
		tplHiddenMismatch   = "parseFilename(.hidden) = %#v."
		tplEmptyMismatch    = "parseFilename(пустая строка) = %#v."
		tplDirNotDetected   = "parseFilename(dir/) не определил директорию: %#v."
	)
	var (
		sep string
		fn  *filename
	)

	sep = string(os.PathSeparator)
	fn = parseFilename("dir/sub/file.tar.gz")
	if len(fn.Path) != 2 || fn.Path[0] != "dir" || fn.Path[1] != "sub" {
		t.Fatalf(tplPathMismatch, fn.Path)
	}
	if fn.Filename != "file.tar" || fn.Extension != "gz" {
		t.Fatalf(tplFilenameMismatch, fn)
	}
	if fn.String() != strings.Join([]string{"dir", "sub", "file.tar.gz"}, sep) {
		t.Fatalf(tplStringMismatch, fn.String())
	}
	if fn.IsPath() || fn.IsDotFirst() || !fn.IsPathPrefix("sub") || fn.IsPathMacos() {
		t.Fatalf(tplHelpersWrong, fn)
	}
	fn = parseFilename("dir/__MACOSX/file.txt")
	if !fn.IsPathMacos() {
		t.Fatalf(tplMacosFalse, fn.Path)
	}
	fn = parseFilename(".hidden")
	if fn.Filename != ".hidden" || fn.Extension != "" || !fn.IsDotFirst() {
		t.Fatalf(tplHiddenMismatch, fn)
	}
	fn = parseFilename("")
	if !fn.IsPath() || fn.String() != "" {
		t.Fatalf(tplEmptyMismatch, fn)
	}
	fn = parseFilename("dir/")
	if !fn.IsPath() {
		t.Fatalf(tplDirNotDetected, fn)
	}
}

func TestFilenameContentTypeFallback(t *testing.T) {
	const (
		tplNoExtension = "ContentType() без расширения = %q."
		tplUnknownExt  = "ContentType() с неизвестным расширением = %q."
		tplPngMismatch = "ContentType(png) = %q."
	)
	var got string

	if got = (filename{}).ContentType(); got != "application/octet-stream" {
		t.Fatalf(tplNoExtension, got)
	}
	if got = (filename{Extension: "unknownextension"}).ContentType(); got != "application/octet-stream" {
		t.Fatalf(tplUnknownExt, got)
	}
	if got = (filename{Extension: "png"}).ContentType(); !strings.HasPrefix(got, "image/png") {
		t.Fatalf(tplPngMismatch, got)
	}
}

func TestFilenameEdgeBranches(t *testing.T) {
	const (
		errPrefixTrue     = "IsPathPrefix() с пустым путём и непустым prefix вернул true."
		errDotNotDetected = "IsDotFirst() не обнаружил точку в сегменте пути."
		tplFilenameFull   = "FilenameFull() без имени файла = %q."
		tplHiddenPath     = "parseFilename(dir/.hidden/file.txt) = %#v."
	)
	var (
		fn  *filename
		got string
	)

	if (filename{}).IsPathPrefix("x") {
		t.Fatal(errPrefixTrue)
	}
	if !(filename{Path: []string{".dir"}, Filename: "file.txt"}).IsDotFirst() {
		t.Fatal(errDotNotDetected)
	}
	if got = (filename{Extension: "png"}).FilenameFull(); got != "" {
		t.Fatalf(tplFilenameFull, got)
	}
	fn = parseFilename("dir/.hidden/file.txt")
	if !fn.IsDotFirst() || fn.Filename != "file" || fn.Extension != "txt" {
		t.Fatalf(tplHiddenPath, fn)
	}
}

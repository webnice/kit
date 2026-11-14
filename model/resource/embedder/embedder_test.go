package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Тип контента тестовых ресурсов.
const testContentType = "text/plain"

func testEmbedderResourceInfo(name string, content []byte) resourceInfo {
	return resourceInfo{
		Size:        uint64(len(content)),
		Time:        time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC),
		TimeFormat:  time.RFC3339Nano,
		Name:        name,
		ContentType: testContentType,
		Content:     getSliceOfBytesAsString(content),
	}
}

func writeTestFile(t *testing.T, name string, content []byte) {
	const (
		tplFatalMkdirAll  = "MkdirAll(%q) вернул ошибку: %v."
		tplFatalWriteFile = "WriteFile(%q) вернул ошибку: %v."
	)

	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatalf(tplFatalMkdirAll, filepath.Dir(name), err)
	}
	if err := os.WriteFile(name, content, 0644); err != nil {
		t.Fatalf(tplFatalWriteFile, name, err)
	}
}

func TestGetSliceOfBytesAsString(t *testing.T) {
	const (
		errEmptyNotEmpty  = "getSliceOfBytesAsString(nil) вернул не пустой builder."
		tplCountHexValues = "getSliceOfBytesAsString() вернул %q, ожидалось %d hex значений."
		tplNoNewLine      = "getSliceOfBytesAsString() не перенёс строку: %q."
		tplEndsNewLine    = "getSliceOfBytesAsString() заканчивается переносом строки: %q."
	)
	var (
		empty   *strings.Builder
		content []byte
		got     string
	)

	empty = getSliceOfBytesAsString(nil)
	if empty == nil || empty.Len() != 0 {
		t.Fatal(errEmptyNotEmpty)
	}
	content = bytes.Repeat([]byte{0x01, 0x02}, 9)
	got = getSliceOfBytesAsString(content).String()
	if strings.Count(got, "0x") != len(content) {
		t.Fatalf(tplCountHexValues, got, len(content))
	}
	if !strings.Contains(got, "\n") {
		t.Fatalf(tplNoNewLine, got)
	}
	if strings.HasSuffix(got, "\n") {
		t.Fatalf(tplEndsNewLine, got)
	}
}

func TestGetFileContentType(t *testing.T) {
	const (
		errNilContentNoError = "getFileContentType(nil) вернул nil ошибку."
		tplContentTypeNotEq  = "getFileContentType(nil) вернул тип контента %q, ожидался пустой."
		tplUnexpectedError   = "getFileContentType() вернул ошибку: %v."
		tplContentTypeWrong  = "getFileContentType() = %q, ожидался text/plain."
	)
	var (
		ret string
		err error
	)

	ret, err = getFileContentType(nil)
	if err == nil {
		t.Fatal(errNilContentNoError)
	}
	if ret != "" {
		t.Fatalf(tplContentTypeNotEq, ret)
	}
	ret, err = getFileContentType([]byte("hello"))
	if err != nil {
		t.Fatalf(tplUnexpectedError, err)
	}
	if ret != "text/plain; charset=utf-8" {
		t.Fatalf(tplContentTypeWrong, ret)
	}
}

func TestGetFileContent(t *testing.T) {
	const (
		tplUnexpectedError  = "getFileContent() вернул ошибку: %v."
		tplContentNotEq     = "getFileContent() = %q, ожидалось %q."
		errMissingFileNoErr = "getFileContent() для отсутствующего файла вернул nil ошибку."
	)
	var (
		tmp     string
		fn      string
		content []byte
		got     []byte
		err     error
	)

	tmp = t.TempDir()
	fn = filepath.Join(tmp, "file.txt")
	content = []byte("file content")
	writeTestFile(t, fn, content)
	got, err = getFileContent(fn)
	if err != nil {
		t.Fatalf(tplUnexpectedError, err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf(tplContentNotEq, got, content)
	}
	if _, err = getFileContent(filepath.Join(tmp, "missing.txt")); err == nil {
		t.Fatal(errMissingFileNoErr)
	}
}

func TestGetArg(t *testing.T) {
	const (
		tplUnexpectedError  = "getArg() вернул ошибку: %v."
		tplPackageNotEq     = "Package = %q, ожидался resource."
		tplPathNotEq        = "Path = %q, ожидался %q."
		tplBaseDirNotEq     = "BaseDir = %q, ожидался %q."
		tplSourceUnexpected = "Source = %v, ожидались group1 и group2."
		tplSourcePathNotEq  = "Sources[group1].Path = %q, ожидался %q."
		tplSourcePath2NotEq = "Sources[group2].Path = %q, ожидался %q."
	)
	var (
		oldFlagSet *flag.FlagSet
		oldArgs    []string
		tmp        string
		cfg        *configuration
		err        error
	)

	oldFlagSet = flag.CommandLine
	oldArgs = os.Args
	tmp = t.TempDir()
	flag.CommandLine = flag.NewFlagSet("embedder", flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	os.Args = []string{"embedder", "--package", "resource", "--path", filepath.Join(tmp, "out")}
	t.Cleanup(func() {
		flag.CommandLine = oldFlagSet
		os.Args = oldArgs
	})
	t.Setenv(envEmbedderStaticBaseDir, tmp)
	t.Setenv(envEmbedderStaticResources, "group1:static,invalid,group2:assets")
	cfg, err = getArg()
	if err != nil {
		t.Fatalf(tplUnexpectedError, err)
	}
	if cfg.Package != "resource" {
		t.Fatalf(tplPackageNotEq, cfg.Package)
	}
	if cfg.Path != filepath.Join(tmp, "out") {
		t.Fatalf(tplPathNotEq, cfg.Path, filepath.Join(tmp, "out"))
	}
	if cfg.BaseDir != tmp {
		t.Fatalf(tplBaseDirNotEq, cfg.BaseDir, tmp)
	}
	if len(cfg.Source) != 2 || cfg.Source[0] != "group1" || cfg.Source[1] != "group2" {
		t.Fatalf(tplSourceUnexpected, cfg.Source)
	}
	if cfg.Sources["group1"].Path != path.Join(tmp, "static") {
		t.Fatalf(tplSourcePathNotEq, cfg.Sources["group1"].Path, path.Join(tmp, "static"))
	}
	if cfg.Sources["group2"].Path != path.Join(tmp, "assets") {
		t.Fatalf(tplSourcePath2NotEq, cfg.Sources["group2"].Path, path.Join(tmp, "assets"))
	}
}

func TestGetSourceInfo(t *testing.T) {
	const (
		tplUnexpectedError     = "getSourceInfo() вернул ошибку: %v."
		tplResourcesCount      = "getSourceInfo() вернул %d ресурсов, ожидалось 2: %+v."
		tplResourceNotFound    = "ресурс %q не найден в %+v."
		tplSizeNotEq           = "Size ресурса %q = %d, ожидалось %d."
		tplContentTypeEmpty    = "ContentType ресурса %q пустой."
		tplContentNotConverted = "Content ресурса %q не преобразован в срез байт: %+v."
		errMissingDirNoError   = "getSourceInfo() для отсутствующей директории вернул nil ошибку."
	)
	var (
		tmp   string
		one   []byte
		two   []byte
		infos []resourceInfo
		found map[string]resourceInfo
		info  resourceInfo
		ok    bool
		name  string
		want  []byte
		err   error
	)

	tmp = t.TempDir()
	one = []byte("hello")
	two = []byte("world")
	writeTestFile(t, filepath.Join(tmp, "one.txt"), one)
	writeTestFile(t, filepath.Join(tmp, "sub", "two.txt"), two)
	writeTestFile(t, filepath.Join(tmp, "empty.txt"), nil)
	infos, err = getSourceInfo(tmp, "")
	if err != nil {
		t.Fatalf(tplUnexpectedError, err)
	}
	if len(infos) != 2 {
		t.Fatalf(tplResourcesCount, len(infos), infos)
	}
	found = make(map[string]resourceInfo)
	for _, info = range infos {
		found[info.Name] = info
	}
	for name, want = range map[string][]byte{"one.txt": one, path.Join("sub", "two.txt"): two} {
		info, ok = found[name]
		if !ok {
			t.Fatalf(tplResourceNotFound, name, found)
		}
		if info.Size != uint64(len(want)) {
			t.Fatalf(tplSizeNotEq, name, info.Size, len(want))
		}
		if info.ContentType == "" {
			t.Fatalf(tplContentTypeEmpty, name)
		}
		if info.Content == nil || !strings.Contains(info.Content.String(), "0x") {
			t.Fatalf(tplContentNotConverted, name, info.Content)
		}
	}
	if _, err = getSourceInfo(filepath.Join(tmp, "missing"), ""); err == nil {
		t.Fatal(errMissingDirNoError)
	}
}

func TestGetResourceList(t *testing.T) {
	const (
		tplUnexpectedError = "getResourceList() вернул ошибку: %v."
		tplFilesUnexpected = "getResourceList() вернул ресурсы %+v, ожидался один файл."
		errMissingDirNoErr = "getResourceList() для отсутствующей директории вернул nil ошибку."
	)
	var (
		tmp string
		cfg *configuration
		err error
	)

	tmp = t.TempDir()
	cfg = &configuration{Sources: map[string]resourceGroup{
		"assets": {Path: tmp},
	}}
	writeTestFile(t, filepath.Join(tmp, "file.txt"), []byte("content"))
	if err = getResourceList(cfg); err != nil {
		t.Fatalf(tplUnexpectedError, err)
	}
	if len(cfg.Sources["assets"].File) != 1 {
		t.Fatalf(tplFilesUnexpected, cfg.Sources["assets"].File)
	}
	cfg = &configuration{Sources: map[string]resourceGroup{
		"assets": {Path: filepath.Join(tmp, "missing")},
	}}
	if err = getResourceList(cfg); err == nil {
		t.Fatal(errMissingDirNoErr)
	}
}

func TestCleanResourceContent(t *testing.T) {
	const (
		tplUnexpectedError  = "cleanResourceContent() вернул ошибку: %v."
		tplFileNotRemoved   = "cleanResourceContent() не удалил старый файл ресурса: %v."
		tplFileRemovedExtra = "cleanResourceContent() удалил лишний файл %q: %v."
		errMissingDirNoErr  = "cleanResourceContent() для отсутствующей директории вернул nil ошибку."
	)
	var (
		tmp       string
		remove    string
		keepGroup string
		keepFile  string
		cfg       *configuration
		fn        string
		err       error
	)

	tmp = t.TempDir()
	remove = filepath.Join(tmp, "resource_content_test_00000000000000000001.go")
	keepGroup = filepath.Join(tmp, "resource_content_other_00000000000000000001.go")
	keepFile = filepath.Join(tmp, "readme.txt")
	cfg = &configuration{Path: tmp, Source: []string{"test"}}
	writeTestFile(t, remove, []byte("remove"))
	writeTestFile(t, keepGroup, []byte("keep group"))
	writeTestFile(t, keepFile, []byte("keep file"))
	if err = cleanResourceContent(cfg); err != nil {
		t.Fatalf(tplUnexpectedError, err)
	}
	if _, err = os.Stat(remove); !os.IsNotExist(err) {
		t.Fatalf(tplFileNotRemoved, err)
	}
	for _, fn = range []string{keepGroup, keepFile} {
		if _, err = os.Stat(fn); err != nil {
			t.Fatalf(tplFileRemovedExtra, fn, err)
		}
	}
	cfg.Path = filepath.Join(tmp, "missing")
	if err = cleanResourceContent(cfg); err == nil {
		t.Fatal(errMissingDirNoErr)
	}
}

func TestMakeResourceContent(t *testing.T) {
	const (
		tplUnexpectedError   = "makeResourceContent() вернул ошибку: %v."
		tplFileNotRead       = "созданный файл ресурса не прочитан: %v."
		tplFileNoPart        = "созданный файл ресурса не содержит %q:\n%s."
		errEmptyPackageNoErr = "makeResourceContent() с пустым package вернул nil ошибку."
		errMissingDirNoErr   = "makeResourceContent() с отсутствующей директорией вернул nil ошибку."
	)
	var (
		tmp  string
		cfg  *configuration
		data []byte
		part string
		err  error
	)

	tmp = t.TempDir()
	cfg = &configuration{
		Package: "resource",
		Path:    tmp,
		Sources: map[string]resourceGroup{
			"assets": {File: []resourceInfo{testEmbedderResourceInfo("hello.txt", []byte("hello"))}},
		},
	}
	if err = makeResourceContent(cfg); err != nil {
		t.Fatalf(tplUnexpectedError, err)
	}
	data, err = os.ReadFile(filepath.Join(tmp, "resource_content_assets_00000000000000000001.go"))
	if err != nil {
		t.Fatalf(tplFileNotRead, err)
	}
	for _, part = range []string{"package resource", "Get().Add(", `"assets"`, `"hello.txt"`, "Size: 5", "ContentType: \"text/plain\""} {
		if !strings.Contains(string(data), part) {
			t.Fatalf(tplFileNoPart, part, data)
		}
	}
	cfg = &configuration{
		Path: tmp,
		Sources: map[string]resourceGroup{
			"broken": {File: []resourceInfo{testEmbedderResourceInfo("broken.txt", []byte("broken"))}},
		},
	}
	if err = makeResourceContent(cfg); err == nil {
		t.Fatal(errEmptyPackageNoErr)
	}
	cfg = &configuration{
		Package: "resource",
		Path:    filepath.Join(tmp, "missing"),
		Sources: map[string]resourceGroup{
			"assets": {File: []resourceInfo{testEmbedderResourceInfo("hello.txt", []byte("hello"))}},
		},
	}
	if err = makeResourceContent(cfg); err == nil {
		t.Fatal(errMissingDirNoErr)
	}
}

func TestSaveFile(t *testing.T) {
	const (
		tplUnexpectedError = "saveFile() вернул ошибку: %v."
		tplFileNotCreated  = "saveFile() не создал файл: %v."
		tplDataNotEq       = "saveFile() записал %q, ожидалось %q."
		errMissingDirNoErr = "saveFile() в отсутствующую директорию вернул nil ошибку."
		fnTemplate         = "resource_content_%s_%020d.go"
	)
	var (
		tmp  string
		data []byte
		fn   string
		got  []byte
		err  error
	)

	tmp = t.TempDir()
	data = []byte("package resource\n")
	if err = saveFile(tmp, "group", 7, data); err != nil {
		t.Fatalf(tplUnexpectedError, err)
	}
	fn = filepath.Join(tmp, fmt.Sprintf(fnTemplate, "group", 7))
	got, err = os.ReadFile(fn)
	if err != nil {
		t.Fatalf(tplFileNotCreated, err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf(tplDataNotEq, got, data)
	}
	if err = saveFile(filepath.Join(tmp, "missing"), "group", 8, data); err == nil {
		t.Fatal(errMissingDirNoErr)
	}
}

func TestGetFileContentReadError(t *testing.T) {
	const errDirNoError = "getFileContent() для директории вернул nil ошибку."

	if _, err := getFileContent(t.TempDir()); err == nil {
		t.Fatal(errDirNoError)
	}
}

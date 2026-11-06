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

func testEmbedderResourceInfo(name string, content []byte) resourceInfo {
	return resourceInfo{
		Size:        uint64(len(content)),
		Time:        time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC),
		TimeFormat:  time.RFC3339Nano,
		Name:        name,
		ContentType: "text/plain",
		Content:     getSliceOfBytesAsString(content),
	}
}

func writeTestFile(t *testing.T, name string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatalf("MkdirAll(%q) вернул ошибку: %v", filepath.Dir(name), err)
	}
	if err := os.WriteFile(name, content, 0644); err != nil {
		t.Fatalf("WriteFile(%q) вернул ошибку: %v", name, err)
	}
}

func TestGetSliceOfBytesAsString(t *testing.T) {
	var empty = getSliceOfBytesAsString(nil)
	if empty == nil || empty.Len() != 0 {
		t.Fatal("getSliceOfBytesAsString(nil) вернул не пустой builder")
	}

	var (
		content = bytes.Repeat([]byte{0x01, 0x02}, 9)
		got     = getSliceOfBytesAsString(content).String()
	)
	if strings.Count(got, "0x") != len(content) {
		t.Fatalf("getSliceOfBytesAsString() вернул %q, ожидалось %d hex значений", got, len(content))
	}
	if !strings.Contains(got, "\n") {
		t.Fatalf("getSliceOfBytesAsString() не перенёс строку: %q", got)
	}
	if strings.HasSuffix(got, "\n") {
		t.Fatalf("getSliceOfBytesAsString() заканчивается переносом строки: %q", got)
	}
}

func TestGetFileContentType(t *testing.T) {
	ret, err := getFileContentType(nil)
	if err == nil {
		t.Fatal("getFileContentType(nil) вернул nil ошибку")
	}
	if ret != "" {
		t.Fatalf("getFileContentType(nil) вернул тип контента %q, ожидался пустой", ret)
	}

	ret, err = getFileContentType([]byte("hello"))
	if err != nil {
		t.Fatalf("getFileContentType() вернул ошибку: %v", err)
	}
	if ret != "text/plain; charset=utf-8" {
		t.Fatalf("getFileContentType() = %q, ожидался text/plain", ret)
	}
}

func TestGetFileContent(t *testing.T) {
	var (
		tmp     = t.TempDir()
		fn      = filepath.Join(tmp, "file.txt")
		content = []byte("file content")
	)

	writeTestFile(t, fn, content)
	got, err := getFileContent(fn)
	if err != nil {
		t.Fatalf("getFileContent() вернул ошибку: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("getFileContent() = %q, ожидалось %q", got, content)
	}

	if _, err = getFileContent(filepath.Join(tmp, "missing.txt")); err == nil {
		t.Fatal("getFileContent() для отсутствующего файла вернул nil ошибку")
	}
}

func TestGetArg(t *testing.T) {
	var (
		oldFlagSet = flag.CommandLine
		oldArgs    = os.Args
		tmp        = t.TempDir()
	)

	flag.CommandLine = flag.NewFlagSet("embedder", flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	os.Args = []string{"embedder", "--package", "resource", "--path", filepath.Join(tmp, "out")}
	t.Cleanup(func() {
		flag.CommandLine = oldFlagSet
		os.Args = oldArgs
	})
	t.Setenv("EMBEDDER_STATIC_BASE_DIR", tmp)
	t.Setenv("EMBEDDER_STATIC_RESOURCES", "group1:static,invalid,group2:assets")

	cfg, err := getArg()
	if err != nil {
		t.Fatalf("getArg() вернул ошибку: %v", err)
	}
	if cfg.Package != "resource" {
		t.Fatalf("Package = %q, ожидался resource", cfg.Package)
	}
	if cfg.Path != filepath.Join(tmp, "out") {
		t.Fatalf("Path = %q, ожидался %q", cfg.Path, filepath.Join(tmp, "out"))
	}
	if cfg.BaseDir != tmp {
		t.Fatalf("BaseDir = %q, ожидался %q", cfg.BaseDir, tmp)
	}
	if len(cfg.Source) != 2 || cfg.Source[0] != "group1" || cfg.Source[1] != "group2" {
		t.Fatalf("Source = %v, ожидались group1 и group2", cfg.Source)
	}
	if cfg.Sources["group1"].Path != path.Join(tmp, "static") {
		t.Fatalf("Sources[group1].Path = %q, ожидался %q", cfg.Sources["group1"].Path, path.Join(tmp, "static"))
	}
	if cfg.Sources["group2"].Path != path.Join(tmp, "assets") {
		t.Fatalf("Sources[group2].Path = %q, ожидался %q", cfg.Sources["group2"].Path, path.Join(tmp, "assets"))
	}
}

func TestGetSourceInfo(t *testing.T) {
	var (
		tmp = t.TempDir()
		one = []byte("hello")
		two = []byte("world")
	)

	writeTestFile(t, filepath.Join(tmp, "one.txt"), one)
	writeTestFile(t, filepath.Join(tmp, "sub", "two.txt"), two)
	writeTestFile(t, filepath.Join(tmp, "empty.txt"), nil)

	infos, err := getSourceInfo(tmp, "")
	if err != nil {
		t.Fatalf("getSourceInfo() вернул ошибку: %v", err)
	}
	if len(infos) != 2 {
		t.Fatalf("getSourceInfo() вернул %d ресурсов, ожидалось 2: %+v", len(infos), infos)
	}
	found := make(map[string]resourceInfo)
	for _, info := range infos {
		found[info.Name] = info
	}
	for name, want := range map[string][]byte{"one.txt": one, path.Join("sub", "two.txt"): two} {
		info, ok := found[name]
		if !ok {
			t.Fatalf("ресурс %q не найден в %+v", name, found)
		}
		if info.Size != uint64(len(want)) {
			t.Fatalf("Size ресурса %q = %d, ожидалось %d", name, info.Size, len(want))
		}
		if info.ContentType == "" {
			t.Fatalf("ContentType ресурса %q пустой", name)
		}
		if info.Content == nil || !strings.Contains(info.Content.String(), "0x") {
			t.Fatalf("Content ресурса %q не преобразован в срез байт: %+v", name, info.Content)
		}
	}

	if _, err = getSourceInfo(filepath.Join(tmp, "missing"), ""); err == nil {
		t.Fatal("getSourceInfo() для отсутствующей директории вернул nil ошибку")
	}
}

func TestGetResourceList(t *testing.T) {
	var (
		tmp = t.TempDir()
		cfg = &configuration{Sources: map[string]resourceGroup{
			"assets": {Path: tmp},
		}}
	)

	writeTestFile(t, filepath.Join(tmp, "file.txt"), []byte("content"))
	if err := getResourceList(cfg); err != nil {
		t.Fatalf("getResourceList() вернул ошибку: %v", err)
	}
	if len(cfg.Sources["assets"].File) != 1 {
		t.Fatalf("getResourceList() вернул ресурсы %+v, ожидался один файл", cfg.Sources["assets"].File)
	}

	cfg = &configuration{Sources: map[string]resourceGroup{
		"assets": {Path: filepath.Join(tmp, "missing")},
	}}
	if err := getResourceList(cfg); err == nil {
		t.Fatal("getResourceList() для отсутствующей директории вернул nil ошибку")
	}
}

func TestCleanResourceContent(t *testing.T) {
	var (
		tmp       = t.TempDir()
		remove    = filepath.Join(tmp, "resource_content_test_00000000000000000001.go")
		keepGroup = filepath.Join(tmp, "resource_content_other_00000000000000000001.go")
		keepFile  = filepath.Join(tmp, "readme.txt")
		cfg       = &configuration{Path: tmp, Source: []string{"test"}}
	)

	writeTestFile(t, remove, []byte("remove"))
	writeTestFile(t, keepGroup, []byte("keep group"))
	writeTestFile(t, keepFile, []byte("keep file"))

	if err := cleanResourceContent(cfg); err != nil {
		t.Fatalf("cleanResourceContent() вернул ошибку: %v", err)
	}
	if _, err := os.Stat(remove); !os.IsNotExist(err) {
		t.Fatalf("cleanResourceContent() не удалил старый файл ресурса: %v", err)
	}
	for _, fn := range []string{keepGroup, keepFile} {
		if _, err := os.Stat(fn); err != nil {
			t.Fatalf("cleanResourceContent() удалил лишний файл %q: %v", fn, err)
		}
	}

	cfg.Path = filepath.Join(tmp, "missing")
	if err := cleanResourceContent(cfg); err == nil {
		t.Fatal("cleanResourceContent() для отсутствующей директории вернул nil ошибку")
	}
}

func TestMakeResourceContent(t *testing.T) {
	var (
		tmp  = t.TempDir()
		info = testEmbedderResourceInfo("hello.txt", []byte("hello"))
		cfg  = &configuration{
			Package: "resource",
			Path:    tmp,
			Sources: map[string]resourceGroup{
				"assets": {File: []resourceInfo{info}},
			},
		}
	)

	if err := makeResourceContent(cfg); err != nil {
		t.Fatalf("makeResourceContent() вернул ошибку: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(tmp, "resource_content_assets_00000000000000000001.go"))
	if err != nil {
		t.Fatalf("созданный файл ресурса не прочитан: %v", err)
	}
	for _, part := range []string{"package resource", "Get().Add(", `"assets"`, `"hello.txt"`, "Size: 5", "ContentType: \"text/plain\""} {
		if !strings.Contains(string(data), part) {
			t.Fatalf("созданный файл ресурса не содержит %q:\n%s", part, data)
		}
	}

	cfg = &configuration{
		Path: tmp,
		Sources: map[string]resourceGroup{
			"broken": {File: []resourceInfo{testEmbedderResourceInfo("broken.txt", []byte("broken"))}},
		},
	}
	if err = makeResourceContent(cfg); err == nil {
		t.Fatal("makeResourceContent() с пустым package вернул nil ошибку")
	}

	cfg = &configuration{
		Package: "resource",
		Path:    filepath.Join(tmp, "missing"),
		Sources: map[string]resourceGroup{
			"assets": {File: []resourceInfo{testEmbedderResourceInfo("hello.txt", []byte("hello"))}},
		},
	}
	if err = makeResourceContent(cfg); err == nil {
		t.Fatal("makeResourceContent() с отсутствующей директорией вернул nil ошибку")
	}
}

func TestSaveFile(t *testing.T) {
	var (
		tmp  = t.TempDir()
		data = []byte("package resource\n")
	)

	if err := saveFile(tmp, "group", 7, data); err != nil {
		t.Fatalf("saveFile() вернул ошибку: %v", err)
	}
	fn := filepath.Join(tmp, fmt.Sprintf("resource_content_%s_%020d.go", "group", 7))
	got, err := os.ReadFile(fn)
	if err != nil {
		t.Fatalf("saveFile() не создал файл: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("saveFile() записал %q, ожидалось %q", got, data)
	}

	if err = saveFile(filepath.Join(tmp, "missing"), "group", 8, data); err == nil {
		t.Fatal("saveFile() в отсутствующую директорию вернул nil ошибку")
	}
}

func TestGetFileContentReadError(t *testing.T) {
	if _, err := getFileContent(t.TempDir()); err == nil {
		t.Fatal("getFileContent() для директории вернул nil ошибку")
	}
}

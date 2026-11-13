package emt

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
	"testing"
)

type testZipEntry struct {
	Name    string
	Content string
	Dir     bool
}

func makeZip(t *testing.T, entries []testZipEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, entry := range entries {
		if entry.Dir {
			if _, err := zw.Create(entry.Name); err != nil {
				t.Fatalf("zw.Create(%q) вернул ошибку: %v", entry.Name, err)
			}
			continue
		}
		w, err := zw.Create(entry.Name)
		if err != nil {
			t.Fatalf("zw.Create(%q) вернул ошибку: %v", entry.Name, err)
		}
		if _, err = w.Write([]byte(entry.Content)); err != nil {
			t.Fatalf("w.Write(%q) вернул ошибку: %v", entry.Name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zw.Close() вернул ошибку: %v", err)
	}

	return buf.Bytes()
}

type testErrorReader struct{ err error }

func (ter *testErrorReader) Read([]byte) (int, error) { return 0, ter.err }

func TestParseZipErrors(t *testing.T) {
	emt := New()
	if _, err := emt.ParseZip(&testErrorReader{err: errors.New("read error")}); err == nil ||
		!strings.Contains(err.Error(), "загрузка файла") {
		t.Fatalf("ParseZip(read error) = %v", err)
	}
	if _, err := emt.ParseZip(bytes.NewReader([]byte("not a zip"))); err == nil ||
		!strings.Contains(err.Error(), "открытие ZIP") {
		t.Fatalf("ParseZip(not zip) = %v", err)
	}
}

func TestParseZipSuccessAndDecomposition(t *testing.T) {
	data := makeZip(t, []testZipEntry{
		{Name: "emptydir/", Dir: true},
		{Name: "index.html", Content: `<html><body><img src="images/logo.png"><img src="https://example.com/a.png"><img src="data:image/png;base64,QUJD"></body></html>`},
		{Name: "images/logo.png", Content: "fake-png"},
		{Name: "attach.bin", Content: "attachment"},
		{Name: "docs/readme.md", Content: "# readme"},
		{Name: "file.unknownext", Content: "unknown"},
		{Name: "__MACOSX/ignore.txt", Content: "macos"},
		{Name: "index.json", Content: `[{"key":"One","value":"1"}]`},
	})
	tpl, err := New().ParseZip(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseZip() вернул ошибку: %v", err)
	}
	if len(tpl.Body) != 1 || !strings.HasPrefix(tpl.Body[0].Type, "text/html") {
		t.Fatalf("ParseZip() Body = %#v", tpl.Body)
	}
	if len(tpl.Tags) != 1 || tpl.Tags[0].Key != "One" || tpl.Tags[0].Value != "1" {
		t.Fatalf("ParseZip() Tags = %#v", tpl.Tags)
	}
	if len(tpl.Embedded) != 2 {
		t.Fatalf("ParseZip() Embedded = %#v", tpl.Embedded)
	}
	var foundLogo, foundData bool
	for n := range tpl.Embedded {
		switch {
		case tpl.Embedded[n].Name == "images/logo.png":
			foundLogo = true
		case strings.HasPrefix(tpl.Embedded[n].Name, "data:image/png;base64,"):
			foundData = true
			if string(tpl.Embedded[n].ValueGet()) != "ABC" {
				t.Fatalf("data-uri ValueGet() = %q", tpl.Embedded[n].ValueGet())
			}
		}
	}
	if !foundLogo || !foundData {
		t.Fatalf("ParseZip() не извлёк ожидаемые вложения: %#v", tpl.Embedded)
	}
	if len(tpl.BodyImgUri) != 1 || tpl.BodyImgUri[0].String() != "https://example.com/a.png" {
		t.Fatalf("ParseZip() BodyImgUri = %#v", tpl.BodyImgUri)
	}
	if len(tpl.Attach) != 3 {
		t.Fatalf("ParseZip() Attach = %#v", tpl.Attach)
	}
	for n := range tpl.Attach {
		if strings.Contains(tpl.Attach[n].Name, "__MACOSX") {
			t.Fatalf("ParseZip() не пропустил __MACOSX: %#v", tpl.Attach[n])
		}
		if tpl.Attach[n].Name == "file.unknownext" && tpl.Attach[n].Type != "application/octet-stream" {
			t.Fatalf("ContentType fallback для unknownext = %q", tpl.Attach[n].Type)
		}
	}
}

func TestParseZipInvalidTags(t *testing.T) {
	data := makeZip(t, []testZipEntry{
		{Name: "index.json", Content: `{"broken":`},
	})
	if _, err := New().ParseZip(bytes.NewReader(data)); err == nil ||
		!strings.Contains(err.Error(), "index.json") {
		t.Fatalf("ParseZip(invalid tags) = %v", err)
	}
}

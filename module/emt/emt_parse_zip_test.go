package emt

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type testZipEntry struct {
	Name    string
	Content string
	Dir     bool
}

func makeZip(t *testing.T, entries []testZipEntry) []byte {
	const (
		tplFatalCreate = "zw.Create(%q) вернул ошибку: %v."
		tplFatalWrite  = "w.Write(%q) вернул ошибку: %v."
		tplFatalClose  = "zw.Close() вернул ошибку: %v."
	)
	var (
		buf   bytes.Buffer
		zw    *zip.Writer
		w     io.Writer
		entry testZipEntry
		err   error
	)

	t.Helper()
	zw = zip.NewWriter(&buf)
	for _, entry = range entries {
		if entry.Dir {
			if _, err = zw.Create(entry.Name); err != nil {
				t.Fatalf(tplFatalCreate, entry.Name, err)
			}
			continue
		}
		if w, err = zw.Create(entry.Name); err != nil {
			t.Fatalf(tplFatalCreate, entry.Name, err)
		}
		if _, err = w.Write([]byte(entry.Content)); err != nil {
			t.Fatalf(tplFatalWrite, entry.Name, err)
		}
	}
	if err = zw.Close(); err != nil {
		t.Fatalf(tplFatalClose, err)
	}

	return buf.Bytes()
}

type testErrorReader struct{ err error }

// Read Реализация интерфейса io.Reader.
func (ter *testErrorReader) Read([]byte) (int, error) { return 0, ter.err }

func TestParseZipErrors(t *testing.T) {
	const (
		errTextReadError   = "read error"
		tplParseZipReadErr = "ParseZip(read error) = %v."
		tplParseZipNotZip  = "ParseZip(not zip) = %v."
	)
	var (
		emt Interface
		err error
	)

	emt = New()
	if _, err = emt.ParseZip(&testErrorReader{err: errors.New(errTextReadError)}); err == nil ||
		!strings.Contains(err.Error(), "загрузка файла") {
		t.Fatalf(tplParseZipReadErr, err)
	}
	if _, err = emt.ParseZip(bytes.NewReader([]byte("not a zip"))); err == nil ||
		!strings.Contains(err.Error(), "открытие ZIP") {
		t.Fatalf(tplParseZipNotZip, err)
	}
}

func TestParseZipSuccessAndDecomposition(t *testing.T) {
	const (
		tplFatalParseZip       = "ParseZip() вернул ошибку: %v."
		tplBodyMismatch        = "ParseZip() Body = %#v."
		tplTagsMismatch        = "ParseZip() Tags = %#v."
		tplEmbeddedMismatch    = "ParseZip() Embedded = %#v."
		tplDataUriValue        = "data-uri ValueGet() = %q."
		tplEmbeddedNotFound    = "ParseZip() не извлёк ожидаемые вложения: %#v."
		tplBodyImgUriMismatch  = "ParseZip() BodyImgUri = %#v."
		tplAttachMismatch      = "ParseZip() Attach = %#v."
		tplMacOSXNotSkipped    = "ParseZip() не пропустил __MACOSX: %#v."
		tplContentTypeFallback = "ContentType fallback для unknownext = %q."
	)
	var (
		data      []byte
		tpl       *Template
		foundLogo bool
		foundData bool
		n         int
		err       error
	)

	data = makeZip(t, []testZipEntry{
		{Name: "emptydir/", Dir: true},
		{Name: "index.html", Content: `<html><body><img src="images/logo.png"><img src="https://example.com/a.png"><img src="data:image/png;base64,QUJD"></body></html>`},
		{Name: "images/logo.png", Content: "fake-png"},
		{Name: "attach.bin", Content: "attachment"},
		{Name: "docs/readme.md", Content: "# readme"},
		{Name: "file.unknownext", Content: "unknown"},
		{Name: "__MACOSX/ignore.txt", Content: "macos"},
		{Name: "index.json", Content: `[{"key":"One","value":"1"}]`},
	})
	tpl, err = New().ParseZip(bytes.NewReader(data))
	if err != nil {
		t.Fatalf(tplFatalParseZip, err)
	}
	if len(tpl.Body) != 1 || !strings.HasPrefix(tpl.Body[0].Type, "text/html") {
		t.Fatalf(tplBodyMismatch, tpl.Body)
	}
	if len(tpl.Tags) != 1 || tpl.Tags[0].Key != "One" || tpl.Tags[0].Value != "1" {
		t.Fatalf(tplTagsMismatch, tpl.Tags)
	}
	if len(tpl.Embedded) != 2 {
		t.Fatalf(tplEmbeddedMismatch, tpl.Embedded)
	}
	for n = range tpl.Embedded {
		switch {
		case tpl.Embedded[n].Name == "images/logo.png":
			foundLogo = true
		case strings.HasPrefix(tpl.Embedded[n].Name, "data:image/png;base64,"):
			foundData = true
			if string(tpl.Embedded[n].ValueGet()) != "ABC" {
				t.Fatalf(tplDataUriValue, tpl.Embedded[n].ValueGet())
			}
		}
	}
	if !foundLogo || !foundData {
		t.Fatalf(tplEmbeddedNotFound, tpl.Embedded)
	}
	if len(tpl.BodyImgUri) != 1 || tpl.BodyImgUri[0].String() != "https://example.com/a.png" {
		t.Fatalf(tplBodyImgUriMismatch, tpl.BodyImgUri)
	}
	if len(tpl.Attach) != 3 {
		t.Fatalf(tplAttachMismatch, tpl.Attach)
	}
	for n = range tpl.Attach {
		if strings.Contains(tpl.Attach[n].Name, "__MACOSX") {
			t.Fatalf(tplMacOSXNotSkipped, tpl.Attach[n])
		}
		if tpl.Attach[n].Name == "file.unknownext" && tpl.Attach[n].Type != "application/octet-stream" {
			t.Fatalf(tplContentTypeFallback, tpl.Attach[n].Type)
		}
	}
}

func TestParseZipInvalidTags(t *testing.T) {
	const (
		tplParseZipInvalidTags = "ParseZip(invalid tags) = %v."
	)
	var (
		data []byte
		err  error
	)

	data = makeZip(t, []testZipEntry{
		{Name: "index.json", Content: `{"broken":`},
	})
	if _, err = New().ParseZip(bytes.NewReader(data)); err == nil ||
		!strings.Contains(err.Error(), "index.json") {
		t.Fatalf(tplParseZipInvalidTags, err)
	}
}

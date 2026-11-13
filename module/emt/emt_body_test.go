package emt

import (
	"bytes"
	"github.com/webnice/dic"
	"net/url"
	"strings"
	"testing"
)

func parseMustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse(%q) вернул ошибку: %v", raw, err)
	}

	return u
}

func TestBodyImageUri(t *testing.T) {
	emt := New()
	tpl := &Template{
		Body:       []Body{{Type: "text/html", Value: encodeBase64([]byte(`<img src="https://example.com/a.png">`))}},
		BodyImgUri: []*url.URL{parseMustURL(t, "https://example.com/a.png")},
	}
	if err := emt.BodyImageUri(tpl, func(_ dic.IMime, _ *url.URL) string { return "cid:image1" }); err != nil {
		t.Fatalf("BodyImageUri() вернул ошибку: %v", err)
	}
	if body := string(tpl.Body[0].ValueGet()); !strings.Contains(body, "cid:image1") {
		t.Fatalf("BodyImageUri() body = %q", body)
	}

	tpl.Body[0].ValueSet([]byte(`<img src="https://example.com/a.png">`))
	if err := emt.BodyImageUri(tpl, func(_ dic.IMime, _ *url.URL) string { return "" }); err != nil {
		t.Fatalf("BodyImageUri(empty) вернул ошибку: %v", err)
	}
	if body := string(tpl.Body[0].ValueGet()); !strings.Contains(body, "https://example.com/a.png") {
		t.Fatalf("BodyImageUri(empty) изменил body: %q", body)
	}

	tpl.Body[0].ValueSet([]byte(`<img src="https://example.com/a.png">`))
	if err := emt.BodyImageUri(tpl, func(_ dic.IMime, _ *url.URL) string { panic("boom") }); err != nil {
		t.Fatalf("BodyImageUri(panic) вернул ошибку: %v", err)
	}
	if body := string(tpl.Body[0].ValueGet()); !strings.Contains(body, "https://example.com/a.png") {
		t.Fatalf("BodyImageUri(panic) изменил body: %q", body)
	}
}

func TestBodyImageEmbed(t *testing.T) {
	emt := New()
	tpl := &Template{
		Body: []Body{{Type: "text/html", Value: encodeBase64([]byte(`<img src="images/logo.png">`))}},
		Embedded: []File{
			{Type: "image/png", Name: "images/logo.png", Value: encodeBase64([]byte("PNG"))},
		},
	}
	if err := emt.BodyImageEmbed(tpl, func(_ dic.IMime, _ *bytes.Buffer) string { return "cid:logo" }); err != nil {
		t.Fatalf("BodyImageEmbed() вернул ошибку: %v", err)
	}
	if body := string(tpl.Body[0].ValueGet()); !strings.Contains(body, "cid:logo") {
		t.Fatalf("BodyImageEmbed() body = %q", body)
	}

	tpl.Body[0].ValueSet([]byte(`<img src="images/logo.png">`))
	if err := emt.BodyImageEmbed(tpl, func(_ dic.IMime, _ *bytes.Buffer) string { panic("boom") }); err != nil {
		t.Fatalf("BodyImageEmbed(panic) вернул ошибку: %v", err)
	}
	if body := string(tpl.Body[0].ValueGet()); !strings.Contains(body, "images/logo.png") {
		t.Fatalf("BodyImageEmbed(panic) изменил body: %q", body)
	}
}

func TestBodyImageUriSkipsUnknownMimeAndEmptyPath(t *testing.T) {
	emt := New()
	tpl := &Template{
		Body: []Body{{Type: "text/html", Value: encodeBase64([]byte(
			`<img src="https://example.com/a.unknownext"><img src="https://example.com">`))}},
		BodyImgUri: []*url.URL{
			parseMustURL(t, "https://example.com/a.unknownext"),
			parseMustURL(t, "https://example.com"),
		},
	}
	called := false
	if err := emt.BodyImageUri(tpl, func(_ dic.IMime, _ *url.URL) string {
		called = true
		return "cid:should-not-be-used"
	}); err != nil {
		t.Fatalf("BodyImageUri() вернул ошибку: %v", err)
	}
	if called {
		t.Fatal("BodyImageUri() вызвал fn для пустого пути или неизвестного MIME типа")
	}
	if body := string(tpl.Body[0].ValueGet()); !strings.Contains(body, "a.unknownext") {
		t.Fatalf("BodyImageUri() изменил body: %q", body)
	}
}

func TestBodyImageEmbedSkipsOctetStream(t *testing.T) {
	emt := New()
	tpl := &Template{
		Body: []Body{{Type: "text/html", Value: encodeBase64([]byte(`<img src="blob.bin">`))}},
		Embedded: []File{
			{Type: "application/octet-stream", Name: "blob.bin", Value: encodeBase64([]byte("BIN"))},
		},
	}
	called := false
	if err := emt.BodyImageEmbed(tpl, func(_ dic.IMime, _ *bytes.Buffer) string {
		called = true
		return "cid:should-not-be-used"
	}); err != nil {
		t.Fatalf("BodyImageEmbed() вернул ошибку: %v", err)
	}
	if called {
		t.Fatal("BodyImageEmbed() вызвал fn для application/octet-stream")
	}
}

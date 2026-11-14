package emt

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	"github.com/webnice/dic"
)

func parseMustURL(t *testing.T, raw string) *url.URL {
	const tplFatalParse = "url.Parse(%q) вернул ошибку: %v."

	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf(tplFatalParse, raw, err)
	}

	return u
}

func TestBodyImageUri(t *testing.T) {
	const (
		tplFatalBodyImageUri      = "BodyImageUri() вернул ошибку: %v."
		tplBodyMismatch           = "BodyImageUri() body = %q."
		tplFatalBodyImageUriEmpty = "BodyImageUri(empty) вернул ошибку: %v."
		tplBodyChangedEmpty       = "BodyImageUri(empty) изменил body: %q."
		tplFatalBodyImageUriPanic = "BodyImageUri(panic) вернул ошибку: %v."
		tplBodyChangedPanic       = "BodyImageUri(panic) изменил body: %q."
		cidImage                  = "cid:image1"
	)
	var (
		emt  Interface
		tpl  *Template
		err  error
		body string
	)

	emt = New()
	tpl = &Template{
		Body:       []Body{{Type: "text/html", Value: encodeBase64([]byte(`<img src="https://example.com/a.png">`))}},
		BodyImgUri: []*url.URL{parseMustURL(t, "https://example.com/a.png")},
	}
	if err = emt.BodyImageUri(tpl, func(_ dic.IMime, _ *url.URL) string { return cidImage }); err != nil {
		t.Fatalf(tplFatalBodyImageUri, err)
	}
	body = string(tpl.Body[0].ValueGet())
	if !strings.Contains(body, cidImage) {
		t.Fatalf(tplBodyMismatch, body)
	}
	tpl.Body[0].ValueSet([]byte(`<img src="https://example.com/a.png">`))
	if err = emt.BodyImageUri(tpl, func(_ dic.IMime, _ *url.URL) string { return "" }); err != nil {
		t.Fatalf(tplFatalBodyImageUriEmpty, err)
	}
	body = string(tpl.Body[0].ValueGet())
	if !strings.Contains(body, "https://example.com/a.png") {
		t.Fatalf(tplBodyChangedEmpty, body)
	}
	tpl.Body[0].ValueSet([]byte(`<img src="https://example.com/a.png">`))
	if err = emt.BodyImageUri(tpl, func(_ dic.IMime, _ *url.URL) string { panic("boom") }); err != nil {
		t.Fatalf(tplFatalBodyImageUriPanic, err)
	}
	body = string(tpl.Body[0].ValueGet())
	if !strings.Contains(body, "https://example.com/a.png") {
		t.Fatalf(tplBodyChangedPanic, body)
	}
}

func TestBodyImageEmbed(t *testing.T) {
	const (
		tplFatalBodyImageEmbed      = "BodyImageEmbed() вернул ошибку: %v."
		tplBodyMismatch             = "BodyImageEmbed() body = %q."
		tplFatalBodyImageEmbedPanic = "BodyImageEmbed(panic) вернул ошибку: %v."
		tplBodyChangedPanic         = "BodyImageEmbed(panic) изменил body: %q."
		cidLogo                     = "cid:logo"
	)
	var (
		emt  Interface
		tpl  *Template
		err  error
		body string
	)

	emt = New()
	tpl = &Template{
		Body: []Body{{Type: "text/html", Value: encodeBase64([]byte(`<img src="images/logo.png">`))}},
		Embedded: []File{
			{Type: "image/png", Name: "images/logo.png", Value: encodeBase64([]byte("PNG"))},
		},
	}
	if err = emt.BodyImageEmbed(tpl, func(_ dic.IMime, _ *bytes.Buffer) string { return cidLogo }); err != nil {
		t.Fatalf(tplFatalBodyImageEmbed, err)
	}
	body = string(tpl.Body[0].ValueGet())
	if !strings.Contains(body, cidLogo) {
		t.Fatalf(tplBodyMismatch, body)
	}
	tpl.Body[0].ValueSet([]byte(`<img src="images/logo.png">`))
	if err = emt.BodyImageEmbed(tpl, func(_ dic.IMime, _ *bytes.Buffer) string { panic("boom") }); err != nil {
		t.Fatalf(tplFatalBodyImageEmbedPanic, err)
	}
	body = string(tpl.Body[0].ValueGet())
	if !strings.Contains(body, "images/logo.png") {
		t.Fatalf(tplBodyChangedPanic, body)
	}
}

func TestBodyImageUriSkipsUnknownMimeAndEmptyPath(t *testing.T) {
	const (
		tplFatalBodyImageUri = "BodyImageUri() вернул ошибку: %v."
		errFnCalled          = "BodyImageUri() вызвал fn для пустого пути или неизвестного MIME типа."
		tplBodyChanged       = "BodyImageUri() изменил body: %q."
		cidNotUsed           = "cid:should-not-be-used"
	)
	var (
		emt    Interface
		tpl    *Template
		called bool
		err    error
		body   string
	)

	emt = New()
	tpl = &Template{
		Body: []Body{{Type: "text/html", Value: encodeBase64([]byte(
			`<img src="https://example.com/a.unknownext"><img src="https://example.com">`))}},
		BodyImgUri: []*url.URL{
			parseMustURL(t, "https://example.com/a.unknownext"),
			parseMustURL(t, "https://example.com"),
		},
	}
	if err = emt.BodyImageUri(tpl, func(_ dic.IMime, _ *url.URL) string {
		called = true

		return cidNotUsed
	}); err != nil {
		t.Fatalf(tplFatalBodyImageUri, err)
	}
	if called {
		t.Fatal(errFnCalled)
	}
	body = string(tpl.Body[0].ValueGet())
	if !strings.Contains(body, "a.unknownext") {
		t.Fatalf(tplBodyChanged, body)
	}
}

func TestBodyImageEmbedSkipsOctetStream(t *testing.T) {
	const (
		tplFatalBodyImageEmbed = "BodyImageEmbed() вернул ошибку: %v."
		errFnCalled            = "BodyImageEmbed() вызвал fn для application/octet-stream."
		cidNotUsed             = "cid:should-not-be-used"
	)
	var (
		emt    Interface
		tpl    *Template
		called bool
		err    error
	)

	emt = New()
	tpl = &Template{
		Body: []Body{{Type: "text/html", Value: encodeBase64([]byte(`<img src="blob.bin">`))}},
		Embedded: []File{
			{Type: "application/octet-stream", Name: "blob.bin", Value: encodeBase64([]byte("BIN"))},
		},
	}
	if err = emt.BodyImageEmbed(tpl, func(_ dic.IMime, _ *bytes.Buffer) string {
		called = true

		return cidNotUsed
	}); err != nil {
		t.Fatalf(tplFatalBodyImageEmbed, err)
	}
	if called {
		t.Fatal(errFnCalled)
	}
}

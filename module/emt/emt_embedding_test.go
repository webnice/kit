package emt

import (
	"strings"
	"testing"
)

func TestEmbedding(t *testing.T) {
	const (
		tplFatalEmbedding  = "Embedding() вернул ошибку: %v."
		tplBodyMismatch    = "Embedding() body = %q."
		tplBodyInvalidMime = "Embedding() заменил MIME с некорректным типом: %q."
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
			{Type: "invalid-mime", Name: "images/skip.png", Value: encodeBase64([]byte("SKIP"))},
		},
	}
	if err = emt.Embedding(tpl); err != nil {
		t.Fatalf(tplFatalEmbedding, err)
	}
	body = string(tpl.Body[0].ValueGet())
	if !strings.Contains(body, `src="data:image/png;base64,UE5H"`) {
		t.Fatalf(tplBodyMismatch, body)
	}
	if strings.Contains(body, "images/skip.png") {
		t.Fatalf(tplBodyInvalidMime, body)
	}
}

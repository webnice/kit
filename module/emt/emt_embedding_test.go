package emt

import (
	"strings"
	"testing"
)

func TestEmbedding(t *testing.T) {
	emt := New()
	tpl := &Template{
		Body: []Body{{Type: "text/html", Value: encodeBase64([]byte(`<img src="images/logo.png">`))}},
		Embedded: []File{
			{Type: "image/png", Name: "images/logo.png", Value: encodeBase64([]byte("PNG"))},
			{Type: "invalid-mime", Name: "images/skip.png", Value: encodeBase64([]byte("SKIP"))},
		},
	}
	if err := emt.Embedding(tpl); err != nil {
		t.Fatalf("Embedding() вернул ошибку: %v", err)
	}
	body := string(tpl.Body[0].ValueGet())
	if !strings.Contains(body, `src="data:image/png;base64,UE5H"`) {
		t.Fatalf("Embedding() body = %q", body)
	}
	if strings.Contains(body, "images/skip.png") {
		t.Fatalf("Embedding() заменил MIME с некорректным типом: %q", body)
	}
}

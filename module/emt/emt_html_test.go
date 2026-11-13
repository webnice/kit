package emt

import (
	"strings"
	"testing"
)

func TestFindImageUriFromHTMLSkipsInvalidURL(t *testing.T) {
	uries, err := New().(*impl).findImageUriFromHTML(strings.NewReader(`<img src="bad%url"><img src="plain.png">`))
	if err != nil {
		t.Fatalf("findImageUriFromHTML() вернул ошибку: %v", err)
	}
	if len(uries) != 0 {
		t.Fatalf("findImageUriFromHTML() = %#v", uries)
	}
}

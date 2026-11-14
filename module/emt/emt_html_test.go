package emt

import (
	"net/url"
	"strings"
	"testing"
)

func TestFindImageUriFromHTMLSkipsInvalidURL(t *testing.T) {
	const (
		tplFatalFindUri = "findImageUriFromHTML() вернул ошибку: %v."
		tplUries        = "findImageUriFromHTML() = %#v."
	)
	var (
		uries []*url.URL
		err   error
	)

	uries, err = New().(*impl).findImageUriFromHTML(strings.NewReader(`<img src="bad%url"><img src="plain.png">`))
	if err != nil {
		t.Fatalf(tplFatalFindUri, err)
	}
	if len(uries) != 0 {
		t.Fatalf(tplUries, uries)
	}
}

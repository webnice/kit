package emt

import (
	"strings"
	"testing"
)

func TestUtilities(t *testing.T) {
	if encodeBase64(nil) != "" || string(decodeBase64("")) != "" {
		t.Fatal("encodeBase64/decodeBase64 для пустого значения некорректны")
	}
	if encodeBase64([]byte("abc")) != "YWJj" || string(decodeBase64("YWJj")) != "abc" {
		t.Fatal("encodeBase64/decodeBase64 некорректны")
	}
	if got := decodeBase64("not-base64"); len(got) != 0 {
		t.Fatalf("decodeBase64(invalid) = %q", got)
	}

	emt := New().(*impl)
	if err := emt.safeCall(func() {}); err != nil {
		t.Fatalf("safeCall() без паники вернул ошибку: %v", err)
	}
	if err := emt.safeCall(func() { panic("boom") }); err == nil || !strings.Contains(err.Error(), "Паника") {
		t.Fatalf("safeCall() с паникой вернул %v", err)
	}

	tpl := &Template{Body: []Body{{Type: "text/plain", Value: encodeBase64([]byte("hello {name} / {name}"))}}}
	emt.bodyReplace(tpl, []*replaceContent{{Old: "{name}", New: "world"}})
	if string(tpl.Body[0].ValueGet()) != "hello world / world" {
		t.Fatalf("bodyReplace() = %q", tpl.Body[0].ValueGet())
	}
}

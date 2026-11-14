package emt

import (
	"strings"
	"testing"
)

func TestUtilities(t *testing.T) {
	const (
		errBase64Empty   = "encodeBase64/decodeBase64 для пустого значения некорректны."
		errBase64        = "encodeBase64/decodeBase64 некорректны."
		tplDecodeInvalid = "decodeBase64(invalid) = %q."
		tplFatalSafeCall = "safeCall() без паники вернул ошибку: %v."
		tplSafeCallPanic = "safeCall() с паникой вернул %v."
		tplBodyReplace   = "bodyReplace() = %q."
	)
	var (
		emt *impl
		tpl *Template
		err error
	)

	if encodeBase64(nil) != "" || string(decodeBase64("")) != "" {
		t.Fatal(errBase64Empty)
	}
	if encodeBase64([]byte("abc")) != "YWJj" || string(decodeBase64("YWJj")) != "abc" {
		t.Fatal(errBase64)
	}
	if got := decodeBase64("not-base64"); len(got) != 0 {
		t.Fatalf(tplDecodeInvalid, got)
	}
	emt = New().(*impl)
	if err = emt.safeCall(func() {}); err != nil {
		t.Fatalf(tplFatalSafeCall, err)
	}
	if err = emt.safeCall(func() { panic("boom") }); err == nil || !strings.Contains(err.Error(), "Паника") {
		t.Fatalf(tplSafeCallPanic, err)
	}
	tpl = &Template{Body: []Body{{Type: "text/plain", Value: encodeBase64([]byte("hello {name} / {name}"))}}}
	emt.bodyReplace(tpl, []*replaceContent{{Old: "{name}", New: "world"}})
	if string(tpl.Body[0].ValueGet()) != "hello world / world" {
		t.Fatalf(tplBodyReplace, tpl.Body[0].ValueGet())
	}
}

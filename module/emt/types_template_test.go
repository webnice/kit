package emt

import (
	"testing"
)

func TestBodyAndFileValues(t *testing.T) {
	const (
		tplBodyIncorrect    = "Body ValueSet/ValueGet некорректны: %#v."
		tplBodyNilIncorrect = "Body ValueSet(nil) некорректен: %#v."
		tplFileIncorrect    = "File ValueSet/ValueGet некорректны: %#v."
		tplJsonUnmarshal    = "JsonUnmarshal() = %#v, %v."
		tplJsonMarshal      = "JsonMarshal() вернул ошибку: %v."
		tplFilenameExt      = "Filename/Extension = %q, %q."
		tplExtHidden        = "FilenameExtension(.hidden) = %q."
	)
	var (
		body    Body
		file    File
		decoded map[string]int
		err     error
	)

	body.ValueSet([]byte("hello"))
	if string(body.ValueGet()) != "hello" || body.Value == "" {
		t.Fatalf(tplBodyIncorrect, body)
	}
	body.ValueSet(nil)
	if len(body.ValueGet()) != 0 || body.Value != "" {
		t.Fatalf(tplBodyNilIncorrect, body)
	}
	file.ValueSet([]byte(`{"a":1}`))
	if string(file.ValueGet()) != `{"a":1}` {
		t.Fatalf(tplFileIncorrect, file)
	}
	if err = file.JsonUnmarshal(&decoded); err != nil || decoded["a"] != 1 {
		t.Fatalf(tplJsonUnmarshal, decoded, err)
	}
	if err = file.JsonMarshal(map[string]int{"b": 2}); err != nil {
		t.Fatalf(tplJsonMarshal, err)
	}
	if file.Name = "dir/sub/archive.tar.gz"; file.Filename() != "archive.tar.gz" || file.FilenameExtension() != "gz" {
		t.Fatalf(tplFilenameExt, file.Filename(), file.FilenameExtension())
	}
	if file.Name = ".hidden"; file.FilenameExtension() != "" {
		t.Fatalf(tplExtHidden, file.FilenameExtension())
	}
}

func TestFileJsonMarshalError(t *testing.T) {
	const errNoError = "JsonMarshal() с неподдерживаемым типом не вернул ошибку."
	var file File

	if err := file.JsonMarshal(make(chan int)); err == nil {
		t.Fatal(errNoError)
	}
}

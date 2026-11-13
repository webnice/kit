package emt

import (
	"testing"
)

func TestBodyAndFileValues(t *testing.T) {
	var body Body
	body.ValueSet([]byte("hello"))
	if string(body.ValueGet()) != "hello" || body.Value == "" {
		t.Fatalf("Body ValueSet/ValueGet некорректны: %#v", body)
	}
	body.ValueSet(nil)
	if len(body.ValueGet()) != 0 || body.Value != "" {
		t.Fatalf("Body ValueSet(nil) некорректен: %#v", body)
	}

	var file File
	file.ValueSet([]byte(`{"a":1}`))
	if string(file.ValueGet()) != `{"a":1}` {
		t.Fatalf("File ValueSet/ValueGet некорректны: %#v", file)
	}
	var decoded map[string]int
	if err := file.JsonUnmarshal(&decoded); err != nil || decoded["a"] != 1 {
		t.Fatalf("JsonUnmarshal() = %#v, %v", decoded, err)
	}
	if err := file.JsonMarshal(map[string]int{"b": 2}); err != nil {
		t.Fatalf("JsonMarshal() вернул ошибку: %v", err)
	}
	if file.Name = "dir/sub/archive.tar.gz"; file.Filename() != "archive.tar.gz" || file.FilenameExtension() != "gz" {
		t.Fatalf("Filename/Extension = %q, %q", file.Filename(), file.FilenameExtension())
	}
	if file.Name = ".hidden"; file.FilenameExtension() != "" {
		t.Fatalf("FilenameExtension(.hidden) = %q", file.FilenameExtension())
	}
}

func TestFileJsonMarshalError(t *testing.T) {
	var file File
	if err := file.JsonMarshal(make(chan int)); err == nil {
		t.Fatal("JsonMarshal() с неподдерживаемым типом не вернул ошибку")
	}
}

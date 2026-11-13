package uuid

import (
	"bytes"
	"testing"
)

// TestImpl_FromBytes Конвертация среза байт в UUID.
func TestImpl_FromBytes(t *testing.T) {
	orig := &uuid{data: [size]byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}}
	bytes1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	u := singleton.FromBytes(bytes1)
	if !bytes.Equal(orig.Bytes(), u.Bytes()) {
		t.Fatal("конвертация байтов в UUID некорректна")
	}
	if singleton.FromBytes(nil) != NULL {
		t.Fatal("FromBytes() с неверной длиной не вернул NULL")
	}
}

// TestImpl_FromString Конвертация строки в UUID.
func TestImpl_FromString(t *testing.T) {
	bytes1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	string1 := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	b1 := singleton.FromBytes(bytes1)
	s1 := singleton.FromString(string1)
	if !bytes.Equal(b1.Bytes(), s1.Bytes()) {
		t.Fatal("конвертация строки в UUID некорректна")
	}
	if singleton.FromString("uuid") != NULL {
		t.Fatal("FromString() с неверной строкой не вернул NULL")
	}
}

// TestUuid_MarshalUnmarshal Сериализация и десериализация UUID.
func TestUuid_MarshalUnmarshal(t *testing.T) {
	u := singleton.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8").(*uuid)

	bin, err := u.MarshalBinary()
	if err != nil || !bytes.Equal(bin, u.Bytes()) {
		t.Fatalf("MarshalBinary() = %v, %v", bin, err)
	}
	var u2 uuid
	if err = u2.UnmarshalBinary(bin); err != nil || !u2.Equal(u) {
		t.Fatalf("UnmarshalBinary() = %v, %v", u2.String(), err)
	}
	if err = u2.UnmarshalBinary(bin[:15]); err == nil {
		t.Fatal("UnmarshalBinary() с неверной длиной не вернул ошибку")
	}

	text, err := u.MarshalText()
	if err != nil || string(text) != u.String() {
		t.Fatalf("MarshalText() = %q, %v", text, err)
	}
	if err = u2.UnmarshalText(text); err != nil || !u2.Equal(u) {
		t.Fatalf("UnmarshalText() = %v, %v", u2.String(), err)
	}
}

// TestUuid_UnmarshalTextFormats Десериализация UUID из строк разных форматов.
func TestUuid_UnmarshalTextFormats(t *testing.T) {
	tests := map[string]string{
		"canonical":    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"upper":        "6BA7B810-9DAD-11D1-80B4-00C04FD430C8",
		"hashlike":     "6ba7b8109dad11d180b400c04fd430c8",
		"bracedCanon":  "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
		"bracedHash":   "{6ba7b8109dad11d180b400c04fd430c8}",
		"urnCanonical": "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"urnHash":      "urn:uuid:6ba7b8109dad11d180b400c04fd430c8",
	}
	for name, src := range tests {
		var u uuid
		if err := u.UnmarshalText([]byte(src)); err != nil {
			t.Fatalf("UnmarshalText(%s) вернул ошибку: %v", name, err)
		}
		if u.String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
			t.Fatalf("UnmarshalText(%s) = %q", name, u.String())
		}
	}

	errors := map[string]string{
		"empty":           "",
		"short":           "6ba7b810-9dad-11d1-80b4-00c04fd430c",
		"badSeparator":    "6ba7b8109dad-11d1-80b4-00c04fd430c8",
		"badSeparatorPos": "6ba7b810-9dadx11d1-80b4-00c04fd430c8",
		"badHexCanon":     "6ba7b810-9dad-11d1-80b4-00c04fd430cg",
		"badHexHash":      "6ba7b8109dad11d180b400c04fd430cg",
		"badBrace":        "6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
		"badBraces":       "x6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
		"badBraceInner":   "{6ba7b810-9dad-11d1-80b4-00c04fd430c}",
		"badURNPrefix":    "url:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"badURNInner":     "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c",
	}
	for name, src := range errors {
		var u uuid
		if err := u.UnmarshalText([]byte(src)); err == nil {
			t.Fatalf("UnmarshalText(%s) не вернул ошибку", name)
		}
	}
	var plain uuid
	if err := plain.decodePlain([]byte("short")); err == nil {
		t.Fatal("decodePlain() с неверной длиной не вернул ошибку")
	}
}

// TestUuid_JSON Сериализация UUID в JSON.
func TestUuid_JSON(t *testing.T) {
	u := singleton.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8").(*uuid)

	data, err := u.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() вернул ошибку: %v", err)
	}
	if string(data) != `"6ba7b810-9dad-11d1-80b4-00c04fd430c8"` {
		t.Fatalf("MarshalJSON() = %q", data)
	}

	var decoded uuid
	if err = decoded.UnmarshalJSON(data); err != nil || !decoded.Equal(u) {
		t.Fatalf("UnmarshalJSON() = %q, %v", decoded.String(), err)
	}
	if err = decoded.UnmarshalJSON([]byte(`123`)); err == nil {
		t.Fatal("UnmarshalJSON() с не-строкой не вернул ошибку")
	}
	if err = decoded.UnmarshalJSON([]byte(`"not-an-uuid"`)); err == nil {
		t.Fatal("UnmarshalJSON() с неверной строкой не вернул ошибку")
	}
}

package uuid

import (
	"bytes"
	"testing"
)

// Тестовое значение UUID в разных представлениях.
const testUUIDString = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

// TestImpl_FromBytes Конвертация среза байт в UUID.
func TestImpl_FromBytes(t *testing.T) {
	const (
		errConvertWrong = "конвертация байтов в UUID некорректна."
		errNoNull       = "FromBytes() с неверной длиной не вернул NULL."
	)
	var (
		orig   *uuid
		bytes1 []byte
		u      UUID
	)

	orig = &uuid{data: [size]byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}}
	bytes1 = []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	u = singleton.FromBytes(bytes1)
	if !bytes.Equal(orig.Bytes(), u.Bytes()) {
		t.Fatal(errConvertWrong)
	}
	if singleton.FromBytes(nil) != NULL {
		t.Fatal(errNoNull)
	}
}

// TestImpl_FromString Конвертация строки в UUID.
func TestImpl_FromString(t *testing.T) {
	const (
		errConvertWrong = "конвертация строки в UUID некорректна."
		errNoNull       = "FromString() с неверной строкой не вернул NULL."
	)
	var (
		bytes1  []byte
		string1 string
		b1      UUID
		s1      UUID
	)

	bytes1 = []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	string1 = testUUIDString
	b1 = singleton.FromBytes(bytes1)
	s1 = singleton.FromString(string1)
	if !bytes.Equal(b1.Bytes(), s1.Bytes()) {
		t.Fatal(errConvertWrong)
	}
	if singleton.FromString("uuid") != NULL {
		t.Fatal(errNoNull)
	}
}

// TestUuid_MarshalUnmarshal Сериализация и десериализация UUID.
func TestUuid_MarshalUnmarshal(t *testing.T) {
	const (
		tplMarshalBinWrong    = "MarshalBinary() = %v, %v."
		tplUnmarshalBinWrong  = "UnmarshalBinary() = %v, %v."
		errBinNoError         = "UnmarshalBinary() с неверной длиной не вернул ошибку."
		tplMarshalTextWrong   = "MarshalText() = %q, %v."
		tplUnmarshalTextWrong = "UnmarshalText() = %v, %v."
	)
	var (
		u    *uuid
		u2   uuid
		bin  []byte
		text []byte
		err  error
	)

	u = singleton.FromString(testUUIDString).(*uuid)
	bin, err = u.MarshalBinary()
	if err != nil || !bytes.Equal(bin, u.Bytes()) {
		t.Fatalf(tplMarshalBinWrong, bin, err)
	}
	if err = u2.UnmarshalBinary(bin); err != nil || !u2.Equal(u) {
		t.Fatalf(tplUnmarshalBinWrong, u2.String(), err)
	}
	if err = u2.UnmarshalBinary(bin[:15]); err == nil {
		t.Fatal(errBinNoError)
	}
	text, err = u.MarshalText()
	if err != nil || string(text) != u.String() {
		t.Fatalf(tplMarshalTextWrong, text, err)
	}
	if err = u2.UnmarshalText(text); err != nil || !u2.Equal(u) {
		t.Fatalf(tplUnmarshalTextWrong, u2.String(), err)
	}
}

// TestUuid_UnmarshalTextFormats Десериализация UUID из строк разных форматов.
func TestUuid_UnmarshalTextFormats(t *testing.T) {
	const (
		tplFatalUnmarshal     = "UnmarshalText(%s) вернул ошибку: %v."
		tplUnmarshalWrong     = "UnmarshalText(%s) = %q."
		errNoError            = "UnmarshalText(%s) не вернул ошибку."
		errDecodePlainNoError = "decodePlain() с неверной длиной не вернул ошибку."
	)
	var (
		name   string
		src    string
		u      uuid
		err    error
		tests  map[string]string
		errors map[string]string
	)

	tests = map[string]string{
		"canonical":    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"upper":        "6BA7B810-9DAD-11D1-80B4-00C04FD430C8",
		"hashlike":     "6ba7b8109dad11d180b400c04fd430c8",
		"bracedCanon":  "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}",
		"bracedHash":   "{6ba7b8109dad11d180b400c04fd430c8}",
		"urnCanonical": "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"urnHash":      "urn:uuid:6ba7b8109dad11d180b400c04fd430c8",
	}
	for name, src = range tests {
		if err = u.UnmarshalText([]byte(src)); err != nil {
			t.Fatalf(tplFatalUnmarshal, name, err)
		}
		if u.String() != testUUIDString {
			t.Fatalf(tplUnmarshalWrong, name, u.String())
		}
	}

	errors = map[string]string{
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
	for name, src = range errors {
		if err = u.UnmarshalText([]byte(src)); err == nil {
			t.Fatalf(errNoError, name)
		}
	}
	var plain uuid
	if err = plain.decodePlain([]byte("short")); err == nil {
		t.Fatal(errDecodePlainNoError)
	}
}

// TestUuid_JSON Сериализация UUID в JSON.
func TestUuid_JSON(t *testing.T) {
	const (
		tplFatalMarshalJSON   = "MarshalJSON() вернул ошибку: %v."
		tplMarshalJSONWrong   = "MarshalJSON() = %q."
		tplUnmarshalJSONWrong = "UnmarshalJSON() = %q, %v."
		errNotStringNoError   = "UnmarshalJSON() с не-строкой не вернул ошибку."
		errBadStringNoError   = "UnmarshalJSON() с неверной строкой не вернул ошибку."
		testJSON              = `"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`
	)
	var (
		u       *uuid
		decoded uuid
		data    []byte
		err     error
	)

	u = singleton.FromString(testUUIDString).(*uuid)
	data, err = u.MarshalJSON()
	if err != nil {
		t.Fatalf(tplFatalMarshalJSON, err)
	}
	if string(data) != testJSON {
		t.Fatalf(tplMarshalJSONWrong, data)
	}
	if err = decoded.UnmarshalJSON(data); err != nil || !decoded.Equal(u) {
		t.Fatalf(tplUnmarshalJSONWrong, decoded.String(), err)
	}
	if err = decoded.UnmarshalJSON([]byte(`123`)); err == nil {
		t.Fatal(errNotStringNoError)
	}
	if err = decoded.UnmarshalJSON([]byte(`"not-an-uuid"`)); err == nil {
		t.Fatal(errBadStringNoError)
	}
}

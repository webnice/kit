// Package uuid
package uuid

import (
	"bytes"
	"testing"
)

func TestImpl_FromBytes(t *testing.T) {
	orig := &uuid{data: [size]byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}}
	bytes1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	u := singleton.FromBytes(bytes1)
	if !bytes.Equal(orig.Bytes(), u.Bytes()) {
		t.Fatalf("конвертация байтов в UUID не корретна")
	}
}

func TestImpl_FromString(t *testing.T) {
	bytes1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	string1 := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	b1 := singleton.FromBytes(bytes1)
	s1 := singleton.FromString(string1)
	if !bytes.Equal(b1.Bytes(), s1.Bytes()) {
		t.Fatalf("конвертация строки в UUID не корретна")
	}
}

func TestUuid_Version(t *testing.T) {
	u := &uuid{data: [size]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}
	if u.Version() != Version.V1 {
		t.Errorf("функция Version не работает")
	}
	if u.Version() == Version.V2 {
		t.Errorf("функция Version не работает")
	}
}

func TestUuid_Equal(t *testing.T) {
	orig := &uuid{data: [size]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}
	bytes1 := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	string1 := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	b1 := singleton.FromBytes(bytes1)
	s1 := singleton.FromString(string1)
	if !b1.Equal(s1) {
		t.Errorf("функция Equal не работает")
	}
	if b1.Equal(orig) {
		t.Errorf("функция Equal не работает")
	}
}
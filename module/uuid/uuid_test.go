package uuid

import "testing"

// TestUuid_VersionVariantSetters Установка и чтение версии и варианта UUID.
func TestUuid_VersionVariantSetters(t *testing.T) {
	var u uuid

	versions := []VersionType{Version.V1, Version.V2, Version.V3, Version.V4, Version.V5, Version.V6, Version.V7, Version.V8}
	for n := range versions {
		u = uuid{}
		u.SetVersion(versions[n])
		if u.Version() != versions[n] {
			t.Fatalf("Version() = %v, ожидалась %v", u.Version(), versions[n])
		}
	}
	u = uuid{data: [size]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x90, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}
	if u.Version() != Version.Unknown {
		t.Fatal("Version() не вернула Unknown")
	}

	variants := []VariantType{Variant.NCS, Variant.RFC4122, Variant.Microsoft, Variant.Future}
	for n := range variants {
		u = uuid{data: [size]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}
		u.SetVariant(variants[n])
		if u.Variant() != variants[n] {
			t.Fatalf("Variant() = %v, ожидался %v", u.Variant(), variants[n])
		}
	}
	u = uuid{data: [size]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}
	u.SetVariant(VariantType(255))
	if u.Variant() != Variant.Future {
		t.Fatal("SetVariant() с неизвестным вариантом не установила Future")
	}
}

// TestUuid_String Строковое представление UUID.
func TestUuid_String(t *testing.T) {
	if NULL.String() != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("NULL.String() = %q", NULL.String())
	}
}

// TestUuid_Equal Сравнение UUID.
func TestUuid_Equal(t *testing.T) {
	orig := &uuid{data: [size]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}
	b1 := singleton.FromBytes([]byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8})
	s1 := singleton.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if !b1.Equal(s1) {
		t.Error("Equal() для одинаковых UUID вернул false")
	}
	if b1.Equal(orig) {
		t.Error("Equal() для разных UUID вернул true")
	}
}

// TestImpl_Get Глобальный объект пакета.
func TestImpl_Get(t *testing.T) {
	if Get() != singleton {
		t.Fatal("Get() вернул не singleton")
	}
}

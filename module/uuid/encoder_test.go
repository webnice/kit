package uuid

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

type errorReader struct{ err error }

func (er *errorReader) Read([]byte) (int, error) { return 0, er.err }

type byteReader struct {
	data []byte
	err  error
	off  int
}

func (br *byteReader) Read(p []byte) (int, error) {
	if br.off >= len(br.data) {
		if br.err != nil {
			return 0, br.err
		}
		return 0, io.EOF
	}
	n := copy(p, br.data[br.off:])
	br.off += n

	return n, nil
}

func newTestImpl(rand io.Reader, epoch time.Time, hw net.HardwareAddr, hwErr error) *impl {
	if rand == nil {
		rand = &byteReader{}
	}

	return &impl{
		rand: rand,
		epochFunc: func() time.Time {
			return epoch
		},
		hwAddrFunc: func() (net.HardwareAddr, error) {
			if hwErr != nil {
				return nil, hwErr
			}
			return hw, nil
		},
	}
}

func testEpoch() time.Time { return time.Unix(1700000000, 123456789).UTC() }

func testV1Data(t *testing.T, ui *impl) []byte {
	t.Helper()

	clockSeq := []byte{0x12, 0x34}
	hardwareAddr := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	ui.rand = &byteReader{data: append(append([]byte{}, clockSeq...), hardwareAddr...)}
	ui.hwAddrFunc = func() (net.HardwareAddr, error) { return hardwareAddr, nil }

	return hardwareAddr
}

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

func TestUuid_StringAndTypes(t *testing.T) {
	if NULL.String() != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("NULL.String() = %q", NULL.String())
	}
	if Namespace.DNS.String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Fatalf("Namespace.DNS.String() = %q", Namespace.DNS.String())
	}
	if Version.V7.String() != "V7" || Version.V8.String() != "V8" {
		t.Fatal("String() версии UUID некорректен")
	}
}

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

func TestImpl_DefaultHWAddrFunc(t *testing.T) {
	oldNetInterfaces := netInterfaces
	t.Cleanup(func() { netInterfaces = oldNetInterfaces })

	netInterfaces = func() ([]net.Interface, error) { return nil, errors.New("ifaces error") }
	if _, err := defaultHWAddrFunc(); err == nil {
		t.Fatal("defaultHWAddrFunc() не вернул ошибку net.Interfaces")
	}

	netInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{{HardwareAddr: net.HardwareAddr{0x01}}}, nil
	}
	if _, err := defaultHWAddrFunc(); err == nil {
		t.Fatal("defaultHWAddrFunc() не вернул errNotFound")
	}

	netInterfaces = oldNetInterfaces
	hw, err := defaultHWAddrFunc()
	if err == nil && len(hw) < 6 {
		t.Fatalf("defaultHWAddrFunc() вернул короткий адрес: %v", hw)
	}
}

func TestImpl_GetEpoch(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(&byteReader{}, epoch, nil, nil)
	if got, want := ui.getEpoch(), epochStart+uint64(epoch.UnixNano()/100); got != want {
		t.Fatalf("getEpoch() = %d, ожидалось %d", got, want)
	}
}

func TestImpl_GetHardwareAddr(t *testing.T) {
	epoch := testEpoch()
	hw := net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	ui := newTestImpl(&byteReader{}, epoch, hw, nil)
	got, err := ui.getHardwareAddr()
	if err != nil || !bytes.Equal(got, hw) {
		t.Fatalf("getHardwareAddr() = %v, %v", got, err)
	}

	ui = newTestImpl(&byteReader{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}}, epoch, nil, errors.New("no hw"))
	got, err = ui.getHardwareAddr()
	if err != nil || len(got) != 6 || got[0] != 0x01 || got[1] != 0x01 {
		t.Fatalf("getHardwareAddr() fallback = %v, %v", got, err)
	}

	ui = newTestImpl(&errorReader{err: errors.New("rand error")}, epoch, nil, errors.New("no hw"))
	if _, err = ui.getHardwareAddr(); err == nil {
		t.Fatal("getHardwareAddr() не вернул ошибку чтения случайного адреса")
	}
}

func TestImpl_GetClockSequence(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(&byteReader{data: []byte{0x12, 0x34}}, epoch, nil, nil)

	ret, seq, err := ui.getClockSequence()
	if err != nil || ret != ui.getEpoch() || seq != 0x1234 {
		t.Fatalf("getClockSequence() = %d, %x, %v", ret, seq, err)
	}
	_, seq, err = ui.getClockSequence()
	if err != nil || seq != 0x1235 {
		t.Fatalf("getClockSequence() без сдвига времени = %x, %v", seq, err)
	}

	ui = newTestImpl(&errorReader{err: errors.New("rand error")}, epoch, nil, nil)
	if _, _, err = ui.getClockSequence(); err == nil {
		t.Fatal("getClockSequence() не вернул ошибку чтения случайной последовательности")
	}
}

func TestImpl_V1(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(nil, epoch, nil, nil)
	hardwareAddr := testV1Data(t, ui)

	got, err := ui.V1()
	if err != nil {
		t.Fatalf("V1() вернул ошибку: %v", err)
	}
	expected := make([]byte, size)
	timeNow := epochStart + uint64(epoch.UnixNano()/100)
	binary.BigEndian.PutUint32(expected[0:], uint32(timeNow))
	binary.BigEndian.PutUint16(expected[4:], uint16(timeNow>>32))
	binary.BigEndian.PutUint16(expected[6:], uint16(timeNow>>48))
	binary.BigEndian.PutUint16(expected[8:], 0x1234)
	copy(expected[10:], hardwareAddr)
	expected[6] = (expected[6] & 0x0f) | (Version.V1.value << 4)
	expected[8] = (expected[8] & 0x3f) | 0x80
	if !bytes.Equal(got.Bytes(), expected) {
		t.Fatalf("V1() = %v, ожидалось %x", got.String(), expected)
	}
	if got.Version() != Version.V1 || got.Variant() != Variant.RFC4122 {
		t.Fatal("V1() установил некорректную версию или вариант")
	}

	ui = newTestImpl(&errorReader{err: errors.New("rand error")}, epoch, nil, nil)
	if got, err = ui.V1(); err == nil || got != NULL {
		t.Fatal("V1() не вернул NULL и ошибку чтения случайной последовательности")
	}

	randomNode := []byte{0x00, 0x10, 0x20, 0x30, 0x40, 0x50}
	ui = newTestImpl(&byteReader{data: append([]byte{0x12, 0x34}, randomNode...)}, epoch, nil, errors.New("no hw"))
	got, err = ui.V1()
	if err != nil || got.(*uuid).data[10] != 0x01 || got.(*uuid).data[11] != 0x10 {
		t.Fatalf("V1() fallback hardware addr = %v, %v", got.String(), err)
	}
}

func TestImpl_V2(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(nil, epoch, nil, nil)
	testV1Data(t, ui)

	got, err := ui.V2(Domain.Person)
	if err != nil || got.Version() != Version.V2 || got.Variant() != Variant.RFC4122 {
		t.Fatalf("V2(Person) = %v, %v", got.String(), err)
	}
	if binary.BigEndian.Uint32(got.Bytes()[0:4]) != posixUID || got.Bytes()[9] != byte(Domain.Person) {
		t.Fatal("V2(Person) не записал UID и домен")
	}

	got, err = ui.V2(Domain.Group)
	if err != nil || binary.BigEndian.Uint32(got.Bytes()[0:4]) != posixGID || got.Bytes()[9] != byte(Domain.Group) {
		t.Fatalf("V2(Group) = %v, %v", got.String(), err)
	}

	got, err = ui.V2(Domain.Org)
	if err != nil || got.Bytes()[9] != byte(Domain.Org) || got.Version() != Version.V2 {
		t.Fatalf("V2(Org) = %v, %v", got.String(), err)
	}

	ui = newTestImpl(&errorReader{err: errors.New("rand error")}, epoch, nil, nil)
	if got, err = ui.V2(Domain.Person); err == nil || got != NULL {
		t.Fatal("V2() не вернул NULL и ошибку")
	}
}

func TestImpl_V3V5(t *testing.T) {
	u3 := singleton.V3(Namespace.DNS, "python.org")
	if u3.String() != "6fa459ea-ee8a-3ca4-894e-db77e160355e" || u3.Version() != Version.V3 || u3.Variant() != Variant.RFC4122 {
		t.Fatalf("V3() = %s", u3.String())
	}

	u5 := singleton.V5(Namespace.DNS, "python.org")
	if u5.String() != "886313e1-3b8a-5372-9b90-0c9aee199e5d" || u5.Version() != Version.V5 || u5.Variant() != Variant.RFC4122 {
		t.Fatalf("V5() = %s", u5.String())
	}
}

func TestImpl_V4(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(&byteReader{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}}, epoch, nil, nil)

	got := ui.V4()
	if got.String() != "00010203-0405-4607-8809-0a0b0c0d0e0f" || got.Version() != Version.V4 || got.Variant() != Variant.RFC4122 {
		t.Fatalf("V4() = %s", got.String())
	}

	ui = newTestImpl(&errorReader{err: errors.New("rand error")}, epoch, nil, nil)
	if got = ui.V4(); got != NULL {
		t.Fatal("V4() не вернул NULL при ошибке чтения случайных данных")
	}
}

func TestImpl_V6(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(nil, epoch, nil, nil)
	hardwareAddr := testV1Data(t, ui)

	got := ui.V6()
	timeNow := epochStart + uint64(epoch.UnixNano()/100)
	timeHighMid := timeNow >> 12
	expected := make([]byte, size)
	binary.BigEndian.PutUint32(expected[0:], uint32(timeHighMid>>16))
	binary.BigEndian.PutUint16(expected[4:], uint16(timeHighMid&0xffff))
	binary.BigEndian.PutUint16(expected[6:], uint16(timeNow&0xfff)|0x6000)
	binary.BigEndian.PutUint16(expected[8:], 0x9234)
	copy(expected[10:], hardwareAddr)
	if !bytes.Equal(got.Bytes(), expected) {
		t.Fatalf("V6() = %v, ожидалось %x", got.String(), expected)
	}
	if got.Version() != Version.V6 || got.Variant() != Variant.RFC4122 {
		t.Fatal("V6() установил некорректную версию или вариант")
	}

	ui = newTestImpl(&errorReader{err: errors.New("rand error")}, epoch, nil, nil)
	if got = ui.V6(); got != NULL {
		t.Fatal("V6() не вернул NULL при ошибке чтения случайной последовательности")
	}
}

func TestImpl_V7(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(&byteReader{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}}, epoch, nil, nil)

	got := ui.V7()
	timestamp := uint64(epoch.UnixNano() / 1e6)
	expectedPrefix := make([]byte, 6)
	binary.BigEndian.PutUint32(expectedPrefix[0:], uint32(timestamp>>16))
	binary.BigEndian.PutUint16(expectedPrefix[4:], uint16(timestamp))
	if !bytes.Equal(got.Bytes()[:6], expectedPrefix) || got.Version() != Version.V7 || got.Variant() != Variant.RFC4122 {
		t.Fatalf("V7() = %s", got.String())
	}

	ui = newTestImpl(&errorReader{err: errors.New("rand error")}, epoch, nil, nil)
	if got = ui.V7(); got != NULL {
		t.Fatal("V7() не вернул NULL при ошибке чтения случайных данных")
	}
}

func TestImpl_V8(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(&byteReader{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}}, epoch, nil, nil)

	got := ui.V8()
	if got.String() != "00010203-0405-8607-8809-0a0b0c0d0e0f" || got.Version() != Version.V8 || got.Variant() != Variant.RFC4122 {
		t.Fatalf("V8() = %s", got.String())
	}

	ui = newTestImpl(&errorReader{err: errors.New("rand error")}, epoch, nil, nil)
	if got = ui.V8(); got != NULL {
		t.Fatal("V8() не вернул NULL при ошибке чтения случайных данных")
	}
}

func TestImpl_Get(t *testing.T) {
	if Get() != singleton {
		t.Fatal("Get() вернул не singleton")
	}
}

func TestImpl_V1V6HardwareError(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(&byteReader{data: []byte{0x12, 0x34}, err: errors.New("rand error")}, epoch, nil, errors.New("no hw"))
	if got, err := ui.V1(); err == nil || got != NULL {
		t.Fatalf("V1() при ошибке hardware address = %v, %v", got, err)
	}

	ui = newTestImpl(&byteReader{data: []byte{0x12, 0x34}, err: errors.New("rand error")}, epoch, nil, errors.New("no hw"))
	if got := ui.V6(); got != NULL {
		t.Fatalf("V6() при ошибке hardware address = %v", got)
	}
}

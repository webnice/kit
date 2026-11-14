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

// errorReader Объект io.Reader всегда возвращающий ошибку.
type errorReader struct{ err error }

func (er *errorReader) Read([]byte) (int, error) { return 0, er.err }

// byteReader Объект io.Reader читающий из среза байт.
type byteReader struct {
	data []byte
	err  error
	off  int
}

func (br *byteReader) Read(p []byte) (n int, err error) {
	if br.off >= len(br.data) {
		if br.err != nil {
			return 0, br.err
		}
		return 0, io.EOF
	}
	n = copy(p, br.data[br.off:])
	br.off += n

	return
}

// newTestImpl Конструктор тестового объекта с подменяемыми зависимостями.
func newTestImpl(rand io.Reader, epoch time.Time, hw net.HardwareAddr, hwErr error) *impl {
	if rand == nil {
		rand = &byteReader{}
	}
	// _ = rand

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

// testEpoch Тестовая эпоха.
func testEpoch() time.Time { return time.Unix(1700000000, 123456789).UTC() }

// testV1Data Подготовка тестовых данных аппаратного адреса и последовательности часов.
func testV1Data(t *testing.T, ui *impl) []byte {
	t.Helper()

	var (
		clockSeq     []byte
		hardwareAddr net.HardwareAddr
	)

	clockSeq = []byte{0x12, 0x34}
	hardwareAddr = net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	ui.rand = &byteReader{data: append(append([]byte{}, clockSeq...), hardwareAddr...)}
	ui.hwAddrFunc = func() (net.HardwareAddr, error) { return hardwareAddr, nil }

	return hardwareAddr
}

// TestImpl_V1 Генерация UUID версии 1.
func TestImpl_V1(t *testing.T) {
	const (
		tplFatalV1       = "V1() вернул ошибку: %v."
		tplV1Wrong       = "V1() = %v, ожидалось %x."
		errVersionWrong  = "V1() установил некорректную версию или вариант."
		errNoNull        = "V1() не вернул NULL и ошибку чтения случайной последовательности."
		tplV1Fallback    = "V1() fallback hardware addr = %v, %v."
		errTextRandError = "rand error."
		errTextNoHw      = "no hw."
	)
	var (
		epoch        time.Time
		ui           *impl
		hardwareAddr net.HardwareAddr
		got          UUID
		err          error
		expected     []byte
		timeNow      uint64
		randomNode   []byte
	)

	epoch = testEpoch()
	ui = newTestImpl(nil, epoch, nil, nil)
	hardwareAddr = testV1Data(t, ui)
	got, err = ui.V1()
	if err != nil {
		t.Fatalf(tplFatalV1, err)
	}
	expected = make([]byte, size)
	timeNow = epochStart + uint64(epoch.UnixNano()/100)
	binary.BigEndian.PutUint32(expected[0:], uint32(timeNow))
	binary.BigEndian.PutUint16(expected[4:], uint16(timeNow>>32))
	binary.BigEndian.PutUint16(expected[6:], uint16(timeNow>>48))
	binary.BigEndian.PutUint16(expected[8:], 0x1234)
	copy(expected[10:], hardwareAddr)
	expected[6] = (expected[6] & 0x0f) | (Version.V1.value << 4)
	expected[8] = (expected[8] & 0x3f) | 0x80
	if !bytes.Equal(got.Bytes(), expected) {
		t.Fatalf(tplV1Wrong, got.String(), expected)
	}
	if got.Version() != Version.V1 || got.Variant() != Variant.RFC4122 {
		t.Fatal(errVersionWrong)
	}
	ui = newTestImpl(&errorReader{err: errors.New(errTextRandError)}, epoch, nil, nil)
	if got, err = ui.V1(); err == nil || got != NULL {
		t.Fatal(errNoNull)
	}
	randomNode = []byte{0x00, 0x10, 0x20, 0x30, 0x40, 0x50}
	ui = newTestImpl(&byteReader{data: append([]byte{0x12, 0x34}, randomNode...)}, epoch, nil, errors.New(errTextNoHw))
	got, err = ui.V1()
	if err != nil || got.(*uuid).data[10] != 0x01 || got.(*uuid).data[11] != 0x10 {
		t.Fatalf(tplV1Fallback, got.String(), err)
	}
}

// TestImpl_V2 Генерация UUID версии 2.
func TestImpl_V2(t *testing.T) {
	const (
		tplV2Wrong       = "V2(Person) = %v, %v."
		tplV2GroupWrong  = "V2(Group) = %v, %v."
		tplV2OrgWrong    = "V2(Org) = %v, %v."
		errNoUID         = "V2(Person) не записал UID и домен."
		errNoNullV2      = "V2() не вернул NULL и ошибку."
		errTextRandError = "rand error."
	)
	var (
		epoch time.Time
		ui    *impl
		got   UUID
		err   error
	)

	epoch = testEpoch()
	ui = newTestImpl(nil, epoch, nil, nil)
	testV1Data(t, ui)
	got, err = ui.V2(Domain.Person)
	if err != nil || got.Version() != Version.V2 || got.Variant() != Variant.RFC4122 {
		t.Fatalf(tplV2Wrong, got.String(), err)
	}
	if binary.BigEndian.Uint32(got.Bytes()[0:4]) != posixUID || got.Bytes()[9] != byte(Domain.Person) {
		t.Fatal(errNoUID)
	}
	got, err = ui.V2(Domain.Group)
	if err != nil || binary.BigEndian.Uint32(got.Bytes()[0:4]) != posixGID || got.Bytes()[9] != byte(Domain.Group) {
		t.Fatalf(tplV2GroupWrong, got.String(), err)
	}
	got, err = ui.V2(Domain.Org)
	if err != nil || got.Bytes()[9] != byte(Domain.Org) || got.Version() != Version.V2 {
		t.Fatalf(tplV2OrgWrong, got.String(), err)
	}
	ui = newTestImpl(&errorReader{err: errors.New(errTextRandError)}, epoch, nil, nil)
	if got, err = ui.V2(Domain.Person); err == nil || got != NULL {
		t.Fatal(errNoNullV2)
	}
}

// TestImpl_V3V5 Генерация UUID версий 3 и 5.
func TestImpl_V3V5(t *testing.T) {
	const (
		tplV3Wrong    = "V3() = %s."
		tplV5Wrong    = "V5() = %s."
		expectedV3    = "6fa459ea-ee8a-3ca4-894e-db77e160355e"
		expectedV5    = "886313e1-3b8a-5372-9b90-0c9aee199e5d"
		testNamespace = "python.org"
	)
	var (
		u3 UUID
		u5 UUID
	)

	u3 = singleton.V3(Namespace.DNS, testNamespace)
	if u3.String() != expectedV3 || u3.Version() != Version.V3 || u3.Variant() != Variant.RFC4122 {
		t.Fatalf(tplV3Wrong, u3.String())
	}
	u5 = singleton.V5(Namespace.DNS, testNamespace)
	if u5.String() != expectedV5 || u5.Version() != Version.V5 || u5.Variant() != Variant.RFC4122 {
		t.Fatalf(tplV5Wrong, u5.String())
	}
}

// TestImpl_V4 Генерация UUID версии 4.
func TestImpl_V4(t *testing.T) {
	const (
		tplV4Wrong       = "V4() = %s."
		errNoNullV4      = "V4() не вернул NULL при ошибке чтения случайных данных."
		errTextRandError = "rand error."
		expectedV4       = "00010203-0405-4607-8809-0a0b0c0d0e0f"
	)
	var (
		epoch time.Time
		ui    *impl
		got   UUID
	)

	epoch = testEpoch()
	ui = newTestImpl(&byteReader{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}}, epoch, nil, nil)
	got = ui.V4()
	if got.String() != expectedV4 || got.Version() != Version.V4 || got.Variant() != Variant.RFC4122 {
		t.Fatalf(tplV4Wrong, got.String())
	}
	ui = newTestImpl(&errorReader{err: errors.New(errTextRandError)}, epoch, nil, nil)
	if got = ui.V4(); got != NULL {
		t.Fatal(errNoNullV4)
	}
}

// TestImpl_V6 Генерация UUID версии 6.
func TestImpl_V6(t *testing.T) {
	const (
		tplV6Wrong       = "V6() = %v, ожидалось %x."
		errVersionWrong6 = "V6() установил некорректную версию или вариант."
		errNoNullV6      = "V6() не вернул NULL при ошибке чтения случайной последовательности."
		errTextRandError = "rand error."
	)
	var (
		epoch        time.Time
		ui           *impl
		hardwareAddr net.HardwareAddr
		got          UUID
		expected     []byte
		timeNow      uint64
		timeHighMid  uint64
	)

	epoch = testEpoch()
	ui = newTestImpl(nil, epoch, nil, nil)
	hardwareAddr = testV1Data(t, ui)
	got = ui.V6()
	timeNow = epochStart + uint64(epoch.UnixNano()/100)
	timeHighMid = timeNow >> 12
	expected = make([]byte, size)
	binary.BigEndian.PutUint32(expected[0:], uint32(timeHighMid>>16))
	binary.BigEndian.PutUint16(expected[4:], uint16(timeHighMid&0xffff))
	binary.BigEndian.PutUint16(expected[6:], uint16(timeNow&0xfff)|0x6000)
	binary.BigEndian.PutUint16(expected[8:], 0x9234)
	copy(expected[10:], hardwareAddr)
	if !bytes.Equal(got.Bytes(), expected) {
		t.Fatalf(tplV6Wrong, got.String(), expected)
	}
	if got.Version() != Version.V6 || got.Variant() != Variant.RFC4122 {
		t.Fatal(errVersionWrong6)
	}
	ui = newTestImpl(&errorReader{err: errors.New(errTextRandError)}, epoch, nil, nil)
	if got = ui.V6(); got != NULL {
		t.Fatal(errNoNullV6)
	}
}

// TestImpl_V7 Генерация UUID версии 7.
func TestImpl_V7(t *testing.T) {
	const (
		tplV7Wrong       = "V7() = %s."
		errNoNullV7      = "V7() не вернул NULL при ошибке чтения случайных данных."
		errTextRandError = "rand error."
	)
	var (
		epoch          time.Time
		ui             *impl
		got            UUID
		timestamp      uint64
		expectedPrefix []byte
	)

	epoch = testEpoch()
	ui = newTestImpl(&byteReader{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}}, epoch, nil, nil)
	got = ui.V7()
	timestamp = uint64(epoch.UnixNano() / 1e6)
	expectedPrefix = make([]byte, 6)
	binary.BigEndian.PutUint32(expectedPrefix[0:], uint32(timestamp>>16))
	binary.BigEndian.PutUint16(expectedPrefix[4:], uint16(timestamp))
	if !bytes.Equal(got.Bytes()[:6], expectedPrefix) || got.Version() != Version.V7 || got.Variant() != Variant.RFC4122 {
		t.Fatalf(tplV7Wrong, got.String())
	}
	ui = newTestImpl(&errorReader{err: errors.New(errTextRandError)}, epoch, nil, nil)
	if got = ui.V7(); got != NULL {
		t.Fatal(errNoNullV7)
	}
}

// TestImpl_V8 Генерация UUID версии 8.
func TestImpl_V8(t *testing.T) {
	const (
		tplV8Wrong       = "V8() = %s."
		errNoNullV8      = "V8() не вернул NULL при ошибке чтения случайных данных."
		errTextRandError = "rand error."
		expectedV8       = "00010203-0405-8607-8809-0a0b0c0d0e0f"
	)
	var (
		epoch time.Time
		ui    *impl
		got   UUID
	)

	epoch = testEpoch()
	ui = newTestImpl(&byteReader{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}}, epoch, nil, nil)
	got = ui.V8()
	if got.String() != expectedV8 || got.Version() != Version.V8 || got.Variant() != Variant.RFC4122 {
		t.Fatalf(tplV8Wrong, got.String())
	}
	ui = newTestImpl(&errorReader{err: errors.New(errTextRandError)}, epoch, nil, nil)
	if got = ui.V8(); got != NULL {
		t.Fatal(errNoNullV8)
	}
}

// TestImpl_V1V6HardwareError Ошибка аппаратного адреса при генерации UUID версий 1 и 6.
func TestImpl_V1V6HardwareError(t *testing.T) {
	const (
		tplV1HwWrong     = "V1() при ошибке hardware address = %v, %v."
		tplV6HwWrong     = "V6() при ошибке hardware address = %v."
		errTextRandError = "rand error."
		errTextNoHw      = "no hw."
	)
	var (
		epoch time.Time
		ui    *impl
		got   UUID
		err   error
	)

	epoch = testEpoch()
	ui = newTestImpl(&byteReader{data: []byte{0x12, 0x34}, err: errors.New(errTextRandError)}, epoch, nil, errors.New(errTextNoHw))
	if got, err = ui.V1(); err == nil || got != NULL {
		t.Fatalf(tplV1HwWrong, got, err)
	}
	ui = newTestImpl(&byteReader{data: []byte{0x12, 0x34}, err: errors.New(errTextRandError)}, epoch, nil, errors.New(errTextNoHw))
	if got = ui.V6(); got != NULL {
		t.Fatalf(tplV6HwWrong, got)
	}
}

package uuid

import (
	"bytes"
	"errors"
	"net"
	"testing"
	"time"
)

// TestImpl_DefaultHWAddrFunc Аппаратный адрес по умолчанию.
func TestImpl_DefaultHWAddrFunc(t *testing.T) {
	const (
		errNoNetInterfacesErr = "defaultHWAddrFunc() не вернул ошибку net.Interfaces."
		errNoNotFoundErr      = "defaultHWAddrFunc() не вернул errNotFound."
		tplShortAddr          = "defaultHWAddrFunc() вернул короткий адрес: %v."
		errTextIfacesError    = "ifaces error."
	)
	var (
		oldNetInterfaces func() ([]net.Interface, error)
		hw               net.HardwareAddr
		err              error
	)

	oldNetInterfaces = netInterfaces
	t.Cleanup(func() { netInterfaces = oldNetInterfaces })
	netInterfaces = func() ([]net.Interface, error) { return nil, errors.New(errTextIfacesError) }
	if _, err = defaultHWAddrFunc(); err == nil {
		t.Fatal(errNoNetInterfacesErr)
	}
	netInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{{HardwareAddr: net.HardwareAddr{0x01}}}, nil
	}
	if _, err = defaultHWAddrFunc(); err == nil {
		t.Fatal(errNoNotFoundErr)
	}
	netInterfaces = oldNetInterfaces
	hw, err = defaultHWAddrFunc()
	if err == nil && len(hw) < 6 {
		t.Fatalf(tplShortAddr, hw)
	}
}

// TestImpl_GetEpoch Эпоха генерации UUID.
func TestImpl_GetEpoch(t *testing.T) {
	const (
		tplEpochWrong = "getEpoch() = %d, ожидалось %d."
	)
	var (
		epoch time.Time
		ui    *impl
		got   uint64
		want  uint64
	)

	epoch = testEpoch()
	ui = newTestImpl(&byteReader{}, epoch, nil, nil)
	got, want = ui.getEpoch(), epochStart+uint64(epoch.UnixNano()/100)
	if got != want {
		t.Fatalf(tplEpochWrong, got, want)
	}
}

// TestImpl_GetHardwareAddr Аппаратный адрес генератора.
func TestImpl_GetHardwareAddr(t *testing.T) {
	const (
		tplGetHwWrong    = "getHardwareAddr() = %v, %v."
		tplGetHwFallback = "getHardwareAddr() fallback = %v, %v."
		errNoRandAddrErr = "getHardwareAddr() не вернул ошибку чтения случайного адреса."
		errTextNoHw      = "no hw."
		errTextRandError = "rand error."
	)
	var (
		epoch time.Time
		ui    *impl
		hw    net.HardwareAddr
		got   net.HardwareAddr
		err   error
	)

	epoch = testEpoch()
	hw = net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	ui = newTestImpl(&byteReader{}, epoch, hw, nil)
	got, err = ui.getHardwareAddr()
	if err != nil || !bytes.Equal(got, hw) {
		t.Fatalf(tplGetHwWrong, got, err)
	}
	ui = newTestImpl(&byteReader{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}}, epoch, nil, errors.New(errTextNoHw))
	got, err = ui.getHardwareAddr()
	if err != nil || len(got) != 6 || got[0] != 0x01 || got[1] != 0x01 {
		t.Fatalf(tplGetHwFallback, got, err)
	}
	ui = newTestImpl(&errorReader{err: errors.New(errTextRandError)}, epoch, nil, errors.New(errTextNoHw))
	if _, err = ui.getHardwareAddr(); err == nil {
		t.Fatal(errNoRandAddrErr)
	}
}

// TestImpl_GetClockSequence Последовательность часов генератора.
func TestImpl_GetClockSequence(t *testing.T) {
	const (
		tplClockSeqWrong   = "getClockSequence() = %d, %x, %v."
		tplClockSeqNoShift = "getClockSequence() без сдвига времени = %x, %v."
		errNoClockSeqErr   = "getClockSequence() не вернул ошибку чтения случайной последовательности."
		errTextRandError   = "rand error."
	)
	var (
		epoch time.Time
		ui    *impl
		ret   uint64
		seq   uint16
		err   error
	)

	epoch = testEpoch()
	ui = newTestImpl(&byteReader{data: []byte{0x12, 0x34}}, epoch, nil, nil)
	ret, seq, err = ui.getClockSequence()
	if err != nil || ret != ui.getEpoch() || seq != 0x1234 {
		t.Fatalf(tplClockSeqWrong, ret, seq, err)
	}
	_, seq, err = ui.getClockSequence()
	if err != nil || seq != 0x1235 {
		t.Fatalf(tplClockSeqNoShift, seq, err)
	}
	ui = newTestImpl(&errorReader{err: errors.New(errTextRandError)}, epoch, nil, nil)
	if _, _, err = ui.getClockSequence(); err == nil {
		t.Fatal(errNoClockSeqErr)
	}
}

package uuid

import (
	"bytes"
	"errors"
	"net"
	"testing"
)

// TestImpl_DefaultHWAddrFunc Аппаратный адрес по умолчанию.
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

// TestImpl_GetEpoch Эпоха генерации UUID.
func TestImpl_GetEpoch(t *testing.T) {
	epoch := testEpoch()
	ui := newTestImpl(&byteReader{}, epoch, nil, nil)
	if got, want := ui.getEpoch(), epochStart+uint64(epoch.UnixNano()/100); got != want {
		t.Fatalf("getEpoch() = %d, ожидалось %d", got, want)
	}
}

// TestImpl_GetHardwareAddr Аппаратный адрес генератора.
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

// TestImpl_GetClockSequence Последовательность часов генератора.
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

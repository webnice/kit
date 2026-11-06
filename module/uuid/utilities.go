package uuid

import (
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"net"
)

var netInterfaces = net.Interfaces

// defaultHWAddrFunc возвращает MAC-адрес сетевой карты.
func defaultHWAddrFunc() (ret net.HardwareAddr, err error) {
	const errNotFound = "MAC HW адрес не найден"
	var (
		ifaces []net.Interface
		iface  net.Interface
	)

	if ifaces, err = netInterfaces(); err != nil {
		return
	}
	for _, iface = range ifaces {
		if len(iface.HardwareAddr) >= 6 {
			ret = iface.HardwareAddr
			return
		}
	}
	err = errors.New(errNotFound)

	return
}

// getEpoch возвращает разницу между началом эпохи и текущим временем с точностью до 100 наносекунд.
func (ui *impl) getEpoch() uint64 { return epochStart + uint64(ui.epochFunc().UnixNano()/100) }

// getHardwareAddr возвращает физический MAC-адрес сетевой карты.
func (ui *impl) getHardwareAddr() (ret []byte, err error) {
	if ui.hardwareAddrOnce.Do(func() {
		var hwAddress []byte

		if hwAddress, err = ui.hwAddrFunc(); err == nil {
			copy(ui.hardwareAddr[:], hwAddress)
			return
		}
		// Инициализация случайным значением в случае отсутствия реальных интерфейсов или невозможности доступа к ним.
		if _, err = io.ReadFull(ui.rand, ui.hardwareAddr[:]); err != nil {
			return
		}
		// Установка мультикаст-бита, рекомендованная в RFC 4122.
		ui.hardwareAddr[0] |= 0x01
	}); err != nil {
		return
	}
	ret = ui.hardwareAddr[:]

	return
}

// getClockSequence возвращает текущее время и clock sequence, основанные на эпохе UUID.
func (ui *impl) getClockSequence() (ret uint64, clockSequence uint16, err error) {
	var buf []byte

	if ui.clockSequenceOnce.Do(func() {
		buf = make([]byte, 2)
		if _, err = io.ReadFull(ui.rand, buf); err != nil {
			return
		}
		ui.clockSequence = binary.BigEndian.Uint16(buf)
	}); err != nil {
		return
	}
	ui.storageMutex.Lock()
	ret = ui.getEpoch()
	if ret <= ui.lastTime {
		ui.clockSequence++
	}
	ui.lastTime, clockSequence = ret, ui.clockSequence
	ui.storageMutex.Unlock()

	return
}

// newFromHash возвращает UUID, основанный на хеше от пространства имён и имени.
func (ui *impl) newFromHash(h hash.Hash, namespace NamespaceType, name string) (ret *uuid) {
	ret = &uuid{data: namespace}
	h.Write(namespace[:])
	h.Write([]byte(name))
	copy(ret.data[:], h.Sum(nil))

	return
}

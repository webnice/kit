//go:build windows

// Package osext
package osext

import (
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	kernel                = syscall.MustLoadDLL("kernel32.dll")
	getModuleFileNameProc = kernel.MustFindProc("GetModuleFileNameW")
)

// GetModuleFileName() with hModule = NULL
func (ost *impl) executable() (ret string, err error) {
	ret, err = ost.getModuleFileName()

	return
}

func (ost *impl) getModuleFileName() (ret string, err error) {
	var (
		b    []uint16
		n    uint32
		size uint32
		r0   uintptr
		e1   error
	)

	b = make([]uint16, syscall.MAX_PATH)
	size = uint32(len(b))

	r0, _, e1 = getModuleFileNameProc.Call(0, uintptr(unsafe.Pointer(&b[0])), uintptr(size))
	if n = uint32(r0); n == 0 {
		err = e1
		return
	}
	ret = string(utf16.Decode(b[0:n]))

	return
}

//go:build linux || netbsd || solaris || dragonfly

// Package osext
package osext

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func (ost *impl) executable() (ret string, err error) {
	const (
		deletedTag          = " (deleted)"
		osLinuxProcPath     = "/proc/self/exe"
		osNetbsdProcPath    = "/proc/curproc/exe"
		osDragonflyProcPath = "/proc/curproc/file"
		osSolarisProcTpl    = "/proc/%d/path/a.out"
	)

	switch runtime.GOOS {
	case osLinux:
		if ret, err = os.Readlink(osLinuxProcPath); err != nil {
			return
		}
		ret = strings.TrimPrefix(strings.TrimSuffix(ret, deletedTag), deletedTag)
	case osNetbsd:
		ret, err = os.Readlink(osNetbsdProcPath)
	case osDragonfly:
		ret, err = os.Readlink(osDragonflyProcPath)
	case osSolaris:
		ret, err = os.Readlink(fmt.Sprintf(osSolarisProcTpl, os.Getpid()))
	default:
		err = errors.New("osext is not implemented for " + runtime.GOOS)
	}

	return
}

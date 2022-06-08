// Package osext
package osext

import (
	"os"
	"path/filepath"
)

const (
	osLinux     = "linux"
	osDarwin    = "darwin"
	osNetbsd    = "netbsd"
	osDragonfly = "dragonfly"
	osSolaris   = "solaris"
	osFreebsd   = "freebsd"
	osOpenbsd   = "openbsd"
)

// Внутренняя структура объекта пакета.
type impl struct {
	cx string
	ce error
	wd string
	we error
}

// New Конструктор объекта пакета.
func New() *impl {
	var ost = new(impl)
	ost.executableClean()
	ost.wd, ost.we = os.Getwd()
	return ost
}

func (ost *impl) executableClean() {
	if ost.cx, ost.ce = ost.executable(); ost.ce != nil {
		return
	}
	ost.cx = filepath.Clean(ost.cx)

	return
}

// Executable returns an absolute path that can be used to re-invoke the current program. It may not be valid after
// the current program exits.
func (ost *impl) Executable() (string, error) { return ost.cx, ost.ce }

// ExecutableFolder Returns same path as Executable, returns just the folder path. Excludes the executable name and
// any trailing slash.
func (ost *impl) ExecutableFolder() (ret string, err error) {
	if ret, err = ost.Executable(); err != nil {
		return
	}
	ret, err = filepath.Dir(ret), nil

	return
}

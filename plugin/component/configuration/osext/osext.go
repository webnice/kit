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
type OsExt struct {
	cx string
	ce error
	wd string
	we error
}

// New Конструктор объекта пакета.
func New() (ost *OsExt) {
	ost = new(OsExt)
	ost.executableClean()
	ost.wd, ost.we = os.Getwd()

	return
}

func (ost *OsExt) executableClean() {
	if ost.cx, ost.ce = ost.executable(); ost.ce != nil {
		return
	}
	ost.cx = filepath.Clean(ost.cx)

	return
}

// Executable returns an absolute path that can be used to re-invoke the current program. It may not be valid after
// the current program exits.
func (ost *OsExt) Executable() (string, error) { return ost.cx, ost.ce }

// ExecutableFolder Returns same path as Executable, returns just the folder path. Excludes the executable name and
// any trailing slash.
func (ost *OsExt) ExecutableFolder() (ret string, err error) {
	if ret, err = ost.Executable(); err != nil {
		return
	}
	ret, err = filepath.Dir(ret), nil

	return
}

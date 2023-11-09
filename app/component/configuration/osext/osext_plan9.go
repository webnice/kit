package osext

import (
	"os"
	"strconv"
	"syscall"
)

func (ost *OsExt) executable() (ret string, err error) {
	const (
		keyProc = "/proc/"
		keyText = "/text"
	)
	var fh *os.File

	if fh, err = os.Open(keyProc + strconv.Itoa(os.Getpid()) + keyText); err != nil {
		return
	}
	defer func() { _ = fh.Close() }()
	ret, err = syscall.Fd2path(int(fh.Fd()))

	return
}

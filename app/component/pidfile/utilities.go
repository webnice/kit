package pidfile

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"

	kitTypes "github.com/webnice/kit/v4/types"
)

// IsProcessExist Проверка существования процесса по данным из PID файла.
func (pid *impl) IsProcessExist() (err kitTypes.ErrorWithCode) {
	var (
		e       error
		fi      os.FileInfo
		buf     []byte
		str     []string
		pds     []int
		pvl     int64
		n       int
		process *os.Process
		exists  bool
	)

	if fi, e = os.Stat(pid.pfnm); os.IsNotExist(e) {
		return
	}
	if e != nil {
		err = pid.cfg.Errors().PidFileError(0, pid.pfnm, e)
		return
	}
	if fi.IsDir() {
		e = fmt.Errorf(tplPidFileIsDirectory, pid.pfnm)
		err = pid.cfg.Errors().PidFileError(0, pid.pfnm, e)
		return
	}
	if buf, e = os.ReadFile(pid.pfnm); e != nil {
		e = fmt.Errorf(tplPidContentRead, pid.pfnm, e)
		err = pid.cfg.Errors().PidFileError(0, pid.pfnm, e)
		return
	}
	str = strings.Split(string(buf), "\n")
	pds = make([]int, 0, len(str))
	for n = range str {
		str[n] = strings.TrimSpace(str[n])
		if pvl, _ = strconv.ParseInt(str[n], 10, 64); pvl > 0 {
			pds = append(pds, int(pvl))
		}
	}
	// Проверка наличия процессов с указанными в PID файле идентификаторами.
	for n = range pds {
		if process, e = os.FindProcess(pds[n]); e != nil {
			continue
		}
		switch e = process.Signal(syscall.Signal(0)); {
		case errors.Is(e, syscall.ESRCH), errors.Is(e, os.ErrProcessDone):
			// Процесс завершён.
		case e == nil, errors.Is(e, syscall.EPERM):
			// Процесс есть или к нему нет доступа, что говорит о том что он есть.
			exists = true
		}
	}
	if exists {
		err = pid.cfg.Errors().PidExistsAnotherProcessOfApplication(0, pds)
	}

	return
}

// Создание директории по имени файла без обработки ошибок.
func (pid *impl) createDirectoryForFile(filename string) {
	var dir = path.Dir(filename)
	_ = os.MkdirAll(dir, defaultModeDir)
}

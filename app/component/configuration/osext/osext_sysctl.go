//go:build darwin || freebsd || openbsd

package osext

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

func (ost *OsExt) executable() (ret string, err error) {
	var (
		errNo syscall.Errno
		mib   [4]int32
		n     uintptr
		buf   []byte
	)

	mib, n = ost.makeMIB(), uintptr(0)
	// Get length.
	if _, _, errNo = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		0,
		uintptr(unsafe.Pointer(&n)),
		0,
		0,
	); errNo != 0 {
		err = errNo
		return
	}
	// This shouldn't happen.
	if n == 0 {
		return
	}
	buf = make([]byte, n)
	if _, _, errNo = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&n)),
		0,
		0,
	); errNo != 0 {
		err = errNo
		return
	}
	// This shouldn't happen.
	if n == 0 {
		return
	}
	// execPath will not be empty due to above checks.
	// Try to get the absolute path if the execPath is not rooted.
	if ret = ost.makeExecPath(buf, n); ret[0] != '/' {
		if ret, err = ost.getAbs(ret); err != nil {
			return
		}
	}
	// For darwin KERN_PROCARGS may return the path to a symlink rather than the actual executable.
	switch runtime.GOOS {
	case osDarwin:
		if ret, err = filepath.EvalSymlinks(ret); err != nil {
			return
		}
	}

	return
}

func (ost *OsExt) makeMIB() (mib [4]int32) {
	switch runtime.GOOS {
	case osFreebsd:
		mib = [4]int32{1 /* CTL_KERN */, 14 /* KERN_PROC */, 12 /* KERN_PROC_PATHNAME */, -1}
	case osDarwin:
		mib = [4]int32{1 /* CTL_KERN */, 38 /* KERN_PROCARGS */, int32(os.Getpid()), -1}
	case osOpenbsd:
		mib = [4]int32{1 /* CTL_KERN */, 55 /* KERN_PROC_ARGS */, int32(os.Getpid()), 1 /* KERN_PROC_ARGV */}
	}

	return
}

func (ost *OsExt) makeExecPath(buf []byte, n uintptr) (ret string) {
	var (
		i int
		v byte
	)

	switch runtime.GOOS {
	case osOpenbsd:
		ret = ost.makeExecPathOpenBSD(buf, n)
	default:
		for i, v = range buf {
			if v == 0 {
				buf = buf[:i]
				break
			}
		}
		ret = string(buf)
	}

	return
}

func (ost *OsExt) makeExecPathOpenBSD(buf []byte, n uintptr) (execPath string) {
	var (
		args []string
		argv uintptr
		argp *[1048576]byte
		i    int
	)

	// buf now contains **argv, with pointers to each of the C-style NULL terminated arguments.
	argv = uintptr(unsafe.Pointer(&buf[0]))
Loop:
	for {
		argp = *(**[1 << 20]byte)(unsafe.Pointer(argv))
		if argp == nil {
			break
		}
		for i = 0; uintptr(i) < n; i++ {
			// we don't want the full arguments list
			if string(argp[i]) == " " {
				break Loop
			}
			if argp[i] != 0 {
				continue
			}
			args = append(args, string(argp[:i]))
			n -= uintptr(i)
			break
		}
		if n < unsafe.Sizeof(argv) {
			break
		}
		argv += unsafe.Sizeof(argv)
		n -= unsafe.Sizeof(argv)
	}
	execPath = args[0]
	// There is no canonical way to get an executable path on OpenBSD, so check PATH in case we are called directly.
	if execPath[0] != '/' && execPath[0] != '.' {
		execIsInPath, err := exec.LookPath(execPath)
		if err == nil {
			execPath = execIsInPath
		}
	}

	return
}

func (ost *OsExt) getAbs(execPath string) (ret string, err error) {
	if ost.we != nil {
		ret, err = execPath, ost.we
		return
	}
	// The execPath may begin with a "../" or a "./" so clean it first.
	// Join the two paths, trailing and starting slashes undetermined, so use the generic Join function.
	ret = filepath.Join(ost.wd, filepath.Clean(execPath))

	return
}

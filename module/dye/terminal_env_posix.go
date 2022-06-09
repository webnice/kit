//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd
// +build darwin dragonfly freebsd linux netbsd openbsd

// Package dye
package dye

import "golang.org/x/sys/unix"

func isForeground(fd int) (ret bool) {
	var (
		err  error
		pgrp int
	)

	if pgrp, err = unix.IoctlGetInt(fd, unix.TIOCGPGRP); err != nil {
		return
	}
	ret = pgrp == unix.Getpgrp()

	return
}

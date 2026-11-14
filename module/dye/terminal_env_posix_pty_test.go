//go:build darwin || linux
// +build darwin linux

package dye

import (
	"os"
	"testing"
)

// TestIsForegroundPty Проверка признака активного терминала на псевдотерминале.
func TestIsForegroundPty(t *testing.T) {
	const errNotTrue = "isForeground(): для псевдотерминала с foreground группой ожидалось true."
	var master, slave *os.File

	master, slave = openTestPty(t)
	defer func() { _ = master.Close() }()
	defer func() { _ = slave.Close() }()
	ptySetForegroundPgrp(t, master)
	if !isForeground(int(slave.Fd())) {
		t.Fatal(errNotTrue)
	}
}

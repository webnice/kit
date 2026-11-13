//go:build darwin || linux
// +build darwin linux

package dye

import (
	"testing"
)

// TestIsForegroundPty Проверка признака активного терминала на псевдотерминале.
func TestIsForegroundPty(t *testing.T) {
	master, slave := openTestPty(t)
	defer func() { _ = master.Close() }()
	defer func() { _ = slave.Close() }()
	ptySetForegroundPgrp(t, master)
	if !isForeground(int(slave.Fd())) {
		t.Fatal("isForeground(): для псевдотерминала с foreground группой ожидалось true")
	}
}

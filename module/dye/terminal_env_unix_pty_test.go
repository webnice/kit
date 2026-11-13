//go:build darwin || linux
// +build darwin linux

package dye

import (
	"errors"
	"testing"
)

// TestTermStatusReportFile Запрос и получение ответа статуса терминала через псевдотерминал.
func TestTermStatusReportFile(t *testing.T) {
	t.Setenv(envTerm, "xterm-256color")
	t.Run("OscResponse", func(t *testing.T) {
		master, slave := openTestPty(t)
		defer func() { _ = master.Close() }()
		defer func() { _ = slave.Close() }()
		ptySetForegroundPgrp(t, master)
		// Ответы терминала записываются в master заранее.
		if _, err := master.Write([]byte("\x1b]11;rgb:1111/2222/3333\a\x1b[42;1R")); err != nil {
			t.Fatalf("Write(): неожиданная ошибка: %v", err)
		}
		ret, err := termStatusReportFile(slave, 11)
		if err != nil {
			t.Fatalf("termStatusReportFile(): неожиданная ошибка: %v", err)
		}
		if ret != "\x1b]11;rgb:1111/2222/3333\a" {
			t.Fatalf("termStatusReportFile(): ожидался ответ %q, получено %q", "\x1b]11;rgb:1111/2222/3333\a", ret)
		}
	})
	t.Run("CursorInsteadOfOsc", func(t *testing.T) {
		master, slave := openTestPty(t)
		defer func() { _ = master.Close() }()
		defer func() { _ = slave.Close() }()
		ptySetForegroundPgrp(t, master)
		if _, err := master.Write([]byte("\x1b[42;1R")); err != nil {
			t.Fatalf("Write(): неожиданная ошибка: %v", err)
		}
		if _, err := termStatusReportFile(slave, 11); !errors.Is(err, errStatusReport) {
			t.Fatalf("termStatusReportFile(): ожидалась ошибка errStatusReport, получено %v", err)
		}
	})
	t.Run("MissingCursorResponse", func(t *testing.T) {
		master, slave := openTestPty(t)
		defer func() { _ = slave.Close() }()
		ptySetForegroundPgrp(t, master)
		if _, err := master.Write([]byte("\x1b]11;rgb:1111/2222/3333\a")); err != nil {
			t.Fatalf("Write(): неожиданная ошибка: %v", err)
		}
		// Закрытие master до чтения ответа позиции курсора.
		_ = master.Close()
		if _, err := termStatusReportFile(slave, 11); err == nil {
			t.Fatal("termStatusReportFile(): при обрыве ответа ожидалась ошибка")
		}
	})
}

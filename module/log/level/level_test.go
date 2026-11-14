package level

import (
	"testing"
)

func TestLevelString(t *testing.T) {
	const (
		tplStringMismatch = "Level(%d).String() = %q, ожидалось %q."
		tplUnknownString  = "не известный уровень String() = %q, ожидалась пустая строка."
	)
	var (
		unknown = Level(100)
		key     Level
		lvs     struct {
			Name  string
			Short rune
		}
	)

	for key, lvs = range level {
		if key.String() != lvs.Name {
			t.Errorf(tplStringMismatch, key.Int(), key.String(), lvs.Name)
		}
	}
	if unknown.String() != "" {
		t.Errorf(tplUnknownString, unknown.String())
	}
}

func TestLevelBytes(t *testing.T) {
	const (
		tplBytesMismatch = "Level(%d).Bytes() = %q, ожидалось %q."
		tplUnknownBytes  = "не известный уровень Bytes() = %v, ожидался nil."
	)
	var (
		unknown = Level(100)
		key     Level
		lvs     struct {
			Name  string
			Short rune
		}
	)

	for key, lvs = range level {
		if string(key.Bytes()) != lvs.Name {
			t.Errorf(tplBytesMismatch, key.Int(), key.Bytes(), lvs.Name)
		}
	}
	if unknown.Bytes() != nil {
		t.Errorf(tplUnknownBytes, unknown.Bytes())
	}
}

func TestLevelShort(t *testing.T) {
	const (
		tplShortMismatch = "Level(%d).Short() = %q, ожидалось %q."
		tplUnknownShort  = "не известный уровень Short() = %q, ожидался 0."
	)
	var (
		unknown = Level(100)
		key     Level
		lvs     struct {
			Name  string
			Short rune
		}
	)

	for key, lvs = range level {
		if key.Short() != lvs.Short {
			t.Errorf(tplShortMismatch, key.Int(), key.Short(), lvs.Short)
		}
	}
	if unknown.Short() != 0 {
		t.Errorf(tplUnknownShort, unknown.Short())
	}
}

func TestLevelInt(t *testing.T) {
	const (
		tplIntMismatch = "Int() = %d/%d/%d, ожидалось -1/0/8."
	)
	var (
		off   Level
		fatal Level
		trace Level
	)

	off, fatal, trace = Off, Fatal, Trace
	if off.Int() != -1 || fatal.Int() != 0 || trace.Int() != 8 {
		t.Errorf(tplIntMismatch, off.Int(), fatal.Int(), trace.Int())
	}
}

func TestLevelMarshalText(t *testing.T) {
	const (
		tplMarshalMismatch = "Info.MarshalText() = %q, %v, ожидалось %q."
		tplMarshalUnknown  = "Level(100).MarshalText() = %q, %v, ожидалась пустая строка."
	)
	var (
		info    Level
		unknown Level
		ret     []byte
		err     error
	)

	info, unknown = Info, Level(100)
	if ret, err = info.MarshalText(); err != nil || string(ret) != "info" {
		t.Errorf(tplMarshalMismatch, ret, err, "info")
	}
	if ret, err = unknown.MarshalText(); err != nil || string(ret) != "" {
		t.Errorf(tplMarshalUnknown, ret, err)
	}
}

func TestLevelUnmarshalText(t *testing.T) {
	const (
		tplUnmarshalMismatch = "UnmarshalText(%q) = %v, %v, ожидалось %v."
		errWrongNoError      = "UnmarshalText(wrong), ошибка не возвращена."
	)
	var l Level

	// Наименование уровня.
	if err := l.UnmarshalText([]byte("Info")); err != nil || l != Info {
		t.Errorf(tplUnmarshalMismatch, "Info", l, err, Info)
	}
	// Краткое наименование уровня.
	if err := l.UnmarshalText([]byte("W")); err != nil || l != Warning {
		t.Errorf(tplUnmarshalMismatch, "W", l, err, Warning)
	}
	// Числовое значение уровня.
	if err := l.UnmarshalText([]byte(" 5 ")); err != nil || l != Notice {
		t.Errorf(tplUnmarshalMismatch, " 5 ", l, err, Notice)
	}
	// Не известное числовое значение преобразуется в Off.
	if err := l.UnmarshalText([]byte("100")); err != nil || l != Off {
		t.Errorf(tplUnmarshalMismatch, "100", l, err, Off)
	}
	// Не верное значение - ошибка.
	if err := l.UnmarshalText([]byte("wrong")); err == nil {
		t.Error(errWrongNoError)
	}
}

func TestParseLevelInt64(t *testing.T) {
	const tplParseMismatch = "ParseLevelInt64(%d) = %v, ожидалось %v."

	if ret := ParseLevelInt64(0); ret != Fatal {
		t.Errorf(tplParseMismatch, 0, ret, Fatal)
	}
	if ret := ParseLevelInt64(8); ret != Trace {
		t.Errorf(tplParseMismatch, 8, ret, Trace)
	}
	if ret := ParseLevelInt64(-1); ret != Off {
		t.Errorf(tplParseMismatch, -1, ret, Off)
	}
	if ret := ParseLevelInt64(100); ret != Off {
		t.Errorf(tplParseMismatch, 100, ret, Off)
	}
}

func TestParseLevelString(t *testing.T) {
	const tplParseMismatch = "ParseLevelString(%s) = %v, ожидалось %v."

	if ret := ParseLevelString("debug"); ret != Debug {
		t.Errorf(tplParseMismatch, "debug", ret, Debug)
	}
	if ret := ParseLevelString(" E "); ret != Error {
		t.Errorf(tplParseMismatch, " E ", ret, Error)
	}
	if ret := ParseLevelString("ALERT"); ret != Alert {
		t.Errorf(tplParseMismatch, "ALERT", ret, Alert)
	}
	if ret := ParseLevelString("wrong"); ret != Off {
		t.Errorf(tplParseMismatch, "wrong", ret, Off)
	}
	if ret := ParseLevelString("  "); ret != Off {
		t.Errorf(tplParseMismatch, "пробелы", ret, Off)
	}
}

package level

import (
	"testing"
)

func TestLevelString(t *testing.T) {
	var unknown = Level(100)

	for key, lvs := range level {
		if key.String() != lvs.Name {
			t.Errorf("Level(%d).String() = %q, ожидалось %q", key.Int(), key.String(), lvs.Name)
		}
	}
	if unknown.String() != "" {
		t.Errorf("не известный уровень String() = %q, ожидалась пустая строка", unknown.String())
	}
}

func TestLevelBytes(t *testing.T) {
	var unknown = Level(100)

	for key, lvs := range level {
		if string(key.Bytes()) != lvs.Name {
			t.Errorf("Level(%d).Bytes() = %q, ожидалось %q", key.Int(), key.Bytes(), lvs.Name)
		}
	}
	if unknown.Bytes() != nil {
		t.Errorf("не известный уровень Bytes() = %v, ожидался nil", unknown.Bytes())
	}
}

func TestLevelShort(t *testing.T) {
	var unknown = Level(100)

	for key, lvs := range level {
		if key.Short() != lvs.Short {
			t.Errorf("Level(%d).Short() = %q, ожидалось %q", key.Int(), key.Short(), lvs.Short)
		}
	}
	if unknown.Short() != 0 {
		t.Errorf("не известный уровень Short() = %q, ожидался 0", unknown.Short())
	}
}

func TestLevelInt(t *testing.T) {
	var (
		off   = Off
		fatal = Fatal
		trace = Trace
	)

	if off.Int() != -1 || fatal.Int() != 0 || trace.Int() != 8 {
		t.Errorf("Int() = %d/%d/%d, ожидалось -1/0/8", off.Int(), fatal.Int(), trace.Int())
	}
}

func TestLevelMarshalText(t *testing.T) {
	var (
		info    = Info
		unknown = Level(100)
	)

	if ret, err := info.MarshalText(); err != nil || string(ret) != "info" {
		t.Errorf("Info.MarshalText() = %q, %v, ожидалось %q", ret, err, "info")
	}
	if ret, err := unknown.MarshalText(); err != nil || string(ret) != "" {
		t.Errorf("Level(100).MarshalText() = %q, %v, ожидалась пустая строка", ret, err)
	}
}

func TestLevelUnmarshalText(t *testing.T) {
	var l Level

	// Наименование уровня.
	if err := l.UnmarshalText([]byte("Info")); err != nil || l != Info {
		t.Errorf("UnmarshalText(%q) = %v, %v, ожидалось %v", "Info", l, err, Info)
	}
	// Краткое наименование уровня.
	if err := l.UnmarshalText([]byte("W")); err != nil || l != Warning {
		t.Errorf("UnmarshalText(%q) = %v, %v, ожидалось %v", "W", l, err, Warning)
	}
	// Числовое значение уровня.
	if err := l.UnmarshalText([]byte(" 5 ")); err != nil || l != Notice {
		t.Errorf("UnmarshalText(%q) = %v, %v, ожидалось %v", " 5 ", l, err, Notice)
	}
	// Не известное числовое значение преобразуется в Off.
	if err := l.UnmarshalText([]byte("100")); err != nil || l != Off {
		t.Errorf("UnmarshalText(%q) = %v, %v, ожидалось %v", "100", l, err, Off)
	}
	// Не верное значение - ошибка.
	if err := l.UnmarshalText([]byte("wrong")); err == nil {
		t.Error("UnmarshalText(wrong), ошибка не возвращена")
	}
}

func TestParseLevelInt64(t *testing.T) {
	if ret := ParseLevelInt64(0); ret != Fatal {
		t.Errorf("ParseLevelInt64(0) = %v, ожидалось %v", ret, Fatal)
	}
	if ret := ParseLevelInt64(8); ret != Trace {
		t.Errorf("ParseLevelInt64(8) = %v, ожидалось %v", ret, Trace)
	}
	if ret := ParseLevelInt64(-1); ret != Off {
		t.Errorf("ParseLevelInt64(-1) = %v, ожидалось %v", ret, Off)
	}
	if ret := ParseLevelInt64(100); ret != Off {
		t.Errorf("ParseLevelInt64(100) = %v, ожидалось %v", ret, Off)
	}
}

func TestParseLevelString(t *testing.T) {
	if ret := ParseLevelString("debug"); ret != Debug {
		t.Errorf("ParseLevelString(debug) = %v, ожидалось %v", ret, Debug)
	}
	if ret := ParseLevelString(" E "); ret != Error {
		t.Errorf("ParseLevelString( E ) = %v, ожидалось %v", ret, Error)
	}
	if ret := ParseLevelString("ALERT"); ret != Alert {
		t.Errorf("ParseLevelString(ALERT) = %v, ожидалось %v", ret, Alert)
	}
	if ret := ParseLevelString("wrong"); ret != Off {
		t.Errorf("ParseLevelString(wrong) = %v, ожидалось %v", ret, Off)
	}
	if ret := ParseLevelString("  "); ret != Off {
		t.Errorf("ParseLevelString(пробелы) = %v, ожидалось %v", ret, Off)
	}
}

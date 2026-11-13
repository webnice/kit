package colorful

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestHexColor(t *testing.T) {
	for _, tc := range []struct {
		hc HexColor
		s  string
	}{
		{HexColor{R: 0, G: 0, B: 0}, "#000000"},
		{HexColor{R: 1, G: 0, B: 0}, "#ff0000"},
		{HexColor{R: 0, G: 1, B: 0}, "#00ff00"},
		{HexColor{R: 0, G: 0, B: 1}, "#0000ff"},
		{HexColor{R: 1, G: 1, B: 1}, "#ffffff"},
	} {
		var gotHC HexColor
		if err := gotHC.Scan(tc.s); err != nil {
			t.Errorf("_.Scan(%q) == %v, want <nil>", tc.s, err)
		}
		if !reflect.DeepEqual(gotHC, tc.hc) {
			t.Errorf("_.Scan(%q) wrote %v, want %v", tc.s, gotHC, tc.hc)
		}
		if gotValue, err := tc.hc.Value(); err != nil || !reflect.DeepEqual(gotValue, tc.s) {
			t.Errorf("%v.Value() == %v, %v, want %v, <nil>", tc.hc, gotValue, err, tc.s)
		}
	}
}

type CompositeType struct {
	Name  string   `json:"name,omitempty"`
	Color HexColor `json:"color,omitempty"`
}

func TestHexColorCompositeJson(t *testing.T) {
	var obj = CompositeType{Name: "John", Color: HexColor{R: 1, G: 0, B: 1}}
	var jsonData, err = json.Marshal(obj)
	if err != nil {
		t.Errorf("json.Marshall(obj) wrote %v", err)
	}
	var obj2 CompositeType
	err = json.Unmarshal(jsonData, &obj2)

	if err != nil {
		t.Errorf("json.Unmarshall(%s) wrote %v", jsonData, err)
	}

	if !reflect.DeepEqual(obj2, obj) {
		t.Errorf("json.Unmarshal(json.Marsrhall(obj)) != obj")
	}

}

// TestHexColorScan Сканирование значения базы данных.
func TestHexColorScan(t *testing.T) {
	var hc HexColor
	if err := hc.Scan(123); err == nil {
		t.Error("Scan(): ожидалась ошибка неподдерживаемого типа")
	} else {
		var e errUnsupportedType
		if !errors.As(err, &e) {
			t.Errorf("Scan(): неожиданная ошибка: %v", err)
		}
		if msg := e.Error(); msg == "" {
			t.Error("Error(): ожидалось непустое сообщение об ошибке")
		}
	}
	if err := hc.Scan("#xyz"); err == nil {
		t.Error("Scan(): ожидалась ошибка разбора цвета")
	}
	if err := hc.Scan("#ff8000"); err != nil {
		t.Fatalf("Scan(): неожиданная ошибка: %v", err)
	}
	if got := Color(hc).Hex(); got != "#ff8000" {
		t.Errorf("Scan(): ожидался цвет #ff8000, получено %v", got)
	}
}

// TestHexColorUnmarshalJSON Десериализация JSON.
func TestHexColorUnmarshalJSON(t *testing.T) {
	var hc HexColor
	if err := hc.UnmarshalJSON([]byte("{invalid json")); err == nil {
		t.Error("UnmarshalJSON(): ожидалась ошибка JSON")
	}
	if err := hc.UnmarshalJSON([]byte("\"#xyz\"")); err == nil {
		t.Error("UnmarshalJSON(): ожидалась ошибка разбора цвета")
	}
	if err := hc.UnmarshalJSON([]byte("\"#ff8000\"")); err != nil {
		t.Fatalf("UnmarshalJSON(): неожиданная ошибка: %v", err)
	}
	if got := Color(hc).Hex(); got != "#ff8000" {
		t.Errorf("UnmarshalJSON(): ожидался цвет #ff8000, получено %v", got)
	}
}

// TestHexColorDecode Десериализация строки.
func TestHexColorDecode(t *testing.T) {
	var hc HexColor
	if err := hc.Decode("#xyz"); err == nil {
		t.Error("Decode(): ожидалась ошибка разбора цвета")
	}
	if err := hc.Decode("#ff8000"); err != nil {
		t.Fatalf("Decode(): неожиданная ошибка: %v", err)
	}
	if got := Color(hc).Hex(); got != "#ff8000" {
		t.Errorf("Decode(): ожидался цвет #ff8000, получено %v", got)
	}
}

// TestHexColorMarshalYAML Сериализация YAML.
func TestHexColorMarshalYAML(t *testing.T) {
	c, err := Hex("#ff8000")
	if err != nil {
		t.Fatalf("Hex(): неожиданная ошибка: %v", err)
	}
	got, err := HexColor(c).MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML(): неожиданная ошибка: %v", err)
	}
	if got != "#ff8000" {
		t.Errorf("MarshalYAML(): ожидалось значение #ff8000, получено %v", got)
	}
}

// TestHexColorUnmarshalYAML Десериализация YAML.
func TestHexColorUnmarshalYAML(t *testing.T) {
	var hc HexColor
	if err := hc.UnmarshalYAML(func(any) error { return errors.New("ошибка") }); err == nil {
		t.Error("UnmarshalYAML(): ожидалась ошибка функции разбора")
	}
	if err := hc.UnmarshalYAML(func(v any) error { *(v.(*string)) = "#xyz"; return nil }); err == nil {
		t.Error("UnmarshalYAML(): ожидалась ошибка разбора цвета")
	}
	if err := hc.UnmarshalYAML(func(v any) error { *(v.(*string)) = "#ff8000"; return nil }); err != nil {
		t.Fatalf("UnmarshalYAML(): неожиданная ошибка: %v", err)
	}
	if got := Color(hc).Hex(); got != "#ff8000" {
		t.Errorf("UnmarshalYAML(): ожидался цвет #ff8000, получено %v", got)
	}
}

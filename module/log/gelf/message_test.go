package gelf

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessageMarshal(t *testing.T) {
	const (
		tplFatalUnexpectedError = "json.Marshal() вернул неожиданную ошибку: %v."
		tplFatalUnmarshalError  = "json.Unmarshal() вернул неожиданную ошибку: %v."
		tplErrorFieldValue      = "поле %q содержит неверное значение: %v."
		tplErrorFieldNotFound   = "поле %q не найдено в сообщении."
	)
	var (
		err  error
		msg  *Message
		buf  []byte
		item map[string]any
	)

	msg = NewMessage("test-host", "short message", 6)
	msg.Full = "full message"
	msg.Facility = "application"
	msg.Keys["key1"] = "value1"
	msg.Keys["_key2"] = "value2"
	if buf, err = json.Marshal(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = json.Unmarshal(buf, &item); err != nil {
		t.Fatalf(tplFatalUnmarshalError, err)
	}
	// Проверка стандартных полей сообщения.
	if item["version"] != "1.1" {
		t.Errorf(tplErrorFieldValue, "version", item["version"])
	}
	if item["host"] != "test-host" {
		t.Errorf(tplErrorFieldValue, "host", item["host"])
	}
	if item["short_message"] != "short message" {
		t.Errorf(tplErrorFieldValue, "short_message", item["short_message"])
	}
	if item["full_message"] != "full message" {
		t.Errorf(tplErrorFieldValue, "full_message", item["full_message"])
	}
	if item["level"] != float64(6) {
		t.Errorf(tplErrorFieldValue, "level", item["level"])
	}
	if item["facility"] != "application" {
		t.Errorf(tplErrorFieldValue, "facility", item["facility"])
	}
	if _, ok := item["timestamp"]; !ok {
		t.Errorf(tplErrorFieldNotFound, "timestamp")
	}
	// Проверка дополнительных полей с префиксом "_".
	if item["_key1"] != "value1" {
		t.Errorf(tplErrorFieldValue, "_key1", item["_key1"])
	}
	if _, ok := item["key1"]; ok {
		t.Errorf(tplErrorFieldValue, "key1", item["key1"])
	}
	// Дополнительное поле с уже существующим префиксом "_" не дублирует префикс.
	if item["_key2"] != "value2" {
		t.Errorf(tplErrorFieldValue, "_key2", item["_key2"])
	}
	if _, ok := item["__key2"]; ok {
		t.Errorf(tplErrorFieldValue, "__key2", item["__key2"])
	}
	// Сообщение без полного текста и категории не содержит поля full_message и facility.
	msg = NewMessage("test-host", "short message", 6)
	if buf, err = json.Marshal(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	item = nil
	if err = json.Unmarshal(buf, &item); err != nil {
		t.Fatalf(tplFatalUnmarshalError, err)
	}
	if _, ok := item["full_message"]; ok {
		t.Errorf(tplErrorFieldValue, "full_message", item["full_message"])
	}
	if _, ok := item["facility"]; ok {
		t.Errorf(tplErrorFieldValue, "facility", item["facility"])
	}
}

func TestSeverity(t *testing.T) {
	const (
		tplErrorSeverityValue = "константа %s содержит неверное значение %d, ожидалось %d."
	)
	type item struct {
		name string // Имя константы.
		got  int32  // Значение константы.
		want int32  // Ожидаемое значение константы.
	}
	var (
		items []item
		n     int
	)

	items = []item{
		{"SeverityEmergency", SeverityEmergency, 0},
		{"SeverityAlert", SeverityAlert, 1},
		{"SeverityCritical", SeverityCritical, 2},
		{"SeverityError", SeverityError, 3},
		{"SeverityWarning", SeverityWarning, 4},
		{"SeverityNotice", SeverityNotice, 5},
		{"SeverityInfo", SeverityInfo, 6},
		{"SeverityDebug", SeverityDebug, 7},
	}
	for n = range items {
		if items[n].got != items[n].want {
			t.Errorf(tplErrorSeverityValue, items[n].name, items[n].got, items[n].want)
		}
	}
}

func TestMessageMarshalValueString(t *testing.T) {
	const (
		tplFatalUnexpectedError = "json.Marshal() вернул неожиданную ошибку: %v."
		tplFatalUnmarshalError  = "json.Unmarshal() вернул неожиданную ошибку: %v."
		tplErrorFieldValue      = "поле %q содержит неверное значение: %v."
	)
	var (
		err  error
		ok   bool
		msg  *Message
		buf  []byte
		item map[string]any
		val  string
	)

	// Значение дополнительного поля недопустимого типа приводится к строке, ошибки кодирования не возникает.
	msg = NewMessage("test-host", "short message", 6)
	msg.Keys["fn"] = func() {}
	if buf, err = json.Marshal(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = json.Unmarshal(buf, &item); err != nil {
		t.Fatalf(tplFatalUnmarshalError, err)
	}
	if val, ok = item["_fn"].(string); !ok || val == "" {
		t.Errorf(tplErrorFieldValue, "_fn", item["_fn"])
	}
}

func TestMessageMarshalANSI(t *testing.T) {
	const (
		tplFatalUnexpectedError = "json.Marshal() вернул неожиданную ошибку: %v."
		tplFatalUnmarshalError  = "json.Unmarshal() вернул неожиданную ошибку: %v."
		tplErrorFieldValue      = "поле %q содержит неверное значение: %v."
	)
	var (
		err  error
		msg  *Message
		buf  []byte
		item map[string]any
	)

	msg = NewMessage("test-host", "\x1b[33mshort message\x1b[0m", 6)
	msg.Full = "\x1b[31mfull message\x1b[0m"
	if buf, err = json.Marshal(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = json.Unmarshal(buf, &item); err != nil {
		t.Fatalf(tplFatalUnmarshalError, err)
	}
	// Проверка очистки текстов сообщения от ANSI escape-последовательностей терминала.
	if item["short_message"] != "short message" {
		t.Errorf(tplErrorFieldValue, "short_message", item["short_message"])
	}
	if item["full_message"] != "full message" {
		t.Errorf(tplErrorFieldValue, "full_message", item["full_message"])
	}
}

func TestMessageMarshalSanitize(t *testing.T) {
	const (
		tplFatalUnexpectedError = "json.Marshal() вернул неожиданную ошибку: %v."
		tplFatalUnmarshalError  = "json.Unmarshal() вернул неожиданную ошибку: %v."
		tplErrorFieldValue      = "поле %q содержит неверное значение: %v."
		tplErrorFieldNotFound   = "поле %q не найдено в сообщении."
	)
	var (
		err      error
		ok       bool
		msg      *Message
		buf      []byte
		item     map[string]any
		expected map[string]any
		key      string
		val      string
	)

	msg = NewMessage("test-host", "short message", 6)
	msg.Keys["Auth-Method"] = "plain"
	msg.Keys["Client.IP"] = "127.0.0.1"
	msg.Keys["wd-mailauth"] = true
	msg.Keys["attempt"] = uint64(5)
	msg.Fields["app-Name"] = "wd-mailauth"
	msg.Fields["when"] = time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	msg.Fields["fn"] = func() {}
	if buf, err = json.Marshal(msg); err != nil {
		t.Fatalf(tplFatalUnexpectedError, err)
	}
	if err = json.Unmarshal(buf, &item); err != nil {
		t.Fatalf(tplFatalUnmarshalError, err)
	}
	// Имена полей приводятся к допустимому формату GELF, логические значения приводятся к строке.
	expected = map[string]any{
		"_Auth_Method": "plain",
		"_Client_IP":   "127.0.0.1",
		"_wd_mailauth": "true",
		"_attempt":     float64(5),
		"app_Name":     "wd-mailauth",
		"when":         "2026-11-25T12:00:00Z",
	}
	for key = range expected {
		if _, ok = item[key]; !ok {
			t.Errorf(tplErrorFieldNotFound, key)
			continue
		}
		if item[key] != expected[key] {
			t.Errorf(tplErrorFieldValue, key, item[key])
		}
	}
	// Значение поля верхнего уровня, которое не кодируется в формат JSON, приводится к строке.
	if val, ok = item["fn"].(string); !ok || val == "" {
		t.Errorf(tplErrorFieldValue, "fn", item["fn"])
	}
}

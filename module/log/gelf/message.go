package gelf

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/webnice/kit/v5/module/dye"
)

// NewMessage Конструктор сообщения GELF.
func NewMessage(host, short string, level int32) (ret *Message) {
	ret = &Message{
		Version: version,
		Host:    host,
		Short:   short,
		Time:    float64(time.Now().UnixNano()) / float64(time.Second),
		Level:   level,
		Fields:  make(map[string]any),
		Keys:    make(map[string]any),
	}

	return
}

// Приведение имени дополнительного поля к допустимому формату GELF: буквы латинского алфавита, цифры
// и символ подчёркивания, остальные символы заменяются на подчёркивание.
func sanitizeKeyName(key string) (ret string) {
	ret = strings.Map(func(ch rune) rune {
		switch {
		case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z', ch >= '0' && ch <= '9', ch == '_':
			return ch
		}

		return '_'
	}, key)

	return
}

// Приведение значения дополнительного поля к допустимому формату GELF: строки и числа передаются без
// изменений, значения остальных типов приводятся к строке.
func sanitizeKeyValue(val any) (ret any) {
	switch val.(type) {
	case string,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		ret = val
	default:
		ret = fmt.Sprint(val)
	}

	return
}

// Приведение значения поля сообщения верхнего уровня к кодированию в формат JSON: значения, которые не
// кодируются в формат JSON, приводятся к строке.
func sanitizeFieldValue(val any) (ret any) {
	if _, err := json.Marshal(val); err != nil {
		ret = fmt.Sprint(val)
		return
	}
	ret = val

	return
}

// MarshalJSON Кодирование сообщения в формат JSON с учётом дополнительных полей.
// Краткий и полный тексты сообщения очищаются от ANSI escape-последовательностей терминала.
func (msg *Message) MarshalJSON() (ret []byte, err error) {
	var (
		item map[string]any
		key  string
	)

	// Заполнение стандартных полей сообщения.
	item = make(map[string]any, len(msg.Keys)+7)
	item["version"] = msg.Version
	item["host"] = msg.Host
	// Краткий и полный тексты очищаются от ANSI escape-последовательностей терминала.
	item["short_message"] = dye.StripANSI(msg.Short)
	if msg.Full != "" {
		item["full_message"] = dye.StripANSI(msg.Full)
	}
	if msg.Facility != "" {
		item["facility"] = msg.Facility
	}
	item["timestamp"] = msg.Time
	item["level"] = msg.Level
	// Дополнительные поля сообщения верхнего уровня, имена полей приводятся к допустимому формату GELF.
	for key = range msg.Fields {
		item[sanitizeKeyName(key)] = sanitizeFieldValue(msg.Fields[key])
	}
	// Дополнительные поля сообщения с префиксом "_", имена и значения полей приводятся к допустимому
	// формату GELF.
	for key = range msg.Keys {
		item[keyPrefix+sanitizeKeyName(strings.TrimPrefix(key, keyPrefix))] = sanitizeKeyValue(msg.Keys[key])
	}
	// Кодирование сообщения в формат JSON.
	if ret, err = json.Marshal(item); err != nil {
		return
	}

	return
}

package gelf

import (
	"encoding/json"
	"strings"
	"time"
)

// NewMessage Конструктор сообщения GELF.
func NewMessage(host, short string, level int32) (ret *Message) {
	ret = &Message{
		Version: version,
		Host:    host,
		Short:   short,
		Time:    float64(time.Now().UnixNano()) / float64(time.Second),
		Level:   level,
		Keys:    make(map[string]any),
	}

	return
}

// MarshalJSON Кодирование сообщения в формат JSON с учётом дополнительных полей.
func (msg *Message) MarshalJSON() (ret []byte, err error) {
	var (
		item map[string]any
		key  string
	)

	// Заполнение стандартных полей сообщения.
	item = make(map[string]any, len(msg.Keys)+7)
	item["version"] = msg.Version
	item["host"] = msg.Host
	item["short_message"] = msg.Short
	if msg.Full != "" {
		item["full_message"] = msg.Full
	}
	if msg.Facility != "" {
		item["facility"] = msg.Facility
	}
	item["timestamp"] = msg.Time
	item["level"] = msg.Level
	// Дополнительные поля сообщения с префиксом "_".
	for key = range msg.Keys {
		item[keyPrefix+strings.TrimPrefix(key, keyPrefix)] = msg.Keys[key]
	}
	// Кодирование сообщения в формат JSON.
	if ret, err = json.Marshal(item); err != nil {
		return
	}

	return
}

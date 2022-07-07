// Package log
package log

import (
	"bytes"
	"runtime"
	"sync"
	"time"

	kitTypes "github.com/webnice/kit/v3/types"
)

// Инициализация бассейна объектов.
func (log *logger) messageInitPool() {
	log.messagePool = new(sync.Pool)
	log.messagePool.New = log.messageNew
}

// Конструктор объектов для бассейна объектов.
func (log *logger) messageNew() interface{} {
	var msg = &Message{
		Timestamp: time.Unix(0, 0),
		Level:     0,
		Pattern:   &bytes.Buffer{},
		Argument:  make([]interface{}, 0, 10),
		Keys:      make(map[string]interface{}),
		Trace:     kitTypes.NewTraceInfo(),
		Fatality:  false,
		Done:      make(chan struct{}),
	}
	runtime.SetFinalizer(msg, log.messageDestructor)

	return msg
}

// Деструктор объекта с закрытием канала обратной связи.
func (log *logger) messageDestructor(msg *Message) {
	defer func() { _ = recover() }()
	close(msg.Done)
}

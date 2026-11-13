package log

import (
	"bytes"
	standardLog "log"
	"strings"
	"testing"
	"time"

	kitModuleLogLevel "github.com/webnice/kit/v4/module/log/level"
)

// standardLogPrint Печать сообщения через стандартный пакет log.
func standardLogPrint(msg string) { standardLog.Print(msg) }

// log Тип для формирования имён функций содержащих log.fatal и log.panic.
// Полное имя методов: github.com/webnice/kit/v4/module/log.log.fatal и ...log.log.panic.
type log struct{}

func (l log) fatal(tl Logger) { l.write(tl) }

func (l log) panic(tl Logger) { l.write(tl) }

func (l log) write(tl Logger) { _, _ = tl.Write([]byte("уровень")) }

// messageSnapshot Копия сообщения лога для проверки значений после возврата объекта в бассейн.
// Функция вызывается только из обработчика сообщения, до возврата объекта в бассейн.
func messageSnapshot(msg *Message) (ret Message) {
	ret = *msg
	if msg.Pattern != nil {
		ret.Pattern = bytes.NewBufferString(msg.Pattern.String())
	}

	return
}

func TestWriteLevelNotice(t *testing.T) {
	var (
		tl      *logger
		receive chan *Message
		err     error
		n       int
	)

	tl, _, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf("функция Initialization(), вернула не ожидаемую ошибку: %s", err)
	}
	defer tl.FlushAndClose()
	receive = make(chan *Message, 2)
	if err = tl.HandlerSubscribe(func(m *Message) { s := messageSnapshot(m); receive <- &s },
		kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf("функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s", err)
	}
	// Вызов не из функций log.Fatal/log.Panic - уровень Notice по умолчанию.
	if n, err = tl.Write([]byte("сообщение notice")); err != nil || n != len("сообщение notice") {
		t.Fatalf("функция Write() = %d, %v", n, err)
	}
	msg := waitMessage(t, receive)
	if msg.Level != kitModuleLogLevel.Notice {
		t.Errorf("уровень сообщения %v, ожидался Notice", msg.Level)
	}
	if msg.Pattern.String() != "сообщение notice" {
		t.Errorf("шаблон сообщения %q, ожидалось %q", msg.Pattern.String(), "сообщение notice")
	}
	if msg.Timestamp.Location() != time.UTC {
		t.Errorf("временная зона метки времени %v, ожидался UTC", msg.Timestamp.Location())
	}
	if msg.Fatality {
		t.Error("флаг фатальности установлен для уровня Notice")
	}
	if msg.Trace == nil {
		t.Error("информация о вызове не заполнена")
	}
}

func TestWriteLevelCritical(t *testing.T) {
	var (
		tl      *logger
		receive chan *Message
		err     error
	)

	tl, _, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf("функция Initialization(), вернула не ожидаемую ошибку: %s", err)
	}
	defer tl.FlushAndClose()
	receive = make(chan *Message, 2)
	if err = tl.HandlerSubscribe(func(m *Message) { s := messageSnapshot(m); receive <- &s },
		kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf("функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s", err)
	}
	// Вызов из функции с именем содержащим log.panic - уровень Critical.
	log{}.panic(tl)
	msg := waitMessage(t, receive)
	if msg.Level != kitModuleLogLevel.Critical {
		t.Errorf("уровень сообщения %v, ожидался Critical", msg.Level)
	}
	if msg.Fatality {
		t.Error("флаг фатальности установлен для уровня Critical")
	}
}

func TestWriteTrimsSpaces(t *testing.T) {
	var (
		tl  *logger
		tsw *testSyncWriter
		n   int
		err error
	)

	tl, _, tsw = newTestLogger()
	if n, err = tl.Write([]byte("\t \nсообщение\r \n")); err != nil || n != len("\t \nсообщение\r \n") {
		t.Fatalf("функция Write() = %d, %v", n, err)
	}
	waitFor(t, func() bool { return tsw.String() == "сообщение\n" })
}

func TestWriteEmpty(t *testing.T) {
	var (
		tl  *logger
		tsw *testSyncWriter
		n   int
		err error
	)

	tl, _, tsw = newTestLogger()
	if n, err = tl.Write(nil); err != nil || n != 0 {
		t.Fatalf("функция Write(nil) = %d, %v, ожидалось 0, nil", n, err)
	}
	if tsw.Len() != 0 {
		t.Errorf("пустое сообщение напечатано в лог: %q", tsw.String())
	}
}

func TestWriteString(t *testing.T) {
	var (
		tl  *logger
		tsw *testSyncWriter
		n   int
		err error
	)

	tl, _, tsw = newTestLogger()
	if n, err = tl.WriteString("  строка сообщения  "); err != nil || n != len("  строка сообщения  ") {
		t.Fatalf("функция WriteString() = %d, %v", n, err)
	}
	waitFor(t, func() bool { return strings.Contains(tsw.String(), "строка сообщения") })
}

func TestWriteStringEmpty(t *testing.T) {
	var (
		tl *logger
		n  int
	)

	tl, _, _ = newTestLogger()
	// Пустая строка - сообщение обрабатывается, паники нет.
	if n, _ = tl.WriteString(""); n != 0 {
		t.Errorf("функция WriteString(пустая) = %d, ожидалось 0", n)
	}
}

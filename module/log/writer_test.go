package log

import (
	"bytes"
	standardLog "log"
	"strings"
	"testing"
	"time"

	kitModuleLogLevel "github.com/webnice/kit/v4/module/log/level"
)

// Печать сообщения через стандартный пакет log.
func standardLogPrint(msg string) { standardLog.Print(msg) }

// log Тип для формирования имён функций содержащих log.fatal и log.panic.
// Полное имя методов: github.com/webnice/kit/v4/module/log.log.fatal и ...log.log.panic.
type log struct{}

func (l log) fatal(tl Logger) { l.write(tl) }

func (l log) panic(tl Logger) { l.write(tl) }

func (l log) write(tl Logger) { _, _ = tl.Write([]byte("уровень")) }

// Копия сообщения лога для проверки значений после возврата объекта в бассейн.
// Функция вызывается только из обработчика сообщения, до возврата объекта в бассейн.
func messageSnapshot(msg *Message) (ret Message) {
	ret = *msg
	if msg.Pattern != nil {
		ret.Pattern = bytes.NewBufferString(msg.Pattern.String())
	}

	return
}

func TestWriteLevelNotice(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		tplFatalSubscribe      = "функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s."
		tplWriteWrong          = "функция Write() = %d, %v."
		tplLevelMismatch       = "уровень сообщения %v, ожидался Notice."
		tplPatternMismatch     = "шаблон сообщения %q, ожидалось %q."
		tplLocationMismatch    = "временная зона метки времени %v, ожидался UTC."
		errFatalitySet         = "флаг фатальности установлен для уровня Notice."
		errTraceNil            = "информация о вызове не заполнена."
		testMessage            = "сообщение notice"
	)
	var (
		tl      *logger
		receive chan *Message
		msg     *Message
		n       int
		err     error
	)

	tl, _, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	defer tl.FlushAndClose()
	receive = make(chan *Message, 2)
	if err = tl.HandlerSubscribe(func(m *Message) {
		var s Message

		s = messageSnapshot(m)
		receive <- &s
	}, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	// Вызов не из функций log.Fatal/log.Panic - уровень Notice по умолчанию.
	if n, err = tl.Write([]byte(testMessage)); err != nil || n != len(testMessage) {
		t.Fatalf(tplWriteWrong, n, err)
	}
	msg = waitMessage(t, receive)
	if msg.Level != kitModuleLogLevel.Notice {
		t.Errorf(tplLevelMismatch, msg.Level)
	}
	if msg.Pattern.String() != testMessage {
		t.Errorf(tplPatternMismatch, msg.Pattern.String(), testMessage)
	}
	if msg.Timestamp.Location() != time.UTC {
		t.Errorf(tplLocationMismatch, msg.Timestamp.Location())
	}
	if msg.Fatality {
		t.Error(errFatalitySet)
	}
	if msg.Trace == nil {
		t.Error(errTraceNil)
	}
}

func TestWriteLevelCritical(t *testing.T) {
	const (
		tplFatalInitialization = "функция Initialization(), вернула не ожидаемую ошибку: %s."
		tplFatalSubscribe      = "функция HandlerSubscribe(), вернула не ожидаемую ошибку: %s."
		tplLevelMismatch       = "уровень сообщения %v, ожидался Critical."
		errFatalitySet         = "флаг фатальности установлен для уровня Critical."
	)
	var (
		tl      *logger
		receive chan *Message
		msg     *Message
		err     error
	)

	tl, _, _ = newTestLogger()
	if err = tl.Initialization(); err != nil {
		t.Fatalf(tplFatalInitialization, err)
	}
	defer tl.FlushAndClose()
	receive = make(chan *Message, 2)
	if err = tl.HandlerSubscribe(func(m *Message) {
		var s Message

		s = messageSnapshot(m)
		receive <- &s
	}, kitModuleLogLevel.Fatal, kitModuleLogLevel.Trace); err != nil {
		t.Fatalf(tplFatalSubscribe, err)
	}
	// Вызов из функции с именем содержащим log.panic - уровень Critical.
	log{}.panic(tl)
	msg = waitMessage(t, receive)
	if msg.Level != kitModuleLogLevel.Critical {
		t.Errorf(tplLevelMismatch, msg.Level)
	}
	if msg.Fatality {
		t.Error(errFatalitySet)
	}
}

func TestWriteTrimsSpaces(t *testing.T) {
	const (
		tplWriteWrong    = "функция Write() = %d, %v."
		testMessageDirty = "\t \nсообщение\r \n"
		testMessageClean = "сообщение\n"
	)
	var (
		tl  *logger
		tsw *testSyncWriter
		n   int
		err error
	)

	tl, _, tsw = newTestLogger()
	if n, err = tl.Write([]byte(testMessageDirty)); err != nil || n != len(testMessageDirty) {
		t.Fatalf(tplWriteWrong, n, err)
	}
	waitFor(t, func() bool { return tsw.String() == testMessageClean })
}

func TestWriteEmpty(t *testing.T) {
	const (
		tplWriteNilWrong = "функция Write(nil) = %d, %v, ожидалось 0, nil."
		tplPrinted       = "пустое сообщение напечатано в лог: %q."
	)
	var (
		tl  *logger
		tsw *testSyncWriter
		n   int
		err error
	)

	tl, _, tsw = newTestLogger()
	if n, err = tl.Write(nil); err != nil || n != 0 {
		t.Fatalf(tplWriteNilWrong, n, err)
	}
	if tsw.Len() != 0 {
		t.Errorf(tplPrinted, tsw.String())
	}
}

func TestWriteString(t *testing.T) {
	const (
		tplWriteStringWrong = "функция WriteString() = %d, %v."
		testMessage         = "  строка сообщения  "
		testMessagePart     = "строка сообщения"
	)
	var (
		tl  *logger
		tsw *testSyncWriter
		n   int
		err error
	)

	tl, _, tsw = newTestLogger()
	if n, err = tl.WriteString(testMessage); err != nil || n != len(testMessage) {
		t.Fatalf(tplWriteStringWrong, n, err)
	}
	waitFor(t, func() bool { return strings.Contains(tsw.String(), testMessagePart) })
}

func TestWriteStringEmpty(t *testing.T) {
	const (
		tplWriteStringEmpty = "функция WriteString(пустая) = %d, ожидалось 0."
	)
	var (
		tl *logger
		n  int
	)

	tl, _, _ = newTestLogger()
	// Пустая строка - сообщение обрабатывается, паники нет.
	if n, _ = tl.WriteString(""); n != 0 {
		t.Errorf(tplWriteStringEmpty, n)
	}
}

package log

import "testing"

// TestMessageDestructor Деструктор объекта сообщения лога.
func TestMessageDestructor(t *testing.T) {
	const (
		errNewWrong      = "функция messageNew(), вернулся не корректный объект."
		errDoneNotClosed = "функция messageDestructor(), канал Done не закрыт."
	)
	var (
		tl  *logger
		msg *Message
	)

	tl, _, _ = newTestLogger()
	msg = tl.messageNew().(*Message)
	if msg == nil || msg.Done == nil {
		t.Fatal(errNewWrong)
	}
	// Деструктор закрывает канал.
	tl.messageDestructor(msg)
	select {
	case <-msg.Done:
	default:
		t.Error(errDoneNotClosed)
	}
	// Повторный вызов деструктора с закрытым каналом - паника перехвачена.
	tl.messageDestructor(msg)
}

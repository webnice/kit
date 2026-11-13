package log

import "testing"

// TestMessageDestructor Деструктор объекта сообщения лога.
func TestMessageDestructor(t *testing.T) {
	var (
		tl  *logger
		msg *Message
	)

	tl, _, _ = newTestLogger()
	msg = tl.messageNew().(*Message)
	if msg == nil || msg.Done == nil {
		t.Fatal("функция messageNew(), вернулся не корректный объект")
	}
	// Деструктор закрывает канал.
	tl.messageDestructor(msg)
	select {
	case <-msg.Done:
	default:
		t.Error("функция messageDestructor(), канал Done не закрыт")
	}
	// Повторный вызов деструктора с закрытым каналом - паника перехвачена.
	tl.messageDestructor(msg)
}

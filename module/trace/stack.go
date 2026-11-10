package trace

import "runtime"

// StackShort Загрузка короткого стека текущей горутины.
func StackShort() (ret []byte) { return stack(make([]byte, defaultBufferLength), false) }

// StackFull Загрузка полного стека вызовов всех горутин приложения.
func StackFull() (ret []byte) { return stack(make([]byte, defaultBufferLength), true) }

// Загрузка стека с увеличением буфера до тех пор, пока весь стек не поместится в буфер.
func stack(buf []byte, all bool) (ret []byte) {
	for {
		if n := runtime.Stack(buf, all); n < len(buf) {
			ret = buf[:n]
			break
		}
		buf = make([]byte, len(buf)*2)
	}

	return
}

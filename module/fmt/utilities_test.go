package fmt

import "testing"

func TestStripNumbers(t *testing.T) {
	const expected = "1234567890"

	src := "+_-)(*&^%$#@!±ЙЦУКЕНГШЩЗХЪФЫВАПРОЛДЖЭЁЯЧСМИТЬБЮ?йцукенгшщзхъфывапролджэёячсмитьбю/QWERTYUIOP{}ASDFGHJK" +
		"L:|ZXCVBNM<>?§1234567890-="
	rsp := stripNumbers(src)
	if rsp != expected {
		t.Errorf("stripNumbers(): ожидалось %q, получено %q", expected, rsp)
		return
	}
}

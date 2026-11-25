package dye

import (
	"testing"
)

// TestStripANSI Проверка очистки текста от ANSI escape-последовательностей терминала.
func TestStripANSI(t *testing.T) {
	const (
		tplErrorStripANSI = "StripANSI() тест %q для текста %q вернула значение %q, ожидалось %q."
	)
	type item struct {
		name string // Имя теста.
		text string // Исходный текст.
		want string // Ожидаемый результат очистки.
	}
	var (
		items []item
		ret   string
		n     int
	)

	items = []item{
		{"пустой текст", "", ""},
		{"текст без escape-последовательностей", "SELECT 1", "SELECT 1"},
		{"цвета текста",
			"sql:\"\x1b[33mSELECT 1\x1b[0m\", ошибка: \x1b[31mrecord not found\x1b[0m",
			"sql:\"SELECT 1\", ошибка: record not found"},
		{"цвет палитры ANSI256", "\x1b[38;5;196mred\x1b[0m", "red"},
		{"управление курсором", "\x1b[2K\x1b[1Gтекст", "текст"},
		{"двухбайтовая последовательность", "\x1bMтекст", "текст"},
		{"OSC с завершением BEL", "\x1b]0;заголовок\x07текст", "текст"},
		{"OSC гиперссылка", "\x1b]8;;https://example.com\x1b\\ссылка\x1b]8;;\x1b\\", "ссылка"},
		{"OSC без завершения", "\x1b]0;заголовок", ""},
		{"символ ESC в конце текста", "текст\x1b", "текст"},
		{"некорректная последовательность CSI", "\x1b[\x01текст", "\x01текст"},
		{"незавершённая последовательность CSI", "текст\x1b[33", "текст"},
	}
	for n = range items {
		if ret = StripANSI(items[n].text); ret != items[n].want {
			t.Errorf(tplErrorStripANSI, items[n].name, items[n].text, ret, items[n].want)
		}
	}
}

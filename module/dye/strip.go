package dye

import "strings"

// StripANSI Очистка текста от ANSI escape-последовательностей терминала:
// последовательности CSI (цвета, управление курсором), OSC (заголовки окон,
// гиперссылки) и двухбайтовых последовательностей.
func StripANSI(text string) (ret string) {
	var (
		buf strings.Builder
		end int
		i   int
	)

	// Быстрая проверка: текст без символа ESC не содержит escape-последовательностей.
	if strings.IndexByte(text, ansiEscape) < 0 {
		ret = text
		return
	}
	buf.Grow(len(text))
	for i = 0; i < len(text); i++ {
		// Копирование обычного символа текста.
		if text[i] != ansiEscape {
			buf.WriteByte(text[i])
			continue
		}
		// Символ ESC в конце текста не образует escape-последовательность и пропускается.
		if i+1 >= len(text) {
			break
		}
		switch text[i+1] {
		case ansiCSI:
			// Последовательность CSI: ESC [ байты параметров, промежуточные байты, финальный байт.
			for end = i + 2; end < len(text) && text[end] >= 0x30 && text[end] <= 0x3f; end++ {
			}
			for ; end < len(text) && text[end] >= 0x20 && text[end] <= 0x2f; end++ {
			}
			// Пропуск последовательности целиком, у некорректной последовательности
			// без финального байта сохраняется остаток текста.
			if end < len(text) && text[end] >= 0x40 && text[end] <= 0x7e {
				i = end
			} else {
				i = end - 1
			}
		case ansiOSC:
			// Последовательность OSC: ESC ] текст, завершается символом BEL
			// или последовательностью ST (ESC \).
			for end = i + 2; end < len(text); end++ {
				if text[end] == ansiBell {
					break
				}
				if text[end] == ansiEscape && end+1 < len(text) && text[end+1] == ansiST {
					break
				}
			}
			// Пропуск последовательности вместе с завершающим символом,
			// незавершённая последовательность пропускается до конца текста.
			switch {
			case end >= len(text):
				i = len(text)
			case text[end] == ansiBell:
				i = end
			default:
				i = end + 1
			}
		default:
			// Двухбайтовая escape-последовательность: ESC и управляющий символ.
			i++
		}
	}
	ret = buf.String()

	return
}

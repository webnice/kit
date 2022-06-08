// Package tpl
package tpl

const (
	chunkUnknown = iota // Кусок ещё не распознан.
	chunkDelete         // Кусок был обработан и больше не требуется.
	chunkText           // Кусок является текстом, выводится как есть.
	chunkData           // Кусок является тегом для вывода данных.
	chunkColor          // Кусок является тегом переключения цвета.
	chunkFormat         // Кусок является тегом форматирования.
)

const (
	tplSuffix        = "..."
	tplTimezoneError = "не найдена зона времени %q, ошибка вывода времени: %s"
	tplNewLine       = "\n"
)

// Определение последовательностей.
const (
	seqCSI               = "\x1b[" // Начало последовательности управления.
	seqOSC               = "\x1b]" // Команда операционной системы.
	seqResetSeq          = "0"     // Сброс всего.
	seqBoldSeq           = "1"     // Жирный.
	seqFaintSeq          = "2"     //
	seqItalicSeq         = "3"     // Курсив.
	seqUnderlineSeq      = "4"     // Подчёркнутый.
	seqBlinkSeq          = "5"     // Мигание.
	seqReverseSeq        = "7"     // Символ назад.
	seqCrossOutSeq       = "9"     // Зачёркнутый.
	seqOverlineSeq       = "53"    // Над линией.
	seqResetBoldSeq      = "21"    // Сбросить жирный.
	seqResetFaintSeq     = "22"    //
	seqResetItalicSeq    = "23"    // Сбросить курсив.
	seqResetUnderlineSeq = "24"    // Сбросить подчёркнутый.
	seqResetBlinkSeq     = "25"    // Сбросить мигание.
	seqResetReverseSeq   = "27"    // Сбросить символ назад.
	seqResetCrossOutSeq  = "29"    // Сбросить зачёркнутый.
)

// Тип куска шаблона.
type chunkType uint

// Интерфейс Stringer.
func (ct chunkType) String() (ret string) {
	switch ct {
	case chunkDelete:
		ret = "Удалённый."
	case chunkText:
		ret = "Текст."
	case chunkData:
		ret = "Данные."
	case chunkColor:
		ret = "Цвет."
	case chunkFormat:
		ret = "Формат."
	default:
		ret = "Неизвестно."
	}

	return
}

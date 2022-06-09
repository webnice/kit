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
	keyQuote         = "\\"
	tplParseFirst    = "перед созданием сессии, необходимо выполнить функцию Parse()"
	tplTimezoneError = "не найдена зона времени %q, ошибка вывода времени: %s"
	tplNewLine       = "\n"
	tplSuffix        = "..."
	tplPanic         = `Работа с шаблоном прервана паникой:` + "\n%v\n%s."
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

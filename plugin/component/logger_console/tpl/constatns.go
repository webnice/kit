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
	keyQuote     = "\\"
	keySep       = ":"
	keyLine      = "line"
	keyTimestamp = "timestamp"
	keyLevel     = "level"
	keyMessage   = "message"
	keyKeys      = "keys"
	keyDye       = "dye"
)

const (
	tplParseFirst    = "перед созданием сессии, необходимо выполнить функцию Parse()"
	tplTimezoneError = "не найдена зона времени %q, ошибка вывода времени: %s"
	tplNewLine       = "\n"
	tplSuffix        = "..."
	tplPanic         = "Работа шаблонизатора прервана паникой:\n%v\n%s."
	tplParsePanic    = "Обработка данных шаблонизатором прервана паникой:\n%v\n%s."
)

// Тип куска шаблона.
type chunkType uint

// Интерфейс Stringer.
func (ct chunkType) String() (ret string) {
	const (
		typeUnknown = "Неизвестно."
		typeDelete  = "Удалённый."
		typeText    = "Текст."
		typeData    = "Данные."
		typeColor   = "Цвет."
		typeFormat  = "Формат."
	)

	switch ct {
	case chunkDelete:
		ret = typeDelete
	case chunkText:
		ret = typeText
	case chunkData:
		ret = typeData
	case chunkColor:
		ret = typeColor
	case chunkFormat:
		ret = typeFormat
	default:
		ret = typeUnknown
	}

	return
}

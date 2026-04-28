package tgbot

// UpdatesIncomingType Тип входящих обновлений получаемых с сервера телеграм.
type UpdatesIncomingType string

// String Строковое представление константы.
func (uit UpdatesIncomingType) String() string { return string(uit) }

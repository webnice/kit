package migration

import "github.com/webnice/dic"

// Все ошибки определены как константы
const (
	cDatabaseIsNotInUse      = "База данных не используется или не настроена."
	cDatabaseUnexpectedError = "Неожиданная ошибка базы данных: %s."
)

// Error Структура справочника ошибок.
type Error struct {
	dic.Errors

	// DatabaseIsNotInUse База данных не используется или не настроена.
	DatabaseIsNotInUse dic.IError

	// DatabaseUnexpectedError Неожиданная ошибка базы данных: ...
	DatabaseUnexpectedError dic.IError
}

var errSingleton = &Error{
	Errors:                  dic.Error(),
	DatabaseIsNotInUse:      dic.NewError(cDatabaseIsNotInUse),
	DatabaseUnexpectedError: dic.NewError(cDatabaseUnexpectedError, "ошибка"),
}

// Errors Справочник ошибок.
func Errors() *Error { return errSingleton }

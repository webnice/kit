package sql

import "github.com/webnice/dic"

// Коды ошибок.
const (
	eConfigurationIsEmpty  uint = iota + 1 // 001
	eUnknownDatabaseDriver                 // 002
	eUsernameIsEmpty                       // 003
	eWrongConnectionType                   // 004
	eConnectError                          // 005
	eDriverUnImplemented                   // 006
	eApplyMigration                        // 007
	eUnknownDialect                        // 008
)

// Текстовые значения кодов ошибок на основном языке приложения.
const (
	cConfigurationIsEmpty  = "Конфигурация подключения к базе данных пустая."
	cUnknownDatabaseDriver = "Указан неизвестный или не поддерживаемый драйвер базы данных: %q."
	cUsernameIsEmpty       = "Не указано имя пользователя, для подключения к базе данных."
	cWrongConnectionType   = "Указан неизвестный или не поддерживаемый способ подключения к базе данных: %q."
	cConnectError          = "Подключение к базе данных завершилось ошибкой: %w."
	cDriverUnImplemented   = "Подключение к базе данных с помощью драйвера %q не создано."
	cApplyMigration        = "Применение новых миграций базы данных прервано ошибкой: %w."
	cUnknownDialect        = "Применение миграций базы данных, настройка диалекта %q прервано ошибкой: %w."
)

type Error struct {
	dic.Errors

	// ConfigurationIsEmpty Конфигурация подключения к базе данных пустая.
	ConfigurationIsEmpty dic.IError

	// UnknownDatabaseDriver Указан неизвестный или не поддерживаемый драйвер базы данных: ...
	UnknownDatabaseDriver dic.IError

	// UsernameIsEmpty Не указано имя пользователя, для подключения к базе данных.
	UsernameIsEmpty dic.IError

	// WrongConnectionType Указан неизвестный или не поддерживаемый способ подключения к базе данных: ...
	WrongConnectionType dic.IError

	// ConnectError Подключение к базе данных завершилось ошибкой: ...
	ConnectError dic.IError

	// DriverUnImplemented Подключение к базе данных с помощью драйвера ... не создано.
	DriverUnImplemented dic.IError

	// ApplyMigration Применение новых миграций базы данных прервано ошибкой: ...
	ApplyMigration dic.IError

	// UnknownDialect Применение миграций базы данных, настройка диалекта ... прервано ошибкой: ...
	UnknownDialect dic.IError
}

var (
	errSingleton = &Error{
		Errors:                dic.Error(),
		ConfigurationIsEmpty:  dic.NewError(cConfigurationIsEmpty).CodeU().Set(eConfigurationIsEmpty),
		UnknownDatabaseDriver: dic.NewError(cUnknownDatabaseDriver, "название драйвера").CodeU().Set(eUnknownDatabaseDriver),
		UsernameIsEmpty:       dic.NewError(cUsernameIsEmpty).CodeU().Set(eUsernameIsEmpty),
		WrongConnectionType:   dic.NewError(cWrongConnectionType, "способ подключения").CodeU().Set(eWrongConnectionType),
		ConnectError:          dic.NewError(cConnectError, "ошибка").CodeU().Set(eConnectError),
		DriverUnImplemented:   dic.NewError(cDriverUnImplemented, "название драйвера").CodeU().Set(eDriverUnImplemented),
		ApplyMigration:        dic.NewError(cApplyMigration, "ошибка").CodeU().Set(eApplyMigration),
		UnknownDialect:        dic.NewError(cUnknownDialect, "диалект", "ошибка").CodeU().Set(eUnknownDialect),
	}

	// Errors Справочник ошибок.
	Errors = func() *Error { return errSingleton }
)

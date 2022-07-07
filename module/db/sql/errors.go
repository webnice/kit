// Package sql
package sql

// Обычные ошибки
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
	cConfigurationIsEmpty  = `Конфигурация подключения к базе данных пустая.`
	cUnknownDatabaseDriver = `Указан неизвестный или не поддерживаемый драйвер базы данных: ` + "%q."
	cUsernameIsEmpty       = `Не указано имя пользователя, для подключения к базе данных.`
	cWrongConnectionType   = `Указан неизвестный или не поддерживаемый способ подключения к базе данных: ` + "%q."
	cConnectError          = `Подключение к базе данных завершилось ошибкой: ` + "%s."
	cDriverUnImplemented   = `Подключение к базе данных с помощью драйвера %q не создано.`
	cApplyMigration        = `Применение новых миграций базы данных прервано ошибкой: ` + "%s."
	cUnknownDialect        = `Применение миграций базы данных, настройка диалекта %q прервано ошибкой: ` + "%s."
)

// Константы указаны в объектах, адрес которых фиксирован всё время работы приложения.
// Это позволяет сравнивать ошибки между собой используя обычное сравнение "==", но сравнивать необходимо только
// якорь "Anchor()" объекта ошибки.
var (
	errSingleton             = &Error{}
	errConfigurationIsEmpty  = err{tpl: cConfigurationIsEmpty, code: eConfigurationIsEmpty}
	errUnknownDatabaseDriver = err{tpl: cUnknownDatabaseDriver, code: eUnknownDatabaseDriver}
	errUsernameIsEmpty       = err{tpl: cUsernameIsEmpty, code: eUsernameIsEmpty}
	errWrongConnectionType   = err{tpl: cWrongConnectionType, code: eWrongConnectionType}
	errConnectError          = err{tpl: cConnectError, code: eConnectError}
	errDriverUnImplemented   = err{tpl: cDriverUnImplemented, code: eDriverUnImplemented}
	errApplyMigration        = err{tpl: cApplyMigration, code: eApplyMigration}
	errUnknownDialect        = err{tpl: cUnknownDialect, code: eUnknownDialect}
)

// ERRORS: Реализация ошибок с возможностью сравнения ошибок между собой.

// ConfigurationIsEmpty Конфигурация подключения к базе данных пустая.
func (e *Error) ConfigurationIsEmpty(code uint) Err { return newErr(&errConfigurationIsEmpty, code) }

// UnknownDatabaseDriver Указан неизвестный или не поддерживаемый драйвер базы данных: ...
func (e *Error) UnknownDatabaseDriver(code uint, driver string) Err {
	return newErr(&errUnknownDatabaseDriver, code, driver)
}

// UsernameIsEmpty Не указано имя пользователя, для подключения к базе данных.
func (e *Error) UsernameIsEmpty(code uint) Err { return newErr(&errUsernameIsEmpty, code) }

// WrongConnectionType Указан неизвестный или не поддерживаемый способ подключения к базе данных: ...
func (e *Error) WrongConnectionType(code uint, connType string) Err {
	return newErr(&errWrongConnectionType, code, connType)
}

// ConnectError Подключение к базе данных завершилось ошибкой: ...
func (e *Error) ConnectError(code uint, err error) Err { return newErr(&errConnectError, code, err) }

// DriverUnImplemented Подключение к базе данных с помощью драйвера ... не создано.
func (e *Error) DriverUnImplemented(code uint, driver string) Err {
	return newErr(&errDriverUnImplemented, code, driver)
}

// ApplyMigration Применение новых миграций базы данных прервано ошибкой: ...
func (e *Error) ApplyMigration(code uint, err error) Err {
	return newErr(&errApplyMigration, code, err)
}

// UnknownDialect Применение миграций базы данных, настройка диалекта ... прервано ошибкой: ...
func (e *Error) UnknownDialect(code uint, dialect string, err error) Err {
	return newErr(&errUnknownDialect, code, dialect, err)
}

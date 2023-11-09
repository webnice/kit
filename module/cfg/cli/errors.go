package cli

// Все ошибки определены как константы.
const (
	cUnexpectedError    = "Неожиданная ошибка."
	cHelpDisplayed      = "Отображение помощи по командам, аргументам и флагам командной строки."
	cRequiredCommand    = "Не указана обязательная команда командной строки."
	cRequiredFlag       = "Не указан обязательный флаг командной строки."
	cUnknownCommand     = "Неизвестная команда командной строки."
	cUnknownArgument    = "Неизвестный аргумент командной строки."
	cNotCorrectArgument = "Не верное значение или тип аргумента, флага или параметра."
)

// Константы указаны в объектах, адрес которых фиксирован всё время работы приложения.
// Ошибку с ошибкой можно сравнивать по телу, по адресу и т.п.
var (
	errSingleton          = &Error{}
	errUnexpectedError    = err(cUnexpectedError)
	errHelpDisplayed      = err(cHelpDisplayed)
	errRequiredCommand    = err(cRequiredCommand)
	errRequiredFlag       = err(cRequiredFlag)
	errUnknownCommand     = err(cUnknownCommand)
	errUnknownArgument    = err(cUnknownArgument)
	errNotCorrectArgument = err(cNotCorrectArgument)
)

type (
	// Error object of package.
	Error struct{}
	err   string
)

// Error The error built-in interface implementation.
func (e err) Error() string { return string(e) }

// Errors Справочник всех ошибок пакета.
func Errors() *Error { return errSingleton }

// ERRORS:

// UnexpectedError Неожиданная ошибка.
func (e *Error) UnexpectedError() error { return &errUnexpectedError }

// HelpDisplayed Отображение помощи по командам, аргументам и флагам командной строки.
func (e *Error) HelpDisplayed() error { return &errHelpDisplayed }

// RequiredCommand Не указана обязательная команда командной строки.
func (e *Error) RequiredCommand() error { return &errRequiredCommand }

// RequiredFlag Не указан обязательный флаг командной строки.
func (e *Error) RequiredFlag() error { return &errRequiredFlag }

// UnknownCommand Неизвестная команда командной строки.
func (e *Error) UnknownCommand() error { return &errUnknownCommand }

// UnknownArgument Неизвестный аргумент командной строки.
func (e *Error) UnknownArgument() error { return &errUnknownArgument }

// NotCorrectArgument Не верное значение или тип аргумента, флага или параметра.
func (e *Error) NotCorrectArgument() error { return &errNotCorrectArgument }

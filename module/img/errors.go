package img

// Все ошибки определены как константы.
const (
	cNotFound = "Not found"
)

// Константы указаны в объектах, адрес которых фиксирован всё время работы приложения.
// Ошибку с ошибкой можно сравнивать по телу, по адресу и т.п.
var (
	errSingleton = &Error{}
	errNotFound  = err(cNotFound)
)

type (
	// Error Объект ошибки.
	Error struct{}
	err   string
)

// Error Представление ошибки в качестве строки.
func (e err) Error() string { return string(e) }

// Errors Справочник ошибок.
func Errors() *Error { return errSingleton }

// ОШИБКИ

// NotFound Not found.
func (e *Error) NotFound() error { return &errNotFound }

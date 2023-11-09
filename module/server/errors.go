package server

// Все ошибки определены как константы. Коды ошибок приложения:

// Обычные ошибки
const (
	eUuidError uint8 = iota + 1
	eServerWithUuidNotFound
)

// Текстовые значения кодов ошибок на основном языке приложения.
const (
	cUuidError              = `Передан ошибочный UUID: %q.`
	cServerWithUuidNotFound = `Сервер с UUID=%q не найден.`
)

// Константы указаны в объектах, адрес которых фиксирован всё время работы приложения.
// Это позволяет сравнивать ошибки между собой используя обычное сравнение "==", но сравнивать необходимо только якорь "Anchor()" объекта ошибки.
var (
	errSingleton              = &Error{}
	errUuidError              = err{tpl: cUuidError, code: eUuidError}
	errServerWithUuidNotFound = err{tpl: cServerWithUuidNotFound, code: eServerWithUuidNotFound}
)

// ERRORS: Реализация ошибок с возможностью сравнения ошибок между собой.

// UuidError Передан ошибочный UUID: ...
func (e *Error) UuidError(uuid string) Err {
	return newErr(&errUuidError, 0, uuid)
}

// ServerWithUuidNotFound Сервер с UUID=... не найден.
func (e *Error) ServerWithUuidNotFound(uuid string) Err {
	return newErr(&errServerWithUuidNotFound, 0, uuid)
}

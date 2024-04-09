package server

// Все ошибки определены как константы. Коды ошибок приложения:

// Обычные ошибки
const (
	eUuidError uint8 = iota + 1
	eServerWithUuidNotFound
	eTypeNotImplemented
	eServerIsStarted
)

// Текстовые значения кодов ошибок на основном языке приложения.
const (
	cUuidError              = `Передан ошибочный UUID: %q.`
	cServerWithUuidNotFound = `Сервер с UUID=%q не найден.`
	cTypeNotImplemented     = "Тип %s не реализован."
	cServerIsStarted        = "Один или несколько серверов выполняется. Удаление запущенных серверов не возможно. Переданы идентификаторы: %v."
)

// Константы указаны в объектах, адрес которых фиксирован всё время работы приложения.
// Это позволяет сравнивать ошибки между собой используя обычное сравнение "==", но сравнивать необходимо только якорь "Anchor()" объекта ошибки.
var (
	errSingleton              = &Error{}
	errUuidError              = err{tpl: cUuidError, code: eUuidError}
	errServerWithUuidNotFound = err{tpl: cServerWithUuidNotFound, code: eServerWithUuidNotFound}
	errTypeNotImplemented     = err{tpl: cTypeNotImplemented, code: eTypeNotImplemented}
	errServerIsStarted        = err{tpl: cServerIsStarted, code: eServerIsStarted}
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

// TypeNotImplemented Тип ... не реализован.
func (e *Error) TypeNotImplemented(t string) Err {
	return newErr(&errTypeNotImplemented, 0, t)
}

// ServerIsStarted Один или несколько серверов выполняется. Удаление запущенных серверов не возможно. Переданы
// идентификаторы: ...
func (e *Error) ServerIsStarted(IDs ...string) Err {
	return newErr(&errServerIsStarted, 0, IDs)
}

package server

// Все ошибки определены как константы. Коды ошибок приложения:

// Обычные ошибки
const (
	eServerByIdNotFound uint8 = iota + 1
	eServersAddedNotEqualLaunched
	eModulePanicException
	eBeforeExitWithError
	eServerByIdAlreadyStarted
	eServerByIdNotStarted
)

// Текстовые значения кодов ошибок на основном языке приложения.
const (
	cServerByIdNotFound           = "Конфигурация сервера с идентификатором %q не найдена."
	cServersAddedNotEqualLaunched = "Запуск серверов завершился ошибкой, добавлено серверов: %d, запущено серверов: %d."
	cModulePanicException         = "Выполнение модуля %q прервано паникой:\n%v\n%s."
	cBeforeExitWithError          = "Функция Before() завершилась с ошибкой: %s"
	cServerByIdAlreadyStarted     = "Сервер с идентификатором %q уже запущен."
	cServerByIdNotStarted         = "Сервер с идентификатором %q не был запущен."
)

// Константы указаны в объектах, адрес которых фиксирован всё время работы приложения.
// Это позволяет сравнивать ошибки между собой используя обычное сравнение "==", но сравнивать необходимо только якорь "Anchor()" объекта ошибки.
var (
	errSingleton                    = &Error{}
	errServerByIdNotFound           = err{tpl: cServerByIdNotFound, code: eServerByIdNotFound}
	errServersAddedNotEqualLaunched = err{tpl: cServersAddedNotEqualLaunched, code: eServersAddedNotEqualLaunched}
	errModulePanicException         = err{tpl: cModulePanicException, code: eModulePanicException}
	errBeforeExitWithError          = err{tpl: cBeforeExitWithError, code: eBeforeExitWithError}
	errServerByIdAlreadyStarted     = err{tpl: cServerByIdAlreadyStarted, code: eServerByIdAlreadyStarted}
	errServerByIdNotStarted         = err{tpl: cServerByIdNotStarted, code: eServerByIdNotStarted}
)

// ERRORS: Реализация ошибок с возможностью сравнения ошибок между собой.

// ServerByIdNotFound Конфигурация сервера с идентификатором ... не найдена.
func (e *Error) ServerByIdNotFound(serverID string) Err {
	return newErr(&errServerByIdNotFound, 0, serverID)
}

// ServersAddedNotEqualLaunched Добавлено ... серверов, запущено ... серверов.
func (e *Error) ServersAddedNotEqualLaunched(added, launched uint64) Err {
	return newErr(&errServersAddedNotEqualLaunched, 0, added, launched)
}

// ModulePanicException Выполнение модуля ... прервано паникой: ...
func (e *Error) ModulePanicException(componentName string, err any, stack []byte) Err {
	return newErr(&errModulePanicException, 0, componentName, err, string(stack))
}

// BeforeExitWithError Функция Before() завершилась с ошибкой: ...
func (e *Error) BeforeExitWithError(err error) Err { return newErr(&errBeforeExitWithError, 0, err) }

// ServerByIdAlreadyStarted Сервер с идентификатором ... уже запущен.
func (e *Error) ServerByIdAlreadyStarted(serverID string) Err {
	return newErr(&errServerByIdAlreadyStarted, 0, serverID)
}

// ServerByIdNotStarted Сервер с идентификатором ... не был запущен.
func (e *Error) ServerByIdNotStarted(serverID string) Err {
	return newErr(&errServerByIdNotStarted, 0, serverID)
}

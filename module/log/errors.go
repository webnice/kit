package log

// Обычные ошибки
const (
	eLogPanicException           uint8 = iota + 1 // 001
	eHandlerAlreadySubscribed                     // 002
	eHandlerSubscriptionNotFound                  // 003
)

// Текстовые значения кодов ошибок на основном языке приложения.
const (
	cLogPanicException           = `Вызов обработчика сообщений лога %q прервана паникой:` + "\n%v\n%s."
	cHandlerAlreadySubscribed    = "Обработчик логов %q уже зарегистрирован."
	cHandlerSubscriptionNotFound = "Регистрация обработчика логов %q не найдена."
)

// Константы указаны в объектах, адрес которых фиксирован всё время работы приложения.
// Это позволяет сравнивать ошибки между собой используя обычное сравнение "==", но сравнивать необходимо только
// якорь "Anchor()" объекта ошибки.
var (
	errSingleton                   = &Error{}
	errLogPanicException           = err{tpl: cLogPanicException, code: eLogPanicException}
	errHandlerAlreadySubscribed    = err{tpl: cHandlerAlreadySubscribed, code: eHandlerAlreadySubscribed}
	errHandlerSubscriptionNotFound = err{tpl: cHandlerSubscriptionNotFound, code: eHandlerSubscriptionNotFound}
)

// ERRORS: Реализация ошибок с возможностью сравнения ошибок между собой.

// LogPanicException Вызов обработчика сообщений лога ... прервана паникой: ... ....
func (e *Error) LogPanicException(code uint8, subscriberName, err interface{}, stack []byte) Err {
	return newErr(&errLogPanicException, code, subscriberName, err, string(stack))
}

// HandlerAlreadySubscribed Обработчик логов ... уже зарегистрирован.
func (e *Error) HandlerAlreadySubscribed(code uint8, subscriberName string) Err {
	return newErr(&errHandlerAlreadySubscribed, code, subscriberName)
}

// HandlerSubscriptionNotFound Регистрация обработчика логов ... не найдена.
func (e *Error) HandlerSubscriptionNotFound(code uint8, subscriberName string) Err {
	return newErr(&errHandlerSubscriptionNotFound, code, subscriberName)
}

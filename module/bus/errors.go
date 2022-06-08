// Package bus
package bus

// Обычные ошибки
const (
	eDatabusRecursivePointer      uint8 = iota + 1 // 001
	eDatabusPanicException                         // 002
	eDatabusSubscribeNotFound                      // 003
	eDatabusInternalError                          // 004
	eDatabusNotSubscribersForType                  // 005
	eDatabusObjectIsNil                            // 006
	eDatabusPoolInternalError                      // 007
)

// Текстовые значения кодов ошибок на основном языке приложения.
const (
	cDatabusRecursivePointer      = `Не возможно определить тип рекурсивного указателя: ` + "%q."
	cDatabusPanicException        = `Работа с подпиской потребителя, в шине данных, прервана паникой:` + "\n%v\n%s."
	cDatabusSubscribeNotFound     = `Потребитель данных %q не был подписан на шину данных.`
	cDatabusInternalError         = `Внутренняя ошибка шины данных: ` + "%s."
	cDatabusNotSubscribersForType = `Отсутствуют потребители данных для типа данных: ` + "%q."
	cDatabusObjectIsNil           = `Передан nil объект.`
	cDatabusPoolInternalError     = `Бассейн объектов вернул не корректный объект.`
)

// Константы указаны в объектах, адрес которых фиксирован всё время работы приложения.
// Это позволяет сравнивать ошибки между собой используя обычное сравнение "==", но сравнивать необходимо только
// якорь "Anchor()" объекта ошибки.
var (
	errSingleton                    = &Error{}
	errDatabusRecursivePointer      = err{tpl: cDatabusRecursivePointer, code: eDatabusRecursivePointer}
	errDatabusPanicException        = err{tpl: cDatabusPanicException, code: eDatabusPanicException}
	errDatabusSubscribeNotFound     = err{tpl: cDatabusSubscribeNotFound, code: eDatabusSubscribeNotFound}
	errDatabusInternalError         = err{tpl: cDatabusInternalError, code: eDatabusInternalError}
	errDatabusNotSubscribersForType = err{tpl: cDatabusNotSubscribersForType, code: eDatabusNotSubscribersForType}
	errDatabusObjectIsNil           = err{tpl: cDatabusObjectIsNil, code: eDatabusObjectIsNil}
	errDatabusPoolInternalError     = err{tpl: cDatabusPoolInternalError, code: eDatabusPoolInternalError}
)

// ERRORS: Реализация ошибок с возможностью сравнения ошибок между собой.

// DatabusRecursivePointer Не возможно определить тип рекурсивного указателя: ...
func (e *Error) DatabusRecursivePointer(code uint8, pointer string) Err {
	return newErr(&errDatabusRecursivePointer, code, pointer)
}

// DatabusPanicException Работа с подпиской потребителя, в шине данных, прервана паникой: ... ....
func (e *Error) DatabusPanicException(code uint8, err interface{}, stack []byte) Err {
	return newErr(&errDatabusPanicException, code, err, string(stack))
}

// DatabusSubscribeNotFound Потребитель данных ... не был подписан на шину данных.
func (e *Error) DatabusSubscribeNotFound(code uint8, databuserName string) Err {
	return newErr(&errDatabusSubscribeNotFound, code, databuserName)
}

// DatabusInternalError Внутренняя ошибка шины данных: ...
func (e *Error) DatabusInternalError(code uint8, err error) Err {
	return newErr(&errDatabusInternalError, code, err)
}

// DatabusNotSubscribersForType Отсутствуют потребители данных для типа данных: ...
func (e *Error) DatabusNotSubscribersForType(code uint8, typeName string) Err {
	return newErr(&errDatabusNotSubscribersForType, code, typeName)
}

// DatabusObjectIsNil Передан nil объект.
func (e *Error) DatabusObjectIsNil(code uint8) Err {
	return newErr(&errDatabusObjectIsNil, code)
}

// DatabusPoolInternalError Бассейн объектов вернул не корректный объект.
func (e *Error) DatabusPoolInternalError(code uint8) Err {
	return newErr(&errDatabusPoolInternalError, code)
}

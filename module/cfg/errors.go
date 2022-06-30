// nolint: lll

// Package cfg
package cfg

import (
	"bytes"
	"strconv"
	"strings"
	"time"
)

// Все ошибки определены как константы. Коды ошибок приложения:

// Особенные ошибки
const (
	eApplicationPanicException uint8 = 255 // 255 - приложение завершилось из-за прерывания по исключению - panic.
	eApplicationUnknownError   uint8 = 254 // 254 - неожиданная ошибка приложения.
	eApplicationHelpDisplayed  uint8 = 253 // 253 - приложение завершилось корректно с отображением помощи по CLI.
	eApplicationfatality       uint8 = 252 // 252 - приложение завершилось из за печати в лог сообщения с уровнем Fatal.
)

// Обычные ошибки
const (
	eApplicationVersion                        uint8 = iota + 1 // 001
	eApplicationMainFuncNotFound                                // 002
	eApplicationMainFuncAlreadyRegistered                       // 003
	eApplicationRegistrationUnknownObject                       // 004
	eComponentIsNull                                            // 005
	eComponentRegistrationProhibited                            // 006
	eComponentRegistrationError                                 // 007
	eComponentPreferencesCallBeforeCompleting                   // 008
	eComponentPanicException                                    // 009
	eComponentRunlevelError                                     // 010
	eComponentRulesError                                        // 011
	eRunlevelCantLessCurrentLevel                               // 012
	eInitLogging                                                // 013
	eComponentConflict                                          // 014
	eComponentRequires                                          // 015
	eComponentInitiateTimeout                                   // 016
	eComponentInitiateExecution                                 // 017
	eComponentInitiatePanicException                            // 018
	eComponentDoExecution                                       // 019
	eComponentDoPanicException                                  // 020
	eComponentDoUnknownError                                    // 021
	eComponentFinalizeExecution                                 // 022
	eComponentFinalizePanicException                            // 023
	eComponentFinalizeUnknownError                              // 024
	eComponentFinalizeWarning                                   // 025
	eRunlevelSubscribeUnsubscribeNilFunction                    // 026
	eRunlevelAlreadySubscribedFunction                          // 027
	eRunlevelSubscriptionNotFound                               // 028
	eRunlevelSubscriptionPanicException                         // 029
	eCommandLineArgumentRequired                                // 030
	eCommandLineArgumentUnknown                                 // 031
	eCommandLineArgumentNotCorrect                              // 032
	eCommandLineRequiredFlag                                    // 033
	eCommandLineUnexpectedError                                 // 034
	eConfigurationBootstrap                                     // 035
	eGetCurrentUser                                             // 036
	eCantChangeWorkDirectory                                    // 037
	ePidExistsAnotherProcessOfApplication                       // 038
	ePidFileError                                               // 039
	eDatabusRecursivePointer                                    // 040
	eDatabusPanicException                                      // 041
	eDatabusSubscribeNotFound                                   // 042
	eDatabusInternalError                                       // 043
	eDatabusNotSubscribersForType                               // 044
	eDatabusObjectIsNil                                         // 045
	eConfigurationApplicationProhibited                         // 046
	eConfigurationApplicationObject                             // 047
	eConfigurationApplicationPanic                              // 048
	eConfigurationFileNotFound                                  // 049
	eConfigurationPermissionDenied                              // 050
	eConfigurationUnexpectedMistakeFileAccess                   // 051
	eConfigurationFileIsDirectory                               // 052
	eConfigurationFileReadingError                              // 053
	eConfigurationSetDefault                                    // 054
	eConfigurationSetDefaultValue                               // 055
	eConfigurationSetDefaultPanic                               // 056
	eConfigurationObjectNotFound                                // 057
	eConfigurationObjectIsNotStructure                          // 058
	eConfigurationObjectIsNil                                   // 059
	eConfigurationObjectIsNotValid                              // 060
	eConfigurationObjectIsNotAddress                            // 061
	eConfigurationObjectCopy                                    // 062
	eConfigurationCallbackAlreadyRegistered                     // 063
	eConfigurationCallbackSubscriptionNotFound                  // 064
	//eCantCreateWorkers                                     //
	//eCantStartWorkers                                      //
	//eAllWorkersStopWithError                               //
	//eCantOpenSocket                                        //
)

// Текстовые значения кодов ошибок на основном языке приложения.
const (
	cApplicationPanicException                 = `Выполнение приложения прервано паникой:` + "\n%v\n%s."
	cApplicationUnknownError                   = "%s"
	cApplicationHelpDisplayed                  = "%s"
	cApplicationVersion                        = `Версии приложения содержит ошибку: ` + "%s."
	cApplicationMainFuncNotFound               = `Не определена основная функция приложения.`
	cApplicationMainFuncAlreadyRegistered      = `Основная функция приложения уже зарегистрирована.`
	cApplicationRegistrationUnknownObject      = `Регистрация не известного компонента, объекта или модуля: ` + "%q."
	cComponentIsNull                           = `В качестве объекта компоненты передан nil.`
	cComponentRegistrationProhibited           = `Регистрация компонентов запрещена. Компонента %q не зарегистрирована.`
	cComponentRegistrationError                = `Регистрация компоненты %q завершилась ошибкой: %s.`
	cComponentPreferencesCallBeforeCompleting  = `Опрос настроек компонентов вызван до завершения регистрации компонентов.`
	cComponentPanicException                   = `Выполнение компоненты ` + "%q" + ` прервано паникой:` + "\n%v\n%s."
	cComponentRunlevelError                    = `Уровень запуска (runlevel), для компоненты %q указан %d, необходимо указать уровень равный 0, либо в интервале от 10 до 65534 включительно.`
	cComponentRulesError                       = `Правила %q для компоненты %q содержат ошибку: ` + "%s."
	cRunlevelCantLessCurrentLevel              = `Новый уровень работы приложения (%d) не может быть меньше текущего уровня работы приложения (%d).`
	cInitLogging                               = `Критическая ошибка в модуле менеджера логирования: ` + "%s."
	cComponentConflict                         = `Компонента ` + "%q" + ` конфликтует с компонентой ` + "%q."
	cComponentRequires                         = `Компонента ` + "%q" + ` имеет не удовлетворённую зависимость ` + "%q."
	cComponentInitiateTimeout                  = `Превышено время ожидание выполнения функции Initiate() компоненты ` + "%q."
	cComponentInitiateExecution                = `Выполнение функции Initiate() компоненты ` + "%q" + ` завершено с ошибкой: ` + "%s."
	cComponentInitiatePanicException           = `Выполнение функции Initiate() компоненты ` + "%q" + ` прервано паникой:` + "\n%v\n%s."
	cComponentDoExecution                      = `Выполнение функции Do() компоненты ` + "%q" + ` завершено с ошибкой: ` + "%s."
	cComponentDoPanicException                 = `Выполнение функции Do() компоненты ` + "%q" + ` прервано паникой:` + "\n%v\n%s."
	cComponentDoUnknownError                   = `Выполнение функций Do() завершилось ошибкой: ` + "%s."
	cComponentFinalizeExecution                = `Выполнение функции Finalize() компоненты ` + "%q" + ` завершено с ошибкой: ` + "%s."
	cComponentFinalizePanicException           = `Выполнение функции Finalize() компоненты ` + "%q" + ` прервано паникой:` + "\n%v\n%s."
	cComponentFinalizeUnknownError             = `Выполнение функций Finalize() завершилось ошибкой: ` + "%s."
	cComponentFinalizeWarning                  = `Выполнение функций Finalize() компоненты ` + "%q" + `, длится дольше отведённого времени (` + "%s" + `).`
	cRunlevelSubscribeUnsubscribeNilFunction   = `Передана nil функция, подписка или отписка nil функции не возможна.`
	cRunlevelAlreadySubscribedFunction         = `Функция ` + "%q" + ` уже подписана на получение событий изменения уровня работы приложения.`
	cRunlevelSubscriptionNotFound              = `Не найдена подписка функции ` + "%q" + ` на события изменения уровня работы приложения.`
	cRunlevelSubscriptionPanicException        = `Вызов функции подписчика на событие изменения уровня работы приложения, прервано паникой:` + "\n%v\n%s."
	cCommandLineArgumentRequired               = `Требуется указать обязательную команду, аргумент или флаг командной строки: ` + "%s."
	cCommandLineArgumentUnknown                = `Неизвестная команда, аргумент или флаг командной строки: ` + "%s."
	cCommandLineArgumentNotCorrect             = `Не верное значение или тип аргумента, флага или параметра: ` + "%s."
	cCommandLineRequiredFlag                   = `Не указан один или несколько обязательных флагов: ` + "%s."
	cCommandLineUnexpectedError                = `Не предвиденная ошибка библиотеки командного интерфейса приложения: ` + "%s; %s."
	cConfigurationBootstrap                    = `Ошибка начально bootstrap конфигурации приложения: ` + "%s."
	cGetCurrentUser                            = `Не удалось загрузить данные о текущем пользователе операционной системы: ` + "%s."
	cCantChangeWorkDirectory                   = `Не удалось сменить рабочую директорию приложения: ` + "%s."
	cPidExistsAnotherProcessOfApplication      = `Существует один или несколько работающих процессов приложения, измените PID файл или остановите экземпляры приложения, PID: ` + "%s."
	cPidFileError                              = `Ошибка работы с PID файлом %q: ` + "%s."
	cDatabusRecursivePointer                   = `Не возможно определить тип рекурсивного указателя: ` + "%q."
	cDatabusPanicException                     = `Работа с подпиской потребителя, в шине данных, прервана паникой:` + "\n%v\n%s."
	cDatabusSubscribeNotFound                  = `Потребитель данных %q не был подписан на шину данных.`
	cDatabusInternalError                      = `Внутренняя ошибка шины данных: ` + "%s."
	cDatabusNotSubscribersForType              = `Отсутствуют потребители данных для типа данных: ` + "%q."
	cDatabusObjectIsNil                        = `Передан nil объект.`
	cConfigurationApplicationProhibited        = `Регистрация объектов конфигурации на текущем уровне работы приложения запрещена. Конфигурация %q не зарегистрирована.`
	cConfigurationApplicationObject            = `Объект конфигурации приложения содержит ошибку: ` + "%s."
	cConfigurationApplicationPanic             = `Непредвиденная ошибка при регистрации объекта конфигурации.` + "\nПаника: %v.\n%s"
	cConfigurationFileNotFound                 = `Указанного файла конфигурации ` + "%q" + ` не существует: ` + "%s."
	cConfigurationPermissionDenied             = `Отсутствует доступ к файлу конфигурации, ошибка: ` + "%s."
	cConfigurationUnexpectedMistakeFileAccess  = `Неожиданная ошибка доступа к файлу конфигурации ` + "%q: %s."
	cConfigurationFileIsDirectory              = `В качестве файла конфигурации указана директория: ` + "%s."
	cConfigurationFileReadingError             = `Чтение фала конфигурации ` + "%q" + ` прервано ошибкой: ` + "%s."
	cConfigurationSetDefault                   = `Установка значений по умолчанию, для переменных конфигурации, прервана ошибкой: ` + "%s."
	cConfigurationSetDefaultValue              = `Установка значения по умолчанию %q, для переменной конфигурации %q, прервана ошибкой: ` + "%s."
	cConfigurationSetDefaultPanic              = `Непредвиденная ошибка, при установке значений по умолчанию, объекта конфигурации.` + "\nПаника: %v.\n%s"
	cConfigurationObjectNotFound               = `Объект конфигурации с типом %q не найден.`
	cConfigurationObjectIsNotStructure         = `Переданный объект %q не является структурой.`
	cConfigurationObjectIsNil                  = `Переданный объект, является nil объектом.`
	cConfigurationObjectIsNotValid             = `Объект конфигурации с типом %q не инициализирован.`
	cConfigurationObjectIsNotAddress           = `Объект конфигурации с типом %q передан не корректно. Необходимо передать адрес объекта.`
	cConfigurationObjectCopy                   = `Копирование объекта конфигурации с типом %q прервано ошибкой: ` + "%s."
	cConfigurationCallbackAlreadyRegistered    = `Подписка функции обратного вызова на изменение конфигурации с типом %q для функции %q уже существует.`
	cConfigurationCallbackSubscriptionNotFound = `Подписка функции обратного вызова на изменение конфигурации с типом %q для функции %q не существует.`
)

// Константы указаны в объектах, адрес которых фиксирован всё время работы приложения.
// Это позволяет сравнивать ошибки между собой используя обычное сравнение "==", но сравнивать необходимо только якорь "Anchor()" объекта ошибки.
var (
	errSingleton                                 = &Error{}
	errApplicationPanicException                 = err{tpl: cApplicationPanicException, code: eApplicationPanicException}
	errApplicationUnknownError                   = err{tpl: cApplicationUnknownError, code: eApplicationUnknownError}
	errApplicationHelpDisplayed                  = err{tpl: cApplicationHelpDisplayed, code: eApplicationHelpDisplayed}
	errApplicationVersion                        = err{tpl: cApplicationVersion, code: eApplicationVersion}
	errApplicationMainFuncNotFound               = err{tpl: cApplicationMainFuncNotFound, code: eApplicationMainFuncNotFound}
	errApplicationMainFuncAlreadyRegistered      = err{tpl: cApplicationMainFuncAlreadyRegistered, code: eApplicationMainFuncAlreadyRegistered}
	errApplicationRegistrationUnknownObject      = err{tpl: cApplicationRegistrationUnknownObject, code: eApplicationRegistrationUnknownObject}
	errComponentIsNull                           = err{tpl: cComponentIsNull, code: eComponentIsNull}
	errComponentRegistrationProhibited           = err{tpl: cComponentRegistrationProhibited, code: eComponentRegistrationProhibited}
	errComponentRegistrationError                = err{tpl: cComponentRegistrationError, code: eComponentRegistrationError}
	errComponentPreferencesCallBeforeCompleting  = err{tpl: cComponentPreferencesCallBeforeCompleting, code: eComponentPreferencesCallBeforeCompleting}
	errComponentPanicException                   = err{tpl: cComponentPanicException, code: eComponentPanicException}
	errComponentRunlevelError                    = err{tpl: cComponentRunlevelError, code: eComponentRunlevelError}
	errComponentRulesError                       = err{tpl: cComponentRulesError, code: eComponentRulesError}
	errRunlevelCantLessCurrentLevel              = err{tpl: cRunlevelCantLessCurrentLevel, code: eRunlevelCantLessCurrentLevel}
	errInitLogging                               = err{tpl: cInitLogging, code: eInitLogging}
	errComponentConflict                         = err{tpl: cComponentConflict, code: eComponentConflict}
	errComponentRequires                         = err{tpl: cComponentRequires, code: eComponentRequires}
	errComponentInitiateTimeout                  = err{tpl: cComponentInitiateTimeout, code: eComponentInitiateTimeout}
	errComponentInitiateExecution                = err{tpl: cComponentInitiateExecution, code: eComponentInitiateExecution}
	errComponentInitiatePanicException           = err{tpl: cComponentInitiatePanicException, code: eComponentInitiatePanicException}
	errComponentDoExecution                      = err{tpl: cComponentDoExecution, code: eComponentDoExecution}
	errComponentDoPanicException                 = err{tpl: cComponentDoPanicException, code: eComponentDoPanicException}
	errComponentDoUnknownError                   = err{tpl: cComponentDoUnknownError, code: eComponentDoUnknownError}
	errComponentFinalizeExecution                = err{tpl: cComponentFinalizeExecution, code: eComponentFinalizeExecution}
	errComponentFinalizePanicException           = err{tpl: cComponentFinalizePanicException, code: eComponentFinalizePanicException}
	errComponentFinalizeUnknownError             = err{tpl: cComponentFinalizeUnknownError, code: eComponentFinalizeUnknownError}
	errComponentFinalizeWarning                  = err{tpl: cComponentFinalizeWarning, code: eComponentFinalizeWarning}
	errRunlevelSubscribeUnsubscribeNilFunction   = err{tpl: cRunlevelSubscribeUnsubscribeNilFunction, code: eRunlevelSubscribeUnsubscribeNilFunction}
	errRunlevelAlreadySubscribedFunction         = err{tpl: cRunlevelAlreadySubscribedFunction, code: eRunlevelAlreadySubscribedFunction}
	errRunlevelSubscriptionNotFound              = err{tpl: cRunlevelSubscriptionNotFound, code: eRunlevelSubscriptionNotFound}
	errRunlevelSubscriptionPanicException        = err{tpl: cRunlevelSubscriptionPanicException, code: eRunlevelSubscriptionPanicException}
	errCommandLineArgumentRequired               = err{tpl: cCommandLineArgumentRequired, code: eCommandLineArgumentRequired}
	errCommandLineArgumentUnknown                = err{tpl: cCommandLineArgumentUnknown, code: eCommandLineArgumentUnknown}
	errCommandLineArgumentNotCorrect             = err{tpl: cCommandLineArgumentNotCorrect, code: eCommandLineArgumentNotCorrect}
	errCommandLineRequiredFlag                   = err{tpl: cCommandLineRequiredFlag, code: eCommandLineRequiredFlag}
	errCommandLineUnexpectedError                = err{tpl: cCommandLineUnexpectedError, code: eCommandLineUnexpectedError}
	errConfigurationBootstrap                    = err{tpl: cConfigurationBootstrap, code: eConfigurationBootstrap}
	errGetCurrentUser                            = err{tpl: cGetCurrentUser, code: eGetCurrentUser}
	errCantChangeWorkDirectory                   = err{tpl: cCantChangeWorkDirectory, code: eCantChangeWorkDirectory}
	errPidExistsAnotherProcessOfApplication      = err{tpl: cPidExistsAnotherProcessOfApplication, code: ePidExistsAnotherProcessOfApplication}
	errPidFileError                              = err{tpl: cPidFileError, code: ePidFileError}
	errDatabusRecursivePointer                   = err{tpl: cDatabusRecursivePointer, code: eDatabusRecursivePointer}
	errDatabusPanicException                     = err{tpl: cDatabusPanicException, code: eDatabusPanicException}
	errDatabusSubscribeNotFound                  = err{tpl: cDatabusSubscribeNotFound, code: eDatabusSubscribeNotFound}
	errDatabusInternalError                      = err{tpl: cDatabusInternalError, code: eDatabusInternalError}
	errDatabusNotSubscribersForType              = err{tpl: cDatabusNotSubscribersForType, code: eDatabusNotSubscribersForType}
	errDatabusObjectIsNil                        = err{tpl: cDatabusObjectIsNil, code: eDatabusObjectIsNil}
	errConfigurationApplicationProhibited        = err{tpl: cConfigurationApplicationProhibited, code: eConfigurationApplicationProhibited}
	errConfigurationApplicationObject            = err{tpl: cConfigurationApplicationObject, code: eConfigurationApplicationObject}
	errConfigurationApplicationPanic             = err{tpl: cConfigurationApplicationPanic, code: eConfigurationApplicationPanic}
	errConfigurationFileNotFound                 = err{tpl: cConfigurationFileNotFound, code: eConfigurationFileNotFound}
	errConfigurationPermissionDenied             = err{tpl: cConfigurationPermissionDenied, code: eConfigurationPermissionDenied}
	errConfigurationUnexpectedMistakeFileAccess  = err{tpl: cConfigurationUnexpectedMistakeFileAccess, code: eConfigurationUnexpectedMistakeFileAccess}
	errConfigurationFileIsDirectory              = err{tpl: cConfigurationFileIsDirectory, code: eConfigurationFileIsDirectory}
	errConfigurationFileReadingError             = err{tpl: cConfigurationFileReadingError, code: eConfigurationFileReadingError}
	errConfigurationSetDefault                   = err{tpl: cConfigurationSetDefault, code: eConfigurationSetDefault}
	errConfigurationSetDefaultValue              = err{tpl: cConfigurationSetDefaultValue, code: eConfigurationSetDefaultValue}
	errConfigurationSetDefaultPanic              = err{tpl: cConfigurationSetDefaultPanic, code: eConfigurationSetDefaultPanic}
	errConfigurationObjectNotFound               = err{tpl: cConfigurationObjectNotFound, code: eConfigurationObjectNotFound}
	errConfigurationObjectIsNotStructure         = err{tpl: cConfigurationObjectIsNotStructure, code: eConfigurationObjectIsNotStructure}
	errConfigurationObjectIsNil                  = err{tpl: cConfigurationObjectIsNil, code: eConfigurationObjectIsNil}
	errConfigurationObjectIsNotValid             = err{tpl: cConfigurationObjectIsNotValid, code: eConfigurationObjectIsNotValid}
	errConfigurationObjectIsNotAddress           = err{tpl: cConfigurationObjectIsNotAddress, code: eConfigurationObjectIsNotAddress}
	errConfigurationObjectCopy                   = err{tpl: cConfigurationObjectCopy, code: eConfigurationObjectCopy}
	errConfigurationCallbackAlreadyRegistered    = err{tpl: cConfigurationCallbackAlreadyRegistered, code: eConfigurationCallbackAlreadyRegistered}
	errConfigurationCallbackSubscriptionNotFound = err{tpl: cConfigurationCallbackSubscriptionNotFound, code: eConfigurationCallbackSubscriptionNotFound}
)

// ERRORS: Реализация ошибок с возможностью сравнения ошибок между собой.

// ApplicationPanicException Выполнение приложения прервано паникой: ...
func (e *Error) ApplicationPanicException(code uint8, err interface{}, stack []byte) Err {
	return newErr(&errApplicationPanicException, code, err, string(stack))
}

// ApplicationUnknownError Любая не описанная и не ожидаемая ошибка приложения.
func (e *Error) ApplicationUnknownError(code uint8, err error) Err {
	return newErr(&errApplicationUnknownError, code, err)
}

// ApplicationHelpDisplayed Отображение помощи по командам, аргументам и флагам запуска приложения.
func (e *Error) ApplicationHelpDisplayed(code uint8, help *bytes.Buffer) Err {
	return newErr(&errApplicationHelpDisplayed, code, help.String())
}

// ApplicationVersion Версии приложения содержит ошибку: ...
func (e *Error) ApplicationVersion(code uint8, arg ...interface{}) Err {
	return newErr(&errApplicationVersion, code, arg...)
}

// ApplicationMainFuncNotFound Не определена основная функция приложения.
func (e *Error) ApplicationMainFuncNotFound(code uint8) Err {
	return newErr(&errApplicationMainFuncNotFound, code)
}

// ApplicationMainFuncAlreadyRegistered Основная функция приложения уже зарегистрирована.
func (e *Error) ApplicationMainFuncAlreadyRegistered(code uint8) Err {
	return newErr(&errApplicationMainFuncAlreadyRegistered, code)
}

// ApplicationRegistrationUnknownObject Регистрация не известного компонента, объекта или модуля: ...
func (e *Error) ApplicationRegistrationUnknownObject(code uint8, objectName string) Err {
	return newErr(&errApplicationRegistrationUnknownObject, code, objectName)
}

// ComponentIsNull В качестве объекта компоненты передан nil.
func (e *Error) ComponentIsNull(code uint8) Err { return newErr(&errComponentIsNull, code) }

// ComponentRegistrationProhibited Регистрация компонентов запрещена. Компонента %q не зарегистрирована.
func (e *Error) ComponentRegistrationProhibited(code uint8, componentName string) Err {
	return newErr(&errComponentRegistrationProhibited, code, componentName)
}

// ComponentRegistrationError Регистрация компоненты ... завершилась ошибкой: ...
func (e *Error) ComponentRegistrationError(code uint8, componentName string, err error) Err {
	return newErr(&errComponentRegistrationError, code, componentName, err)
}

// ComponentPreferencesCallBeforeCompleting Опрос настроек компонентов вызван до завершения регистрации компонентов.
func (e *Error) ComponentPreferencesCallBeforeCompleting(code uint8) Err {
	return newErr(&errComponentPreferencesCallBeforeCompleting, code)
}

// ComponentPanicException Выполнение компоненты ... прервано паникой: ...
func (e *Error) ComponentPanicException(code uint8, componentName string, err interface{}, stack []byte) Err {
	return newErr(&errComponentPanicException, code, componentName, err, string(stack))
}

// ComponentRunlevelError Уровень запуска (runlevel), для компоненты %q указан %d, необходимо указать уровень равный 0,
// либо в интервале от 10 до 65534 включительно.
func (e *Error) ComponentRunlevelError(code uint8, componentName string, runlevel uint16) Err {
	return newErr(&errComponentRunlevelError, code, componentName, runlevel)
}

// ComponentRulesError Правила %q для компоненты ... содержат ошибку: ...
func (e *Error) ComponentRulesError(code uint8, rulesName string, componentName string, err error) Err {
	return newErr(&errComponentRulesError, code, rulesName, componentName, err)
}

// RunlevelCantLessCurrentLevel Новый уровень работы приложения (...) не может быть меньше текущего уровня работы
// приложения (...).
func (e *Error) RunlevelCantLessCurrentLevel(code uint8, currentRunlevel uint16, newRunlevel uint16) Err {
	return newErr(&errRunlevelCantLessCurrentLevel, code, newRunlevel, currentRunlevel)
}

// InitLogging Критическая ошибка в модуле менеджера логирования: ...
func (e *Error) InitLogging(code uint8, err error) Err { return newErr(&errInitLogging, code, err) }

// ComponentConflict Компонента ... конфликтует с компонентой ...
func (e *Error) ComponentConflict(code uint8, nameComponent, nameConflict string) Err {
	return newErr(&errComponentConflict, code, nameComponent, nameConflict)
}

// ComponentRequires Компонента ... имеет не удовлетворённую зависимость ...
func (e *Error) ComponentRequires(code uint8, nameComponent, nameRequires string) Err {
	return newErr(&errComponentRequires, code, nameComponent, nameRequires)
}

// ComponentInitiateTimeout Превышено время ожидание выполнения функции Initiate() компоненты ...
func (e *Error) ComponentInitiateTimeout(code uint8, nameComponent string) Err {
	return newErr(&errComponentInitiateTimeout, code, nameComponent)
}

// ComponentInitiateExecution Выполнение функции Initiate() компоненты ... завершено с ошибкой: ...
func (e *Error) ComponentInitiateExecution(code uint8, nameComponent string, err error) Err {
	return newErr(&errComponentInitiateExecution, code, nameComponent, err)
}

// ComponentInitiatePanicException Выполнение функции Initiate() компоненты ... прервано паникой: ...
func (e *Error) ComponentInitiatePanicException(code uint8, nameComponent string, err interface{}, stack []byte) Err {
	return newErr(&errComponentInitiatePanicException, code, nameComponent, err, string(stack))
}

// ComponentDoExecution Выполнение функции Do() компоненты ... завершено с ошибкой: ...
func (e *Error) ComponentDoExecution(code uint8, nameComponent string, err error) Err {
	return newErr(&errComponentDoExecution, code, nameComponent, err)
}

// ComponentDoPanicException Выполнение функции Do() компоненты ... прервано паникой: ...
func (e *Error) ComponentDoPanicException(code uint8, nameComponent string, err interface{}, stack []byte) Err {
	return newErr(&errComponentDoPanicException, code, nameComponent, err, string(stack))
}

// ComponentDoUnknownError Выполнение функций Do() завершилось ошибкой: ...
func (e *Error) ComponentDoUnknownError(code uint8, err error) Err {
	return newErr(&errComponentDoUnknownError, code, err)
}

// ComponentFinalizeExecution Выполнение функции Finalize() компоненты ... завершено с ошибкой: ...
func (e *Error) ComponentFinalizeExecution(code uint8, nameComponent string, err error) Err {
	return newErr(&errComponentFinalizeExecution, code, nameComponent, err)
}

// ComponentFinalizePanicException Выполнение функции Finalize() компоненты ... прервано паникой: ...
func (e *Error) ComponentFinalizePanicException(code uint8, nameComponent string, err interface{}, stack []byte) Err {
	return newErr(&errComponentFinalizePanicException, code, nameComponent, err, string(stack))
}

// ComponentFinalizeUnknownError Выполнение функций Finalize() завершилось ошибкой: ...
func (e *Error) ComponentFinalizeUnknownError(code uint8, err error) Err {
	return newErr(&errComponentFinalizeUnknownError, code, err)
}

// ComponentFinalizeWarning Выполнение функций Finalize() компоненты ..., длится дольше отведённого времени ...
func (e *Error) ComponentFinalizeWarning(code uint8, nameComponent string, tout time.Duration) Err {
	return newErr(&errComponentFinalizeWarning, code, nameComponent, tout)
}

// RunlevelSubscribeUnsubscribeNilFunction Передана nil функция, подписка или отписка nil функции не возможна.
func (e *Error) RunlevelSubscribeUnsubscribeNilFunction(code uint8) Err {
	return newErr(&errRunlevelSubscribeUnsubscribeNilFunction, code)
}

// RunlevelAlreadySubscribedFunction Функция ... уже подписана на получение событий изменения уровня работы приложения.
func (e *Error) RunlevelAlreadySubscribedFunction(code uint8, nameFn string) Err {
	return newErr(&errRunlevelAlreadySubscribedFunction, code, nameFn)
}

// RunlevelSubscriptionNotFound Не найдена подписка функции ... на события изменения уровня работы приложения.
func (e *Error) RunlevelSubscriptionNotFound(code uint8, nameFn string) Err {
	return newErr(&errRunlevelSubscriptionNotFound, code, nameFn)
}

// RunlevelSubscriptionPanicException Вызов функции подписчика на событие изменения уровня работы приложения,
// прервано паникой: ...
func (e *Error) RunlevelSubscriptionPanicException(code uint8, err interface{}, stack []byte) Err {
	return newErr(&errRunlevelSubscriptionPanicException, code, err, string(stack))
}

// CommandLineArgumentRequired Требуется указать обязательную команду, аргумент или флаг командной строки: ...
func (e *Error) CommandLineArgumentRequired(code uint8, description string) Err {
	return newErr(&errCommandLineArgumentRequired, code, description)
}

// CommandLineArgumentUnknown Неизвестная команда, аргумент или флаг командной строки: ...
func (e *Error) CommandLineArgumentUnknown(code uint8, description string) Err {
	return newErr(&errCommandLineArgumentUnknown, code, description)
}

// CommandLineArgumentNotCorrect Не верное значение или тип аргумента, флага или параметра: ...
func (e *Error) CommandLineArgumentNotCorrect(code uint8, description string) Err {
	return newErr(&errCommandLineArgumentNotCorrect, code, description)
}

// CommandLineRequiredFlag Не указан один или несколько обязательных флагов: ...
func (e *Error) CommandLineRequiredFlag(code uint8, description string) Err {
	return newErr(&errCommandLineRequiredFlag, code, description)
}

// CommandLineUnexpectedError Не предвиденная ошибка библиотеки командного интерфейса приложения: ...
func (e *Error) CommandLineUnexpectedError(code uint8, description string, err error) Err {
	return newErr(&errCommandLineUnexpectedError, code, err, description)
}

// ConfigurationBootstrap Ошибка начально bootstrap конфигурации приложения: ...
func (e *Error) ConfigurationBootstrap(code uint8, err error) Err {
	return newErr(&errConfigurationBootstrap, code, err)
}

// GetCurrentUser Не удалось загрузить данные о текущем пользователе операционной системы: ...
func (e *Error) GetCurrentUser(code uint8, err error) Err {
	return newErr(&errGetCurrentUser, code, err)
}

// CantChangeWorkDirectory Не удалось сменить рабочую директорию приложения: ...
func (e *Error) CantChangeWorkDirectory(code uint8, err error) Err {
	return newErr(&errCantChangeWorkDirectory, code, err)
}

// PidExistsAnotherProcessOfApplication Существует один или несколько работающих процессов приложения, измените
// PID файл или остановите экземпляры приложения, PID: ...
func (e *Error) PidExistsAnotherProcessOfApplication(code uint8, pids []int) (err Err) {
	var (
		tmp []string
		n   int
	)

	tmp = make([]string, 0, len(pids))
	for n = range pids {
		tmp = append(tmp, strconv.FormatInt(int64(pids[n]), 10))
	}
	err = newErr(&errPidExistsAnotherProcessOfApplication, code, strings.Join(tmp, ", "))

	return
}

// PidFileError Ошибка работы с PID файлом ...: ...
func (e *Error) PidFileError(code uint8, filename string, err error) Err {
	return newErr(&errPidFileError, code, filename, err)
}

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

// ConfigurationApplicationProhibited Регистрация объектов конфигурации на текущем уровне работы приложения запрещена.
// Конфигурация ... не зарегистрирована.
func (e *Error) ConfigurationApplicationProhibited(code uint8, objectName string) Err {
	return newErr(&errConfigurationApplicationProhibited, code, objectName)
}

// ConfigurationApplicationObject Объект конфигурации приложения содержит ошибку: ...
func (e *Error) ConfigurationApplicationObject(code uint8, err error) Err {
	return newErr(&errConfigurationApplicationObject, code, err)
}

// ConfigurationApplicationPanic Непредвиденная ошибка при регистрации объекта конфигурации. Паника: ... ...
func (e *Error) ConfigurationApplicationPanic(code uint8, err interface{}, stack []byte) Err {
	return newErr(&errConfigurationApplicationPanic, code, err, string(stack))
}

// ConfigurationFileNotFound Указанного файла конфигурации ... не существует: ...
func (e *Error) ConfigurationFileNotFound(code uint8, filename string, err error) Err {
	return newErr(&errConfigurationFileNotFound, code, filename, err)
}

// ConfigurationPermissionDenied Отсутствует доступ к файлу конфигурации, ошибка: ...
func (e *Error) ConfigurationPermissionDenied(code uint8, _ string, err error) Err {
	return newErr(&errConfigurationPermissionDenied, code, err)
}

// ConfigurationUnexpectedMistakeFileAccess Неожиданная ошибка доступа к файлу конфигурации ...: ...
func (e *Error) ConfigurationUnexpectedMistakeFileAccess(code uint8, filename string, err error) Err {
	return newErr(&errConfigurationUnexpectedMistakeFileAccess, code, filename, err)
}

// ConfigurationFileIsDirectory В качестве файла конфигурации указана директория: ...
func (e *Error) ConfigurationFileIsDirectory(code uint8, filename string) Err {
	return newErr(&errConfigurationFileIsDirectory, code, filename)
}

// ConfigurationFileReadingError Чтение фала конфигурации ... прервано ошибкой: ...
func (e *Error) ConfigurationFileReadingError(code uint8, filename string, err error) Err {
	return newErr(&errConfigurationFileReadingError, code, filename, err)
}

// ConfigurationSetDefault Установка значений по умолчанию, для переменных конфигурации, прервана ошибкой: ...
func (e *Error) ConfigurationSetDefault(code uint8, err error) Err {
	return newErr(&errConfigurationSetDefault, code, err)
}

// ConfigurationSetDefaultValue Установка значения по умолчанию ..., для переменной конфигурации ..., прервана
// ошибкой: ...
func (e *Error) ConfigurationSetDefaultValue(code uint8, value string, name string, err error) Err {
	return newErr(&errConfigurationSetDefaultValue, code, value, name, err)
}

// ConfigurationSetDefaultPanic Непредвиденная ошибка, при установке значений по умолчанию, объекта
// конфигурации. Паника: ...
func (e *Error) ConfigurationSetDefaultPanic(code uint8, err interface{}, stack []byte) Err {
	return newErr(&errConfigurationSetDefaultPanic, code, err, string(stack))
}

// ConfigurationObjectNotFound Объект конфигурации с типом ... не найден.
func (e *Error) ConfigurationObjectNotFound(code uint8, typeName string) Err {
	return newErr(&errConfigurationObjectNotFound, code, typeName)
}

// ConfigurationObjectIsNotStructure Переданный объект ... не является структурой.
func (e *Error) ConfigurationObjectIsNotStructure(code uint8, typeName string) Err {
	return newErr(&errConfigurationObjectIsNotStructure, code, typeName)
}

// ConfigurationObjectIsNil Переданный объект, является nil объектом.
func (e *Error) ConfigurationObjectIsNil(code uint8) Err {
	return newErr(&errConfigurationObjectIsNil, code)
}

// ConfigurationObjectIsNotValid Объект конфигурации с типом ... не инициализирован.
func (e *Error) ConfigurationObjectIsNotValid(code uint8, typeName string) Err {
	return newErr(&errConfigurationObjectIsNotValid, code, typeName)
}

// ConfigurationObjectIsNotAddress Объект конфигурации с типом ... передан не корректно. Необходимо передать
// адрес объекта.
func (e *Error) ConfigurationObjectIsNotAddress(code uint8, typeName string) Err {
	return newErr(&errConfigurationObjectIsNotAddress, code, typeName)
}

// ConfigurationObjectCopy Копирование объекта конфигурации я типом ... прервано ошибкой: ...
func (e *Error) ConfigurationObjectCopy(code uint8, typeName string, err error) Err {
	return newErr(&errConfigurationObjectCopy, code, typeName, err)
}

// ConfigurationCallbackAlreadyRegistered Подписка функции обратного вызова на изменение конфигурации с
// типом ... для функции ... уже существует.
func (e *Error) ConfigurationCallbackAlreadyRegistered(code uint8, typeName string, fnName string) Err {
	return newErr(&errConfigurationCallbackAlreadyRegistered, code, typeName, fnName)
}

// ConfigurationCallbackSubscriptionNotFound Подписка функции обратного вызова на изменение конфигурации с
// типом ... для функции ... не существует.
func (e *Error) ConfigurationCallbackSubscriptionNotFound(code uint8, typeName string, fnName string) Err {
	return newErr(&errConfigurationCallbackSubscriptionNotFound, code, typeName, fnName)
}

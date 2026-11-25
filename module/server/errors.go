package server

import "github.com/webnice/dic"

// Все ошибки определены как константы. Коды ошибок приложения:

// Коды ошибок.
const (
	eServerByIdNotFound uint8 = iota + 1
	eServersAddedNotEqualLaunched
	eModulePanicException
	eBeforeExitWithError
	eServerByIdAlreadyStarted
	eServerByIdNotStarted
	eServersNotRegistered
	eDomainRefused
)

// Текстовые значения кодов ошибок на основном языке приложения.
const (
	cServerByIdNotFound           = "Конфигурация сервера с идентификатором %q не найдена."
	cServersAddedNotEqualLaunched = "Запуск серверов завершился ошибкой, добавлено серверов: %d, запущено серверов: %d."
	cModulePanicException         = "Выполнение модуля %q прервано паникой:\n%v\n%s."
	cBeforeExitWithError          = "Функция Before() завершилась с ошибкой: %s"
	cServerByIdAlreadyStarted     = "Сервер с идентификатором %q уже запущен."
	cServerByIdNotStarted         = "Сервер с идентификатором %q не был запущен."
	cServersNotRegistered         = "Не зарегистрирован ни один сервер."
	cDomainRefused                = "Отказ в обслуживании запроса от %q, домен %q не разрешён конфигурацией сервера."
)

type Error struct {
	dic.Errors

	// ServerByIdNotFound Конфигурация сервера с идентификатором ... не найдена.
	ServerByIdNotFound dic.IError

	// ServersAddedNotEqualLaunched Добавлено ... серверов, запущено ... серверов.
	ServersAddedNotEqualLaunched dic.IError

	// ModulePanicException Выполнение модуля ... прервано паникой: ...
	ModulePanicException dic.IError

	// BeforeExitWithError Функция Before() завершилась с ошибкой: ...
	BeforeExitWithError dic.IError

	// ServerByIdAlreadyStarted Сервер с идентификатором ... уже запущен.
	ServerByIdAlreadyStarted dic.IError

	// ServerByIdNotStarted Сервер с идентификатором ... не был запущен.
	ServerByIdNotStarted dic.IError

	// ServersNotRegistered Не зарегистрирован ни один сервер.
	ServersNotRegistered dic.IError

	// DomainRefused Отказ в обслуживании запроса от ..., домен ... не разрешён конфигурацией сервера.
	DomainRefused dic.IError
}

var (
	errSingleton = &Error{
		Errors:                       dic.Error(),
		ServerByIdNotFound:           dic.NewError(cServerByIdNotFound, "идентификатор").CodeU8().Set(eServerByIdNotFound),
		ServersAddedNotEqualLaunched: dic.NewError(cServersAddedNotEqualLaunched, "добавлено", "запущено").CodeU8().Set(eServersAddedNotEqualLaunched),
		ModulePanicException:         dic.NewError(cModulePanicException, "название модуля", "стек паники").CodeU8().Set(eModulePanicException),
		BeforeExitWithError:          dic.NewError(cBeforeExitWithError, "ошибка").CodeU8().Set(eBeforeExitWithError),
		ServerByIdAlreadyStarted:     dic.NewError(cServerByIdAlreadyStarted, "идентификатор").CodeU8().Set(eServerByIdAlreadyStarted),
		ServerByIdNotStarted:         dic.NewError(cServerByIdNotStarted, "идентификатор").CodeU8().Set(eServerByIdNotStarted),
		ServersNotRegistered:         dic.NewError(cServersNotRegistered).CodeU8().Set(eServersNotRegistered),
		DomainRefused:                dic.NewError(cDomainRefused, "адрес", "домен").CodeU8().Set(eDomainRefused),
	}

	// Errors Справочник ошибок.
	Errors = func() *Error { return errSingleton }
)

package tgbot

import "github.com/webnice/dic"

// Константы ошибок.
const (
	cPanic                 = "Перехвачена паника:\n%v\n%s."
	cLogger                = "Установка функционала журналирования прервана ошибкой: %s."
	cBusUnknownEvent       = "Не обработанное событие шины данных: %v."
	cProxyError            = "Указан не корректный адрес прокси сервера %q, ошибка: %s."
	cSendMessageToUser     = "Отправка сообщения пользователю телеграм прервана ошибкой: %s."
	cGetWebhookInfo        = "Загрузка информации о вебхуке телеграм бота прервана ошибкой: %s."
	cWebhookCreate         = "Создание URI для получения webhook прервано ошибкой: %s."
	cWebhookRegistration   = "Регистрация URI webhook %q прервано ошибкой: %s."
	cRegistrationNilObject = "Регистрируемый объект не может быть nil."
	cRegistrationInterface = "Регистрируемый объект не реализует ни один обязательный интерфейс."
	cSubscriptionIncorrect = "Не корректная подписка %q, ожидался интерфейс: %q."
)

// Error Структура справочника ошибок.
type Error struct {
	dic.Errors

	// Panic Перехвачена паника: ... ...
	Panic dic.IError

	// Logger Установка функционала журналирования прервана ошибкой: ...
	Logger dic.IError

	// BusUnknownEvent Не обработанное событие шины данных: ...
	BusUnknownEvent dic.IError

	// ProxyError Указан не корректный адрес прокси сервера ..., ошибка: ...
	ProxyError dic.IError

	// SendMessageToUser Отправка сообщения пользователю телеграм прервана ошибкой: ...
	SendMessageToUser dic.IError

	// GetWebhookInfo Загрузка информации о вебхуке телеграм бота прервана ошибкой: ...
	GetWebhookInfo dic.IError

	// WebhookCreate Создание URI для получения webhook прервано ошибкой: ...
	WebhookCreate dic.IError

	// WebhookRegistration Регистрация URI webhook ... прервано ошибкой: ...
	WebhookRegistration dic.IError

	// RegistrationNilObject "Регистрируемый объект не может быть nil."
	RegistrationNilObject dic.IError

	// RegistrationInterface Регистрируемый объект не реализует ни один обязательный интерфейс.
	RegistrationInterface dic.IError

	// SubscriptionFnIncorrect Не корректная подписка %q, ожидался интерфейс: ...
	SubscriptionIncorrect dic.IError
}

var errSingleton = &Error{
	Errors:                dic.Error(),
	Panic:                 dic.NewError(cPanic, "ошибка", "короткий стек вызова"),
	Logger:                dic.NewError(cLogger, "ошибка"),
	BusUnknownEvent:       dic.NewError(cBusUnknownEvent, "событие"),
	ProxyError:            dic.NewError(cProxyError, "адрес", "ошибка"),
	SendMessageToUser:     dic.NewError(cSendMessageToUser, "ошибка"),
	GetWebhookInfo:        dic.NewError(cGetWebhookInfo, "ошибка"),
	WebhookCreate:         dic.NewError(cWebhookCreate, "ошибка"),
	WebhookRegistration:   dic.NewError(cWebhookRegistration, "адрес", "ошибка"),
	RegistrationNilObject: dic.NewError(cRegistrationNilObject),
	RegistrationInterface: dic.NewError(cRegistrationInterface),
	SubscriptionIncorrect: dic.NewError(cSubscriptionIncorrect, "объект", "ожидаемый интерфейс"),
}

// Errors Справочник ошибок.
//
//goland:noinspection ALL
func Errors() *Error { return errSingleton }

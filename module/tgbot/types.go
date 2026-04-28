/*

	Инструментарий создания и работы с телеграм ботами.

*/

package tgbot

import (
	"context"
	"sync"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitTypes "github.com/webnice/kit/v4/types"
	tgbotapi "github.com/webnice/tba/v9"
)

// Interface Интерфейс пакета.
type Interface interface {
	// Databuser Встраивание интерфейса шины данных.
	kitTypes.Databuser

	// BusSubscribe Подписка на сообщения из шины данных.
	BusSubscribe()

	// BusUnsubscribe Отписка от сообщений из шины данных.
	BusUnsubscribe()

	// ТЕЛЕГРАМ БОТ

	// Registration Регистрация объектов, отвечающих за функционал бота.
	// К регистрации допускаются объекты, которые удовлетворяют хотя бы один интерфейс из перечисленных в IUpdatesFull.
	// Регистрируемый объект должен удовлетворять как минимум один интерфейс обновлений.
	//
	// Если регистрируемый объект не удовлетворяет ни одному интерфейсу, тем самым не содержит ни одного
	// метода, тогда функция регистрации вернёт ошибку, а сам объект не будет зарегистрирован.
	Registration(obj any) (err error)

	// Use Подключение промежуточного.
	Use(middlewares ...func(Handler) Handler)

	// Initialization Инициализация и запуск телеграм бота.
	Initialization(ctx context.Context, cfg *Configuration) (err error)

	LogOut() (err error)

	// IsReady Флаг готовности телеграм бота к работе.
	IsReady() bool

	// ОШИБКИ

	// Errors Справочник всех ошибок пакета.
	Errors() *Error

	// BotLogger Встраивание интерфейса журналирования.
	tgbotapi.BotLogger
}

// Объект сущности пакета.
type impl struct {
	cfg              kitModuleCfg.Interface  // Интерфейс конфигурации приложения.
	api              *tgbotapi.BotAPI        // Телеграм бот API.
	msgInp           chan *tgbotapi.Update   // Канал входящих сообщений телеграм бота.
	botCfg           *Configuration          // Конфигурация телеграм бота.
	botUser          *tgbotapi.User          // Объект пользователя телеграм бота.
	botReady         bool                    // Флаг готовности телеграм бота к работе.
	subscription     map[int][]any           // Подписка зарегистрированных объектов функционала бота.
	subscriptionSync *sync.RWMutex           // Защита карты подписки.
	middlewares      []func(Handler) Handler // Срез промежуточного.
}

type serverHandler struct {
	tbt *impl
}

// Middlewares Описание функций промежуточного.
type Middlewares []func(Handler) Handler

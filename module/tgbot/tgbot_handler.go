package tgbot

import (
	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	tgbotapi "github.com/webnice/tba/v9"
)

// Handler Стандартный обработчик обновлений телеграм.
type Handler interface {
	ServeTelegram(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// HandlerFunc Тип стандартного обработчика обновлений телеграм.
type HandlerFunc func(api *tgbotapi.BotAPI, upd *tgbotapi.Update)

// ServeTelegram Функция стандартного обработчика обновлений телеграм.
func (f HandlerFunc) ServeTelegram(api *tgbotapi.BotAPI, upd *tgbotapi.Update) { f(api, upd) }

// ServeTelegram Обработчик всех входящих обновлений.
func (sh serverHandler) ServeTelegram(upd *tgbotapi.Update) {
	var (
		err     error
		handler Handler
	)

	defer func() {
		if e := recover(); e != nil {
			sh.tbt.log().Critical(sh.tbt.Errors().Panic.Bind(e.(error), kitModuleTrace.StackShort()))
		}
	}()
	// Событие появления данных пользователя.
	if err = sh.tbt.
		callSubscriptionTelegramUserEvent(upd); err != nil {
		sh.tbt.log().Error(err.Error())
	}
	// Запуск всех промежуточных и затем запуск основного обработчика.
	// На этом этапе, обработка обновлений телеграм может быть прервана
	// из промежуточного путём пропуска вызова цепочки промежуточного.
	handler = Chain(sh.tbt.middlewares...).Handler(sh.tbt)
	handler.ServeTelegram(sh.tbt.api, upd)
}

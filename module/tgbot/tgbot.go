package tgbot

import (
	"sync"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitTypes "github.com/webnice/kit/v4/types"
	tgbotapi "github.com/webnice/tba/v9"
)

// New Конструктор объекта сущности пакета, возвращается интерфейс пакета.
//
//goland:noinspection GoUnusedExportedFunction
func New() Interface {
	var tbt = &impl{
		cfg:              kitModuleCfg.Get(),
		msgInp:           make(chan *tgbotapi.Update, inputChannelBufferSize),
		subscription:     make(map[int][]any),
		subscriptionSync: new(sync.RWMutex),
	}
	// Установка объекта пакета в качестве менеджера журнала.
	tbt.setLogger()

	return tbt
}

// Ссылка на менеджер логирования.
func (tbt *impl) log() kitTypes.Logger { return tbt.cfg.Log() }

// Errors Справочник ошибок.
func (tbt *impl) Errors() *Error { return Errors() }

// Println Реализация интерфейса tgbotapi.BotLogger.
func (tbt *impl) Println(v ...interface{}) { tbt.log().Warning(v...) }

// Printf Реализация интерфейса tgbotapi.BotLogger.
func (tbt *impl) Printf(format string, v ...interface{}) { tbt.log().Warningf(format, v...) }

// IsReady Флаг готовности телеграм бота к работе.
func (tbt *impl) IsReady() bool { return tbt.botReady }

// Установка объекта пакета в качестве менеджера журнала.
func (tbt *impl) setLogger() {
	if err := tgbotapi.
		SetLogger(tbt); err != nil {
		err = tbt.Errors().Logger.Bind(err)
		tbt.log().Critical(err)
	}
}

// Разрешённые по умолчанию обновления, передаваемые сервером телеграм, телеграм боту.
func (tbt *impl) defaultAllowedUpdates() []string {
	return []string{
		AllowedUpdateMessage.String(),
		AllowedUpdateEditedMessage.String(),
		AllowedUpdateChannelPost.String(),
		AllowedUpdateEditedChannelPost.String(),
		AllowedUpdateBusinessConnection.String(),
		AllowedUpdateBusinessMessage.String(),
		AllowedUpdateEditedBusinessMessage.String(),
		AllowedUpdateDeletedBusinessMessages.String(),
		AllowedUpdateMessageReaction.String(),
		AllowedUpdateMessageReactionCount.String(),
		AllowedUpdateInlineQuery.String(),
		AllowedUpdateChosenInlineResult.String(),
		AllowedUpdateCallbackQuery.String(),
		AllowedUpdateShippingQuery.String(),
		AllowedUpdatePreCheckoutQuery.String(),
		AllowedUpdatePurchasedPaidMedia.String(),
		AllowedUpdatePoll.String(),
		AllowedUpdatePollAnswer.String(),
		AllowedUpdateMyChatMember.String(),
		AllowedUpdateChatMember.String(),
		AllowedUpdateChatJoinRequest.String(),
		AllowedUpdateChatBoost.String(),
		AllowedUpdateRemovedChatBoost.String(),
		AllowedUpdateManagedBot.String(),
	}
}

// Use Подключение промежуточного.
func (tbt *impl) Use(middlewares ...func(Handler) Handler) {
	tbt.middlewares = append(tbt.middlewares, middlewares...)
}

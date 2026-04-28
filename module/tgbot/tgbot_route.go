package tgbot

import (
	"github.com/webnice/debug"
	tgbotapi "github.com/webnice/tba/v9"
)

// ServeTelegram Основной стандартный обработчик обновлений телеграм.
func (tbt *impl) ServeTelegram(_ *tgbotapi.BotAPI, upd *tgbotapi.Update) {
	var err error

	// Разбор полученного обновления на основные типы.
	switch {
	// Новое входящее сообщение любого типа — текст, фото, стикер и т.д.
	case upd.Message != nil:
		err = tbt.makeSubscriptionMessageEvent(eventOnMessage, upd)
	// Новая версия сообщения, которое известно боту и было отредактировано.
	case upd.EditedMessage != nil:
		err = tbt.makeSubscriptionMessageEvent(eventOnEditedMessage, upd)
	// Новая версия сообщения в канал, которое известно боту и было отредактировано.
	case upd.ChannelPost != nil:
		err = tbt.callSubscriptionEvent(eventOnChannelPost, upd)
	// Новая входящая публикация на канале любого рода — текст, фотография, стикер и т.д.
	case upd.EditedChannelPost != nil:
		err = tbt.callSubscriptionEvent(eventOnEditedChannelPost, upd)
	// Бот был подключен к бизнес-аккаунту или отключен от него, или пользователь отредактировал существующее
	// соединение с ботом.
	case upd.BusinessConnection != nil:
		err = tbt.callSubscriptionEvent(eventOnBusinessConnection, upd)
	// Новое неслужебное сообщение от подключенного бизнес-аккаунта.
	case upd.BusinessMessage != nil:
		err = tbt.callSubscriptionEvent(eventOnBusinessMessage, upd)
	// Новая версия сообщения от подключенного бизнес-аккаунта.
	case upd.EditedBusinessMessage != nil:
		err = tbt.callSubscriptionEvent(eventOnEditedBusinessMessage, upd)
	// Сообщения были удалены из подключенного корпоративного аккаунта.
	case upd.DeletedBusinessMessages != nil:
		err = tbt.callSubscriptionEvent(eventOnDeletedBusinessMessages, upd)
	// Пользователь изменил реакцию на сообщение.
	case upd.MessageReaction != nil:
		err = tbt.callSubscriptionEvent(eventOnMessageReaction, upd)
	// Изменены реакции на сообщение с анонимными реакциями.
	case upd.MessageReactionCount != nil:
		err = tbt.callSubscriptionEvent(eventOnMessageReactionCount, upd)
	// Новый входящий встроенный запрос.
	case upd.InlineQuery != nil:
		err = tbt.callSubscriptionEvent(eventOnInlineQuery, upd)
	// Результат встроенного запроса, который был выбран пользователем и отправлен его партнеру по чату.
	case upd.ChosenInlineResult != nil:
		err = tbt.callSubscriptionEvent(eventOnChosenInlineResult, upd)
	// Новый входящий запрос обратного вызова.
	case upd.CallbackQuery != nil:
		err = tbt.callSubscriptionEvent(eventOnCallbackQuery, upd)
	// Новый запрос на доставку. Только для счетов с ценой зависящей от доставки.
	case upd.ShippingQuery != nil:
		err = tbt.callSubscriptionEvent(eventOnShippingQuery, upd)
	// Новый входящий запрос на предварительный заказ.
	// Содержит полную информацию об оформлении заказа.
	case upd.PreCheckoutQuery != nil:
		err = tbt.callSubscriptionEvent(eventOnPreCheckoutQuery, upd)
	// Пользователь приобрел платный контент с полезной нагрузкой, отправленной ботом в чате без канала.
	case upd.PurchasedPaidMedia != nil:
		err = tbt.callSubscriptionEvent(eventOnPurchasedPaidMedia, upd)
	// Новое состояние опроса. Боты получают только уведомления об опросах, остановленных вручную, и об опросах,
	// отправленных ботом.
	case upd.Poll != nil:
		err = tbt.callSubscriptionEvent(eventOnPoll, upd)
	// Пользователь изменил свой ответ в не анонимном опросе.
	// Боты получают новые голоса только в опросах, отправленных ими самими.
	case upd.PollAnswer != nil:
		err = tbt.callSubscriptionEvent(eventOnPollAnswer, upd)
	// Статус участника чата с ботом был обновлен в чате.
	// В личных чатах это обновление отображается только в том случае, если пользователь
	// заблокировал или разблокировал бота.
	case upd.MyChatMember != nil:
		err = tbt.callSubscriptionEvent(eventOnMyChatMember, upd)
	// Статус участника чата был обновлен в чате.
	// Чтобы получать такие обновления, бот должен быть администратором чата и явно указать "chat_member" в
	// списке allowed_updates.
	case upd.ChatMember != nil:
		err = tbt.callSubscriptionEvent(eventOnChatMember, upd)
	// Отправлен запрос на присоединение к чату.
	// Чтобы получать эти обновления, бот должен иметь право администратора can_invite_users в чате.
	case upd.ChatJoinRequest != nil:
		err = tbt.callSubscriptionEvent(eventOnChatJoinRequest, upd)
	// Добавлена или изменена функция ускорения чата.
	// Чтобы получать эти обновления, бот должен быть администратором чата.
	case upd.ChatBoost != nil:
		err = tbt.callSubscriptionEvent(eventOnChatBoost, upd)
	// Буст был удален из чата.
	// Чтобы получать эти обновления, бот должен быть администратором чата.
	case upd.RemovedChatBoost != nil:
		err = tbt.callSubscriptionEvent(eventOnRemovedChatBoost, upd)
	// Был создан новый бот, которым будет управлять ботом, или был изменен токен или владелец управляемого бота.
	case upd.ManagedBot != nil:
		err = tbt.callSubscriptionEvent(eventOnManagedBot, upd)
	// Неожиданный, не реализованный или ошибочный тип обновления.
	default:
		if tbt.cfg.Debug() {
			tbt.log().Debug(debug.DumperString(upd))
		}
	}
	if err != nil {
		tbt.log().Error(err.Error())
	}
}

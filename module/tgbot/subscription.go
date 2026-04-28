package tgbot

import (
	"reflect"

	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	tgbotapi "github.com/webnice/tba/v9"
)

// Поиск в обновлении объекта пользователя и вызов функций подписки на событие OnTelegramUser.
//
//goland:noinspection DuplicatedCode
func (tbt *impl) callSubscriptionTelegramUserEvent(upd *tgbotapi.Update) (err error) {
	const event = eventOnTelegramUser
	var (
		fn   IUpdatesTelegramUser
		user *tgbotapi.User
		n    int
		ok   bool
	)

	// Предотвращение падения приложения из-за ошибки в алгоритме.
	defer func() {
		if e := recover(); e != nil {
			tbt.log().Critical(tbt.Errors().Panic.Bind(e.(error), kitModuleTrace.StackShort()))
		}
	}()
	if user = FindTelegramUser(upd); user == nil {
		return
	}
	tbt.subscriptionSync.RLock()
	defer tbt.subscriptionSync.RUnlock()
	if len(tbt.subscription[event]) > 0 {
		for n = range tbt.subscription[event] {
			if fn, ok = tbt.subscription[event][n].(IUpdatesTelegramUser); ok {
				err = fn.OnTelegramUser(tbt.api, upd, user)
			}
		}
	}

	return
}

// Вызов функций подписки.
func (tbt *impl) callSubscriptionEvent(event int, upd *tgbotapi.Update) (err error) {
	var n int

	// Вызов функций.
	tbt.subscriptionSync.RLock()
	defer tbt.subscriptionSync.RUnlock()
	if len(tbt.subscription[event]) > 0 {
		for n = range tbt.subscription[event] {
			if err = tbt.safeCallSubscription(event, tbt.subscription[event][n], upd); err != nil {
				tbt.log().Critical(err.Error())
				err = nil
			}
		}
	}

	return
}

// Безопасный вызов функции.
func (tbt *impl) safeCallSubscription(event int, fnAny any, upd *tgbotapi.Update) (err error) {
	var user *tgbotapi.User

	// Предотвращение падения приложения из-за ошибки в алгоритме.
	defer func() {
		if e := recover(); e != nil {
			err = tbt.Errors().Panic.Bind(e.(error), kitModuleTrace.StackShort())
		}
	}()
	switch event {
	case eventOnAudio:
		switch fn, ok := fnAny.(IUpdatesMessageAudio); ok {
		case true:
			fn.OnMessageAudio(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageAudio")
		}
	case eventOnDocument:
		switch fn, ok := fnAny.(IUpdatesMessageDocument); ok {
		case true:
			fn.OnMessageDocument(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageDocument")
		}
	case eventOnPhoto:
		switch fn, ok := fnAny.(IUpdatesMessagePhoto); ok {
		case true:
			fn.OnMessagePhoto(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessagePhoto")
		}
	case eventOnSticker:
		switch fn, ok := fnAny.(IUpdatesMessageSticker); ok {
		case true:
			fn.OnMessageSticker(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageSticker")
		}
	case eventOnStory:
		switch fn, ok := fnAny.(IUpdatesMessageStory); ok {
		case true:
			fn.OnMessageStory(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageStory")
		}
	case eventOnVideo:
		switch fn, ok := fnAny.(IUpdatesMessageVideo); ok {
		case true:
			fn.OnMessageVideo(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageVideo")
		}
	case eventOnVideoNote:
		switch fn, ok := fnAny.(IUpdatesMessageVideoNote); ok {
		case true:
			fn.OnMessageVideoNote(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageVideoNote")
		}
	case eventOnVoice:
		switch fn, ok := fnAny.(IUpdatesMessageVoice); ok {
		case true:
			fn.OnMessageVoice(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageVoice")
		}
	case eventOnChecklist:
		switch fn, ok := fnAny.(IUpdatesMessageChecklist); ok {
		case true:
			fn.OnMessageChecklist(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageChecklist")
		}
	case eventOnContact:
		switch fn, ok := fnAny.(IUpdatesMessageContact); ok {
		case true:
			fn.OnMessageContact(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageContact")
		}
	case eventOnDice:
		switch fn, ok := fnAny.(IUpdatesMessageDice); ok {
		case true:
			fn.OnMessageDice(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageDice")
		}
	case eventOnGame:
		switch fn, ok := fnAny.(IUpdatesMessageGame); ok {
		case true:
			fn.OnMessageGame(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageGame")
		}
	case eventOnVenue:
		switch fn, ok := fnAny.(IUpdatesMessageVenue); ok {
		case true:
			fn.OnMessageVenue(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageVenue")
		}
	case eventOnLocation:
		switch fn, ok := fnAny.(IUpdatesMessageLocation); ok {
		case true:
			fn.OnMessageLocation(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageLocation")
		}
	case eventOnInvoice:
		switch fn, ok := fnAny.(IUpdatesMessageInvoice); ok {
		case true:
			fn.OnMessageInvoice(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageInvoice")
		}
	case eventOnMessage:
		switch fn, ok := fnAny.(IUpdatesMessage); ok {
		case true:
			fn.OnMessage(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessage")
		}
	case eventOnEditedMessage:
		switch fn, ok := fnAny.(IUpdatesEditedMessage); ok {
		case true:
			fn.OnEditedMessage(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesEditedMessage")
		}
	case eventOnChannelPost:
		switch fn, ok := fnAny.(IUpdatesChannelPost); ok {
		case true:
			fn.OnChannelPost(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesChannelPost")
		}
	case eventOnEditedChannelPost:
		switch fn, ok := fnAny.(IUpdatesEditedChannelPost); ok {
		case true:
			fn.OnEditedChannelPost(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesEditedChannelPost")
		}
	case eventOnBusinessConnection:
		switch fn, ok := fnAny.(IUpdatesBusinessConnection); ok {
		case true:
			fn.OnBusinessConnection(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesBusinessConnection")
		}
	case eventOnBusinessMessage:
		switch fn, ok := fnAny.(IUpdatesBusinessMessage); ok {
		case true:
			fn.OnBusinessMessage(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesBusinessMessage")
		}
	case eventOnEditedBusinessMessage:
		switch fn, ok := fnAny.(IUpdatesEditedBusinessMessage); ok {
		case true:
			fn.OnEditedBusinessMessage(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesEditedBusinessMessage")
		}
	case eventOnDeletedBusinessMessages:
		switch fn, ok := fnAny.(IUpdatesDeletedBusinessMessages); ok {
		case true:
			fn.OnDeletedBusinessMessages(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesDeletedBusinessMessages")
		}
	case eventOnMessageReaction:
		switch fn, ok := fnAny.(IUpdatesMessageReaction); ok {
		case true:
			fn.OnMessageReaction(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageReaction")
		}
	case eventOnMessageReactionCount:
		switch fn, ok := fnAny.(IUpdatesMessageReactionCount); ok {
		case true:
			fn.OnMessageReactionCount(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMessageReactionCount")
		}
	case eventOnInlineQuery:
		switch fn, ok := fnAny.(IUpdatesInlineQuery); ok {
		case true:
			fn.OnInlineQuery(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesInlineQuery")
		}
	case eventOnChosenInlineResult:
		switch fn, ok := fnAny.(IUpdatesChosenInlineResult); ok {
		case true:
			fn.OnChosenInlineResult(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesChosenInlineResult")
		}
	case eventOnCallbackQuery:
		switch fn, ok := fnAny.(IUpdatesCallbackQuery); ok {
		case true:
			fn.OnCallbackQuery(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesCallbackQuery")
		}
	case eventOnShippingQuery:
		switch fn, ok := fnAny.(IUpdatesShippingQuery); ok {
		case true:
			fn.OnShippingQuery(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesShippingQuery")
		}
	case eventOnPreCheckoutQuery:
		switch fn, ok := fnAny.(IUpdatesPreCheckoutQuery); ok {
		case true:
			fn.OnPreCheckoutQuery(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesPreCheckoutQuery")
		}
	case eventOnPurchasedPaidMedia:
		switch fn, ok := fnAny.(IUpdatesPurchasedPaidMedia); ok {
		case true:
			fn.OnPurchasedPaidMedia(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesPurchasedPaidMedia")
		}
	case eventOnPoll:
		switch fn, ok := fnAny.(IUpdatesPoll); ok {
		case true:
			fn.OnPoll(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesPoll")
		}
	case eventOnPollAnswer:
		switch fn, ok := fnAny.(IUpdatesPollAnswer); ok {
		case true:
			fn.OnPollAnswer(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesPollAnswer")
		}
	case eventOnMyChatMember:
		switch fn, ok := fnAny.(IUpdatesMyChatMember); ok {
		case true:
			if user = FindTelegramUser(upd); user != nil {
				fn.OnMyChatMember(tbt.api, upd, user)
			}
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesMyChatMember")
		}
	case eventOnChatMember:
		switch fn, ok := fnAny.(IUpdatesChatMember); ok {
		case true:
			fn.OnChatMember(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesChatMember")
		}
	case eventOnChatJoinRequest:
		switch fn, ok := fnAny.(IUpdatesChatJoinRequest); ok {
		case true:
			fn.OnChatJoinRequest(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesChatJoinRequest")
		}
	case eventOnChatBoost:
		switch fn, ok := fnAny.(IUpdatesChatBoost); ok {
		case true:
			fn.OnChatBoost(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesChatBoost")
		}
	case eventOnRemovedChatBoost:
		switch fn, ok := fnAny.(IUpdatesRemovedChatBoost); ok {
		case true:
			fn.OnRemovedChatBoost(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesRemovedChatBoost")
		}
	case eventOnManagedBot:
		switch fn, ok := fnAny.(IUpdatesManagedBot); ok {
		case true:
			fn.OnManagedBot(tbt.api, upd)
		default:
			err = tbt.Errors().SubscriptionIncorrect.
				Bind(reflect.TypeOf(fnAny).Elem(), "IUpdatesManagedBot")
		}
	default:
		tbt.log().Warningf("не обработанное событие %d", reflect.TypeOf(event).String())
		return
	}

	return
}

// Дополнительный разбор входящего обновления.
func (tbt *impl) makeSubscriptionMessageEvent(event int, upd *tgbotapi.Update) (err error) {
	var msg *tgbotapi.Message

	switch event {
	case eventOnMessage: // Новое входящее сообщение любого типа — текст, фото, стикер и т.д.
		msg = upd.Message
	case eventOnEditedMessage: // Новая версия сообщения, известная боту и отредактированная.
		msg = upd.EditedMessage
	default:
		return
	}
	switch {
	case msg.Audio != nil:
		err = tbt.callSubscriptionEvent(eventOnAudio, upd)
	case msg.Document != nil:
		err = tbt.callSubscriptionEvent(eventOnDocument, upd)
	case len(msg.Photo) > 0:
		err = tbt.callSubscriptionEvent(eventOnPhoto, upd)
	case msg.Sticker != nil:
		err = tbt.callSubscriptionEvent(eventOnSticker, upd)
	case msg.Story != nil:
		err = tbt.callSubscriptionEvent(eventOnStory, upd)
	case msg.Video != nil:
		err = tbt.callSubscriptionEvent(eventOnVideo, upd)
	case msg.VideoNote != nil:
		err = tbt.callSubscriptionEvent(eventOnVideoNote, upd)
	case msg.Voice != nil:
		err = tbt.callSubscriptionEvent(eventOnVoice, upd)
	case msg.Checklist != nil:
		err = tbt.callSubscriptionEvent(eventOnChecklist, upd)
	case msg.Contact != nil:
		err = tbt.callSubscriptionEvent(eventOnContact, upd)
	case msg.Dice != nil:
		err = tbt.callSubscriptionEvent(eventOnDice, upd)
	case msg.Game != nil:
		err = tbt.callSubscriptionEvent(eventOnGame, upd)
	case msg.Venue != nil:
		err = tbt.callSubscriptionEvent(eventOnVenue, upd)
	case msg.Location != nil:
		err = tbt.callSubscriptionEvent(eventOnLocation, upd)
	case msg.Invoice != nil:
		err = tbt.callSubscriptionEvent(eventOnInvoice, upd)
	default:
		err = tbt.callSubscriptionEvent(event, upd)
	}

	return
}

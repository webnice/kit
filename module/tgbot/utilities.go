package tgbot

import (
	"time"

	"github.com/webnice/dic"
	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
	tgbotapi "github.com/webnice/tba/v9"
)

// FindTelegramUser Поиск пользователя телеграм в объекте обновления.
//
//goland:noinspection DuplicatedCode
func FindTelegramUser(upd *tgbotapi.Update) (ret *tgbotapi.User) {
	if upd.Message != nil && upd.Message.From != nil {
		ret = upd.Message.From
	}
	if ret == nil && upd.EditedMessage != nil && upd.EditedMessage.From != nil {
		ret = upd.EditedMessage.From
	}
	if ret == nil && upd.ChannelPost != nil && upd.ChannelPost.From != nil {
		ret = upd.ChannelPost.From
	}
	if ret == nil && upd.EditedChannelPost != nil && upd.EditedChannelPost.From != nil {
		ret = upd.EditedChannelPost.From
	}
	if ret == nil && upd.BusinessConnection != nil {
		ret = &upd.BusinessConnection.User
	}
	if ret == nil && upd.BusinessMessage != nil && upd.BusinessMessage.From != nil {
		ret = upd.BusinessMessage.From
	}
	if ret == nil && upd.EditedBusinessMessage != nil && upd.EditedBusinessMessage.From != nil {
		ret = upd.EditedBusinessMessage.From
	}
	if ret == nil && upd.MessageReaction != nil && upd.MessageReaction.User != nil {
		ret = upd.MessageReaction.User
	}
	if ret == nil && upd.InlineQuery != nil && upd.InlineQuery.From != nil {
		ret = upd.InlineQuery.From
	}
	if ret == nil && upd.ChosenInlineResult != nil && upd.ChosenInlineResult.From != nil {
		ret = upd.ChosenInlineResult.From
	}
	if ret == nil && upd.CallbackQuery != nil && upd.CallbackQuery.From != nil {
		ret = upd.CallbackQuery.From
	}
	if ret == nil && upd.ShippingQuery != nil && upd.ShippingQuery.From != nil {
		ret = upd.ShippingQuery.From
	}
	if ret == nil && upd.PreCheckoutQuery != nil && upd.PreCheckoutQuery.From != nil {
		ret = upd.PreCheckoutQuery.From
	}
	if ret == nil && upd.PurchasedPaidMedia != nil {
		ret = &upd.PurchasedPaidMedia.From
	}
	if ret == nil && upd.PollAnswer != nil && upd.PollAnswer.User != nil {
		ret = upd.PollAnswer.User
	}
	if ret == nil && upd.MyChatMember != nil {
		ret = &upd.MyChatMember.From
	}
	if ret == nil && upd.ChatMember != nil {
		ret = &upd.ChatMember.From
	}
	if ret == nil && upd.ChatJoinRequest != nil {
		ret = &upd.ChatJoinRequest.From
	}

	return
}

// Функция повтора в случае если возвращается ошибка TooManyRequests.
// Запрос оборачивается в функцию которая возвращает следующие значения:
//   - Ok ------ Флаг успешности запроса.
//   - Status -- Код статуса запроса.
//   - After --- Количество секунд ожидания до выполнения следующего запроса.
//   - Error --- Ошибка.
func requestRetryAfter(fn func() (bool, int64, int64, error)) (err error) {
	var (
		end      bool
		ok       bool
		rqStatus int64
		rqAfter  int64
	)

	defer func() {
		if e := recover(); e != nil {
			err = Errors().Panic.Bind(e.(error), kitModuleTrace.StackShort())
		}
	}()
	for !end {
		switch ok, rqStatus, rqAfter, err = fn(); {
		case !ok && rqStatus == int64(dic.Status().TooManyRequests.Code()):
			if rqAfter <= 0 {
				rqAfter = 1
			}
			<-time.After(time.Second * time.Duration(rqAfter))
			err = nil
			continue
		default:
			end = true
		}
	}

	return
}

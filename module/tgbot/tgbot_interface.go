package tgbot

import tgbotapi "github.com/webnice/tba/v9"

// IUpdatesFull Описание всех возможных методов пакета, отвечающего за функционал телеграм бота.
// К регистрации допускаются объекты, которые удовлетворяют хотя бы один интерфейс.
type IUpdatesFull interface {
	IUpdatesTelegramUser
	IUpdatesMessageAudio
	IUpdatesMessageDocument
	IUpdatesMessagePhoto
	IUpdatesMessageSticker
	IUpdatesMessageStory
	IUpdatesMessageVideo
	IUpdatesMessageVideoNote
	IUpdatesMessageVoice
	IUpdatesMessageChecklist
	IUpdatesMessageContact
	IUpdatesMessageDice
	IUpdatesMessageGame
	IUpdatesMessagePoll
	IUpdatesMessageVenue
	IUpdatesMessageLocation
	IUpdatesMessageInvoice
	IUpdatesMessage
	IUpdatesEditedMessage
	IUpdatesChannelPost
	IUpdatesEditedChannelPost
	IUpdatesBusinessConnection
	IUpdatesBusinessMessage
	IUpdatesEditedBusinessMessage
	IUpdatesDeletedBusinessMessages
	IUpdatesMessageReaction
	IUpdatesMessageReactionCount
	IUpdatesInlineQuery
	IUpdatesChosenInlineResult
	IUpdatesCallbackQuery
	IUpdatesShippingQuery
	IUpdatesPreCheckoutQuery
	IUpdatesPurchasedPaidMedia
	IUpdatesPoll
	IUpdatesPollAnswer
	IUpdatesMyChatMember
	IUpdatesChatMember
	IUpdatesChatJoinRequest
	IUpdatesChatBoost
	IUpdatesRemovedChatBoost
	IUpdatesManagedBot
}

// IUpdatesTelegramUser Метод получающий объект пользователя из любого входящего сообщения обновления.
type IUpdatesTelegramUser interface {
	// OnTelegramUser Метод получающий объект пользователя из любого входящего сообщения обновления.
	OnTelegramUser(api *tgbotapi.BotAPI, upd *tgbotapi.Update, user *tgbotapi.User) (err error)
}

// IUpdatesMessageAudio Метод получающий обновления входящего сообщения:
// Сообщение - это аудиофайл, информация о файле.
type IUpdatesMessageAudio interface {
	// OnMessageAudio Сообщение - это аудиофайл, информация о файле.
	OnMessageAudio(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageDocument Метод получающий обновления входящего сообщения:
// Сообщение - это общий файл, информация о файле.
type IUpdatesMessageDocument interface {
	// OnMessageDocument Сообщение - это общий файл, информация о файле.
	OnMessageDocument(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessagePhoto Метод получающий обновления входящего сообщения:
// Сообщение - это графические файл или файлы.
type IUpdatesMessagePhoto interface {
	// OnMessagePhoto Сообщение - это графические файл или файлы.
	OnMessagePhoto(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageSticker Метод получающий обновления входящего сообщения:
// Сообщение - это наклейка, информация о наклейке.
type IUpdatesMessageSticker interface {
	// OnMessageSticker Сообщение - это наклейка, информация о наклейке.
	OnMessageSticker(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageStory Метод получающий обновления входящего сообщения:
// Сообщение - это пересланная история.
type IUpdatesMessageStory interface {
	// OnMessageStory Сообщение - это пересланная история.
	OnMessageStory(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageVideo Метод получающий обновления входящего сообщения:
// Сообщение - это видео, информация о видео.
type IUpdatesMessageVideo interface {
	// OnMessageVideo Сообщение - это видео, информация о видео.
	OnMessageVideo(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageVideoNote Метод получающий обновления входящего сообщения:
// Сообщение - это видеозапись, информация о видео сообщении.
type IUpdatesMessageVideoNote interface {
	// OnMessageVideoNote Сообщение - это видеозапись, информация о видео сообщении.
	OnMessageVideoNote(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageVoice Метод получающий обновления входящего сообщения:
// Сообщение - это голосовое сообщение, информация о файле.
type IUpdatesMessageVoice interface {
	// OnMessageVoice Сообщение - это голосовое сообщение, информация о файле.
	OnMessageVoice(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageChecklist Метод получающий обновления входящего сообщения:
// Сообщение - это контрольный список.
type IUpdatesMessageChecklist interface {
	// OnMessageChecklist Сообщение - это контрольный список.
	OnMessageChecklist(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageContact Метод получающий обновления входящего сообщения:
// Сообщение - это общий контакт, информация о контакте.
type IUpdatesMessageContact interface {
	// OnMessageContact Сообщение - это общий контакт, информация о контакте.
	OnMessageContact(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageDice Метод получающий обновления входящего сообщения:
// Сообщение - это игральная кость со случайным значением.
type IUpdatesMessageDice interface {
	// OnMessageDice Сообщение - это игральная кость со случайным значением.
	OnMessageDice(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageGame Метод получающий обновления входящего сообщения:
// Сообщение - это игра, информация об игре.
type IUpdatesMessageGame interface {
	// OnMessageGame Сообщение - это игра, информация об игре.
	OnMessageGame(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessagePoll Метод получающий обновления входящего сообщения:
// Сообщение - это нативный опрос, информация о нем.
type IUpdatesMessagePoll interface {
	// OnMessagePoll Сообщение - это нативный опрос, информация о нем.
	OnMessagePoll(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageVenue Метод получающий обновления входящего сообщения:
// Сообщение - это место проведения, информация о месте проведения.
// Для обеспечения обратной совместимости, когда задано поле "место проведения", также будет задано поле местоположения,
// поэтому необходимо использовать либо OnMessageVenue() либо OnMessageLocation(), но не оба одновременно.
type IUpdatesMessageVenue interface {
	// OnMessageVenue Сообщение - это место проведения, информация о месте проведения.
	OnMessageVenue(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageLocation Метод получающий обновления входящего сообщения:
// Сообщение - это общее местоположение, информация о местоположении.
type IUpdatesMessageLocation interface {
	// OnMessageLocation Сообщение - это общее местоположение, информация о местоположении.
	OnMessageLocation(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageInvoice Метод получающий обновления входящего сообщения:
// Сообщение - это счет на оплату, информация о счете-фактуре.
type IUpdatesMessageInvoice interface {
	// OnMessageInvoice Сообщение - это счет на оплату, информация о счете-фактуре.
	OnMessageInvoice(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessage Метод получающий обновления:
// Новое входящее сообщение любого типа — текст, фото, стикер и т.д.
type IUpdatesMessage interface {
	// OnMessage Новое входящее сообщение любого типа — текст, фото, стикер и т.д.
	OnMessage(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesEditedMessage Метод получающий обновления:
// Новая версия сообщения, известная боту и отредактированная.
type IUpdatesEditedMessage interface {
	// OnEditedMessage Новая версия сообщения, известная боту и отредактированная.
	OnEditedMessage(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesChannelPost Метод получающий обновления:
// Новый входящий пост в канале любого типа — текст, фото, стикер и т.д.
type IUpdatesChannelPost interface {
	// OnChannelPost Новый входящий пост в канале любого типа — текст, фото, стикер и т.д.
	OnChannelPost(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesEditedChannelPost Метод получающий обновления:
// Новая версия публикации в канале, известная боту и отредактированная.
type IUpdatesEditedChannelPost interface {
	// OnEditedChannelPost Новая версия публикации в канале, известная боту и отредактированная.
	OnEditedChannelPost(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesBusinessConnection Метод получающий обновления:
// Бот был подключен к бизнес-аккаунту или отключен от него, или пользователь отредактировал
// существующее соединение с ботом.
type IUpdatesBusinessConnection interface {
	// OnBusinessConnection Бот был подключен к бизнес-аккаунту или отключен от него, или пользователь отредактировал
	// существующее соединение с ботом.
	OnBusinessConnection(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesBusinessMessage Метод получающий обновления:
// Новое сообщение от подключенного корпоративного аккаунта.
type IUpdatesBusinessMessage interface {
	// OnBusinessMessage Новое сообщение от подключенного корпоративного аккаунта.
	OnBusinessMessage(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesEditedBusinessMessage Метод получающий обновления:
// Новая версия сообщения от подключенного бизнес-аккаунта.
type IUpdatesEditedBusinessMessage interface {
	// OnEditedBusinessMessage Новая версия сообщения от подключенного бизнес-аккаунта.
	OnEditedBusinessMessage(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesDeletedBusinessMessages Метод получающий обновления:
// Сообщения были удалены из подключенного корпоративного аккаунта.
type IUpdatesDeletedBusinessMessages interface {
	// OnDeletedBusinessMessages Сообщения были удалены из подключенного корпоративного аккаунта.
	OnDeletedBusinessMessages(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageReaction Метод получающий обновления:
// Пользователь изменил реакцию на сообщение.
type IUpdatesMessageReaction interface {
	// OnMessageReaction Пользователь изменил реакцию на сообщение.
	OnMessageReaction(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMessageReactionCount Метод получающий обновления:
// Изменены реакции на сообщение с анонимными реакциями.
type IUpdatesMessageReactionCount interface {
	// OnMessageReactionCount Изменены реакции на сообщение с анонимными реакциями.
	OnMessageReactionCount(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesInlineQuery Метод получающий обновления:
// Новый входящий встроенный запрос.
type IUpdatesInlineQuery interface {
	// OnInlineQuery Новый входящий встроенный запрос.
	OnInlineQuery(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesChosenInlineResult Метод получающий обновления:
// Результат встроенного запроса, который пользователь выбрал и отправил своему собеседнику в чате.
type IUpdatesChosenInlineResult interface {
	// OnChosenInlineResult Результат встроенного запроса, который пользователь выбрал и отправил своему собеседнику в чате.
	OnChosenInlineResult(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesCallbackQuery Метод получающий обновления:
// Новый входящий запрос обратного вызова.
type IUpdatesCallbackQuery interface {
	// OnCallbackQuery Новый входящий запрос обратного вызова.
	OnCallbackQuery(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesShippingQuery Метод получающий обновления:
// Новый запрос на доставку. Только для счетов с ценой зависящей от доставки.
type IUpdatesShippingQuery interface {
	// OnShippingQuery Новый запрос на доставку. Только для счетов с ценой зависящей от доставки.
	OnShippingQuery(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesPreCheckoutQuery Метод получающий обновления:
// Новый входящий запрос перед оформлением заказа.
// Содержит полную информацию об оформлении заказа.
type IUpdatesPreCheckoutQuery interface {
	// OnPreCheckoutQuery Новый входящий запрос перед оформлением заказа.
	// Содержит полную информацию об оформлении заказа.
	OnPreCheckoutQuery(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesPurchasedPaidMedia Метод получающий обновления:
// Пользователь приобрел платный контент с полезной нагрузкой, отправленной ботом в чате без канала.
type IUpdatesPurchasedPaidMedia interface {
	// OnPurchasedPaidMedia Пользователь приобрел платный контент с полезной нагрузкой, отправленной ботом
	// в чате без канала.
	OnPurchasedPaidMedia(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesPoll Метод получающий обновления:
// Новое состояние опроса. Боты получают только уведомления об опросах, остановленных вручную, и об опросах,
// отправленных ботом.
type IUpdatesPoll interface {
	// OnPoll Новое состояние опроса. Боты получают только уведомления об опросах, остановленных вручную, и об опросах,
	// отправленных ботом.
	OnPoll(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesPollAnswer Метод получающий обновления:
// Пользователь изменил свой ответ в не анонимном опросе. Боты получают новые голоса только в опросах,
// отправленных ими самими.
type IUpdatesPollAnswer interface {
	// OnPollAnswer Пользователь изменил свой ответ в не анонимном опросе. Боты получают новые голоса только в опросах,
	// отправленных ими самими.
	OnPollAnswer(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesMyChatMember Метод получающий обновления:
// Статус участника чата с ботом был обновлен в чате. В личных чатах это обновление отображается
// только в том случае, если пользователь заблокировал или разблокировал бота.
type IUpdatesMyChatMember interface {
	// OnMyChatMember Статус участника чата с ботом был обновлен в чате. В личных чатах это обновление отображается
	// только в том случае, если пользователь заблокировал или разблокировал бота.
	OnMyChatMember(api *tgbotapi.BotAPI, upd *tgbotapi.Update, user *tgbotapi.User)
}

// IUpdatesChatMember Метод получающий обновления:
// Статус участника чата был обновлен в чате. Чтобы получать такие обновления, бот должен быть
// администратором чата и явно указать "chat_member" в списке allowed_updates.
type IUpdatesChatMember interface {
	// OnChatMember Статус участника чата был обновлен в чате. Чтобы получать такие обновления, бот должен
	// быть администратором чата и явно указать "chat_member" в списке allowed_updates.
	OnChatMember(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesChatJoinRequest Метод получающий обновления:
// Отправлен запрос на присоединение к чату. Чтобы получать эти обновления, бот должен
// иметь права администратора can_invite_users в чате.
type IUpdatesChatJoinRequest interface {
	// OnChatJoinRequest Отправлен запрос на присоединение к чату. Чтобы получать эти обновления, бот должен
	// иметь права администратора can_invite_users в чате.
	OnChatJoinRequest(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesChatBoost Метод получающий обновления:
// Добавлена или изменена функция ускорения чата. Чтобы получать эти обновления, бот должен
// быть администратором чата.
type IUpdatesChatBoost interface {
	// OnChatBoost Добавлена или изменена функция ускорения чата. Чтобы получать эти обновления, бот должен
	// быть администратором чата.
	OnChatBoost(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesRemovedChatBoost Метод получающий обновления:
// Буст был удален из чата. Чтобы получать эти обновления, бот должен быть администратором чата.
type IUpdatesRemovedChatBoost interface {
	// OnRemovedChatBoost Буст был удален из чата. Чтобы получать эти обновления, бот должен быть администратором чата.
	OnRemovedChatBoost(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

// IUpdatesManagedBot Метод получающий обновления:
// Обновления о создании управляемых ботов и смене их токена.
type IUpdatesManagedBot interface {
	OnManagedBot(api *tgbotapi.BotAPI, upd *tgbotapi.Update)
}

package tgbot

/*

	Константы, описывают типы сообщений, приходящих от сервера телеграм, которые можно разрешить для получения ботом.

*/

const (
	// AllowedUpdateMessage Новое входящее сообщение любого типа — текст, фото, стикер и т.д.
	AllowedUpdateMessage = UpdatesIncomingType("message")

	// AllowedUpdateEditedMessage Новая версия сообщения, известная боту и отредактированная.
	// Иногда это обновление может быть вызвано изменениями в полях сообщения, которые либо недоступны,
	// либо не используются вашим ботом.
	AllowedUpdateEditedMessage = UpdatesIncomingType("edited_message")

	// AllowedUpdateChannelPost Новый входящий пост в канале любого типа — текст, фото, стикер и т.д.
	AllowedUpdateChannelPost = UpdatesIncomingType("channel_post")

	// AllowedUpdateEditedChannelPost Новая версия публикации в канале, известная боту и отредактированная.
	// Иногда это обновление может быть вызвано изменениями в полях сообщения, которые либо недоступны,
	// либо не используются вашим ботом.
	AllowedUpdateEditedChannelPost = UpdatesIncomingType("edited_channel_post")

	// AllowedUpdateBusinessConnection Бот был подключен к бизнес-аккаунту или отключен от него,
	// либо пользователь изменил существующее подключение к боту.
	AllowedUpdateBusinessConnection = UpdatesIncomingType("business_connection")

	// AllowedUpdateBusinessMessage Новое сообщение от подключенного корпоративного аккаунта.
	AllowedUpdateBusinessMessage = UpdatesIncomingType("business_message")

	// AllowedUpdateEditedBusinessMessage Новая версия сообщения от подключенного бизнес-аккаунта.
	AllowedUpdateEditedBusinessMessage = UpdatesIncomingType("edited_business_message")

	// AllowedUpdateDeletedBusinessMessages Сообщения были удалены из подключенного корпоративного аккаунта.
	AllowedUpdateDeletedBusinessMessages = UpdatesIncomingType("deleted_business_messages")

	// AllowedUpdateMessageReaction Пользователь изменил реакцию на сообщение. Чтобы получать такие обновления,
	// бот должен быть администратором чата и явно указать "message_reaction" в списке allowed_updates.
	// Обновления не принимаются для реакций, установленных ботами.
	AllowedUpdateMessageReaction = UpdatesIncomingType("message_reaction")

	// AllowedUpdateMessageReactionCount Изменены реакции на сообщение с анонимными реакциями.
	// Чтобы получать эти обновления, бот должен быть администратором чата и явно указать "message_reaction_count"
	// в списке allowed_updates. Обновления группируются и могут отправляться с задержкой до нескольких минут.
	AllowedUpdateMessageReactionCount = UpdatesIncomingType("message_reaction_count")

	// AllowedUpdateInlineQuery Новый входящий встроенный запрос.
	AllowedUpdateInlineQuery = UpdatesIncomingType("inline_query")

	// AllowedUpdateChosenInlineResult Результат встроенного запроса, который пользователь выбрал и отправил своему
	// собеседнику в чате. Подробнее о том, как включить эти обновления для вашего бота, читайте в нашей документации
	// по сбору отзывов.
	AllowedUpdateChosenInlineResult = UpdatesIncomingType("chosen_inline_result")

	// AllowedUpdateCallbackQuery Новый входящий запрос обратного вызова.
	AllowedUpdateCallbackQuery = UpdatesIncomingType("callback_query")

	// AllowedUpdateShippingQuery Новый запрос на доставку. Только для счетов с ценой зависящей от доставки.
	AllowedUpdateShippingQuery = UpdatesIncomingType("shipping_query")

	// AllowedUpdatePreCheckoutQuery Новый входящий запрос перед оформлением заказа.
	// Содержит полную информацию об оформлении заказа.
	AllowedUpdatePreCheckoutQuery = UpdatesIncomingType("pre_checkout_query")

	// AllowedUpdatePurchasedPaidMedia Пользователь приобрел платный контент с полезной нагрузкой,
	// отправленной ботом в чате без канала.
	AllowedUpdatePurchasedPaidMedia = UpdatesIncomingType("purchased_paid_media")

	// AllowedUpdatePoll Новое состояние опроса. Боты получают только уведомления об опросах, остановленных вручную,
	// и об опросах, отправленных ботом.
	AllowedUpdatePoll = UpdatesIncomingType("poll")

	// AllowedUpdatePollAnswer Пользователь изменил свой ответ в не анонимном опросе.
	// Боты получают новые голоса только в опросах, отправленных ими самими.
	AllowedUpdatePollAnswer = UpdatesIncomingType("poll_answer")

	// AllowedUpdateMyChatMember Статус участника чата с ботом был обновлен в чате.
	// В личных чатах это обновление отображается только в том случае, если пользователь
	// заблокировал или разблокировал бота.
	AllowedUpdateMyChatMember = UpdatesIncomingType("my_chat_member")

	// AllowedUpdateChatMember Статус участника чата был обновлен в чате.
	// Чтобы получать такие обновления, бот должен быть администратором чата и явно указать "chat_member" в
	// списке allowed_updates.
	AllowedUpdateChatMember = UpdatesIncomingType("chat_member")

	// AllowedUpdateChatJoinRequest Отправлен запрос на присоединение к чату.
	// Чтобы получать эти обновления, бот должен иметь право администратора can_invite_users в чате.
	AllowedUpdateChatJoinRequest = UpdatesIncomingType("chat_join_request")

	// AllowedUpdateChatBoost Добавлена или изменена функция ускорения чата.
	// Чтобы получать эти обновления, бот должен быть администратором чата.
	AllowedUpdateChatBoost = UpdatesIncomingType("chat_boost")

	// AllowedUpdateRemovedChatBoost Буст был удален из чата.
	// Чтобы получать эти обновления, бот должен быть администратором чата.
	AllowedUpdateRemovedChatBoost = UpdatesIncomingType("removed_chat_boost")

	// AllowedUpdateManagedBot Был создан новый бот, которым будет управлять ботом, или был изменен токен
	// или владелец управляемого бота.
	AllowedUpdateManagedBot = UpdatesIncomingType("managed_bot")
)

package tgbot

// Registration Регистрация объектов, отвечающих за функционал бота.
// К регистрации допускаются объекты, которые удовлетворяют хотя бы один интерфейс из перечисленных в IUpdatesFull.
// Регистрируемый объект должен удовлетворять как минимум один интерфейс обновлений.
//
// Если регистрируемый объект не удовлетворяет ни одному интерфейсу, тем самым не содержит ни одного
// метода, тогда функция регистрации вернёт ошибку, а сам объект не будет зарегистрирован.
//
//goland:noinspection DuplicatedCode
func (tbt *impl) Registration(obj any) (err error) {
	var countSubscription uint64

	tbt.subscriptionSync.Lock()
	defer tbt.subscriptionSync.Unlock()
	if obj == nil {
		err = tbt.Errors().RegistrationNilObject.Bind()
		return
	}
	// Проверка, какие интерфейсы реализованы в объекте и подписка объекта на соответствующие интерфейсам события.
	if item, ok := obj.(IUpdatesTelegramUser); ok {
		tbt.subscription[eventOnTelegramUser], countSubscription =
			append(tbt.subscription[eventOnTelegramUser], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageAudio); ok {
		tbt.subscription[eventOnAudio], countSubscription =
			append(tbt.subscription[eventOnAudio], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageDocument); ok {
		tbt.subscription[eventOnDocument], countSubscription =
			append(tbt.subscription[eventOnDocument], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessagePhoto); ok {
		tbt.subscription[eventOnPhoto], countSubscription =
			append(tbt.subscription[eventOnPhoto], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageSticker); ok {
		tbt.subscription[eventOnSticker], countSubscription =
			append(tbt.subscription[eventOnSticker], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageStory); ok {
		tbt.subscription[eventOnStory], countSubscription =
			append(tbt.subscription[eventOnStory], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageVideo); ok {
		tbt.subscription[eventOnVideo], countSubscription =
			append(tbt.subscription[eventOnVideo], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageVideoNote); ok {
		tbt.subscription[eventOnVideoNote], countSubscription =
			append(tbt.subscription[eventOnVideoNote], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageVoice); ok {
		tbt.subscription[eventOnVoice], countSubscription =
			append(tbt.subscription[eventOnVoice], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageChecklist); ok {
		tbt.subscription[eventOnChecklist], countSubscription =
			append(tbt.subscription[eventOnChecklist], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageContact); ok {
		tbt.subscription[eventOnContact], countSubscription =
			append(tbt.subscription[eventOnContact], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageDice); ok {
		tbt.subscription[eventOnDice], countSubscription =
			append(tbt.subscription[eventOnDice], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageGame); ok {
		tbt.subscription[eventOnGame], countSubscription =
			append(tbt.subscription[eventOnGame], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessagePoll); ok {
		tbt.subscription[eventOnPoll], countSubscription =
			append(tbt.subscription[eventOnPoll], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageVenue); ok {
		tbt.subscription[eventOnVenue], countSubscription =
			append(tbt.subscription[eventOnVenue], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageLocation); ok {
		tbt.subscription[eventOnLocation], countSubscription =
			append(tbt.subscription[eventOnLocation], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageInvoice); ok {
		tbt.subscription[eventOnInvoice], countSubscription =
			append(tbt.subscription[eventOnInvoice], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessage); ok {
		tbt.subscription[eventOnMessage], countSubscription =
			append(tbt.subscription[eventOnMessage], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesEditedMessage); ok {
		tbt.subscription[eventOnEditedMessage], countSubscription =
			append(tbt.subscription[eventOnEditedMessage], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesChannelPost); ok {
		tbt.subscription[eventOnChannelPost], countSubscription =
			append(tbt.subscription[eventOnChannelPost], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesEditedChannelPost); ok {
		tbt.subscription[eventOnEditedChannelPost], countSubscription =
			append(tbt.subscription[eventOnEditedChannelPost], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesBusinessConnection); ok {
		tbt.subscription[eventOnBusinessConnection], countSubscription =
			append(tbt.subscription[eventOnBusinessConnection], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesBusinessMessage); ok {
		tbt.subscription[eventOnBusinessMessage], countSubscription =
			append(tbt.subscription[eventOnBusinessMessage], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesEditedBusinessMessage); ok {
		tbt.subscription[eventOnEditedBusinessMessage], countSubscription =
			append(tbt.subscription[eventOnEditedBusinessMessage], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesDeletedBusinessMessages); ok {
		tbt.subscription[eventOnDeletedBusinessMessages], countSubscription =
			append(tbt.subscription[eventOnDeletedBusinessMessages], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageReaction); ok {
		tbt.subscription[eventOnMessageReaction], countSubscription =
			append(tbt.subscription[eventOnMessageReaction], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMessageReactionCount); ok {
		tbt.subscription[eventOnMessageReactionCount], countSubscription =
			append(tbt.subscription[eventOnMessageReactionCount], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesInlineQuery); ok {
		tbt.subscription[eventOnInlineQuery], countSubscription =
			append(tbt.subscription[eventOnInlineQuery], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesChosenInlineResult); ok {
		tbt.subscription[eventOnChosenInlineResult], countSubscription =
			append(tbt.subscription[eventOnChosenInlineResult], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesCallbackQuery); ok {
		tbt.subscription[eventOnCallbackQuery], countSubscription =
			append(tbt.subscription[eventOnCallbackQuery], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesShippingQuery); ok {
		tbt.subscription[eventOnShippingQuery], countSubscription =
			append(tbt.subscription[eventOnShippingQuery], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesPreCheckoutQuery); ok {
		tbt.subscription[eventOnPreCheckoutQuery], countSubscription =
			append(tbt.subscription[eventOnPreCheckoutQuery], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesPurchasedPaidMedia); ok {
		tbt.subscription[eventOnPurchasedPaidMedia], countSubscription =
			append(tbt.subscription[eventOnPurchasedPaidMedia], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesPoll); ok {
		tbt.subscription[eventOnPoll], countSubscription =
			append(tbt.subscription[eventOnPoll], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesPollAnswer); ok {
		tbt.subscription[eventOnPollAnswer], countSubscription =
			append(tbt.subscription[eventOnPollAnswer], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesMyChatMember); ok {
		tbt.subscription[eventOnMyChatMember], countSubscription =
			append(tbt.subscription[eventOnMyChatMember], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesChatMember); ok {
		tbt.subscription[eventOnChatMember], countSubscription =
			append(tbt.subscription[eventOnChatMember], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesChatJoinRequest); ok {
		tbt.subscription[eventOnChatJoinRequest], countSubscription =
			append(tbt.subscription[eventOnChatJoinRequest], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesChatBoost); ok {
		tbt.subscription[eventOnChatBoost], countSubscription =
			append(tbt.subscription[eventOnChatBoost], item), countSubscription+1
	}
	if item, ok := obj.(IUpdatesRemovedChatBoost); ok {
		tbt.subscription[eventOnRemovedChatBoost], countSubscription =
			append(tbt.subscription[eventOnRemovedChatBoost], item), countSubscription+1
	}
	if countSubscription == 0 {
		err = tbt.Errors().RegistrationInterface.Bind()
		return
	}

	return
}

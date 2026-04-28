package tgbot

const (
	// Размер буфера канала входящих обновлений.
	inputChannelBufferSize = 500

	// Тайм-аут в секундах для длительного опроса сервера, телеграм для получения обновлений.
	// По умолчанию 0, то есть используется обычный короткий опрос.
	// Значение должно быть положительным, короткий опрос следует использовать только для тестирования.
	getUpdatesRequestTimeout = 60
)

// Константы событий входящих сообщений.
const (
	eventOnTelegramUser            = iota + 1 //  1
	eventOnAudio                              //  2
	eventOnDocument                           //  3
	eventOnPhoto                              //  4
	eventOnSticker                            //  5
	eventOnStory                              //  6
	eventOnVideo                              //  7
	eventOnVideoNote                          //  8
	eventOnVoice                              //  9
	eventOnChecklist                          // 10
	eventOnContact                            // 11
	eventOnDice                               // 12
	eventOnGame                               // 13
	eventOnVenue                              // 14
	eventOnLocation                           // 15
	eventOnInvoice                            // 16
	eventOnMessage                            // 17
	eventOnEditedMessage                      // 18
	eventOnChannelPost                        // 19
	eventOnEditedChannelPost                  // 20
	eventOnBusinessConnection                 // 21
	eventOnBusinessMessage                    // 22
	eventOnEditedBusinessMessage              // 23
	eventOnDeletedBusinessMessages            // 24
	eventOnMessageReaction                    // 25
	eventOnMessageReactionCount               // 26
	eventOnInlineQuery                        // 27
	eventOnChosenInlineResult                 // 28
	eventOnCallbackQuery                      // 29
	eventOnShippingQuery                      // 30
	eventOnPreCheckoutQuery                   // 31
	eventOnPurchasedPaidMedia                 // 32
	eventOnPoll                               // 33
	eventOnPollAnswer                         // 34
	eventOnMyChatMember                       // 35
	eventOnChatMember                         // 36
	eventOnChatJoinRequest                    // 37
	eventOnChatBoost                          // 38
	eventOnRemovedChatBoost                   // 39
	eventOnManagedBot                         // 40
)

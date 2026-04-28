package tgbot

import tgbotapi "github.com/webnice/tba/v9"

// Chain возвращает тип промежуточного программного обеспечения из набора обработчиков
// промежуточного программного обеспечения.
func Chain(middlewares ...func(Handler) Handler) Middlewares {
	return middlewares
}

// Handler создает и возвращает обработчик из цепочки промежуточных программ с конечным обработчиком.
func (mws Middlewares) Handler(h Handler) Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// HandlerFunc создает и возвращает обработчик из цепочки промежуточных программ обработчиком в качестве
// конечного обработчика.
func (mws Middlewares) HandlerFunc(h HandlerFunc) Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// ChainHandler является обработчиком с поддержкой составления и выполнения обработчика.
type ChainHandler struct {
	Endpoint    Handler
	chain       Handler
	Middlewares Middlewares
}

func (c *ChainHandler) ServeTelegram(api *tgbotapi.BotAPI, upd *tgbotapi.Update) {
	c.chain.ServeTelegram(api, upd)
}

// Создаёт обработчик, состоящий из встроенного стека промежуточного программного обеспечения и обработчика
// конечной точки в порядке их передачи.
func chain(middlewares []func(Handler) Handler, endpoint Handler) (ret Handler) {
	var n int

	// Завершаем выполнение, если в цепочке нет промежуточных.
	if len(middlewares) == 0 {
		ret = endpoint
		return
	}
	// Оборачивание конечного обработчика цепочкой промежуточного программного обеспечения.
	ret = middlewares[len(middlewares)-1](endpoint)
	for n = len(middlewares) - 2; n >= 0; n-- {
		ret = middlewares[n](ret)
	}

	return
}

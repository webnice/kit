package tgbot

import tgbotapi "github.com/webnice/tba/v9"

// BusSubscribe Подписка на сообщения из шины данных.
func (tbt *impl) BusSubscribe() {
	var err error

	switch err = tbt.cfg.Bus().
		Subscribe(tbt); {
	case err == nil:
	default:
		tbt.cfg.Gist().ErrorAppend(err)
		return
	}
}

// BusUnsubscribe Отписка от сообщений из шины данных.
func (tbt *impl) BusUnsubscribe() {
	var err error

	switch err = tbt.cfg.Bus().
		Unsubscribe(tbt); {
	case err == nil:
	default:
		tbt.cfg.Gist().ErrorAppend(err)
		return
	}
}

// KnownType Функция вызывается один раз, при регистрации подписчика в шине данных и должна вернуть срез структур
// данных, которые готов получать подписчик. Для получения данных любого типа, необходимо вернуть срез нулевой длины.
func (tbt *impl) KnownType() (ret []interface{}) {
	return []interface{}{
		new(tgbotapi.Update), // Новое сообщение от сервера телеграм.
	}
}

// Consumer Функция получения данных из шины данных.
// В функцию передаются объекты, типы структур которых были получены через вызов функции KnownType()
// Для синхронного вызова, функция должна вернуть ответа, он будет передан издателю.
// Для асинхронного вызова, функция не должна возвращать никакие данные, ничего не будет передаваться издателю.
func (tbt *impl) Consumer(_ bool, data interface{}) (ret []interface{}, errs []error) {
	switch d := data.(type) {
	// Новое обновление от сервера телеграм.
	case *tgbotapi.Update:
		tbt.msgInp <- d
	// Неожиданные данные.
	default:
		tbt.log().Warning(tbt.Errors().BusUnknownEvent.Bind(d))
		return
	}

	return
}

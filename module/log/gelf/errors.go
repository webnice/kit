package gelf

import "github.com/webnice/dic"

// Коды ошибок.
const (
	eUnknownProtocol         uint8 = iota + 1 // 001
	eInvalidAddress                           // 002
	eUnknownCompression                       // 003
	eCompressionNotSupported                  // 004
	eMessageTooLarge                          // 005
	eNotConnected                             // 006
)

// Текстовые значения кодов ошибок на основном языке приложения.
const (
	cUnknownProtocol         = "Не известный протокол %q, поддерживаются протоколы: udp, tcp, tls."
	cInvalidAddress          = "Адрес сервера %q не является корректным: %s."
	cUnknownCompression      = "Не известный тип сжатия сообщений %d."
	cCompressionNotSupported = "Сжатие сообщений %q не поддерживается протоколом %q."
	cMessageTooLarge         = "Размер сообщения %d байт превышает максимально допустимый размер %d байт для отправки по протоколу UDP."
	cNotConnected            = "Соединение с сервером не установлено."
)

// Error Справочник всех ошибок известного состояния, которые может вернуть приложение или функция.
type Error struct {
	dic.Errors

	// UnknownProtocol Не известный протокол ..., поддерживаются протоколы: udp, tcp, tls.
	UnknownProtocol dic.IError

	// InvalidAddress Адрес сервера ... не является корректным: ....
	InvalidAddress dic.IError

	// UnknownCompression Не известный тип сжатия сообщений ....
	UnknownCompression dic.IError

	// CompressionNotSupported Сжатие сообщений ... не поддерживается протоколом ....
	CompressionNotSupported dic.IError

	// MessageTooLarge Размер сообщения ... байт превышает максимально допустимый размер ... байт для отправки по протоколу UDP.
	MessageTooLarge dic.IError

	// NotConnected Соединение с сервером не установлено.
	NotConnected dic.IError
}

var (
	errSingleton = &Error{
		Errors:                  dic.Error(),
		UnknownProtocol:         dic.NewError(cUnknownProtocol, "протокол").CodeU8().Set(eUnknownProtocol),
		InvalidAddress:          dic.NewError(cInvalidAddress, "адрес", "причина").CodeU8().Set(eInvalidAddress),
		UnknownCompression:      dic.NewError(cUnknownCompression, "тип сжатия").CodeU8().Set(eUnknownCompression),
		CompressionNotSupported: dic.NewError(cCompressionNotSupported, "тип сжатия", "протокол").CodeU8().Set(eCompressionNotSupported),
		MessageTooLarge:         dic.NewError(cMessageTooLarge, "размер сообщения", "максимальный размер").CodeU8().Set(eMessageTooLarge),
		NotConnected:            dic.NewError(cNotConnected).CodeU8().Set(eNotConnected),
	}

	// Errors Справочник ошибок.
	Errors = func() *Error { return errSingleton }
)

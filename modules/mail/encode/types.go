package encode

import "mime"

const (
	// DefaultCharset Кодировка по умолчанию
	DefaultCharset string = `UTF-8`

	// _encodeBase64 base64 encoding as defined in RFC 2045
	_encodeBase64 string = "base64"
	// _encodeQuotedPrintable quoted-printable encoding as defined in RFC 2045
	_encodeQuotedPrintable string = "quoted-printable"
	// _encodeUnencoded no using encoding
	_encodeUnencoded string = "8bit"
)

// Interface encode interface
type Interface interface {
	Encoder(Encoding) Interface                      // Назначение кодировщика
	Charset(string) Interface                        // Назначение кодировки
	GetCharset() string                              // Получение текущей кодеровки
	EncodeString(string) string                      // Кодирование строки
	EmailAddress(address string, name string) string // Кодирование электронного адреса
	String() string                                  // Возвращает название установленного кодировщика
}

// impl encode implementation
type impl struct {
	encoder  Encoder // Кодировщик
	charset  string  // Кодировка
	encoding string  // Название кодировщика для MIME
}

// Encoder type
type Encoder struct {
	mime.WordEncoder
}

// Encoding is a MIME encoding scheme
type Encoding struct {
	encoding string  // Название кодировщика для MIME
	encoder  Encoder // Кодировщик
}

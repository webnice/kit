package multipart

import (
	"io"
	"mime/multipart"

	"github.com/webnice/kit/module/mail/encode"
)

const (
	// PartBody type of Part
	PartBody PartType = `body`

	// PartAttach type of Part
	PartAttach PartType = `attach`

	// PartEmbed type of Part
	PartEmbed PartType = `embed`

	_Inline                  = `inline`
	_Attachment              = `attachment`
	_ContentType             = `Content-Type`
	_ContentTransferEncoding = `Content-Transfer-Encoding`
	_ContentDisposition      = `Content-Disposition`
	_ContentID               = `Content-ID`
	_EncodeBase64            = `base64`
	_EncodeQuotedPrintable   = `quoted-printable`
	_Encode8bit              = `8bit`
)

// Interface multipart interface
type Interface interface {
	Count() int64 // Возвращает количество записанных байт во врайтер
	Error() error // Возвращает последнюю ошибку

	SetWriter(wr io.Writer) Interface                               // Назначение io.Writer
	SetStringWriter(f func(io.Writer, string) int) Interface        // Назначение функции WriteHeader
	SetHeaderWriter(f func(io.Writer, string, ...string)) Interface // Назначение функции WriteHeader
	CreateNewMultipart(string)                                      // Создание нового вложенного мультипарта
	CloseMultipart()                                                // Закрытие мультипарт в порядке вложенности
	WritePart(*Part)                                                // Запись куска
}

// impl multipart implementation
type impl struct {
	count     int64 // Количество записанных байт во врайтер
	lastError error // Последняя ошибка

	writer io.Writer // Врайтер

	nesting    uint8                // Вложенность мультипартов
	multiparts [3]*multipart.Writer // Все вложенные мультипарты
	partWriter io.Writer            // Текущий part в мультипарт

	writeHeader func(io.Writer, string, ...string)
	writeString func(io.Writer, string) int
}

// Part of message
type Part struct {
	FileName    string           // Название файла для attach,embed
	Type        PartType         // Название типа (body,attach,embed)
	ContentType string           // Content-Type куска
	Body        io.Reader        // Тело сообщения
	Encoding    encode.Interface // Кодировщик куска
	Save        bool             // Save to writer
}

// PartType Название типа
type PartType string

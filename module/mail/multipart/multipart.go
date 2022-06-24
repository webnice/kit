package multipart

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/webnice/kit/v2/module/mail/linelimit"
	"github.com/webnice/kit/v2/module/mail/qp"
)

// New interface
func New(writer io.Writer) Interface {
	var mpt = new(impl)
	mpt.writer = writer
	return mpt
}

// CreateNewMultipart Создание нового вложенного мультипарта
func (mpt *impl) CreateNewMultipart(mimeTypeName string) {
	var writer *multipart.Writer
	var contentType string

	writer = multipart.NewWriter(mpt.writer)
	contentType = fmt.Sprintf(`multipart/%s;`+"\r\n\t"+`boundary="%s"`, mimeTypeName, writer.Boundary())
	mpt.multiparts[mpt.nesting] = writer
	if mpt.nesting == 0 {
		mpt.WriteHeader(_ContentType, contentType)
		mpt.WriteString("\r\n")
	} else {
		mpt.CreateNewPart(map[string][]string{_ContentType: {contentType}})
	}
	mpt.nesting++
}

// CreateNewPart Создание единичного part
func (mpt *impl) CreateNewPart(mimeHeader map[string][]string) {
	mpt.partWriter, mpt.lastError = mpt.multiparts[mpt.nesting-1].CreatePart(mimeHeader)
}

// CloseMultipart Закрытие мультипарт в порядке вложенности
func (mpt *impl) CloseMultipart() {
	if mpt.nesting <= 0 {
		return
	}
	mpt.lastError = mpt.multiparts[mpt.nesting-1].Close()
	mpt.nesting--
}

// WritePart Запись куска
func (mpt *impl) WritePart(part *Part) {
	var contentType, contentDisposition string

	switch part.Type {
	case PartEmbed:
		contentDisposition = _Inline
	case PartAttach:
		contentDisposition = _Attachment
	}

	if mpt.nesting > 0 {
		mpt.WriteString(fmt.Sprintf("\r\n--%s\r\n", mpt.multiparts[mpt.nesting-1].Boundary()))
	}

	switch part.Type {
	case PartBody:
		contentType = fmt.Sprintf("%s;\r\n\tcharset=%s", part.ContentType, part.Encoding.GetCharset())
		mpt.WriteHeader(_ContentType, contentType)
		mpt.WriteHeader(_ContentTransferEncoding, part.Encoding.String())
	case PartEmbed:
		contentType = fmt.Sprintf(`%s; name="%s"`, part.ContentType, part.FileName)
		mpt.WriteHeader(_ContentType, contentType)
		mpt.WriteHeader(_ContentTransferEncoding, part.Encoding.String())
		mpt.WriteHeader(_ContentDisposition, contentDisposition)
		mpt.WriteHeader(_ContentID, fmt.Sprintf(`<%s>`, part.FileName))
	case PartAttach:
		contentType = fmt.Sprintf(`%s; name="%s"`, part.ContentType, part.FileName)
		mpt.WriteHeader(_ContentType, contentType)
		mpt.WriteHeader(_ContentTransferEncoding, part.Encoding.String())
		mpt.WriteHeader(_ContentDisposition, contentDisposition)
	}
	mpt.WriteData(part.Body, part.Encoding)
}

// WriteData Запись тела данных
func (mpt *impl) WriteData(rdr io.Reader, enc fmt.Stringer) {
	var writer io.Writer
	var wc io.WriteCloser
	var qpwr *qp.Writer

	if mpt.nesting <= 1 {
		writer = mpt.writer
	} else {
		writer = mpt.partWriter
	}
	mpt.WriteString("\r\n")

	switch enc.String() {
	case _EncodeBase64:
		wc = base64.NewEncoder(base64.StdEncoding, linelimit.New(writer))
		defer func() { _ = wc.Close() }()
		_, mpt.lastError = io.Copy(wc, rdr)
	case _EncodeQuotedPrintable:
		qpwr = qp.NewWriter(writer)
		_, mpt.lastError = io.Copy(qpwr, rdr)
		_ = qpwr.Close()
	case _Encode8bit:
		_, mpt.lastError = io.Copy(writer, rdr)
	default:
		mpt.lastError = fmt.Errorf("Unknown encodig scheme")
	}
}

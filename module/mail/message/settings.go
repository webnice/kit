package message

import (
	"bytes"
	"io"
	"mime"
	"net/mail"
	"strings"

	"github.com/webnice/kit/module/mail/encode"
	"github.com/webnice/kit/module/mail/multipart"
)

// Encoder Set encoder
func (msg *impl) Encoder(encoder encode.Interface) Interface {
	msg.encoder = encoder
	return msg
}

// isUnique =true-unique, =false-not unique
func (msg *impl) isUnique(str string) (ret string, ok bool) {
	var address string

	ret = str

	// Монопольный доступ
	msg.uniqueTo.Lock()
	defer msg.uniqueTo.Unlock()

	if msg.uniqueTo.String == nil {
		msg.uniqueTo.String = make(map[string]bool)
	}
	address = msg.extractEMail(str)
	_, ok = msg.uniqueTo.String[address]
	if ok = !ok; ok {
		msg.uniqueTo.String[address] = true
	}

	return
}

// extractEMail Выделение из строки e-mail адреса
func (msg *impl) extractEMail(str string) string {
	var address string
	var from, to int
	address = strings.TrimSpace(str)
	address = strings.ToLower(address)
	from = strings.Index(address, "<")
	to = strings.Index(address, ">")
	if from >= 0 && to > from {
		address = address[from+1 : to]
	}
	return address
}

// From Set source address
func (msg *impl) From(address string) Interface {
	msg.from = address
	return msg
}

// To Set destination addresses
func (msg *impl) To(toAddresses ...string) Interface {
	var tmp string
	var ok bool
	for i := range toAddresses {
		if tmp, ok = msg.isUnique(toAddresses[i]); ok {
			msg.to = append(msg.to, tmp)
		}
	}
	return msg
}

// ToCC Set destination addresses carbon copy
func (msg *impl) ToCC(toCCAddresses ...string) Interface {
	for i := range toCCAddresses {
		if tmp, ok := msg.isUnique(toCCAddresses[i]); ok {
			msg.toCC = append(msg.toCC, tmp)
		}
	}
	return msg
}

// ToBCC Set destination addresses blind carbon copy
func (msg *impl) ToBCC(toBCCAddresses ...string) Interface {
	for i := range toBCCAddresses {
		tmp, ok := msg.isUnique(toBCCAddresses[i])
		if ok {
			msg.toCC = append(msg.toCC, tmp)
		}
	}
	return msg
}

// Subject Set subject
func (msg *impl) Subject(subject string) Interface {
	msg.subject = subject
	return msg
}

// Body Set body data. Default Content-Type is text/plain
func (msg *impl) Body(buf *bytes.Buffer, contentType ...string) Interface {
	return msg.BodyReader(bytes.NewReader(buf.Bytes()), contentType...)
}

// BodyReader Set body reader. Default Content-Type is text/plain
func (msg *impl) BodyReader(rdr io.Reader, contentTypes ...string) Interface {
	msg.AddPart(rdr, "", multipart.PartBody, contentTypes...)
	return msg
}

// Attach Add attach file. Default Content-Type is text/plain
func (msg *impl) Attach(rdr io.Reader, fileName string, contentTypes ...string) Interface {
	var contentType string
	for i := range contentTypes {
		if contentTypes[i] != "" {
			contentType = contentTypes[i]
		}
	}
	if contentType == "" {
		contentType = mime.TypeByExtension(fileName)
	}
	if contentType == "" {
		contentType = DeafultAttachContentType
	}
	msg.AddPart(rdr, fileName, multipart.PartAttach, contentType)
	return msg
}

// Embed Add embeded file. Default Content-Type is text/plain
func (msg *impl) Embed(rdr io.Reader, fileName string, contentTypes ...string) Interface {
	var contentType string
	var i int
	for i = range contentTypes {
		if contentTypes[i] != "" {
			contentType = contentTypes[i]
		}
	}
	if contentType == "" {
		contentType = mime.TypeByExtension(fileName)
	}
	if contentType == "" {
		contentType = DeafultAttachContentType
	}
	msg.AddPart(rdr, fileName, multipart.PartEmbed, contentType)
	return msg
}

// AddPart Добавление куска
func (msg *impl) AddPart(body io.Reader, fileName string, typeName multipart.PartType, contentTypes ...string) {
	var contentType string
	var i int
	for i = range contentTypes {
		if contentTypes[i] != "" {
			contentType = contentTypes[i]
		}
	}
	if contentType == "" {
		contentType = DeafultContentType
	}

	msg.Parts = append(msg.Parts, &multipart.Part{
		FileName:    fileName,
		Type:        typeName,
		ContentType: contentType,
		Body:        body,
		Encoding:    msg.encoder,
	})

}

// Error Return last error
func (msg *impl) Error() error { return msg.lastError }

// EmailAddress Create formated e-mail addres RFC 5322
func (msg *impl) EmailAddress(address string, name string) (ret string) {
	return encode.New().Encoder(encode.EncoderQuotedPrintableScheme()).EmailAddress(address, name)
}

// GetFrom Get source address
func (msg *impl) GetFrom() (ret string) {
	var address *mail.Address
	var err error
	if address, err = mail.ParseAddress(msg.extractEMail(msg.from)); err == nil {
		ret = address.Address
	}
	return
}

// GetTo Get first valid destination addresses
func (msg *impl) GetTo() (ret string) {
	var i int
	var err error
	var address *mail.Address
	if len(msg.to) <= 0 {
		return
	}
	for i = range msg.to {
		if address, err = mail.ParseAddress(msg.extractEMail(msg.to[i])); err != nil {
			continue
		}
		ret = address.Address
		break
	}

	return
}

// Header Add custom header to message
func (msg *impl) Header(name string, values ...string) Interface {
	msg.Headers[name] = values
	return msg
}

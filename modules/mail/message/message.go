package message // import "gopkg.in/webnice/kit.v1/modules/mail/message"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"io"
	"time"

	"gopkg.in/webnice/kit.v1/modules/mail/encode"
	"gopkg.in/webnice/kit.v1/modules/mail/multipart"
)

// New new message
func New() Interface {
	var msg = new(impl)
	msg.Headers = make(map[string][]string)
	msg.encoder = encode.New()
	return msg
}

// WriteTo Write message to io.Writer
func (msg *impl) WriteTo(wr io.Writer) (int64, error) {
	var part *multipart.Part
	var mpt multipart.Interface
	var ok bool

	mpt = multipart.New(wr).SetStringWriter(msg.WriteString).SetHeaderWriter(msg.WriteHeader)

	// Версия MIME
	if _, ok = msg.Headers[HeaderMimeVersion.String()]; !ok {
		_ = msg.WriteString(wr, HeaderMimeVersion.String()+": 1.0\r\n")
	}

	// Дата и время сообщения
	if _, ok = msg.Headers[HeaderDate.String()]; !ok {
		msg.WriteHeader(wr, HeaderDate.String(), time.Now().In(time.Local).Format(time.RFC1123Z))
	}

	// From
	msg.WriteHeader(wr, HeaderFrom.String(), msg.from)

	// To
	msg.WriteHeader(wr, HeaderTo.String(), msg.to...)

	// Cc
	if len(msg.toCC) > 0 {
		msg.WriteHeader(wr, HeaderToCC.String(), msg.toCC...)
	}

	// Bcc
	if len(msg.toBCC) > 0 {
		msg.WriteHeader(wr, HeaderToBCC.String(), msg.toBCC...)
	}

	// Subject
	msg.WriteHeader(wr, HeaderSubject.String(), msg.subject)

	// Headers
	for headerName := range msg.Headers {
		msg.WriteHeader(wr, headerName, msg.Headers[headerName]...)
	}

	// Parts
	msg.calculateParts()

	// Mixed Part
	if msg.PartsMixed {
		mpt.CreateNewMultipart("mixed")
	}

	// Related Part
	if msg.PartsRelated {
		mpt.CreateNewMultipart("related")
	}

	// Alternative Part
	if msg.PartsAlternative {
		mpt.CreateNewMultipart("alternative")
	}

	// Write all body parts
	for _, part = range msg.Parts {
		if part.Type == multipart.PartBody {
			mpt.WritePart(part)
		}
	}

	// Mixed Part
	if msg.PartsMixed {
		mpt.CloseMultipart()
	}

	// Write all embeded parts
	for _, part = range msg.Parts {
		if part.Type == multipart.PartEmbed {
			mpt.WritePart(part)
		}
	}

	// Related Part
	if msg.PartsRelated {
		mpt.CloseMultipart()
	}

	// Write all attachment part
	for _, part = range msg.Parts {
		if part.Type == multipart.PartAttach {
			mpt.WritePart(part)
		}
	}

	// Alternative Part
	if msg.PartsAlternative {
		mpt.CloseMultipart()
	}

	msg.count += mpt.Count()
	if mpt.Error() != nil {
		msg.lastError = mpt.Error()
	}

	return msg.count, msg.lastError
}

// WriteString Запись строки
func (msg *impl) WriteString(wr io.Writer, str string) (count int) {
	count, msg.lastError = io.WriteString(wr, str)
	msg.count += int64(count)
	return
}

// WriteHeader Запись ключа в заголовок
func (msg *impl) WriteHeader(wr io.Writer, name string, sections ...string) {
	const lineLength = 76
	var i, lineLengthLeft int
	var section string
	var noSpace bool

	lineLengthLeft = lineLength - msg.WriteString(wr, name+":")
	for i, section = range sections {
		if i > 0 {
			lineLengthLeft -= msg.WriteString(wr, ", ")
			noSpace = true
		}

		if !noSpace {
			lineLengthLeft -= msg.WriteString(wr, " ")
			noSpace = false
		}

		if lineLengthLeft < 1 {
			msg.WriteString(wr, "\r\n\t")
			lineLengthLeft = lineLength
		}
		lineLengthLeft -= msg.WriteString(wr, section)
	}
	_ = msg.WriteString(wr, "\r\n")
}

// calculateParts Расстановка флагов для создания multipart
func (msg *impl) calculateParts() {
	var attach, embed, i int

	for i = range msg.Parts {
		switch msg.Parts[i].Type {
		case multipart.PartAttach:
			attach++
		case multipart.PartEmbed:
			embed++
		}
	}
	// Mixed
	if len(msg.Parts) > 0 && attach > 0 || attach > 1 {
		msg.PartsMixed = true
	}
	// Related
	if len(msg.Parts) > 0 && embed > 0 || embed > 1 {
		msg.PartsRelated = true
	}
	// Alternative
	if len(msg.Parts) > 1 {
		msg.PartsAlternative = true
	}
}

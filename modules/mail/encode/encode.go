package encode // import "gopkg.in/webnice/kit.v1/modules/mail/encode"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"bytes"
	"mime"
)

// New interface
func New() Interface {
	var enc = new(impl)
	var scheme = EncoderQuotedPrintableScheme()
	enc.encoder = scheme.encoder
	enc.encoding = scheme.encoding
	enc.charset = DefaultCharset
	return enc
}

// EncoderBase64Scheme Base64 encoding scheme
func EncoderBase64Scheme() Encoding {
	return Encoding{
		encoding: _encodeBase64,
		encoder:  Encoder{mime.BEncoding},
	}
}

// EncoderQuotedPrintableScheme Q-encoding scheme
func EncoderQuotedPrintableScheme() Encoding {
	return Encoding{
		encoding: _encodeQuotedPrintable,
		encoder:  Encoder{mime.QEncoding},
	}
}

// EncoderUnencodedScheme no using encoding
func EncoderUnencodedScheme() Encoding {
	return Encoding{
		encoding: _encodeUnencoded,
		encoder:  Encoder{},
	}
}

// Encoder set encoder
func (enc *impl) Encoder(scheme Encoding) Interface {
	enc.encoder = scheme.encoder
	enc.encoding = scheme.encoding
	return enc
}

// String Возвращает название установленного кодировщика
func (enc *impl) String() string {
	return enc.encoding
}

// Charset set charset
func (enc *impl) Charset(charset string) Interface {
	enc.charset = charset
	return enc
}

// GetCharset get charset
func (enc *impl) GetCharset() string {
	return enc.charset
}

// EncodeString Кодирование отдельно взятой строки
func (enc *impl) EncodeString(value string) (ret string) {
	ret = enc.encoder.Encode(enc.charset, value)
	return
}

// EmailAddress Создание адреса в стандарте RFC 5322
func (enc *impl) EmailAddress(address string, name string) (ret string) {
	var addressEncoded string
	var buf *bytes.Buffer
	var b byte
	if name == "" {
		return address
	}
	buf = bytes.NewBufferString(``)
	addressEncoded = enc.EncodeString(name)
	if addressEncoded == name {
		_ = buf.WriteByte('"')
		for i := 0; i < len(name); i++ {
			b = name[i]
			if b == '\\' || b == '"' {
				_ = buf.WriteByte('\\')
			}
			_ = buf.WriteByte(b)
		}
		_ = buf.WriteByte('"')
	} else if enc.isSpecialChars(name) {
		_, _ = buf.WriteString(EncoderBase64Scheme().encoder.Encode(enc.charset, name))
	} else {
		_, _ = buf.WriteString(addressEncoded)
	}
	_, _ = buf.WriteString(" <")
	_, _ = buf.WriteString(address)
	_ = buf.WriteByte('>')
	ret = buf.String()
	return
}

// isSpecialChars Проверка строки на наличие спец. символов
func (enc *impl) isSpecialChars(text string) (ret bool) {
	var i int
	var b byte
	for i = 0; i < len(text); i++ {
		switch b = text[i]; b {
		case '{', '}', '(', ')', '<', '>', '[', ']', ':', ';', '@', '\\', ',', '.', '"':
			ret = true
		}
	}
	return
}

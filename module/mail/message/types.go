package message

import (
	"bytes"
	"io"
	"sync"

	"github.com/webnice/kit/v2/module/mail/encode"
	"github.com/webnice/kit/v2/module/mail/multipart"
)

// Interface is an interface of message
type Interface interface {
	Encoder(encode.Interface) Interface            // Set encoder
	From(string) Interface                         // Set source address
	To(...string) Interface                        // Set destination addresses
	ToCC(...string) Interface                      // Set destination addresses carbon copy
	ToBCC(...string) Interface                     // Set destination addresses blind carbon copy
	Subject(string) Interface                      // Set subject
	BodyReader(io.Reader, ...string) Interface     // Add body reader. Default Content-Type is text/plain
	Body(*bytes.Buffer, ...string) Interface       // Add body data. Default Content-Type is text/plain
	Attach(io.Reader, string, ...string) Interface // Add attach file. Default Content-Type is text/plain
	Embed(io.Reader, string, ...string) Interface  // Add embeded file. Default Content-Type is text/plain
	Header(string, ...string) Interface            // Add custom header to message

	GetFrom() string                                 // Get source address
	GetTo() string                                   // Get first valid destination addresses
	EmailAddress(address string, name string) string // Create formated e-mail addres RFC 5322
	WriteTo(io.Writer) (int64, error)                // Write message to io.Writer
	Error() error                                    // Return last error
}

// impl is an implementation of message
type impl struct {
	count     int64
	lastError error

	encoder  encode.Interface // Кодировщик
	from     string           // Source address
	to       []string         // Destination addresses
	toCC     []string         // Destination addresses carbon copy
	toBCC    []string         // Destination addresses blind carbon copy
	uniqueTo unique           // Уникальные адреса назначения
	subject  string           // Тема сообщения

	Headers          map[string][]string // Все заголовки сообщения
	Parts            []*multipart.Part   // Куски сообщения
	PartsMixed       bool                // create mixed multipart
	PartsRelated     bool                // create related multipart
	PartsAlternative bool                // create alternative multipart
}

// unique Карта уникальных строк с защитой от отдновременного доступа
type unique struct {
	sync.Mutex
	String map[string]bool
}

// Header names
type Header string

// String return header name
func (hdr Header) String() string { return string(hdr) }

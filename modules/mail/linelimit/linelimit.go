package linelimit // import "github.com/webnice/kit/v1/modules/mail/linelimit"

import "io"

const (
	lineLimit = 76
)

// Interface auth interface
type Interface interface {
	Write([]byte) (int, error)
}

// impl auth implementation
type impl struct {
	writer io.Writer
	length int
}

// New interface
func New(wr io.Writer) Interface {
	var ll = new(impl)
	ll.writer = wr
	return ll
}

// Write func
func (ll *impl) Write(bytes []byte) (count int, err error) {
	for len(bytes)+ll.length > lineLimit {
		_, err = ll.writer.Write(bytes[:lineLimit-ll.length])
		_, err = ll.writer.Write([]byte("\r\n"))
		bytes = bytes[lineLimit-ll.length:]
		count += lineLimit - ll.length
		ll.length = 0
	}
	_, err = ll.writer.Write(bytes)
	ll.length += len(bytes)
	count += len(bytes)
	return
}

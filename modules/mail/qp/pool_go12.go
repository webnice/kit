// +build !go1.3

package qp // import "gopkg.in/webnice/kit.v1/modules/mail/qp"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import "bytes"

var ch = make(chan *bytes.Buffer, 32)

func getBuffer() *bytes.Buffer {
	select {
	case buf := <-ch:
		return buf
	default:
	}
	return new(bytes.Buffer)
}

func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	select {
	case ch <- buf:
	default:
	}
}

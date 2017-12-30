package file // import "gopkg.in/webnice/kit.v1/models/file"

//import "gopkg.in/webnice/log.v2"
//import "gopkg.in/webnice/debug.v1"
import (
	"io"
	"os"
)

// Copy Копирует один файл в другой
func (fl *impl) Copy(dst, src string) (size int64, err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func() {
		var lerr error
		lerr = out.Sync()
		if err == nil {
			err = lerr
		}
		lerr = out.Close()
		if err == nil {
			err = lerr
		}
	}()

	size, err = io.Copy(out, in)
	return
}

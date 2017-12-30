package file // import "gopkg.in/webnice/kit.v1/models/file"

//import "gopkg.in/webnice/log.v2"
//import "gopkg.in/webnice/debug.v1"
import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// LoadFile Загрузка файла в память и возврат в виде *bytes.Buffer
func (fl *impl) LoadFile(fileName string) (ret *bytes.Buffer, err error) {
	var fh *os.File
	fh, err = os.OpenFile(fileName, os.O_RDONLY, 0755)
	if err != nil {
		err = fmt.Errorf("Error open file '%s': %v", fileName, err)
		return
	}
	defer func() { _ = fh.Close() }()

	ret = bytes.NewBufferString(``)
	if _, err = io.Copy(ret, fh); err != nil {
		err = fmt.Errorf("Error read data from file '%s': %v", fileName, err)
		return
	}
	return
}

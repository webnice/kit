package file

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// LoadFile Загрузка файла в память и возврат в виде *bytes.Buffer
func (fl *impl) LoadFile(fileName string) (ret *bytes.Buffer, err error) {
	var fh *os.File

	if fh, err = os.OpenFile(fileName, os.O_RDONLY, 0755); err != nil {
		err = fmt.Errorf("open file %q, error: %s", fileName, err)
		return
	}
	defer func() { _ = fh.Close() }()
	ret = bytes.NewBufferString(``)
	if _, err = io.Copy(ret, fh); err != nil {
		err = fmt.Errorf("read data from file %q, error: %s", fileName, err)
		return
	}

	return
}

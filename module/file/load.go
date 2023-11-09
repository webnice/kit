package file

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// LoadFile Загрузка файла в память и возврат в виде *bytes.Buffer.
func (fl *impl) LoadFile(filename string) (data *bytes.Buffer, info os.FileInfo, err error) {
	const defaultFileMode = os.FileMode(0755)
	var fh *os.File

	if fh, err = os.OpenFile(filename, os.O_RDONLY, defaultFileMode); err != nil {
		err = fmt.Errorf("открытие файла %q прервано ошибкой: %w", filename, err)
		return
	}
	defer func() { _ = fh.Close() }()
	if info, err = fh.Stat(); err != nil {
		err = fmt.Errorf("получение атрибутов файла %q прервано ошибкой: %w", filename, err)
		return
	}
	data = new(bytes.Buffer)
	if _, err = io.Copy(data, fh); err != nil {
		err = fmt.Errorf("чтение данных из файла %q прервано ошибкой: %w", filename, err)
		return
	}

	return
}

package file

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// GetInfoSha512 Считывание информации о файле с контрольной суммой.
func (fl *impl) GetInfoSha512(filename string) (inf *InfoSha512, err error) {
	const defaultFileMode = 0755
	var (
		fh   *os.File
		s512 hash.Hash
		fi   os.FileInfo
	)

	inf = new(InfoSha512)
	if fh, err = os.OpenFile(filename, os.O_RDONLY, os.FileMode(defaultFileMode)); err != nil {
		err = fmt.Errorf("open file %q, error: %s", filename, err)
		return
	}
	defer func() { _ = fh.Close() }()
	s512 = sha512.New()
	if inf.Size, err = io.Copy(s512, fh); err != nil {
		err = fmt.Errorf("чтение файла %q прервано ошибкой: %s", filename, err)
		return
	}
	if fi, err = fh.Stat(); err != nil {
		err = fmt.Errorf("получение атрибутов файла %q прервано ошибкой: %s", filename, err)
		return
	}
	if inf.Size != fi.Size() {
		err = fmt.Errorf("считан файл размером %d байт, ожидался размер %d байт", fi.Size(), inf.Size)
		return
	}
	inf.Sha512, inf.Name = hex.EncodeToString(s512.Sum(nil)[:]), fl.GetFilename(filename)

	return
}

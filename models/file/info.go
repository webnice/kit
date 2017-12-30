package file // import "gopkg.in/webnice/kit.v1/models/file"

//import "gopkg.in/webnice/log.v2"
//import "gopkg.in/webnice/debug.v1"
import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// GetFileName Выделение из полного пути и имени файла, имя файла
func (fl *impl) GetFileName(fileName string) string {
	var ch []string
	ch = strings.Split(fileName, string(os.PathSeparator))
	if len(ch) > 0 {
		return ch[len(ch)-1]
	}
	return fileName
}

// GetInfoSha512 Считывание информации о файле с контрольной суммой
func (fl *impl) GetInfoSha512(fn string) (inf *InfoSha512, err error) {
	var fh *os.File
	var s512 hash.Hash
	var fi os.FileInfo

	inf = new(InfoSha512)
	fh, err = os.OpenFile(fn, os.O_RDONLY, 0755)
	if err != nil {
		err = fmt.Errorf("Error open file '%s': %v", fn, err)
		return
	}
	defer func() {
		_ = fh.Close()
	}()

	s512 = sha512.New()
	inf.Size, err = io.Copy(s512, fh)
	if err != nil {
		err = fmt.Errorf("Error read file '%s': %v", fn, err)
		return
	}
	fi, err = fh.Stat()
	if err != nil {
		err = fmt.Errorf("Failed to get information about a file '%s': %v", fn, err)
		return
	}
	if inf.Size != fi.Size() {
		err = fmt.Errorf("Error create SHA-512 summm, error reading file data")
		return
	}
	inf.Sha512 = hex.EncodeToString(s512.Sum(nil)[:])
	inf.Name = fl.GetFileName(fn)

	return
}

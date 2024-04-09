package file

import (
	"os"
	"strings"

	// Регистрация расширенных типов MIME для файлов.
	_ "github.com/webnice/dic"
)

// New creates new object and return Interface
func New() Interface {
	var obj = new(impl)
	return obj
}

// GetFilename Выделение из полного пути к файлу, имени файла.
func (fl *impl) GetFilename(filename string) (ret string) {
	var ch []string

	if ch = strings.Split(filename, string(os.PathSeparator)); len(ch) > 0 {
		ret = ch[len(ch)-1]
		return
	}
	ret = filename

	return
}

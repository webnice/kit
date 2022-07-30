// Package file
package file

//go:generate go run mime_generate.go
//go:generate gofmt -w mime.go

import (
	stdMime "mime"
	"os"
	"strings"
)

// Добавление всех mime types
func init() { mimeAddAll() }

// New creates new object and return Interface
func New() Interface {
	var obj = new(impl)
	return obj
}

// Добавление всех mime types
func mimeAddAll() {
	const prefix = `.`
	var mt string

	for mt = range mimeTypeExtension {
		_ = stdMime.AddExtensionType(prefix+mimeTypeExtension[mt], mt)
	}
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

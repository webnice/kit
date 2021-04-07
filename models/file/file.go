package file // import "github.com/webnice/kit/models/file"

//go:generate go run mime_generate.go
//go:generate gofmt -w mime.go

import (
	stdMime "mime"
	"os"
	"strings"
)

func init() {
	// Добавление всех mime types
	mimeAddAll()
}

// New creates new object and return Interface
func New() Interface {
	var obj = new(impl)
	return obj
}

// Добавление всех mime types
func mimeAddAll() {
	const preffix = `.`
	var mt string

	for mt = range mimeTypeExtension {
		_ = stdMime.AddExtensionType(preffix+mimeTypeExtension[mt], mt) // nolint: gosec
	}
}

// GetFilename Выделение из полного пути и имени файла, имени файла
func (fl *impl) GetFilename(filename string) (ret string) {
	var ch []string

	if ch = strings.Split(filename, string(os.PathSeparator)); len(ch) > 0 {
		ret = ch[len(ch)-1]
		return
	}
	ret = filename

	return
}

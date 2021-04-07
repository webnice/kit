package file

//go:generate go run mime_generate.go
//go:generate gofmt -w mime.go

import stdMime "mime"

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

package img

import "github.com/webnice/dic"

const (
	cFileNotExist = "открытие файла %q прервано ошибкой: %w"
)

type Error struct {
	dic.Errors
	// NotFound Открытие файла ... прервано ошибкой: ...
	FileNotExist dic.IError
}

var (
	errSingleton = &Error{
		Errors:       dic.Error(),
		FileNotExist: dic.NewError(cFileNotExist, "открываемый объект", "ошибка"),
	}

	// Errors Справочник ошибок.
	Errors = func() *Error { return errSingleton }
)

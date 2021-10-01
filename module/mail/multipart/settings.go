package multipart

import (
	"io"
)

// SetWriter Назначение io.Writer
func (mpt *impl) SetWriter(wr io.Writer) Interface {
	mpt.writer = wr
	return mpt
}

// SetHeaderWriter Назначение функции WriteHeader
func (mpt *impl) SetStringWriter(f func(io.Writer, string) int) Interface {
	mpt.writeString = f
	return mpt
}

// WriteString Запись строки
func (mpt *impl) WriteString(str string) (count int) {
	return mpt.writeString(mpt.writer, str)
}

// SetHeaderWriter Назначение функции WriteHeader
func (mpt *impl) SetHeaderWriter(f func(io.Writer, string, ...string)) Interface {
	mpt.writeHeader = f
	return mpt
}

// WriteHeader Запись ключа в заголовок
func (mpt *impl) WriteHeader(name string, sections ...string) {
	mpt.writeHeader(mpt.writer, name, sections...)
}

// Count Возвращает количество записанных байт во врайтер
func (mpt *impl) Count() int64 {
	return mpt.count
}

// Error Возвращает последнюю ошибку
func (mpt *impl) Error() error {
	return mpt.lastError
}

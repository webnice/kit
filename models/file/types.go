package file

import "bytes"

// Interface is an interface
type Interface interface {
	// CleanEmptyFolder Удаление пустых папок
	CleanEmptyFolder(string) error

	// Copy Копирует один файл в другой
	Copy(string, string) (int64, error)

	// GetInfoSha512 Считывание информации о файле с контрольной суммой
	GetInfoSha512(string) (*InfoSha512, error)

	// RecursiveFileList Поиск всех файлов начиная от path рекурсивно
	RecursiveFileList(string) ([]string, error)

	// GetFileName Выделение из полного пути и имени файла, имя файла
	GetFileName(string) string

	// LoadFile Загрузка файла в память и возврат в виде *bytes.Buffer
	LoadFile(string) (*bytes.Buffer, error)
}

// is an implementation
type impl struct {
}

// InfoSha512 Структура возвращаемой информации о файле
type InfoSha512 struct {
	Name   string // Название файла
	Size   int64  // Размер файла
	Sha512 string // Контрольная сумма файла
}

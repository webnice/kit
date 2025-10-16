package file

import (
	"bytes"
	"io"
	"os"
)

// Interface Интерфейс пакета.
type Interface interface {
	// CleanEmptyFolder Удаление пустых директорий.
	CleanEmptyFolder(folderPath string) (err error)

	// Copy Копирует один файл в другой.
	Copy(dst string, src string) (size int64, err error)

	// CopyWithSha512Sum Копирование контента с параллельным вычислением контрольной суммы алгоритмом SHA512.
	CopyWithSha512Sum(dst io.Writer, src io.Reader) (written int64, sha512sum string, err error)

	// GetInfoSha512 Считывание информации о файле с контрольной суммой.
	GetInfoSha512(filename string) (inf *InfoSha512, err error)

	// RecursiveFileList Поиск всех файлов начиная от folderPath рекурсивно.
	// Возвращается слайс относительных имён файлов.
	RecursiveFileList(folderPath string) (ret []string, err error)

	// GetFilename Выделение из полного пути к файлу, имени файла.
	GetFilename(filename string) (ret string)

	// LoadFile Загрузка файла в память и возврат в виде *bytes.Buffer.
	LoadFile(filename string) (data *bytes.Buffer, info os.FileInfo, err error)
}

// Объект сущности пакета.
type impl struct {
}

// InfoSha512 Структура возвращаемой информации о файле.
type InfoSha512 struct {
	Name   string // Название файла.
	Size   int64  // Размер файла.
	Sha512 string // Контрольная сумма файла.
}

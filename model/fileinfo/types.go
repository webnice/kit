/*

	Модель работы с os.Fileinfo проецируя его на память.
	Предназначена для работы с файлами полностью размещёнными в памяти.

*/

package fileinfo

import (
	"os"
	"time"
)

// Interface Интерфейс пакета.
type Interface interface {
	os.FileInfo

	// CopyFrom Копирование значений из интерфейса переданного объекта и присваивание собственному объекту.
	CopyFrom(src os.FileInfo) Interface

	// SetName Установка нового значения для Name().
	SetName(name string) Interface

	// SetSize Установка нового значения для Size().
	SetSize(size int64) Interface

	// SetMode Установка нового значения для Mode().
	SetMode(mode os.FileMode) Interface

	// SetModTime Установка нового значения для ModTime().
	SetModTime(modTime time.Time) Interface

	// SetIsDir Установка нового значения для IsDir().
	SetIsDir(isDir bool) Interface

	// SetSys Установка нового значения для Sys().
	SetSys(sys interface{}) Interface
}

// Объект сущности пакета.
type impl struct {
	name    string      // Базовое имя файла.
	size    int64       // Длина в байтах для обычных файлов, для других файлов зависит от os.
	mode    os.FileMode // Биты разрешения файла.
	modTime time.Time   // Дата и время модификации файла.
	isDir   bool        // Синоним для Mode().IsDir().
	sys     interface{} // Базовый источник данных (может возвращать значение nil).
}

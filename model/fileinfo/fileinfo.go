package fileinfo

import (
	"os"
	"path"
	"time"
)

// New Конструктор объекта сущности пакета, возвращается интерфейс пакета.
//
//goland:noinspection GoUnusedExportedFunction
func New() Interface {
	var fif = new(impl)
	return fif
}

// Name Базовое имя файла.
func (fif *impl) Name() string { return fif.name }

// Size Длина в байтах для обычных файлов, для других файлов зависит от os.
func (fif *impl) Size() int64 { return fif.size }

// Mode Биты разрешения файла.
func (fif *impl) Mode() os.FileMode { return fif.mode }

// ModTime Дата и время модификации файла.
func (fif *impl) ModTime() time.Time { return fif.modTime }

// IsDir Синоним для Mode().IsDir().
func (fif *impl) IsDir() bool { return fif.isDir }

// Sys Базовый источник данных (может возвращать значение nil).
func (fif *impl) Sys() interface{} { return fif.sys }

// CopyFrom Копирование значений из интерфейса переданного объекта и присваивание собственному объекту.
func (fif *impl) CopyFrom(src os.FileInfo) Interface {
	fif.SetName(src.Name()).
		SetSize(src.Size()).
		SetMode(src.Mode()).
		SetModTime(src.ModTime()).
		SetIsDir(src.IsDir()).
		SetSys(src.Sys())

	return fif
}

// SetName Установка нового значения для Name().
func (fif *impl) SetName(name string) Interface { fif.name = path.Base(name); return fif }

// SetSize Установка нового значения для Size().
func (fif *impl) SetSize(size int64) Interface { fif.size = size; return fif }

// SetMode Установка нового значения для Mode().
func (fif *impl) SetMode(mode os.FileMode) Interface { fif.mode = mode; return fif }

// SetModTime Установка нового значения для ModTime().
func (fif *impl) SetModTime(modTime time.Time) Interface { fif.modTime = modTime; return fif }

// SetIsDir Установка нового значения для IsDir().
func (fif *impl) SetIsDir(isDir bool) Interface { fif.isDir = isDir; return fif }

// SetSys Установка нового значения для Sys().
func (fif *impl) SetSys(sys interface{}) Interface { fif.sys = sys; return fif }

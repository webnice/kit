package img

import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"

	"git.webdesk.ru/wd/kit/v2/module/img/ico" // Расширение image для формата ico.
	"golang.org/x/image/bmp"                  // Расширение image для формата bmp.
	"golang.org/x/image/tiff"                 // Расширение image для формата tiff.
	"image/gif"                               // Расширение image для формата gif.
	"image/jpeg"                              // Расширение image для формата jpeg.
	"image/png"                               // Расширение image для формата png.
)

// Write Реализация интерфейса io.Writer.
func (i *imgItem) Write(p []byte) (n int, err error) { return i.wr.Write(p) }

// Close Реализация интерфейса io.Closer.
func (i *imgItem) Close() (err error) {
	_, err = i.ReadFrom(i.wr)
	i.wr.Reset()
	return
}

// Filename Имя файла из которого было загружено графическое изображение.
func (i *imgItem) Filename() string { return i.fileName }

// FileInfo Информация о файле из которого было загружено графическое изображение.
func (i *imgItem) FileInfo() os.FileInfo { return i.fileInfo }

// Image Графическое изображение в виде стандартного интерфейса image.Image.
func (i *imgItem) Image() image.Image { return i.image }

// Config Конфигурация стандартного графического изображения.
func (i *imgItem) Config() image.Config { return i.config }

// Type Формат исходного графического файла или бинарных данных из которых было загружено графическое изображение.
func (i *imgItem) Type() Type { return i.tpe }

// Reset Полная очистка графического объекта.
func (i *imgItem) Reset() {
	i.fileName = i.fileName[:0]
	i.fileInfo = nil
	i.tpe = TypeUnknown
	i.image = nil
	i.config = image.Config{}
	i.wr.Reset()
}

// SetConfig Установка конфигурации графического образа.
func (i *imgItem) SetConfig(cfg image.Config) Image { i.config = cfg; return i }

// SetFileInfo Установка информации о файле из которого было загружено графическое изображение.
func (i *imgItem) SetFileInfo(fi os.FileInfo) Image { i.fileInfo = fi; return i }

// SetFilename Установка имени файла.
func (i *imgItem) SetFilename(fn string) Image { i.fileName = fn; return i }

// SetOptionsTIFF Установка опций сохранения для формата TIFF.
func (i *imgItem) SetOptionsTIFF(opt *tiff.Options) Image { i.optTIFF = opt; return i }

// SetOptionsGIF Установка опций сохранения для формата GIF.
func (i *imgItem) SetOptionsGIF(opt *gif.Options) Image { i.optGIF = opt; return i }

// SetOptionsJPEG Установка опций сохранения для формата JPEG.
func (i *imgItem) SetOptionsJPEG(opt *jpeg.Options) Image { i.optJPEG = opt; return i }

// WriteTo Реализация интерфейса io.WriterTo.
func (i *imgItem) WriteTo(w io.Writer) (n int64, err error) {
	const (
		errIsNil = "в качестве объекта передан nil"
		errEmpty = "изображение пустое, нет данных для записи"
		errType  = "Не реализованный тип графического изображения %q"
	)
	if w == nil {
		err = errors.New(errIsNil)
		return
	}
	if i.image == nil {
		err = errors.New(errEmpty)
		return
	}
	switch i.tpe {
	case TypeICO:
		err = ico.Encode(w, i.image)
	case TypeBMP:
		err = bmp.Encode(w, i.image)
	case TypeTIFF:
		err = tiff.Encode(w, i.image, i.optTIFF)
	case TypeGIF:
		err = gif.Encode(w, i.image, i.optGIF)
	case TypeJPEG:
		err = jpeg.Encode(w, i.image, i.optJPEG)
	case TypePNG:
		err = png.Encode(w, i.image)
	default:
		err = fmt.Errorf(errType, i.tpe.String())
	}

	return
}

/*

    Модуль работы с графическими изображениями.

*/

package img

import (
	"bytes"
	"image/gif"
	"image/jpeg"
	"os"

	"golang.org/x/image/tiff"
)

// New Конструктор.
func New() Interface {
	var img = new(impl)
	return img
}

// Errors Справочник ошибок.
func (img *impl) Errors() *Error { return Errors() }

// New Создаёт пустой объект графического изображения.
// Объект обладает интерфейсом io.WriteCloser который можно использовать для загрузки данных графического
// объекта, после вызова Close(), записанные во Writer данные обрабатываются независимо от формата
// графического образа и присваиваются объекту.
func (img *impl) New() Image {
	var ret = &imgItem{
		wr:      &bytes.Buffer{},
		optTIFF: new(tiff.Options),
		optGIF:  new(gif.Options),
		optJPEG: new(jpeg.Options),
	}
	return ret
}

// Open Загрузка объекта графического изображения из файла.
func (img *impl) Open(filename string) (ret Image, err error) {
	var (
		fh *os.File
		i  *imgItem
	)

	fh, err = os.Open(filename) // nolint: gosec
	if os.IsNotExist(err) {
		err = img.Errors().NotFound()
		return
	} else if err != nil {
		return
	}
	i = img.New().(*imgItem)
	if _, err = i.ReadFrom(fh); err != nil {
		return
	}
	i.fileName = filename
	ret = i

	return
}

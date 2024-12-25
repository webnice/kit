package img

import (
	"bytes"
	"image"
	"io"
	"os"

	"golang.org/x/image/tiff"
	"image/gif"
	"image/jpeg"
)

// Interface Интерфейс пакета.
type Interface interface {
	// New Создаёт пустой объект графического изображения.
	// Объект обладает интерфейсом io.WriteCloser который можно использовать для загрузки данных графического
	// объекта, после вызова Close(), записанные во Writer данные обрабатываются независимо от формата
	// графического образа и присваиваются объекту.
	New() (ret Image)

	// Open Загрузка объекта графического изображения из файла.
	Open(filename string) (ret Image, err error)

	// Resize Изменение графических размеров изображения.
	Resize(im Image, w, h uint) Image

	// Errors Справочник ошибок.
	Errors() *Error
}

// Объект сущности пакета.
type impl struct {
	io.Writer
}

// Image Интерфейс графического объекта.
type Image interface {
	io.WriteCloser
	io.ReaderFrom
	io.WriterTo

	// Filename Имя файла из которого было загружено графическое изображение.
	Filename() string

	// FileInfo Информация о файле из которого было загружено графическое изображение.
	FileInfo() os.FileInfo

	// Image Графическое изображение в виде стандартного интерфейса image.Image.
	Image() image.Image

	// Config Конфигурация стандартного графического изображения.
	Config() image.Config

	// Type Формат исходного графического файла или бинарных данных из которых было загружено графическое изображение.
	Type() Type

	// Reset Полная очистка графического объекта.
	Reset()

	// SetImage Загрузка объекта графического изображения из стандартного интерфейса image.Image.
	SetImage(im image.Image) Image

	// SetConfig Установка конфигурации графического образа.
	SetConfig(cfg image.Config) Image

	// SetType Установка формата графического изображения.
	SetType(t Type) Image

	// SetFileInfo Установка информации о файле из которого было загружено графическое изображение.
	SetFileInfo(fi os.FileInfo) Image

	// SetFilename Установка имени файла.
	SetFilename(fn string) Image

	// SetOptionsTIFF Установка опций сохранения для формата TIFF.
	SetOptionsTIFF(opt *tiff.Options) Image

	// SetOptionsGIF Установка опций сохранения для формата GIF.
	SetOptionsGIF(opt *gif.Options) Image

	// SetOptionsJPEG Установка опций сохранения для формата JPEG.
	SetOptionsJPEG(opt *jpeg.Options) Image
}

// imgItem Объект графического изображения.
type imgItem struct {
	fileName string        // Имя файла картинки.
	fileInfo os.FileInfo   // Информация об оригинальном файле.
	tpe      Type          // Тип картинки.
	image    image.Image   // Загруженная и декодированная картинка.
	config   image.Config  // Информация о картинке.
	wr       *bytes.Buffer // Данные записанные через Write(), парсятся при вызове Close().
	optTIFF  *tiff.Options // Опций сохранения для формата TIFF.
	optGIF   *gif.Options  // Опций сохранения для формата GIF.
	optJPEG  *jpeg.Options // Опций сохранения для формата JPEG.
}

/*

	// ConvertGrayscale Convert image to grayscale image
	ConvertGrayscale(*Image) *Image

	// ConvertBlackAndWhite Convert image to grayscale image
	ConvertBlackAndWhite(*Image) *Image

*/

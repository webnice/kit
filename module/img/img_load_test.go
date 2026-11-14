package img

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/webnice/kit/v4/module/img/ico"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

// panicReader Объект io.Reader вызывающий панику.
type panicReader struct{}

// Read Реализация интерфейса io.Reader.
func (panicReader) Read([]byte) (int, error) { panic("string panic") }

// TestImageReadFromFormats Загрузка изображений разных форматов.
func TestImageReadFromFormats(t *testing.T) {
	const (
		tplReadFromWrong = "ReadFrom(%s) = %d, %v."
		tplStateWrong    = "ReadFrom(%s) state = type:%s config:%v image:%v."
	)
	// Описание одного теста загрузки формата.
	type formatCase struct {
		name string
		typ  Type
		data []byte
	}
	var (
		formats []formatCase
		format  formatCase
		im      Image
		n       int64
		err     error
		i       int
	)

	formats = []formatCase{
		{"png", TypePNG, encodeBytes(t, func(w io.Writer) error { return png.Encode(w, testImage(2, 2)) })},
		{"jpeg", TypeJPEG, encodeBytes(t, func(w io.Writer) error { return jpeg.Encode(w, testImage(2, 2), nil) })},
		{"gif", TypeGIF, encodeBytes(t, func(w io.Writer) error { return gif.Encode(w, testImage(2, 2), nil) })},
		{"bmp", TypeBMP, encodeBytes(t, func(w io.Writer) error { return bmp.Encode(w, testImage(2, 2)) })},
		{"tiff", TypeTIFF, encodeBytes(t, func(w io.Writer) error { return tiff.Encode(w, testImage(2, 2), nil) })},
		{"ico", TypeICO, encodeBytes(t, func(w io.Writer) error { return ico.Encode(w, testImage(2, 2)) })},
	}
	for i = range formats {
		format = formats[i]
		t.Run(format.name, func(t *testing.T) {
			im = New().New()
			n, err = im.ReadFrom(bytes.NewBuffer(format.data))
			if err != nil || n != int64(len(format.data)) {
				t.Fatalf(tplReadFromWrong, format.name, n, err)
			}
			if im.Type() != format.typ || im.Config().Width != 2 || im.Config().Height != 2 || im.Image() == nil {
				t.Fatalf(tplStateWrong, format.name, im.Type(), im.Config(), im.Image() != nil)
			}
		})
	}
}

// TestImageReadFromErrors Ошибки загрузки изображения.
func TestImageReadFromErrors(t *testing.T) {
	const (
		errInvalidNoError = "ReadFrom(invalid) не вернул ошибку."
		tplPanicWrong     = "ReadFrom(panic) = %v."
	)
	var (
		im  Image
		err error
	)

	im = New().New()
	if _, err = im.ReadFrom(bytes.NewBuffer([]byte("not an image"))); err == nil {
		t.Fatal(errInvalidNoError)
	}
	if _, err = im.ReadFrom(panicReader{}); err == nil || !strings.Contains(err.Error(), "panic recovery") {
		t.Fatalf(tplPanicWrong, err)
	}
}

// TestImageConfigUsesBoundsSize Конфигурация изображения по границам смещённого изображения.
func TestImageConfigUsesBoundsSize(t *testing.T) {
	const (
		tplFatalReadFrom  = "ReadFrom(shifted) вернул ошибку: %v."
		tplConfigMismatch = "Config для shifted bounds = %v."
	)
	var (
		shifted *image.RGBA
		data    []byte
		im      Image
		x, y    int
		err     error
	)

	shifted = image.NewRGBA(image.Rect(10, 10, 12, 14))
	for y = 10; y < 14; y++ {
		for x = 10; x < 12; x++ {
			shifted.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	data = encodeBytes(t, func(w io.Writer) error { return png.Encode(w, shifted) })
	im = New().New()
	if _, err = im.ReadFrom(bytes.NewBuffer(data)); err != nil {
		t.Fatalf(tplFatalReadFrom, err)
	}
	if im.Config().Width != 2 || im.Config().Height != 4 {
		t.Fatalf(tplConfigMismatch, im.Config())
	}
}

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

func (panicReader) Read([]byte) (int, error) { panic("string panic") }

// TestImageReadFromFormats Загрузка изображений разных форматов.
func TestImageReadFromFormats(t *testing.T) {
	var formats = []struct {
		name string
		typ  Type
		data []byte
	}{
		{"png", TypePNG, encodeBytes(t, func(w io.Writer) error { return png.Encode(w, testImage(2, 2)) })},
		{"jpeg", TypeJPEG, encodeBytes(t, func(w io.Writer) error { return jpeg.Encode(w, testImage(2, 2), nil) })},
		{"gif", TypeGIF, encodeBytes(t, func(w io.Writer) error { return gif.Encode(w, testImage(2, 2), nil) })},
		{"bmp", TypeBMP, encodeBytes(t, func(w io.Writer) error { return bmp.Encode(w, testImage(2, 2)) })},
		{"tiff", TypeTIFF, encodeBytes(t, func(w io.Writer) error { return tiff.Encode(w, testImage(2, 2), nil) })},
		{"ico", TypeICO, encodeBytes(t, func(w io.Writer) error { return ico.Encode(w, testImage(2, 2)) })},
	}
	for _, format := range formats {
		t.Run(format.name, func(t *testing.T) {
			im := New().New()
			n, err := im.ReadFrom(bytes.NewBuffer(format.data))
			if err != nil || n != int64(len(format.data)) {
				t.Fatalf("ReadFrom(%s) = %d, %v", format.name, n, err)
			}
			if im.Type() != format.typ || im.Config().Width != 2 || im.Config().Height != 2 || im.Image() == nil {
				t.Fatalf("ReadFrom(%s) state = type:%s config:%v image:%v", format.name, im.Type(), im.Config(), im.Image() != nil)
			}
		})
	}
}

// TestImageReadFromErrors Ошибки загрузки изображения.
func TestImageReadFromErrors(t *testing.T) {
	im := New().New()
	if _, err := im.ReadFrom(bytes.NewBuffer([]byte("not an image"))); err == nil {
		t.Fatal("ReadFrom(invalid) не вернул ошибку")
	}
	if _, err := im.ReadFrom(panicReader{}); err == nil || !strings.Contains(err.Error(), "panic recovery") {
		t.Fatalf("ReadFrom(panic) = %v", err)
	}
}

// TestImageConfigUsesBoundsSize Конфигурация изображения по границам смещённого изображения.
func TestImageConfigUsesBoundsSize(t *testing.T) {
	var shifted = image.NewRGBA(image.Rect(10, 10, 12, 14))
	for y := 10; y < 14; y++ {
		for x := 10; x < 12; x++ {
			shifted.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	data := encodeBytes(t, func(w io.Writer) error { return png.Encode(w, shifted) })
	im := New().New()
	if _, err := im.ReadFrom(bytes.NewBuffer(data)); err != nil {
		t.Fatalf("ReadFrom(shifted) вернул ошибку: %v", err)
	}
	if im.Config().Width != 2 || im.Config().Height != 4 {
		t.Fatalf("Config для shifted bounds = %v", im.Config())
	}
}

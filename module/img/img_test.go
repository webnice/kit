package img

import (
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestImplNewAndSetters Конструктор сервиса изображений.
func TestImplNewAndSetters(t *testing.T) {
	const (
		errNewErrorsWrong = "New()/Errors() некорректны."
		tplStateWrong     = "New() image state некорректен: %#v."
		errOptionsNil     = "New() не инициализировал опции кодирования."
	)
	var (
		service Interface
		im      *imgItem
	)

	service = New()
	if service == nil || service.Errors() != Errors() {
		t.Fatal(errNewErrorsWrong)
	}
	im = service.New().(*imgItem)
	if im.Filename() != "" || im.FileInfo() != nil || im.Image() != nil || im.Type() != TypeUnknown || im.Config().Width != 0 {
		t.Fatalf(tplStateWrong, im)
	}
	if im.optTIFF == nil || im.optGIF == nil || im.optJPEG == nil {
		t.Fatal(errOptionsNil)
	}
}

// TestImageOpen Загрузка изображения из файла.
func TestImageOpen(t *testing.T) {
	const (
		tplFatalOpenMissing = "Open(missing) = %v."
		tplFatalWriteFile   = "WriteFile(): неожиданная ошибка: %v."
		tplFatalOpen        = "Open() вернул ошибку: %v."
		tplOpenState        = "Open() state = %#v."
	)
	var (
		fn  string
		im  Image
		err error
	)

	if _, err = New().Open(filepath.Join(t.TempDir(), "missing.png")); err == nil ||
		!strings.Contains(err.Error(), "открытие файла") {
		t.Fatalf(tplFatalOpenMissing, err)
	}
	fn = filepath.Join(t.TempDir(), "image.png")
	if err = os.WriteFile(fn, encodeBytes(t, func(w io.Writer) error { return png.Encode(w, testImage(2, 2)) }), 0644); err != nil {
		t.Fatalf(tplFatalWriteFile, err)
	}
	im, err = New().Open(fn)
	if err != nil {
		t.Fatalf(tplFatalOpen, err)
	}
	if im.Filename() != fn || im.FileInfo() == nil || im.FileInfo().Name() != "image.png" || im.Type() != TypePNG || im.Image() == nil {
		t.Fatalf(tplOpenState, im)
	}
}

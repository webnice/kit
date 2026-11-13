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
	var service = New()
	if service == nil || service.Errors() != Errors() {
		t.Fatal("New()/Errors() некорректны")
	}

	im := service.New().(*imgItem)
	if im.Filename() != "" || im.FileInfo() != nil || im.Image() != nil || im.Type() != TypeUnknown || im.Config().Width != 0 {
		t.Fatalf("New() image state некорректен: %#v", im)
	}
	if im.optTIFF == nil || im.optGIF == nil || im.optJPEG == nil {
		t.Fatal("New() не инициализировал опции кодирования")
	}
}

// TestImageOpen Загрузка изображения из файла.
func TestImageOpen(t *testing.T) {
	if _, err := New().Open(filepath.Join(t.TempDir(), "missing.png")); err == nil ||
		!strings.Contains(err.Error(), "открытие файла") {
		t.Fatalf("Open(missing) = %v", err)
	}

	fn := filepath.Join(t.TempDir(), "image.png")
	if err := os.WriteFile(fn, encodeBytes(t, func(w io.Writer) error { return png.Encode(w, testImage(2, 2)) }), 0644); err != nil {
		t.Fatal(err)
	}
	im, err := New().Open(fn)
	if err != nil {
		t.Fatalf("Open() вернул ошибку: %v", err)
	}
	if im.Filename() != fn || im.FileInfo() == nil || im.FileInfo().Name() != "image.png" || im.Type() != TypePNG || im.Image() == nil {
		t.Fatalf("Open() state = %#v", im)
	}
}

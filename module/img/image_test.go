package img

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// testImage Создание тестового изображения.
func testImage(width, height int) *image.RGBA {
	var im = image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			im.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	return im
}

// encodeBytes Кодирование изображения в срез байт.
func encodeBytes(t *testing.T, fn func(io.Writer) error) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := fn(&buf); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

// TestImageSettersAndReset Сеттеры и сброс состояния изображения.
func TestImageSettersAndReset(t *testing.T) {
	im := New().New().(*imgItem)

	var (
		fi  os.FileInfo
		err error
	)
	tmp := filepath.Join(t.TempDir(), "file.png")
	if err := os.WriteFile(tmp, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if fi, err = os.Stat(tmp); err != nil {
		t.Fatal(err)
	}
	cfg := image.Config{Width: 2, Height: 3, ColorModel: color.RGBAModel}
	ret := im.SetFilename("name.png").SetFileInfo(fi).SetImage(testImage(2, 3)).SetConfig(cfg).SetType(TypePNG)
	if ret != im || im.Filename() != "name.png" || im.FileInfo() != fi || im.Image() == nil || im.Config() != cfg || im.Type() != TypePNG {
		t.Fatalf("setters некорректны: %#v", im)
	}
	im.SetOptionsTIFF(nil).SetOptionsGIF(nil).SetOptionsJPEG(nil)
	if im.optTIFF != nil || im.optGIF != nil || im.optJPEG != nil {
		t.Fatal("SetOptions* некорректны")
	}
	im.Reset()
	if im.Filename() != "" || im.FileInfo() != nil || im.Image() != nil || im.Type() != TypeUnknown || im.Config().Width != 0 || im.wr.Len() != 0 {
		t.Fatalf("Reset() некорректен: %#v", im)
	}
}

// TestImageWriteCloseAndWriteTo Запись, закрытие и экспорт изображения.
func TestImageWriteCloseAndWriteTo(t *testing.T) {
	pngData := encodeBytes(t, func(w io.Writer) error { return png.Encode(w, testImage(3, 2)) })
	im := New().New()
	if _, err := im.Write(pngData); err != nil {
		t.Fatalf("Write() вернул ошибку: %v", err)
	}
	if err := im.Close(); err != nil {
		t.Fatalf("Close() вернул ошибку: %v", err)
	}
	if im.Type() != TypePNG || im.Config().Width != 3 || im.Config().Height != 2 || im.Image() == nil {
		t.Fatalf("Close() state = type:%s config:%v image:%v", im.Type(), im.Config(), im.Image() != nil)
	}
	if im.(*imgItem).wr.Len() != 0 {
		t.Fatal("Close() не очистил буфер")
	}

	var writers = []struct {
		name string
		typ  Type
	}{
		{"ico", TypeICO}, {"bmp", TypeBMP}, {"tiff", TypeTIFF}, {"gif", TypeGIF}, {"jpeg", TypeJPEG}, {"png", TypePNG},
	}
	for _, writer := range writers {
		t.Run("WriteTo_"+writer.name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := New().New().SetImage(testImage(2, 2)).SetType(writer.typ).WriteTo(&buf)
			if err != nil || n == 0 || int(n) != buf.Len() {
				t.Fatalf("WriteTo(%s) = %d, %v, len=%d", writer.name, n, err, buf.Len())
			}
		})
	}

	if _, err := New().New().SetImage(testImage(1, 1)).SetType(TypePNG).WriteTo(nil); err == nil {
		t.Fatal("WriteTo(nil) не вернул ошибку")
	}
	if _, err := New().New().SetType(TypePNG).WriteTo(&bytes.Buffer{}); err == nil {
		t.Fatal("WriteTo(empty image) не вернул ошибку")
	}
	if _, err := New().New().SetImage(testImage(1, 1)).SetType(Type("unknown")).WriteTo(&bytes.Buffer{}); err == nil {
		t.Fatal("WriteTo(unknown type) не вернул ошибку")
	}
}

package img

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/webnice/kit/v4/module/img/ico"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

func testImage(width, height int) *image.RGBA {
	var im = image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			im.Set(x, y, color.RGBA{R: uint8(100 + x), G: uint8(50 + y), B: 200, A: 255})
		}
	}
	return im
}

func encodeBytes(t *testing.T, fn func(io.Writer) error) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := fn(&buf); err != nil {
		t.Fatalf("encode error: %v", err)
	}
	return buf.Bytes()
}

type panicReader struct{}

func (panicReader) Read([]byte) (int, error) { panic("string panic") }

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
	if Type("abc").String() != "abc" {
		t.Fatal("Type.String() некорректен")
	}
}

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

func TestImageResize(t *testing.T) {
	if got := New().Resize(nil, 1, 1); got != nil {
		t.Fatalf("Resize(nil) = %#v", got)
	}
	empty := New().New()
	if got := New().Resize(empty, 1, 1); got != empty || got.Image() != nil {
		t.Fatalf("Resize(empty image) = %#v", got)
	}
	var (
		fi  os.FileInfo
		err error
	)
	tmp := filepath.Join(t.TempDir(), "x.png")
	if err = os.WriteFile(tmp, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if fi, err = os.Stat(tmp); err != nil {
		t.Fatal(err)
	}
	src := New().New().SetImage(testImage(5, 2)).SetType(TypePNG).SetFilename("src.png").SetFileInfo(fi)
	resized := New().Resize(src, 4, 2)
	if resized.Type() != TypePNG || resized.Filename() != "src.png" || resized.FileInfo() != fi ||
		resized.Config().Width != 4 || resized.Config().Height != 2 || resized.Image() == nil {
		t.Fatalf("Resize() state = %#v", resized)
	}
	square := New().Resize(src, 3, 3)
	if square.Config().Width != 3 || square.Config().Height != 3 {
		t.Fatalf("Resize(square) config = %v", square.Config())
	}
}

func TestImageReadFromErrors(t *testing.T) {
	im := New().New()
	if _, err := im.ReadFrom(bytes.NewBuffer([]byte("not an image"))); err == nil {
		t.Fatal("ReadFrom(invalid) не вернул ошибку")
	}
	if _, err := im.ReadFrom(panicReader{}); err == nil || !strings.Contains(err.Error(), "panic recovery") {
		t.Fatalf("ReadFrom(panic) = %v", err)
	}
}

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

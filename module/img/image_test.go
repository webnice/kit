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

// Создание тестового изображения.
func testImage(width, height int) (im *image.RGBA) {
	var x, y int

	im = image.NewRGBA(image.Rect(0, 0, width, height))
	for y = 0; y < height; y++ {
		for x = 0; x < width; x++ {
			im.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	return
}

// Кодирование изображения в срез байт.
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
	const (
		tplFatalWrite   = "WriteFile(): неожиданная ошибка: %v."
		tplFatalStat    = "Stat(): неожиданная ошибка: %v."
		tplSettersWrong = "setters некорректны: %#v."
		errOptionsWrong = "SetOptions* некорректны."
		tplResetWrong   = "Reset() некорректен: %#v."
	)
	var (
		im  *imgItem
		fi  os.FileInfo
		tmp string
		cfg image.Config
		ret Image
		err error
	)

	im = New().New().(*imgItem)
	tmp = filepath.Join(t.TempDir(), "file.png")
	if err := os.WriteFile(tmp, []byte("x"), 0644); err != nil {
		t.Fatalf(tplFatalWrite, err)
	}
	if fi, err = os.Stat(tmp); err != nil {
		t.Fatalf(tplFatalStat, err)
	}
	cfg = image.Config{Width: 2, Height: 3, ColorModel: color.RGBAModel}
	ret = im.SetFilename("name.png").SetFileInfo(fi).SetImage(testImage(2, 3)).SetConfig(cfg).SetType(TypePNG)
	if ret != im || im.Filename() != "name.png" || im.FileInfo() != fi || im.Image() == nil || im.Config() != cfg || im.Type() != TypePNG {
		t.Fatalf(tplSettersWrong, im)
	}
	im.SetOptionsTIFF(nil).SetOptionsGIF(nil).SetOptionsJPEG(nil)
	if im.optTIFF != nil || im.optGIF != nil || im.optJPEG != nil {
		t.Fatal(errOptionsWrong)
	}
	im.Reset()
	if im.Filename() != "" || im.FileInfo() != nil || im.Image() != nil || im.Type() != TypeUnknown || im.Config().Width != 0 || im.wr.Len() != 0 {
		t.Fatalf(tplResetWrong, im)
	}
}

// TestImageWriteCloseAndWriteTo Запись, закрытие и экспорт изображения.
func TestImageWriteCloseAndWriteTo(t *testing.T) {
	const (
		tplFatalWrite        = "Write() вернул ошибку: %v."
		tplFatalClose        = "Close() вернул ошибку: %v."
		tplCloseState        = "Close() state = type:%s config:%v image:%v."
		errBufferNotCleared  = "Close() не очистил буфер."
		tplWriteToWrong      = "WriteTo(%s) = %d, %v, len=%d."
		errWriteToNilNoError = "WriteTo(nil) не вернул ошибку."
		errWriteToEmptyNoErr = "WriteTo(empty image) не вернул ошибку."
		errWriteToUnknown    = "WriteTo(unknown type) не вернул ошибку."
	)
	// Описание одного теста записи изображения.
	type writerCase struct {
		name string
		typ  Type
	}
	var (
		pngData []byte
		im      Image
		writers []writerCase
		writer  writerCase
		buf     bytes.Buffer
		i       int
		n       int64
		err     error
	)

	pngData = encodeBytes(t, func(w io.Writer) error { return png.Encode(w, testImage(3, 2)) })
	im = New().New()
	if _, err = im.Write(pngData); err != nil {
		t.Fatalf(tplFatalWrite, err)
	}
	if err = im.Close(); err != nil {
		t.Fatalf(tplFatalClose, err)
	}
	if im.Type() != TypePNG || im.Config().Width != 3 || im.Config().Height != 2 || im.Image() == nil {
		t.Fatalf(tplCloseState, im.Type(), im.Config(), im.Image() != nil)
	}
	if im.(*imgItem).wr.Len() != 0 {
		t.Fatal(errBufferNotCleared)
	}
	writers = []writerCase{
		{"ico", TypeICO}, {"bmp", TypeBMP}, {"tiff", TypeTIFF}, {"gif", TypeGIF}, {"jpeg", TypeJPEG}, {"png", TypePNG},
	}
	for i = range writers {
		writer = writers[i]
		t.Run("WriteTo_"+writer.name, func(t *testing.T) {
			buf = bytes.Buffer{}
			if n, err = New().New().SetImage(testImage(2, 2)).SetType(writer.typ).WriteTo(&buf); err != nil || n == 0 || int(n) != buf.Len() {
				t.Fatalf(tplWriteToWrong, writer.name, n, err, buf.Len())
			}
		})
	}
	if _, err = New().New().SetImage(testImage(1, 1)).SetType(TypePNG).WriteTo(nil); err == nil {
		t.Fatal(errWriteToNilNoError)
	}
	if _, err = New().New().SetType(TypePNG).WriteTo(&bytes.Buffer{}); err == nil {
		t.Fatal(errWriteToEmptyNoErr)
	}
	if _, err = New().New().SetImage(testImage(1, 1)).SetType(Type("unknown")).WriteTo(&bytes.Buffer{}); err == nil {
		t.Fatal(errWriteToUnknown)
	}
}

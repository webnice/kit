package ico

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"testing"
)

// testImage Создание тестового изображения.
func testImage(size int) *image.RGBA {
	var im = image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			im.Set(x, y, color.RGBA{R: uint8(100 + x), G: uint8(50 + y), B: 200, A: 255})
		}
	}
	return im
}

// failWriter Объект io.Writer всегда возвращающий ошибку.
type failWriter struct {
	err error
}

func (fw *failWriter) Write([]byte) (int, error) { return 0, fw.err }

// failAfterWriter Объект io.Writer возвращающий ошибку после первого вызова.
type failAfterWriter struct {
	calls int
	err   error
}

func (faw *failAfterWriter) Write([]byte) (int, error) {
	faw.calls++
	if faw.calls > 1 {
		return 0, faw.err
	}
	return 0, nil
}

// TestEncodeErrors Ошибки кодирования изображения в формат ICO.
func TestEncodeErrors(t *testing.T) {
	if err := Encode(&bytes.Buffer{}, image.NewRGBA(image.Rect(0, 0, 0, 0))); err == nil {
		t.Fatal("Encode(0x0) не вернул ошибку")
	}
	if err := Encode(&bytes.Buffer{}, testImage(257)); err == nil {
		t.Fatal("Encode(257x257) не вернул ошибку")
	}
	expected := errors.New("write error")
	if err := Encode(&failWriter{err: expected}, testImage(1)); !errors.Is(err, expected) {
		t.Fatalf("Encode(header write error) = %v", err)
	}
	if err := Encode(&failAfterWriter{err: expected}, testImage(1)); !errors.Is(err, expected) {
		t.Fatalf("Encode(body write error) = %v", err)
	}
}

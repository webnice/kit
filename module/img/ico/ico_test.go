package ico

import (
	"bytes"
	"encoding/binary"
	"image"
	"strings"
	"testing"
)

// TestEncodeDecodeRoundTrip Кодирование и декодирование изображения ICO.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	var encoded bytes.Buffer
	if err := Encode(&encoded, testImage(2)); err != nil {
		t.Fatalf("Encode() вернул ошибку: %v", err)
	}
	decoded, err := Decode(bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatalf("Decode() вернул ошибку: %v", err)
	}
	if decoded.Bounds().Dx() != 2 || decoded.Bounds().Dy() != 2 {
		t.Fatalf("Decode() bounds = %v", decoded.Bounds())
	}
	cfg, err := DecodeConfig(bytes.NewReader(encoded.Bytes()))
	if err != nil || cfg.Width != 2 || cfg.Height != 2 {
		t.Fatalf("DecodeConfig() = %v, %v", cfg, err)
	}
	im, format, err := image.Decode(bytes.NewReader(encoded.Bytes()))
	if err != nil || format != "ico" || im.Bounds().Dx() != 2 {
		t.Fatalf("image.Decode() = %v, %q, %v", im.Bounds(), format, err)
	}
}

// TestDecodeHeaderErrors Ошибки заголовка ICO.
func TestDecodeHeaderErrors(t *testing.T) {
	if _, err := Decode(bytes.NewReader(nil)); err == nil {
		t.Fatal("Decode(empty) не вернул ошибку")
	}
	if _, err := DecodeConfig(bytes.NewReader(nil)); err == nil {
		t.Fatal("DecodeConfig(empty) не вернул ошибку")
	}

	var badFirst bytes.Buffer
	_ = binary.Write(&badFirst, binary.LittleEndian, uint16(1))
	_ = binary.Write(&badFirst, binary.LittleEndian, uint16(1))
	_ = binary.Write(&badFirst, binary.LittleEndian, uint16(1))
	if _, err := Decode(bytes.NewReader(badFirst.Bytes())); err == nil || !strings.Contains(err.Error(), "first byte") {
		t.Fatalf("Decode(bad first) = %v", err)
	}

	var badSecond bytes.Buffer
	_ = binary.Write(&badSecond, binary.LittleEndian, uint16(0))
	_ = binary.Write(&badSecond, binary.LittleEndian, uint16(2))
	_ = binary.Write(&badSecond, binary.LittleEndian, uint16(1))
	if _, err := Decode(bytes.NewReader(badSecond.Bytes())); err == nil || !strings.Contains(err.Error(), "second byte") {
		t.Fatalf("Decode(bad second) = %v", err)
	}

	var noImages bytes.Buffer
	_ = binary.Write(&noImages, binary.LittleEndian, uint16(0))
	_ = binary.Write(&noImages, binary.LittleEndian, uint16(1))
	_ = binary.Write(&noImages, binary.LittleEndian, uint16(0))
	if _, err := Decode(bytes.NewReader(noImages.Bytes())); err == nil || !strings.Contains(err.Error(), "no images") {
		t.Fatalf("Decode(no images) = %v", err)
	}

	var truncatedDir bytes.Buffer
	_ = binary.Write(&truncatedDir, binary.LittleEndian, uint16(0))
	_ = binary.Write(&truncatedDir, binary.LittleEndian, uint16(1))
	_ = binary.Write(&truncatedDir, binary.LittleEndian, uint16(1))
	if _, err := Decode(bytes.NewReader(truncatedDir.Bytes())); err == nil {
		t.Fatal("Decode(truncated dir) не вернул ошибку")
	}
}

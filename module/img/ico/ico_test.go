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
	const (
		tplFatalEncode    = "Encode() вернул ошибку: %v."
		tplFatalDecode    = "Decode() вернул ошибку: %v."
		tplBoundsMismatch = "Decode() bounds = %v."
		tplConfigMismatch = "DecodeConfig() = %v, %v."
		tplImageDecode    = "image.Decode() = %v, %q, %v."
	)
	var (
		encoded bytes.Buffer
		decoded image.Image
		cfg     image.Config
		im      image.Image
		format  string
		err     error
	)

	if err = Encode(&encoded, testImage(2)); err != nil {
		t.Fatalf(tplFatalEncode, err)
	}
	decoded, err = Decode(bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatalf(tplFatalDecode, err)
	}
	if decoded.Bounds().Dx() != 2 || decoded.Bounds().Dy() != 2 {
		t.Fatalf(tplBoundsMismatch, decoded.Bounds())
	}
	cfg, err = DecodeConfig(bytes.NewReader(encoded.Bytes()))
	if err != nil || cfg.Width != 2 || cfg.Height != 2 {
		t.Fatalf(tplConfigMismatch, cfg, err)
	}
	im, format, err = image.Decode(bytes.NewReader(encoded.Bytes()))
	if err != nil || format != "ico" || im.Bounds().Dx() != 2 {
		t.Fatalf(tplImageDecode, im.Bounds(), format, err)
	}
}

// TestDecodeHeaderErrors Ошибки заголовка ICO.
func TestDecodeHeaderErrors(t *testing.T) {
	const (
		errEmptyDecode       = "Decode(empty) не вернул ошибку."
		errEmptyConfig       = "DecodeConfig(empty) не вернул ошибку."
		tplFatalBadFirst     = "Decode(bad first) = %v."
		tplFatalBadSecond    = "Decode(bad second) = %v."
		tplFatalNoImages     = "Decode(no images) = %v."
		errTruncatedDirNoErr = "Decode(truncated dir) не вернул ошибку."
	)
	var (
		badFirst     bytes.Buffer
		badSecond    bytes.Buffer
		noImages     bytes.Buffer
		truncatedDir bytes.Buffer
	)

	if _, err := Decode(bytes.NewReader(nil)); err == nil {
		t.Fatal(errEmptyDecode)
	}
	if _, err := DecodeConfig(bytes.NewReader(nil)); err == nil {
		t.Fatal(errEmptyConfig)
	}
	_ = binary.Write(&badFirst, binary.LittleEndian, uint16(1))
	_ = binary.Write(&badFirst, binary.LittleEndian, uint16(1))
	_ = binary.Write(&badFirst, binary.LittleEndian, uint16(1))
	if _, err := Decode(bytes.NewReader(badFirst.Bytes())); err == nil || !strings.Contains(err.Error(), "first byte") {
		t.Fatalf(tplFatalBadFirst, err)
	}
	_ = binary.Write(&badSecond, binary.LittleEndian, uint16(0))
	_ = binary.Write(&badSecond, binary.LittleEndian, uint16(2))
	_ = binary.Write(&badSecond, binary.LittleEndian, uint16(1))
	if _, err := Decode(bytes.NewReader(badSecond.Bytes())); err == nil || !strings.Contains(err.Error(), "second byte") {
		t.Fatalf(tplFatalBadSecond, err)
	}
	_ = binary.Write(&noImages, binary.LittleEndian, uint16(0))
	_ = binary.Write(&noImages, binary.LittleEndian, uint16(1))
	_ = binary.Write(&noImages, binary.LittleEndian, uint16(0))
	if _, err := Decode(bytes.NewReader(noImages.Bytes())); err == nil || !strings.Contains(err.Error(), "no images") {
		t.Fatalf(tplFatalNoImages, err)
	}
	_ = binary.Write(&truncatedDir, binary.LittleEndian, uint16(0))
	_ = binary.Write(&truncatedDir, binary.LittleEndian, uint16(1))
	_ = binary.Write(&truncatedDir, binary.LittleEndian, uint16(1))
	if _, err := Decode(bytes.NewReader(truncatedDir.Bytes())); err == nil {
		t.Fatal(errTruncatedDirNoErr)
	}
}

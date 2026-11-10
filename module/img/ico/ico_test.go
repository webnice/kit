package ico

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"strings"
	"testing"

	"golang.org/x/image/bmp"
)

type testEntry struct {
	entry entry
	data  []byte
}

type failWriter struct {
	err error
}

func (fw *failWriter) Write([]byte) (int, error) { return 0, fw.err }

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

func testImage(size int) *image.RGBA {
	var im = image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			im.Set(x, y, color.RGBA{R: uint8(100 + x), G: uint8(50 + y), B: 200, A: 255})
		}
	}
	return im
}

func makeICO(entries []testEntry) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(len(entries)))
	for _, item := range entries {
		_ = binary.Write(&buf, binary.LittleEndian, item.entry)
	}
	for _, item := range entries {
		buf.Write(item.data)
	}
	return buf.Bytes()
}

func entryFor(data []byte, width, height uint8, bits uint16) entry {
	return entry{Width: width, Height: height, Plane: 1, Bits: bits, Size: uint32(len(data))}
}

func TestFormatErrorMessage(t *testing.T) {
	if got := FormatError("test").Error(); got != "invalid ICO format: test" {
		t.Fatalf("FormatError.Error() = %q", got)
	}
}

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

func TestDecodeImageErrors(t *testing.T) {
	truncated := makeICO([]testEntry{{entry: entryFor(make([]byte, 10), 1, 1, 32), data: []byte("123")}})
	if _, err := Decode(bytes.NewReader(truncated)); err == nil {
		t.Fatal("Decode(truncated image) не вернул ошибку")
	}
	if _, err := DecodeConfig(bytes.NewReader(truncated)); err == nil {
		t.Fatal("DecodeConfig(truncated image) не вернул ошибку")
	}

	shortData := makeICO([]testEntry{{entry: entryFor([]byte("abcd"), 1, 1, 32), data: []byte("abcd")}})
	if _, err := Decode(bytes.NewReader(shortData)); err == nil {
		t.Fatal("Decode(short png/bmp data) не вернул ошибку")
	}
	if _, err := DecodeConfig(bytes.NewReader(shortData)); err == nil {
		t.Fatal("DecodeConfig(short png/bmp data) не вернул ошибку")
	}

	invalidSize := makeICO([]testEntry{{entry: entry{Width: 1, Height: 1, Bits: 32, Size: 0}}})
	if _, err := Decode(bytes.NewReader(invalidSize)); err == nil {
		t.Fatal("Decode(size=0) не вернул ошибку")
	}
}

func TestDecodeBMPImage(t *testing.T) {
	var bmpBuf bytes.Buffer
	if err := bmp.Encode(&bmpBuf, testImage(2)); err != nil {
		t.Fatalf("bmp.Encode() вернул ошибку: %v", err)
	}
	bmpBytes := bmpBuf.Bytes()
	if len(bmpBytes) <= 14 {
		t.Fatalf("bmp.Encode() too short: %d", len(bmpBytes))
	}
	dib := bmpBytes[14:]
	data := makeICO([]testEntry{{entry: entryFor(dib, 2, 2, 32), data: dib}})
	decoded, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode(BMP ICO) вернул ошибку: %v", err)
	}
	if decoded.Bounds().Dx() != 2 || decoded.Bounds().Dy() != 2 {
		t.Fatalf("Decode(BMP ICO) bounds = %v", decoded.Bounds())
	}
	cfg, err := DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width != 2 || cfg.Height != 2 {
		t.Fatalf("DecodeConfig(BMP ICO) = %v, %v", cfg, err)
	}
}

func TestDecodeConfigBMPSetupErrorPropagates(t *testing.T) {
	data := makeICO([]testEntry{{entry: entryFor([]byte("tiny"), 1, 1, 24), data: []byte("tiny")}})
	if _, err := DecodeConfig(bytes.NewReader(data)); err == nil {
		t.Fatal("DecodeConfig(BMP setup error) не вернул ошибку")
	}
}

func TestSetupBMPColorsBranches(t *testing.T) {
	var d decoder
	if got := d.setupBMPColors(entry{Bits: 8}, 0, 40); got != 256*4 {
		t.Fatalf("setupBMPColors(8bpp default) = %d", got)
	}
	if got := d.setupBMPColors(entry{Bits: 8}, 300, 40); got != 256*4 {
		t.Fatalf("setupBMPColors(8bpp too many colors) = %d", got)
	}
	if got := d.setupBMPColors(entry{Bits: 8}, 10, 40); got != 10*4 {
		t.Fatalf("setupBMPColors(8bpp explicit colors) = %d", got)
	}
	if got := d.setupBMPColors(entry{Bits: 4}, 0, 12); got != 16*3 {
		t.Fatalf("setupBMPColors(4bpp core header) = %d", got)
	}
	if got := d.setupBMPColors(entry{Bits: 1}, 0, 64); got != 2*3 {
		t.Fatalf("setupBMPColors(1bpp v4 header) = %d", got)
	}
	if got := d.setupBMPColors(entry{Bits: 16}, 100, 40); got != 0 {
		t.Fatalf("setupBMPColors(16bpp) = %d", got)
	}
}

func dib24WithMask() []byte {
	data := make([]byte, 48)
	binary.LittleEndian.PutUint32(data[0:4], 40) // dibSize
	binary.LittleEndian.PutUint32(data[4:8], 1)  // width
	binary.LittleEndian.PutUint32(data[8:12], 2) // height includes AND mask
	binary.LittleEndian.PutUint16(data[12:14], 1)
	binary.LittleEndian.PutUint16(data[14:16], 24)
	binary.LittleEndian.PutUint32(data[20:24], 4)
	copy(data[40:44], []byte{0x00, 0x00, 0xFF, 0x00}) // BGR0: red pixel
	// data[44:48] already zero: opaque pixel in AND mask.
	return data
}

func TestDecode24BitBMPWithMask(t *testing.T) {
	dib := dib24WithMask()
	data := makeICO([]testEntry{{entry: entryFor(dib, 1, 1, 24), data: dib}})
	decoded, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode(24bit BMP ICO) вернул ошибку: %v", err)
	}
	if decoded.Bounds().Dx() != 1 || decoded.Bounds().Dy() != 1 {
		t.Fatalf("Decode(24bit BMP ICO) bounds = %v", decoded.Bounds())
	}
	if _, _, _, a := decoded.At(0, 0).RGBA(); a != 0xFFFF {
		t.Fatalf("Decode(24bit BMP ICO) alpha = %#x", a)
	}
}

func TestDecodeBMPUnsupportedDIBSize(t *testing.T) {
	dib := make([]byte, 44)
	binary.LittleEndian.PutUint32(dib[0:4], 12) // BITMAPCOREHEADER не поддерживается текущим декодером.
	data := makeICO([]testEntry{{entry: entryFor(dib, 1, 1, 32), data: dib}})
	if _, err := Decode(bytes.NewReader(data)); err == nil || !strings.Contains(err.Error(), "unsupported BMP DIB") {
		t.Fatalf("Decode(unsupported DIB) = %v", err)
	}
}

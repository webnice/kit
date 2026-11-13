package ico

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"golang.org/x/image/bmp"
)

// testEntry Тестовая запись каталога ICO с данными изображения.
type testEntry struct {
	entry entry
	data  []byte
}

// makeICO Сборка тестового файла ICO из записей.
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

// entryFor Формирование записи каталога ICO.
func entryFor(data []byte, width, height uint8, bits uint16) entry {
	return entry{Width: width, Height: height, Plane: 1, Bits: bits, Size: uint32(len(data))}
}

// TestDecodeImageErrors Ошибки данных изображения ICO.
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

// TestDecodeBMPImage Декодирование BMP данных внутри ICO.
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

// TestDecodeConfigBMPSetupErrorPropagates Ошибка подготовки BMP при чтении конфигурации.
func TestDecodeConfigBMPSetupErrorPropagates(t *testing.T) {
	data := makeICO([]testEntry{{entry: entryFor([]byte("tiny"), 1, 1, 24), data: []byte("tiny")}})
	if _, err := DecodeConfig(bytes.NewReader(data)); err == nil {
		t.Fatal("DecodeConfig(BMP setup error) не вернул ошибку")
	}
}

// TestSetupBMPColorsBranches Ветки расчёта таблицы цветов BMP.
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

// dib24WithMask Формирование 24-битного DIB с маской.
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

// TestDecode24BitBMPWithMask Декодирование 24-битного BMP с маской прозрачности.
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

// TestDecodeBMPUnsupportedDIBSize Неподдерживаемый размер заголовка DIB.
func TestDecodeBMPUnsupportedDIBSize(t *testing.T) {
	dib := make([]byte, 44)
	binary.LittleEndian.PutUint32(dib[0:4], 12) // BITMAPCOREHEADER не поддерживается текущим декодером.
	data := makeICO([]testEntry{{entry: entryFor(dib, 1, 1, 32), data: dib}})
	if _, err := Decode(bytes.NewReader(data)); err == nil || !strings.Contains(err.Error(), "unsupported BMP DIB") {
		t.Fatalf("Decode(unsupported DIB) = %v", err)
	}
}

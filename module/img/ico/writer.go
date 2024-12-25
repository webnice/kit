package ico

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"image"
	"image/draw"
	"image/png"
	"io"
)

type icondir struct {
	reserved  uint16
	imageType uint16
	numImages uint16
}

type icondirentry struct {
	imageWidth   uint8
	imageHeight  uint8
	numColors    uint8
	reserved     uint8
	colorPlanes  uint16
	bitsPerPixel uint16
	sizeInBytes  uint32
	offset       uint32
}

func newIcondir() icondir {
	var id = icondir{
		imageType: 1, // Тип картинки.
		numImages: 1, // Количество картинок.
		reserved:  0, // Резервный байт.
	}
	return id
}

func newIcondirentry() icondirentry {
	var ide = icondirentry{
		colorPlanes:  1,  // Предполагается, что Windows не возражает против 0 или 1, но в других файлах указано 1.
		bitsPerPixel: 32, // Может быть указано 24, для растрового изображения, или 24/32 для png.
		offset:       22, // 6 icondir + 16 icondirentry = 22.
		reserved:     0,  // Резервный байт.
		numColors:    0,  // Резервный байт.
	}
	return ide
}

// Encode Кодирование изображения в  ico формат.
func Encode(w io.Writer, im image.Image) (err error) {
	var (
		b, bounds image.Rectangle
		m         *image.RGBA
		bb, pngBb *bytes.Buffer
		pngWriter *bufio.Writer
		id        icondir
		ide       icondirentry
	)

	b = im.Bounds()
	m = image.NewRGBA(b)
	draw.Draw(m, b, im, b.Min, draw.Src)
	id, ide, pngBb = newIcondir(), newIcondirentry(), &bytes.Buffer{}
	pngWriter = bufio.NewWriter(pngBb)
	if err = png.Encode(pngWriter, m); err != nil {
		return
	}
	_ = pngWriter.Flush()
	ide.sizeInBytes = uint32(len(pngBb.Bytes()))
	bounds = m.Bounds()
	ide.imageWidth, ide.imageHeight, bb = uint8(bounds.Dx()), uint8(bounds.Dy()), &bytes.Buffer{}
	_ = binary.Write(bb, binary.LittleEndian, id)
	_ = binary.Write(bb, binary.LittleEndian, ide)
	if _, err = w.Write(bb.Bytes()); err != nil {
		return
	}
	if _, err = w.Write(pngBb.Bytes()); err != nil {
		return
	}

	return
}

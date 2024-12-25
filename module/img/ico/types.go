package ico

import (
	"image"
	"io"
)

const pngHeader = "\x89PNG\r\n\x1a\n"

type reader interface {
	io.Reader
	io.ByteReader
}

type decoder struct {
	r     reader
	num   uint16
	dir   []entry
	image []image.Image
	cfg   image.Config
}

type entry struct {
	Width   uint8
	Height  uint8
	Palette uint8
	_       uint8 // Резервный байт.
	Plane   uint16
	Bits    uint16
	Size    uint32
	Offset  uint32
}

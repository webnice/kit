package ico

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"

	"golang.org/x/image/bmp"
)

func (d *decoder) decode(r io.Reader, configOnly bool) (err error) {
	var (
		ok bool
		rr reader
	)

	if rr, ok = r.(reader); ok {
		d.r = rr
	} else {
		d.r = bufio.NewReader(r)
	}
	if err = d.readHeader(); err != nil {
		return
	}
	if err = d.readImageDir(configOnly); err != nil {
		return
	}
	if configOnly {
		d.cfg, err = d.parseConfig(d.dir[0])
		if err != nil {
			return
		}
	} else {
		d.image = make([]image.Image, d.num)
		for i, e := range d.dir {
			d.image[i], err = d.parseImage(e)
			if err != nil {
				return
			}
		}
	}

	return
}

func (d *decoder) readHeader() (err error) {
	var first, second uint16

	_ = binary.Read(d.r, binary.LittleEndian, &first)
	_ = binary.Read(d.r, binary.LittleEndian, &second)
	if err = binary.Read(d.r, binary.LittleEndian, &d.num); err != nil {
		return
	}
	if first != 0 {
		return FormatError(fmt.Sprintf("first byte is %d instead of 0", first))
	}
	if second != 1 {
		return FormatError(fmt.Sprintf("second byte is %d instead of 1", second))
	}

	return
}

func (d *decoder) readImageDir(configOnly bool) (err error) {
	var n, i int

	if n = int(d.num); configOnly {
		n = 1
	}
	for i = 0; i < n; i++ {
		var e entry
		if err = binary.Read(d.r, binary.LittleEndian, &e); err != nil {
			return
		}
		d.dir = append(d.dir, e)
	}

	return
}

func (d *decoder) parseImage(e entry) (ret image.Image, err error) {
	var (
		data         []byte
		bmpBytes     []byte
		maskBytes    []byte
		b            []byte
		offset       int
		rowSize      int
		imageRowSize int
		r, c         int
		src          image.Image
		bnd          image.Rectangle
		mask         *image.Alpha
		dst          *image.NRGBA
		alpha        byte
	)

	data = make([]byte, e.Size)
	if _, err = io.ReadFull(d.r, data); err != nil {
		return
	}
	// Проверка, является ли изображение форматом PNG, по первым 8 байтам данных изображения.
	if string(data[:len(pngHeader)]) == pngHeader {
		return png.Decode(bytes.NewReader(data))
	}
	// Декодирование как BMP.
	if bmpBytes, maskBytes, offset, err = d.setupBMP(e, data); err != nil {
		return
	}
	if src, err = bmp.Decode(bytes.NewReader(bmpBytes)); err != nil {
		return
	}
	bnd = src.Bounds()
	mask = image.NewAlpha(image.Rect(0, 0, bnd.Dx(), bnd.Dy()))
	dst = image.NewNRGBA(image.Rect(0, 0, bnd.Dx(), bnd.Dy()))
	rowSize = ((int(e.Width) + 31) / 32) * 4
	b = make([]byte, 4)
	_, _ = offset, b
	for r = 0; r < int(e.Height); r++ {
		for c = 0; c < int(e.Width); c++ {
			_, _ = maskBytes, rowSize
			if len(maskBytes) > 0 {
				if alpha = (maskBytes[r*rowSize+c/8] >> (1 * (7 - uint(c)%8))) & 0x01; alpha != 1 {
					mask.SetAlpha(c, int(e.Height)-r-1, color.Alpha{255})
				}
			}
			// 32-битные bmp-файлы делают хитрые вещи с альфа-каналом, он включен в качестве 4-го байта цветов.
			if e.Bits == 32 {
				imageRowSize = ((int(e.Bits)*int(e.Width) + 31) / 32) * 4
				_, err = io.ReadFull(bytes.NewReader(bmpBytes[offset+r*imageRowSize+c*4:]), b)
				mask.SetAlpha(c, int(e.Height)-r-1, color.Alpha{A: b[3]})
			}
		}
	}
	draw.DrawMask(dst, dst.Bounds(), src, bnd.Min, mask, bnd.Min, draw.Src)
	ret = dst

	return
}

func (d *decoder) parseConfig(e entry) (cfg image.Config, err error) {
	const errSize = "прочитано %d байт из %d байт"
	var (
		tmp []byte
		n   int
	)

	tmp = make([]byte, e.Size)
	switch n, err = io.ReadFull(d.r, tmp); {
	case n != int(e.Size):
		err = fmt.Errorf(errSize, n, e.Size)
		return
	case err != nil:
		return
	}
	if cfg, err = png.DecodeConfig(bytes.NewReader(tmp)); err != nil {
		tmp, _, _, _ = d.setupBMP(e, tmp)
		cfg, err = bmp.DecodeConfig(bytes.NewReader(tmp))
	}

	return cfg, err
}

func (d *decoder) setupBMP(e entry, data []byte) ([]byte, []byte, int, error) {
	const errSize = "прочитано %d байт из %d байт"
	var (
		err                 error
		offset              uint32
		imageSize, maskSize int
		rowSize             int
		numColors           uint32
		n                   int
		dibSize, w, h       uint32
		bpp                 uint16
		size                uint32
		numColorsSize       = d.setupBMPColors(e, numColors, dibSize)
		img, mask           []byte
		iccSize             uint32
	)

	// Вычисление размера изображения.
	// Документация: en.wikipedia.org/wiki/BMP_file_format.
	if imageSize = len(data); int(e.Size) < len(data) {
		imageSize = int(e.Size)
	}
	if e.Bits != 32 {
		rowSize = (1 * (int(e.Width) + 31) / 32) * 4
		maskSize = rowSize * int(e.Height)
		imageSize -= maskSize
	}
	img = make([]byte, 14+imageSize)
	mask = make([]byte, maskSize)
	// Копирование изображения.
	n = copy(img[14:], data[:imageSize])
	if n != imageSize {
		return nil, nil, 0, FormatError(fmt.Sprintf(errSize, n, imageSize))
	}
	// Копирование маски.
	if n = copy(mask, data[imageSize:]); n != maskSize {
		return nil, nil, 0, FormatError(fmt.Sprintf(errSize, n, maskSize))
	}
	_ = binary.Read(bytes.NewReader(img[14:14+4]), binary.LittleEndian, &dibSize)
	_ = binary.Read(bytes.NewReader(img[14+4:14+8]), binary.LittleEndian, &w)
	if err = binary.Read(bytes.NewReader(img[14+8:14+12]), binary.LittleEndian, &h); err != nil {
		return nil, nil, 0, FormatError(fmt.Sprintf("only %d of %d bytes read.", n, maskSize))
	}
	if h > w {
		binary.LittleEndian.PutUint32(img[14+8:14+12], h/2)
	}
	// Магическиое число.
	copy(img[0:2], "\x42\x4D")
	// Размер файла.
	binary.LittleEndian.PutUint32(img[2:6], uint32(imageSize+14))
	// Вычисление позиции данных изображения.
	_ = binary.Read(bytes.NewReader(img[14+32:14+36]), binary.LittleEndian, &numColors)
	if err = binary.Read(bytes.NewReader(img[14+14:14+16]), binary.LittleEndian, &bpp); err != nil {
		return img, mask, int(offset), err
	}
	e.Bits = bpp
	if err = binary.Read(bytes.NewReader(img[14+20:14+24]), binary.LittleEndian, &size); err != nil {
		return img, mask, int(offset), err
	}
	e.Size = size
	offset = 14 + dibSize + numColorsSize
	if dibSize > 40 {
		err = binary.Read(bytes.NewReader(img[14+dibSize-8:14+dibSize-4]), binary.LittleEndian, &iccSize)
		if err != nil {
			return img, mask, int(offset), err
		}
		offset += iccSize
	}
	binary.LittleEndian.PutUint32(img[10:14], offset)

	return img, mask, int(offset), nil
}

func (d *decoder) setupBMPColors(e entry, numColors uint32, dibSize uint32) (numColorsSize uint32) {
	var x uint32

	switch int(e.Bits) {
	case 1, 2, 4, 8:
		x = uint32(1 << e.Bits)
		if numColors == 0 || numColors > x {
			numColors = x
		}
	default:
		numColors = 0
	}

	switch int(dibSize) {
	case 12, 64:
		numColorsSize = numColors * 3
	default:
		numColorsSize = numColors * 4
	}

	return
}

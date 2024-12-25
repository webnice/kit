package ico

import (
	"image"
	"io"
)

func init() {
	image.RegisterFormat("ico", "", Decode, DecodeConfig)
}

// Decode Декодирование картинки в image.Image.
func Decode(r io.Reader) (im image.Image, err error) {
	var d decoder

	if err = d.decode(r, false); err != nil {
		return
	}
	im = d.image[0]

	return
}

// DecodeConfig Декодирование потенциальной картинки с целью узнать являются ли данные картинкой.
func DecodeConfig(r io.Reader) (ret image.Config, err error) {
	var d decoder

	if err = d.decode(r, true); err != nil {
		return
	}
	ret = d.cfg

	return
}

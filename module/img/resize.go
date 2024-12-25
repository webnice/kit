package img

import (
	"image"
	"image/color"

	_ "github.com/webnice/kit/v4/module/img/ico" // Расширение image для формата ico.
	_ "golang.org/x/image/bmp"                   // Расширение image для формата bmp.
	_ "golang.org/x/image/tiff"                  // Расширение image для формата tiff.
	_ "image/gif"                                // Расширение image для формата gif.
	_ "image/jpeg"                               // Расширение image для формата jpeg.
	_ "image/png"                                // Расширение image для формата png.

	"github.com/disintegration/imaging"
)

// Resize Изменение графических размеров изображения.
func (img *impl) Resize(im Image, w, h uint) (ret Image) {
	var (
		minSize int
		newImg  image.Image
	)

	if w != h {
		newImg, minSize = imaging.New(int(w), int(h), color.Transparent), int(w)
		if w > h {
			minSize = int(h)
		}
		newImg = imaging.PasteCenter(newImg, imaging.Resize(im.Image(), minSize, minSize, imaging.Lanczos))
	} else {
		newImg = imaging.Resize(im.Image(), int(w), int(h), imaging.Lanczos)
	}
	ret = img.New().
		SetImage(newImg).
		SetType(im.Type()).
		SetConfig(image.Config{
			Width:      newImg.Bounds().Max.X,
			Height:     newImg.Bounds().Max.Y,
			ColorModel: newImg.ColorModel(),
		}).
		SetFileInfo(im.FileInfo()).
		SetFilename(im.Filename())

	return
}

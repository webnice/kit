package img

const (
	// TypeUnknown Не известный формат графического файла.
	TypeUnknown = Type("")

	// TypeICO Формат графического файла ico.
	TypeICO = Type("ico")

	// TypeBMP Формат графического файла bmp.
	TypeBMP = Type("bmp")

	// TypeTIFF Формат графического файла tiff.
	TypeTIFF = Type("tiff")

	// TypeGIF Формат графического файла gif.
	TypeGIF = Type("gif")

	// TypeJPEG Формат графического файла jpeg.
	TypeJPEG = Type("jpeg")

	// TypePNG Формат графического файла png.
	TypePNG = Type("png")
)

// Type Графический тип изображения.
type Type string

// String Строковое представление графического типа изображения.
func (t Type) String() string { return string(t) }

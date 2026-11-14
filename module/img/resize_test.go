package img

import (
	"os"
	"path/filepath"
	"testing"
)

// TestImageResize Изменение размера изображения.
func TestImageResize(t *testing.T) {
	const (
		tplNilNotNil    = "Resize(nil) = %#v."
		tplEmptyWrong   = "Resize(empty image) = %#v."
		tplFatalWrite   = "WriteFile(): неожиданная ошибка: %v."
		tplFatalStat    = "Stat(): неожиданная ошибка: %v."
		tplResizeState  = "Resize() state = %#v."
		tplSquareConfig = "Resize(square) config = %v."
	)
	var (
		empty   Image
		src     Image
		resized Image
		square  Image
		fi      os.FileInfo
		tmp     string
		err     error
	)

	if got := New().Resize(nil, 1, 1); got != nil {
		t.Fatalf(tplNilNotNil, got)
	}
	empty = New().New()
	if got := New().Resize(empty, 1, 1); got != empty || got.Image() != nil {
		t.Fatalf(tplEmptyWrong, got)
	}
	tmp = filepath.Join(t.TempDir(), "x.png")
	if err = os.WriteFile(tmp, []byte("x"), 0644); err != nil {
		t.Fatalf(tplFatalWrite, err)
	}
	if fi, err = os.Stat(tmp); err != nil {
		t.Fatalf(tplFatalStat, err)
	}
	src = New().New().SetImage(testImage(5, 2)).SetType(TypePNG).SetFilename("src.png").SetFileInfo(fi)
	resized = New().Resize(src, 4, 2)
	if resized.Type() != TypePNG || resized.Filename() != "src.png" || resized.FileInfo() != fi ||
		resized.Config().Width != 4 || resized.Config().Height != 2 || resized.Image() == nil {
		t.Fatalf(tplResizeState, resized)
	}
	square = New().Resize(src, 3, 3)
	if square.Config().Width != 3 || square.Config().Height != 3 {
		t.Fatalf(tplSquareConfig, square.Config())
	}
}

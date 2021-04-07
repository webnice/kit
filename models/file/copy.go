package file

import (
	"fmt"
	"io"
	"os"
)

// Copy Копирует один файл в другой
func (fl *impl) Copy(dst, src string) (size int64, err error) {
	const defaultFileMode = 0644
	var (
		inp *os.File
		out *os.File
	)

	if inp, err = os.Open(src); err != nil {
		err = fmt.Errorf("open file %q, error: %s", src, err)
		return
	}
	defer func() { _ = inp.Close() }()
	if out, err = os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(defaultFileMode)); err != nil {
		err = fmt.Errorf("create file %q, error: %s", dst, err)
		return
	}
	defer func() {
		var le error

		if le = out.Sync(); err == nil {
			err = fmt.Errorf("sync file %q, error: %s", dst, le)
			return
		}
		if le = out.Close(); err == nil {
			err = fmt.Errorf("close file %q, error: %s", dst, le)
			return
		}
	}()
	if size, err = io.Copy(out, inp); err != nil {
		err = fmt.Errorf("copy data from file %q to file %q, error: %s", dst, src, err)
		return
	}

	return
}

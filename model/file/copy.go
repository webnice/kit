package file // import "github.com/webnice/kit/model/file"

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// Copy Копирует один файл в другой
func (fl *impl) Copy(dst string, src string) (size int64, err error) {
	const defaultFileMode = os.FileMode(0644)
	var (
		inp *os.File
		out *os.File
	)

	if inp, err = os.Open(src); err != nil {
		err = fmt.Errorf("open file %q, error: %s", src, err)
		return
	}
	defer func() { _ = inp.Close() }()
	if out, err = os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, defaultFileMode); err != nil {
		err = fmt.Errorf("create file %q, error: %s", dst, err)
		return
	}
	defer func() { _ = out.Sync(); _ = out.Close() }()
	if size, err = io.Copy(out, inp); err != nil {
		err = fmt.Errorf("copy data from file %q to file %q, error: %s", dst, src, err)
		return
	}

	return
}

// CopyWithSha512Sum Копирование контента с параллельным вычислением контрольной суммы алгоритмом SHA512
func (fl *impl) CopyWithSha512Sum(dst io.Writer, src io.Reader) (written int64, sha512sum string, err error) {
	var (
		er, ew       error
		size, nr, nw int
		buf          []byte
		lr           *io.LimitedReader
		ok           bool
		sha          hash.Hash
		end          bool
	)

	size = 32 * 1024
	if lr, ok = src.(*io.LimitedReader); ok && int64(size) > lr.N {
		if size = int(lr.N); lr.N < 1 {
			size = 1
		}
	}
	buf, sha = make([]byte, size), sha512.New()
	for {
		if end {
			break
		}
		if nr, er = src.Read(buf); nr > 0 {
			if _, err = sha.Write(buf[0:nr]); err != nil {
				err = fmt.Errorf("calculate sha-512 sum error: %s", err)
				return
			}
			if nw, ew = dst.Write(buf[0:nr]); nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		switch er {
		case nil:
		case io.EOF:
			end, er = true, nil
		default:
			end, err = true, fmt.Errorf("read source, error: %s", er)
		}
	}
	sha512sum = hex.EncodeToString(sha.Sum(nil))

	return
}

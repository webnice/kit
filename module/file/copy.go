package file

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// fileStat Безопасное получение атрибутов открытого файла.
func fileStat(f *os.File) (fi os.FileInfo) {
	fi, _ = f.Stat()

	return
}

// Copy Копирует один файл в другой.
func (fl *impl) Copy(dst string, src string) (size int64, err error) {
	const defaultFileMode = os.FileMode(0644)
	var (
		inp *os.File
		out *os.File
	)

	if inp, err = os.Open(src); err != nil {
		err = fmt.Errorf("открытие файла %q прервано ошибкой: %s", src, err)
		return
	}
	defer func() { _ = inp.Close() }()
	// Защита от копирования файла в самого себя с потерей данных, проверка до усечения файла.
	if di, e := os.Stat(dst); e == nil && os.SameFile(fileStat(inp), di) {
		err = fmt.Errorf("копирование файла %q в самого себя невозможно", src)
		return
	}
	if out, err = os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, defaultFileMode); err != nil {
		err = fmt.Errorf("создание файла %q прервано ошибкой: %s", dst, err)
		return
	}
	defer func() { _ = out.Sync(); _ = out.Close() }()
	if size, err = io.Copy(out, inp); err != nil {
		err = fmt.Errorf("копирование данных из файла %q в файл %q завершилось ошибкой: %s", dst, src, err)
		return
	}

	return
}

// CopyWithSha512Sum Копирование контента с параллельным вычислением контрольной суммы алгоритмом SHA512.
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
			// В соответствии с контрактом hash.Hash, запись в хеш ошибок не возвращает.
			_, _ = sha.Write(buf[0:nr])
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
			end, err = true, fmt.Errorf("чтение данных прервано ошибкой: %s", er)
		}
	}
	sha512sum = hex.EncodeToString(sha.Sum(nil))

	return
}

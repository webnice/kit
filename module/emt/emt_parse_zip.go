package emt

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

// ParseZip Загрузка и разбор шаблона из ZIP файла.
func (emt *impl) ParseZip(file io.Reader) (ret *Template, err error) {
	const (
		errRead  = "загрузка файла прервана ошибкой: %s"
		errOpen  = "открытие ZIP файла прервано ошибкой: %s"
		errFile  = "извлечение файла %q прервано ошибкой: %s"
		errUnzip = "распаковка файла %q прервано ошибкой: %s"
		errSize  = "размер файла %q не верный"
		errLimit = "размер файла %q превышает допустимый лимит %d байт"
	)
	var (
		buf *bytes.Buffer
		brr *bytes.Reader
		zrc *zip.Reader
		frc io.ReadCloser
		att *File
		n   int
		dat []byte
	)

	buf = &bytes.Buffer{}
	if _, err = io.Copy(buf, file); err != nil {
		err = fmt.Errorf(errRead, err)
		return
	}
	brr = bytes.NewReader(buf.Bytes())
	ret = &Template{
		Body:     []Body{},
		Tags:     []Tags{},
		Embedded: []File{},
		Attach:   []File{},
	}
	if zrc, err = zip.NewReader(brr, brr.Size()); err != nil {
		err = fmt.Errorf(errOpen, err)
		return
	}
	defer func() {
		dat = dat[:0]
		brr.Reset([]byte{})
	}()
	for n = range zrc.File {
		if zrc.File[n].FileInfo().IsDir() {
			continue
		}
		if zrc.File[n].UncompressedSize64 > size1gb {
			err = fmt.Errorf(errLimit, zrc.File[n].Name, uint64(size1gb))
			return
		}
		if frc, err = zrc.File[n].Open(); err != nil {
			err = fmt.Errorf(errFile, zrc.File[n].Name, err)
			return
		}
		dat = make([]byte, 0, int(min(zrc.File[n].UncompressedSize64, 64*1024)))
		if dat, err = io.ReadAll(frc); err != nil {
			_ = frc.Close()
			err = fmt.Errorf(errUnzip, zrc.File[n].Name, err)
			return
		}
		if err = frc.Close(); err != nil {
			err = fmt.Errorf(errUnzip, zrc.File[n].Name, err)
			return
		}
		if uint64(len(dat)) != zrc.File[n].UncompressedSize64 {
			err = fmt.Errorf(errSize, zrc.File[n].Name)
			return
		}
		att = &File{Name: zrc.File[n].Name}
		att.ValueSet(dat)
		ret.Attach = append(ret.Attach, *att)
	}
	// Определение Content-Type по расширению имени файла.
	for n = range ret.Attach {
		ret.Attach[n].Type = parseFilename(ret.Attach[n].Name).ContentType()
	}
	// Раскладывание найденных кусочков по полочкам.
	if err = emt.Decomposition(ret); err != nil {
		return
	}
	// Извлечение из HTML тела встраиваемого контента.
	emt.extractUri(ret)

	return
}

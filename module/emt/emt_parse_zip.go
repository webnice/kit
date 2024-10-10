package emt

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime"

	"github.com/webnice/dic"
)

// ParseZip Загрузка и разбор шаблона из ZIP файла.
func (emt *impl) ParseZip(file io.Reader) (ret *Template, err error) {
	const (
		errRead  = "загрузка файла прервана ошибкой: %s"
		errOpen  = "открытие ZIP файла прервано ошибкой: %s"
		errFile  = "извлечение файла %q прервано ошибкой: %s"
		errUnzip = "распаковка файла %q прервана ошибкой: %s"
		errSize  = "размер файла %q не верный"
	)
	var (
		buf *bytes.Buffer
		brr *bytes.Reader
		zrc *zip.Reader
		frc io.ReadCloser
		att *File
		n   int
		dat []byte
		tmp []string
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
	dat = make([]byte, 0, size1gb)
	defer func() {
		dat = dat[:0]
		brr.Reset([]byte{})
	}()
	for n = range zrc.File {
		if frc, err = zrc.File[n].Open(); err != nil {
			err = fmt.Errorf(errFile, zrc.File[n].Name, err)
			return
		}
		dat = dat[:0]
		if dat, err = io.ReadAll(frc); err != nil {
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
		tmp = rexExt.FindStringSubmatch(ret.Attach[n].Name)
		ret.Attach[n].Type = dic.Mime().ApplicationOctetStream.String()
		if len(tmp) > 1 {
			ret.Attach[n].Type = mime.TypeByExtension("." + tmp[1])
		}
	}
	// Раскладывание найденных кусочков по полочкам.
	if err = emt.Decomposition(ret); err != nil {
		return
	}
	// Извлечение из HTML тела встраиваемого контента.
	emt.extractUri(ret)

	return
}

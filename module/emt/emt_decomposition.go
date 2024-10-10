package emt

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/webnice/dic"
)

// Decomposition Распределение файлов по типам контента.
// Body     - Шаблоны.
// Embedded - Встраиваемые вложения.
// Tags     - Теги.
func (emt *impl) Decomposition(tpl *Template) (err error) {
	const (
		indexTemplate = "index"
		tagsFilename  = "index.json"
		errCopy       = "копирование данных прервано, скопировано %d из %d объектов"
		errTags       = "загрузка файла тегов %q прервана ошибкой: %s"
	)
	var (
		fn   *filename
		tmp  []File
		next []File
		tags *File
		n, j int
		buf  []byte
		find []byte
	)

	tmp = make([]File, len(tpl.Attach))
	next = make([]File, 0, len(tpl.Attach))
	if n = copy(tmp, tpl.Attach); n != len(tpl.Attach) {
		err = fmt.Errorf(errCopy, n, len(tpl.Attach))
		return
	}
	defer func() { tmp, next, buf, find = tmp[:0], next[:0], buf[:0], find[:0] }()
	tpl.Attach = tpl.Attach[:0]
	// Поиск шаблонов, имя файла должно быть index, типы: txt, htm, html, markdown.
	for n = range tmp {
		fn = parseFilename(tmp[n].Name)
		if fn.IsPath() || fn.IsPathMacos() || fn.IsDotFirst() {
			continue
		}
		if len(tmp[n].Value) == 0 {
			continue
		}
		if strings.EqualFold(fn.Filename, indexTemplate) {
			if strings.HasPrefix(fn.ContentType(), dic.Mime().TextHtml.String()) ||
				strings.HasPrefix(fn.ContentType(), dic.Mime().TextPlain.String()) ||
				strings.HasPrefix(fn.ContentType(), dic.Mime().TextMarkdown.String()) {
				tpl.Body = append(tpl.Body, Body{
					Type:  tmp[n].Type,
					Value: tmp[n].Value,
				})
				continue
			}
			if strings.HasPrefix(fn.ContentType(), dic.Mime().ApplicationJson.String()) {
				tags = &File{
					Type:  tmp[n].Type,
					Name:  tmp[n].Name,
					Value: tmp[n].Value,
				}
				continue
			}
		}
		next = append(next, tmp[n])
	}
	// Если нашлись теги, попытка их декодировать.
	if tags != nil {
		if err = tags.JsonUnmarshal(&tpl.Tags); err != nil {
			err = fmt.Errorf(errTags, tagsFilename, err)
			return
		}
	}
	// Поиск в шаблоне с типом text/html упоминаний файлов, найденные файлы перемещаются во встраиваемые вложения.
	for n = range tpl.Body {
		if !strings.HasPrefix(tpl.Body[n].Type, dic.Mime().TextHtml.String()) {
			continue
		}
		if buf = tpl.Body[n].ValueGet(); len(buf) == 0 {
			continue
		}
		tmp = tmp[:0]
		for j = range next {
			find = []byte(fmt.Sprintf("%q", next[j].Name))
			if bytes.Contains(buf, find) {
				tpl.Embedded = append(tpl.Embedded, next[j])
				continue
			}
			tmp = append(tmp, next[j])
		}
		next = next[:0]
		next = append(next, tmp...)
	}
	if len(next) > 0 {
		tpl.Attach = make([]File, len(next))
		if n = copy(tpl.Attach, next); n != len(next) {
			err = fmt.Errorf(errCopy, n, len(tpl.Attach))
			return
		}
	}

	return
}

// Поиск в теле шаблонов встроенных изображений методом data:url и извлечение их в срез Embedded.
func (emt *impl) extractUri(tpl *Template) {
	const (
		sepComma, sepSemicolon = ",", ";"
		keyData                = "data"
		errParse               = "разбор HTML прерван ошибкой: %s"
	)
	var (
		err        error
		n, j       int
		buf        *bytes.Buffer
		uri, uries []*url.URL
		duT1, duT2 []string
		embedded   *File
		slice      []byte
	)

	for n = range tpl.Body {
		buf = bytes.NewBuffer(tpl.Body[n].ValueGet())
		if uries, err = emt.findImageUriFromHTML(buf); err != nil {
			emt.log().Warningf(errParse, err)
			err = nil
			continue
		}
		uri = make([]*url.URL, 0, len(uries))
		for j = range uries {
			uri = append(uri, uries[j])
		}
		for j = range uri {
			switch strings.ToLower(uri[j].Scheme) {
			case keyData:
				// Преобразование data-uri в Embedded контент и добавление в список объектов.
				if duT1 = strings.SplitN(uri[j].Opaque, sepComma, 2); len(duT1) == 2 {
					if duT2 = strings.SplitN(duT1[0], sepSemicolon, 2); len(duT2) == 2 {
						embedded = &File{
							Type: duT2[0],
							Name: uri[j].String(),
						}
						if slice, err = base64.StdEncoding.DecodeString(duT1[1]); err == nil {
							embedded.ValueSet(slice)
							tpl.Embedded = append(tpl.Embedded, *embedded)
						}
						err = nil
					}
				}
			default:
				// Добавление ссылки в список ссылок.
				tpl.BodyImgUri = append(tpl.BodyImgUri, uri[j])
			}
		}
	}
}

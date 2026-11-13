package emt

import (
	"bytes"

	"github.com/webnice/dic"
)

// BodyImageUri Обход, найденных в теле шаблонов с типом HTML, ссылок на изображения и замена их.
// Для каждого найденного изображения, происходит вызов функции, которая принимает решение о замене изображения.
// В случае положительного решения, функции возвращает новый URI адрес изображения, который вставляется в шаблон.
func (emt *impl) BodyImageUri(tpl *Template, fn ImageUriFn) (err error) {
	var (
		cts     string
		mme     dic.IMime
		newUri  string
		n       int
		replace []*replaceContent
	)

	// Обход.
	for n = range tpl.BodyImgUri {
		newUri = ""
		// Всю ссылку нельзя распознавать, мешает urn-param часть.
		cts = parseFilename(tpl.BodyImgUri[n].Path).ContentType()
		if mme = dic.ParseMime(cts); mme == nil || mme.IsEqual(dic.Mime().ApplicationOctetStream) {
			continue
		}
		if err = emt.safeCall(func() { newUri = fn(mme, tpl.BodyImgUri[n]) }); err != nil {
			emt.log().Warning(err.Error())
			err = nil
			continue
		}
		if newUri == "" {
			continue
		}
		replace = append(replace, &replaceContent{
			Old:  tpl.BodyImgUri[n].String(),
			New:  newUri,
			Type: mme,
		})
	}
	// Замена значений в шаблоне.
	emt.bodyReplace(tpl, replace)

	return
}

// BodyImageEmbed Обход, найденных в теле шаблонов с типом HTML, встроенных изображений методом data:url, а так же
// найденных в архиве встраиваемых изображений.
// Для каждого изображения, происходит вызов функции, которая принимает решение о замене изображения.
// В случае положительного решения, функции возвращает новый URI адрес изображения, который вставляется в шаблон.
func (emt *impl) BodyImageEmbed(tpl *Template, fn ImageEmbedFn) (err error) {
	var (
		buf     *bytes.Buffer
		mme     dic.IMime
		n       int
		newUri  string
		replace []*replaceContent
	)

	for n = range tpl.Embedded {
		newUri = ""
		mme = dic.ParseMime(tpl.Embedded[n].Type)
		if mme == nil || mme.IsEqual(dic.Mime().ApplicationOctetStream) {
			continue
		}
		buf = bytes.NewBuffer(tpl.Embedded[n].ValueGet())
		if err = emt.safeCall(func() { newUri = fn(mme, buf) }); err != nil {
			emt.log().Warning(err.Error())
			err = nil
			continue
		}
		if newUri == "" {
			continue
		}
		replace = append(replace, &replaceContent{
			Old:  tpl.Embedded[n].Name,
			New:  newUri,
			Type: mme,
		})
	}
	// Замена значений в шаблоне.
	emt.bodyReplace(tpl, replace)

	return
}

package emt

import (
	"io"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// Поиск в HTML ссылок на изображения.
func (emt *impl) findImageUriFromHTML(d io.Reader) (ret []*url.URL, err error) {
	const errUri = "поиск URI прерван ошибкой: %s"
	var (
		doc *goquery.Document
		uri *url.URL
	)

	if doc, err = goquery.NewDocumentFromReader(d); err != nil {
		return
	}
	doc.Find("img").Each(func(n int, s *goquery.Selection) {
		var (
			tmp string
			ok  bool
		)

		if tmp, ok = s.Attr("src"); ok {
			if uri, err = url.Parse(tmp); err != nil {
				emt.log().Warningf(errUri, err)
				err = nil
				return
			}
			if uri.Scheme == "" {
				return
			}
			ret = append(ret, uri)
		}
	})

	return
}

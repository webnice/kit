package emt

import (
	"encoding/base64"
	"fmt"

	"github.com/webnice/dic"
)

// Embedding Встраивание встраиваемого контента в шаблоны сообщений.
func (emt *impl) Embedding(tpl *Template) (err error) {
	const dataUrlTemplate = "data:%s;base64,%s"
	var (
		mme     dic.IMime
		n       int
		replace []*replaceContent
	)

	// Встраивание контента в тело шаблона методом data-url. Подготовка данных для замены.
	for n = range tpl.Embedded {
		if mme = dic.ParseMime(tpl.Embedded[n].Type); mme == nil {
			continue
		}
		replace = append(replace, &replaceContent{
			Type: mme,
			Old:  tpl.Embedded[n].Name,
			New: fmt.Sprintf(
				dataUrlTemplate,
				tpl.Embedded[n].Type,
				base64.StdEncoding.EncodeToString(tpl.Embedded[n].ValueGet()),
			),
		})
	}
	// Замена значений в шаблоне.
	emt.bodyReplace(tpl, replace)

	return
}

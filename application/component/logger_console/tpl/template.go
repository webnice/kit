// Package tpl
package tpl

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	kitModuleLog "github.com/webnice/kit/module/log"
	kitModuleTrace "github.com/webnice/kit/module/trace"

	"github.com/muesli/termenv"
)

// Template Представления шаблона в виде text/template.
func (tpl *impl) Template(name string) (ret *template.Template, err error) {
	const tplPanic = `Работа с шаблоном прервана паникой:` + "\n%v\n%s."
	var (
		s strings.Builder
		n int
	)

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(tplPanic, e, kitModuleTrace.StackShort())
		}
	}()
	for n = range tpl.chunks {
		_, _ = s.WriteString(tpl.chunks[n].Template())
	}
	if ret, err = template.
		New(name).
		Funcs((&session{parent: tpl}).makeFunc()).
		Parse(s.String()); err != nil {
		return
	}

	return
}

// NewSession Создание сесии обработки данных по шаблону.
func (tpl *impl) NewSession(data *kitModuleLog.Message) (ret Session, err error) {
	const tplParseFirst = "перед созданием сессии, необходимо выполнить функцию Parse()"
	var ses = &session{
		parent:  tpl,
		writer:  &bytes.Buffer{},
		profile: termenv.ColorProfile(),
		Data:    data,
	}

	if tpl.tpl == nil {
		err = fmt.Errorf(tplParseFirst)
		return
	}
	ses.Tpl, err = tpl.tpl.Clone()
	ret = ses

	return
}

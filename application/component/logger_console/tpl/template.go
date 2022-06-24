// Package tpl
package tpl

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	kitModuleDye "github.com/webnice/kit/v3/module/dye"
	kitModuleLog "github.com/webnice/kit/v3/module/log"
	kitModuleTrace "github.com/webnice/kit/v3/module/trace"
)

// Template Представления шаблона в виде text/template.
func (tpl *impl) Template(name string) (ret *template.Template, err error) {
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
	var ses = &session{
		parent: tpl,
		writer: &bytes.Buffer{},
		Data:   data,
	}

	if tpl.tpl == nil {
		err = fmt.Errorf(tplParseFirst)
		return
	}
	ses.Tpl, err = tpl.tpl.Clone()
	ret = ses

	return
}

// Do Выполнение обработки данных через шаблонизатор.
func (ses *session) Do() (ret *bytes.Buffer, err error) {
	var fnc template.FuncMap

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(tplParsePanic, e, kitModuleTrace.StackShort())
		}
	}()
	fnc = ses.makeFunc()
	if err = ses.Tpl.
		Funcs(fnc).
		Execute(ses.writer, nil); err != nil {
		return
	}
	ses.writer.Write(kitModuleDye.New().Reset().Done().Byte())
	ret = ses.writer

	return
}

// Создание функций для шаблонизатора.
func (ses *session) makeFunc() (ret template.FuncMap) {
	var (
		tdi map[string]*tagDataInfo
		key string
	)

	ret = make(map[string]interface{})
	tdi = ses.tagData()
	for key = range tdi {
		ret[tdi[key].Name] = tdi[key].Func
	}
	tdi = ses.tagColor()
	for key = range tdi {
		ret[tdi[key].Name] = tdi[key].Func
	}

	return
}

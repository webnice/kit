// Package tpl
package tpl

import (
	"bytes"
	"fmt"
	"text/template"

	kitModuleTrace "github.com/webnice/kit/module/trace"
)

// Do Выполнение обработки данных через шаблонизатор.
func (ses *session) Do() (ret *bytes.Buffer, err error) {
	const tplPanic = `Обработка данных шаблонизатором прервана паникой:` + "\n%v\n%s."
	var fnc template.FuncMap

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(tplPanic, e, kitModuleTrace.StackShort())
		}
	}()
	fnc = ses.makeFunc()
	if err = ses.Tpl.
		Funcs(fnc).
		Execute(ses.writer, nil); err != nil {
		return
	}
	ses.writer.WriteString(fmt.Sprintf("%sm", seqCSI+seqResetSeq))
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

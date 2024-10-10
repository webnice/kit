package emt

import (
	"bytes"
	"encoding/base64"
	"fmt"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitModuleTrace "github.com/webnice/kit/v4/module/trace"
)

// Вызов функции с защитой от паники.
func (emt *impl) safeCall(fn func()) (err error) {
	const errPanic = "Паника: %q\nСтек вызовов, в момент паники:\n%s."
	// Функция защиты от паники.
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(errPanic, e, kitModuleTrace.StackShort())
		}
	}()
	fn()
	return
}

// Замена значений в шаблоне.
func (emt *impl) bodyReplace(tpl *Template, replace []*replaceContent) {
	var (
		buf   *bytes.Buffer
		slice []byte
		n, j  int
	)

	for n = range tpl.Body {
		buf = bytes.NewBuffer(tpl.Body[n].ValueGet())
		for j = range replace {
			slice = bytes.ReplaceAll(buf.Bytes(), []byte(replace[j].Old), []byte(replace[j].New))
			buf = bytes.NewBuffer(slice)
		}
		tpl.Body[n].ValueSet(buf.Bytes())
	}
}

// Кодирование среза в BASE64.
func encodeBase64(b []byte) (ret string) {
	if len(b) == 0 {
		return
	}
	ret = base64.StdEncoding.EncodeToString(b)

	return
}

// Декодирование BASE64 в срез.
func decodeBase64(s string) (ret []byte) {
	var err error

	if s == "" {
		return
	}
	if ret, err = base64.StdEncoding.DecodeString(s); err != nil {
		kitModuleCfg.Get().Log().Errorf("декодирование BASE64 в срез байт прерван ошибкой: %s", err)
		ret = []byte{}
	}

	return
}

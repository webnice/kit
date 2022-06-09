// Package tpl
package tpl

import "fmt"

// Куски шаблона.
type chunk struct {
	Src  string      // Исходный текст куска.
	Arg  string      // Исходные аргументы тега куска.
	Type chunkType   // Тип куска.
	Tag  interface{} // Распознанный тег. Принимает значение *tagDataInfo или tagFormatInfo.
}

// String Интерфейс Stringer.
func (c chunk) String() (ret string) {
	switch {
	case c.Type == chunkText:
		ret = c.Src
	case c.Src != "" && c.Arg != "":
		ret = fmt.Sprintf("${%s:%s}", c.Src, c.Arg)
	case c.Src != "":
		ret = fmt.Sprintf("${%s}", c.Src)
	default:
		ret = "${}"
	}

	return
}

func (c chunk) makeCallFnWithArgs(numberArgs int) (ret string) {
	const prefix, tplarg, suffix = `{{ %s `, `%q `, `}}`
	var n int

	for ret, n = prefix, 0; n < numberArgs; n++ {
		ret += tplarg
	}
	ret += suffix

	return
}

// Template Представления куска как шаблона template/text.
func (c chunk) Template() (ret string) {
	const tplQuoteString = "{{%q}}"
	var (
		args  []string
		argn  int
		argt  string
		arg   string
		arr   []interface{}
		tdi   *tagDataInfo
		ok    bool
		quote bool
	)

	switch c.Type {
	case chunkText:
		// Кусок является текстом, выводим как есть, но на всякий случай экранируем, если есть скобки.
		if rexCurlyBraceOpen.MatchString(c.Src) {
			quote = true
		}
		if rexCurlyBraceClose.MatchString(c.Src) {
			quote = true
		}
		if ret = c.Src; quote {
			ret = fmt.Sprintf(tplQuoteString, c.Src)
		}
	case chunkData:
		// Кусок является тегом для вывода данных.
		if tdi, ok = c.Tag.(*tagDataInfo); !ok {
			return
		}
		switch tdi.Name {
		case keyLine:
			argn, argt = 0, c.makeCallFnWithArgs(0)
		case keyTimestamp, keyLevel, keyMessage, keyKeys:
			argn, argt = 2, c.makeCallFnWithArgs(2)
		case keyDye:
			argn, argt = 3, c.makeCallFnWithArgs(3)
		default:
			argn, argt = 1, c.makeCallFnWithArgs(1)
		}
		args = splitBySeparator(c.Arg, keySep)
		arr = append(arr, tdi.Name)
		for _, arg = range fixSlice(args, argn, keySep) {
			arr = append(arr, arg)
		}
		ret = fmt.Sprintf(argt, arr...)
	case chunkColor:
		// Кусок является тегом переключения цвета.
		if tdi, ok = c.Tag.(*tagDataInfo); !ok {
			return
		}
		args = splitBySeparator(c.Arg, keySep)
		arr = append(arr, tdi.Name)
		for _, arg = range fixSlice(args, 3, keySep) {
			arr = append(arr, arg)
		}
		ret = fmt.Sprintf(c.makeCallFnWithArgs(3), arr...)
	}

	return
}

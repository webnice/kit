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

// Template Представления куска как шаблона template/text.
func (c chunk) Template() (ret string) {
	const sep = ":"
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
			ret = fmt.Sprintf("{{%q}}", c.Src)
		}
	case chunkData:
		// Кусок является тегом для вывода данных.
		if tdi, ok = c.Tag.(*tagDataInfo); !ok {
			return
		}
		switch tdi.Name {
		case "line":
			argn, argt = 0, "{{ %s }}"
		case "timestamp", "level", "message", "keys":
			argn, argt = 2, "{{ %s %q %q }}"
		case "dye":
			argn, argt = 3, "{{ %s %q %q %q }}"
		default:
			argn, argt = 1, "{{ %s %q }}"
		}
		args = splitBySeparator(c.Arg, sep)
		arr = append(arr, tdi.Name)
		for _, arg = range fixSlice(args, argn, sep) {
			arr = append(arr, arg)
		}
		ret = fmt.Sprintf(argt, arr...)
	case chunkColor:
		// Кусок является тегом переключения цвета.
		if tdi, ok = c.Tag.(*tagDataInfo); !ok {
			return
		}
		args = splitBySeparator(c.Arg, sep)
		arr = append(arr, tdi.Name)
		for _, arg = range fixSlice(args, 3, sep) {
			arr = append(arr, arg)
		}
		ret = fmt.Sprintf("{{ %s %q %q %q }}", arr...)
	}

	return
}

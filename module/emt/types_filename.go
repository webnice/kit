package emt

import (
	"mime"
	"os"
	"strings"

	"github.com/webnice/dic"
)

// Путь, имя файла и расширение файла разобранные на части.
type filename struct {
	Path      []string
	Filename  string
	Extension string
}

// String Реализация интерфейса Stringify.
func (fn filename) String() string {
	var sb = &strings.Builder{}

	_, _ = sb.WriteString(strings.Join(fn.Path, string(os.PathSeparator)))
	if sb.Len() > 0 {
		_ = sb.WriteByte(os.PathSeparator)
	}
	_, _ = sb.WriteString(fn.FilenameFull())

	return sb.String()
}

// FilenameFull Имя файла с расширением.
func (fn filename) FilenameFull() string {
	var sb = &strings.Builder{}

	if fn.Filename != "" && fn.Extension != "" {
		_, _ = sb.WriteString(strings.Join([]string{fn.Filename, fn.Extension}, keyExt))
	} else if fn.Filename != "" && fn.Extension == "" {
		_, _ = sb.WriteString(fn.Filename)
	}

	return sb.String()
}

// IsPath Вернётся значение "истина", если объект является директорией.
func (fn filename) IsPath() bool {
	switch {
	case fn.Filename == "" && fn.Extension == "":
		return true
	default:
		return false
	}
}

// IsDotFirst Вернётся значение "истина", если имя файла или один из сегментов пути начинается на точку.
func (fn filename) IsDotFirst() (ret bool) {
	var n int

	for n = range fn.Path {
		if len(fn.Path[n]) > 0 && string(fn.Path[n][0]) == keyExt {
			ret = true
		}
	}
	if len(fn.Filename) > 0 && string(fn.Filename[0]) == keyExt {
		ret = true
	}

	return
}

// IsPathPrefix Вернётся значение "истина", если любой из сегментов пути содержит строку prefix.
func (fn filename) IsPathPrefix(prefix string) (ret bool) {
	if len(prefix) == 0 {
		return
	}
	for n := range fn.Path {
		if len(fn.Path[n]) > 0 && strings.HasPrefix(fn.Path[n], prefix) {
			ret = true
			break
		}
	}

	return
}

// IsPathMacos Вернётся значение "истина", если любой из сегментов пути содержит ключевое слово "__MACOS".
func (fn filename) IsPathMacos() (ret bool) { return fn.IsPathPrefix(keyPathMacos) }

// ContentType Определение типа контента по расширению.
func (fn filename) ContentType() (ret string) {
	if ret = dic.Mime().ApplicationOctetStream.String(); fn.Extension != "" {
		ret = mime.TypeByExtension("." + fn.Extension)
	}

	return
}

// Разбор строки в структуру имени файла.
func parseFilename(s string) (ret *filename) {
	var tmp []string

	ret = new(filename)
	if tmp = strings.Split(s, string(os.PathSeparator)); len(tmp) == 0 {
		return
	}
	ret.Filename = tmp[len(tmp)-1]
	if tmp = tmp[:len(tmp)-1]; len(tmp) > 0 {
		ret.Path = make([]string, len(tmp))
		_ = copy(ret.Path, tmp)
	}
	if tmp = strings.Split(ret.Filename, keyExt); len(tmp) == 0 || (len(tmp) == 1 && tmp[0] == "") {
		return
	}
	if tmp[0] == "" && len(tmp) == 2 {
		tmp = tmp[1:]
		ret.Filename = strings.Join([]string{keyExt, tmp[0]}, "")
		return
	}
	if len(tmp) == 1 {
		ret.Filename = tmp[0]
		return
	}
	ret.Extension = tmp[len(tmp)-1]
	ret.Filename = strings.Join(tmp[:len(tmp)-1], keyExt)

	return
}

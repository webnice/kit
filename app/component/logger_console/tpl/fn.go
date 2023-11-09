package tpl

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Функция шаблонизатора - Время записи.
func (ses *session) fnTimestamp(tz string, format string) (ret string) {
	var (
		err error
		loc *time.Location
		tmo time.Time
	)

	switch tz {
	case "":
		tmo = ses.Data.Timestamp
	default:
		if loc, err = time.LoadLocation(tz); err != nil {
			err = fmt.Errorf(tplTimezoneError, tz, err)
			ret = err.Error()
			return
		}
		tmo = ses.Data.Timestamp.In(loc)
	}
	switch format {
	case "":
		ret = tmo.String()
	default:
		ret = tmo.Format(format)
	}

	return
}

// Функция шаблонизатора - Уровень логирования сообщения.
func (ses *session) fnLevel(t string, l string) (ret string) {
	const (
		keySUp, keySDown = "S", "s"
		keyDUp, keyDDown = "D", "d"
	)
	var (
		li  int
		buf []rune
	)

	switch li = stringToInt(l); t {
	case keySUp:
		switch li {
		case 0, 1:
			ret = strings.ToUpper(string(ses.Data.Level.Short()))
		default:
			buf = runeMax(stringToRune(ses.Data.Level.String()), li, "")
			ret = strings.ToUpper(runeToString(buf))
		}
	case keySDown:
		switch li {
		case 0, 1:
			ret = strings.ToLower(string(ses.Data.Level.Short()))
		default:
			buf = runeMax(stringToRune(ses.Data.Level.String()), li, "")
			ret = strings.ToLower(runeToString(buf))
		}
	case keyDUp, keyDDown:
		switch li {
		case 0:
			ret = strconv.Itoa(ses.Data.Level.Int())
		default:
			ret = fmt.Sprintf(`%0`+strconv.Itoa(li)+`d`, ses.Data.Level.Int())
		}
	}

	return
}

// Функция шаблонизатора - Шаблон и параметры шаблона, сообщения лога в виде сформированного сообщения.
func (ses *session) fnMessage(mn string, mx string) (ret string) {
	var (
		src      string
		msg      []rune
		min, max int
	)

	switch min, max = stringToInt(mn), stringToInt(mx); ses.Data.Pattern.Len() {
	case 0:
		src = fmt.Sprint(ses.Data.Argument...)
	default:
		src = fmt.Sprintf(ses.Data.Pattern.String(), ses.Data.Argument...)
	}
	msg = stringToRune(src)
	msg = runeMin(msg, min)
	msg = runeMax(msg, max, tplSuffix)
	ret = runeToString(msg)

	return
}

// Функция шаблонизатора - Ключи сообщения.
func (ses *session) fnKeys(s string, skv string) (ret string) {
	const defaultS, defaultSkv = "=", ", "
	var (
		key string
		buf strings.Builder
	)

	if len(ses.Data.Keys) == 0 {
		return
	}
	if s == "" {
		s = defaultS
	}
	if skv == "" {
		skv = defaultSkv
	}
	for key = range ses.Data.Keys {
		_, _ = buf.WriteString(key)
		_, _ = buf.WriteString(s)
		_, _ = buf.WriteString(fmt.Sprint(ses.Data.Keys[key]))
		_, _ = buf.WriteString(skv)
	}
	ret = strings.TrimSuffix(buf.String(), skv)

	return
}

// Функция шаблонизатора - Стек вызовов активного процесса, обрезанный до функции вызова.
func (ses *session) fnStacktrace(l string) (ret string) {
	ret = runeToString(
		runeMax(
			stringToRune(ses.Data.Trace.StackTrace.String()),
			stringToInt(l),
			tplSuffix,
		),
	)

	return
}

// Функция шаблонизатора - Путь и имя файла приложения из которого был совершён вызов.
func (ses *session) fnLongfile(l string) (ret string) {
	ret = runeToString(
		runeMax(
			stringToRune(ses.Data.Trace.FilenameLong.String()),
			stringToInt(l),
			tplSuffix,
		),
	)

	return
}

// Функция шаблонизатора - Название файла из которого был совершён вызов.
func (ses *session) fnShortfile(l string) (ret string) {
	ret = runeToString(
		runeMax(
			stringToRune(ses.Data.Trace.FilenameShort.String()),
			stringToInt(l),
			tplSuffix,
		),
	)

	return
}

// Функция шаблонизатора - Название функции совершившей вызов.
func (ses *session) fnFunction(l string) (ret string) {
	ret = runeToString(
		runeMax(
			stringToRune(ses.Data.Trace.Function.String()),
			stringToInt(l),
			tplSuffix,
		),
	)

	return
}

// Функция шаблонизатора - Номер строки файла из которого был совершён вызов.
func (ses *session) fnLine() (ret string) { ret = strconv.Itoa(ses.Data.Trace.Line); return }

// Функция шаблонизатора - Название пакета.
func (ses *session) fnPackage(l string) (ret string) {
	ret = runeToString(
		runeMax(
			stringToRune(ses.Data.Trace.Package.String()),
			stringToInt(l),
			tplSuffix,
		),
	)

	return
}

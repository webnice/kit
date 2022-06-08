// Package tpl
package tpl

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	kmll "github.com/webnice/kit/module/log/level"

	"github.com/muesli/termenv"
)

func (ses *session) colorByLevel() (fg string, bg string) {
	type ansiStyle = struct {
		Bg string
		Fg string
	}
	var (
		lm    map[kmll.Level]ansiStyle
		style ansiStyle
		ok    bool
	)

	lm = map[kmll.Level]ansiStyle{
		kmll.Fatal:    {Bg: termenv.ANSIRed.Sequence(true), Fg: termenv.ANSIBrightYellow.Sequence(false)},
		kmll.Alert:    {Bg: termenv.ANSIMagenta.Sequence(true), Fg: termenv.ANSIBrightWhite.Sequence(false)},
		kmll.Critical: {Bg: termenv.ANSIBlue.Sequence(true), Fg: termenv.ANSIBrightMagenta.Sequence(false)},
		kmll.Error:    {Bg: termenv.ANSIBlack.Sequence(true), Fg: termenv.ANSIBrightRed.Sequence(false)},
		kmll.Warning:  {Bg: termenv.ANSIBlack.Sequence(true), Fg: termenv.ANSIBrightYellow.Sequence(false)},
		kmll.Notice:   {Bg: termenv.ANSIBlack.Sequence(true), Fg: termenv.ANSIGreen.Sequence(false)},
		kmll.Info:     {Bg: termenv.ANSIBlack.Sequence(true), Fg: termenv.ANSIWhite.Sequence(false)},
		kmll.Debug:    {Bg: termenv.ANSIBlack.Sequence(true), Fg: termenv.ANSICyan.Sequence(false)},
		kmll.Trace:    {Bg: termenv.ANSIBlack.Sequence(true), Fg: termenv.ANSIBrightWhite.Sequence(false)},
	}
	if style, ok = lm[ses.Data.Level]; ok {
		fg, bg = style.Fg, style.Bg
	}

	return
}

// Функция шаблонизатора для управления цветом.
func (ses *session) fnColorSet(dst string, opt string, brc string) (ret string) {
	const tplBadParam = "не верный параметр %q"
	const (
		tagAll, tagReset, tagLevel             = "all", "reset", "level"
		tagText, tagBack                       = "text", "back"
		tagNormal, tagBright                   = "normal", "bright"
		tagBlack, tagRed, tagGreen, tagYellow  = "black", "red", "green", "yellow"
		tagBlue, tagMagenta, tagCyan, tagWhite = "blue", "magenta", "cyan", "white"
	)
	var (
		isBask   bool
		isBright bool
		hexRgb   []string
		styles   []string
		styleFg  string
		styleBg  string
		seq      string
	)

	dst, opt, brc = strings.ToLower(dst), strings.ToLower(opt), strings.ToLower(brc)
	switch dst {
	case tagAll:
		switch opt {
		case tagLevel:
			styleFg, styleBg = ses.colorByLevel()
			styles = append(styles, styleFg)
			styles = append(styles, styleBg)
		case tagReset:
			ret = fmt.Sprintf("%sm", seqCSI+seqResetSeq)
			return
		}
	case tagText:
		isBask = false
	case tagBack:
		isBask = true
	default:
		ret = fmt.Sprintf(tplBadParam, dst)
		return
	}
	switch brc {
	case tagBright:
		isBright = true
	case tagNormal:
		isBright = false
	default:
		isBright = false
	}
	switch opt {
	case tagLevel:
		styleFg, styleBg = ses.colorByLevel()
		switch isBask {
		case true:
			// Цвет фона зависящий от уровня логирования.
			styles = append(styles, styleBg)
		default:
			// Цвет текста зависящий от уровня логирования.
			styles = append(styles, styleFg)
		}
	case tagBlack:
		switch isBright {
		case true:
			styles = append(styles, termenv.ANSIBrightBlack.Sequence(isBask))
		default:
			styles = append(styles, termenv.ANSIBlack.Sequence(isBask))
		}
	case tagRed:
		switch isBright {
		case true:
			styles = append(styles, termenv.ANSIBrightRed.Sequence(isBask))
		default:
			styles = append(styles, termenv.ANSIRed.Sequence(isBask))
		}
	case tagGreen:
		switch isBright {
		case true:
			styles = append(styles, termenv.ANSIBrightGreen.Sequence(isBask))
		default:
			styles = append(styles, termenv.ANSIGreen.Sequence(isBask))
		}
	case tagYellow:
		switch isBright {
		case true:
			styles = append(styles, termenv.ANSIBrightYellow.Sequence(isBask))
		default:
			styles = append(styles, termenv.ANSIYellow.Sequence(isBask))
		}
	case tagBlue:
		switch isBright {
		case true:
			styles = append(styles, termenv.ANSIBrightBlue.Sequence(isBask))
		default:
			styles = append(styles, termenv.ANSIBlue.Sequence(isBask))
		}
	case tagMagenta:
		switch isBright {
		case true:
			styles = append(styles, termenv.ANSIBrightMagenta.Sequence(isBask))
		default:
			styles = append(styles, termenv.ANSIMagenta.Sequence(isBask))
		}
	case tagCyan:
		switch isBright {
		case true:
			styles = append(styles, termenv.ANSIBrightCyan.Sequence(isBask))
		default:
			styles = append(styles, termenv.ANSICyan.Sequence(isBask))
		}
	case tagWhite:
		switch isBright {
		case true:
			styles = append(styles, termenv.ANSIBrightWhite.Sequence(isBask))
		default:
			styles = append(styles, termenv.ANSIWhite.Sequence(isBask))
		}
	default:
		if hexRgb = rexHexColor.FindStringSubmatch(opt); len(hexRgb) != 2 {
			ret = fmt.Sprintf(tplBadParam, opt)
			return
		}
		styles = append(styles, ses.profile.Color(hexRgb[1]).Sequence(isBask))
	}
	seq = strings.Join(styles, ";")
	ret = fmt.Sprintf("%s%sm", seqCSI, seq)

	return
}

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

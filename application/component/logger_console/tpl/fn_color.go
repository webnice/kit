// Package tpl
package tpl

import (
	"fmt"
	"strings"

	kitModuleDye "github.com/webnice/kit/v3/module/dye"
	kmll "github.com/webnice/kit/v3/module/log/level"
)

func (ses *session) colorByLevel() (fg kitModuleDye.Interface, bg kitModuleDye.Interface) {
	var (
		lm    map[kmll.Level]colorAnsiStyle
		style colorAnsiStyle
		ok    bool
	)

	lm = map[kmll.Level]colorAnsiStyle{
		kmll.Fatal:    {Bg: kitModuleDye.New().Background().Red(), Fg: kitModuleDye.New().Bright().Yellow()},
		kmll.Alert:    {Bg: kitModuleDye.New().Background().Magenta(), Fg: kitModuleDye.New().Bright().White()},
		kmll.Critical: {Bg: kitModuleDye.New().Background().Blue(), Fg: kitModuleDye.New().Bright().Magenta()},
		kmll.Error:    {Bg: kitModuleDye.New().Background().Black(), Fg: kitModuleDye.New().Bright().Red()},
		kmll.Warning:  {Bg: kitModuleDye.New().Background().Black(), Fg: kitModuleDye.New().Bright().Yellow()},
		kmll.Notice:   {Bg: kitModuleDye.New().Background().Black(), Fg: kitModuleDye.New().Green()},
		kmll.Info:     {Bg: kitModuleDye.New().Background().Black(), Fg: kitModuleDye.New().White()},
		kmll.Debug:    {Bg: kitModuleDye.New().Background().Black(), Fg: kitModuleDye.New().Cyan()},
		kmll.Trace:    {Bg: kitModuleDye.New().Background().Black(), Fg: kitModuleDye.New().Bright().White()},
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
		tagSet, tagReset                       = "set", "reset"
		tagAll, tagLevel                       = "all", "level"
		tagText, tagBack                       = "text", "back"
		tagNormal, tagBright                   = "normal", "bright"
		tagBlack, tagRed, tagGreen, tagYellow  = "black", "red", "green", "yellow"
		tagBlue, tagMagenta, tagCyan, tagWhite = "blue", "magenta", "cyan", "white"
		tagBold, tagFaded, tagItalic           = "bold", "faded", "italic"
		tagUnderline, tagReverse, tagCrossout  = "underline", "reverse", "crossout"
	)
	var (
		err                   error
		seq, styleFg, styleBg kitModuleDye.Interface
		isBask                bool
	)

	seq, dst, opt, brc = kitModuleDye.New(), strings.ToLower(dst), strings.ToLower(opt), strings.ToLower(brc)
	switch dst {
	case tagAll:
		switch opt {
		case tagLevel:
			styleFg, styleBg = ses.colorByLevel()
			seq.Add(styleFg.Source()...).Add(styleBg.Source()...)
		}
	case tagText:
		seq.Foreground()
	case tagBack:
		isBask = true
		seq.Background()
	case tagSet:
		switch opt {
		case tagBold:
			ret = seq.Bold().Done().String()
		case tagFaded:
			ret = seq.Faded().Done().String()
		case tagItalic:
			ret = seq.Italic().Done().String()
		case tagUnderline:
			ret = seq.Underline().Done().String()
		case tagReverse:
			ret = seq.Reverse().Done().String()
		case tagCrossout:
			ret = seq.CrossOut().Done().String()
		default:
			ret = fmt.Sprintf(tplBadParam, dst)
		}
		return
	case tagReset:
		switch opt {
		case tagAll:
			ret = seq.Reset().Done().String()
		case tagBold:
			ret = seq.ResetBold().Done().String()
		case tagFaded:
			ret = seq.ResetFaded().Done().String()
		case tagItalic:
			ret = seq.ResetItalic().Done().String()
		case tagUnderline:
			ret = seq.ResetUnderline().Done().String()
		case tagReverse:
			ret = seq.ResetReverse().Done().String()
		case tagCrossout:
			ret = seq.ResetCrossOut().Done().String()
		default:
			ret = fmt.Sprintf(tplBadParam, dst)
		}
		return
	default:
		ret = fmt.Sprintf(tplBadParam, dst)
		return
	}
	switch brc {
	case tagBright:
		seq.Bright()
	case tagNormal:
		seq.Normal()
	default:
		seq.Normal()
	}
	switch opt {
	case tagLevel:
		styleFg, styleBg = ses.colorByLevel()
		switch isBask {
		case true:
			seq.Add(styleBg.Source()...) // Цвет фона зависящий от уровня логирования.
		default:
			seq.Add(styleFg.Source()...) // Цвет текста зависящий от уровня логирования.
		}
	case tagBlack:
		seq.Black()
	case tagRed:
		seq.Red()
	case tagGreen:
		seq.Green()
	case tagYellow:
		seq.Yellow()
	case tagBlue:
		seq.Blue()
	case tagMagenta:
		seq.Magenta()
	case tagCyan:
		seq.Cyan()
	case tagWhite:
		seq.White()
	default:
		if err = seq.HexRgb(opt); err != nil {
			ret = err.Error()
			return
		}
	}
	ret = seq.Done().String()

	return
}

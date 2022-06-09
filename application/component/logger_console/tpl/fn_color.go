// Package tpl
package tpl

import (
	"fmt"
	"strings"

	"github.com/webnice/kit/application/component/logger_console/dye"

	kmll "github.com/webnice/kit/module/log/level"
)

func (ses *session) colorByLevel() (fg dye.Interface, bg dye.Interface) {
	type ansiStyle = struct {
		Bg dye.Interface
		Fg dye.Interface
	}
	var (
		lm    map[kmll.Level]ansiStyle
		style ansiStyle
		ok    bool
	)

	lm = map[kmll.Level]ansiStyle{
		kmll.Fatal:    {Bg: dye.New().Background().Red(), Fg: dye.New().Bright().Yellow()},
		kmll.Alert:    {Bg: dye.New().Background().Magenta(), Fg: dye.New().Bright().White()},
		kmll.Critical: {Bg: dye.New().Background().Blue(), Fg: dye.New().Bright().Magenta()},
		kmll.Error:    {Bg: dye.New().Background().Black(), Fg: dye.New().Bright().Red()},
		kmll.Warning:  {Bg: dye.New().Background().Black(), Fg: dye.New().Bright().Yellow()},
		kmll.Notice:   {Bg: dye.New().Background().Black(), Fg: dye.New().Green()},
		kmll.Info:     {Bg: dye.New().Background().Black(), Fg: dye.New().White()},
		kmll.Debug:    {Bg: dye.New().Background().Black(), Fg: dye.New().Cyan()},
		kmll.Trace:    {Bg: dye.New().Background().Black(), Fg: dye.New().Bright().White()},
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
		err     error
		styleFg dye.Interface
		styleBg dye.Interface
		seq     dye.Interface
		isBask  bool
	)

	dst, opt, brc = strings.ToLower(dst), strings.ToLower(opt), strings.ToLower(brc)
	seq = dye.New()
	switch dst {
	case tagAll:
		switch opt {
		case tagLevel:
			styleFg, styleBg = ses.colorByLevel()
			seq.Add(styleFg.Source()...).Add(styleBg.Source()...)
		case tagReset:
			ret = seq.Reset().Done().String()
			return
		}
	case tagText:
		seq.Foreground()
	case tagBack:
		isBask = true
		seq.Background()
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

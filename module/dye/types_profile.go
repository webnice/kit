package dye

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/webnice/kit/v4/module/dye/colorful"
)

// Профиль терминала.
type terminalProfile int

// Convert Конвертация цвета в цвет, поддерживаемый профилем терминала.
func (tp terminalProfile) Convert(c Color) (ret Color) {
	var (
		h   colorful.Color
		err error
		ac  colorAnsi256
	)

	if tp == terminalAscii {
		return colorNone{}
	}
	switch v := c.(type) {
	case colorAnsi:
		return v
	case colorAnsi256:
		if tp == terminalANSI {
			return ansi256ToANSIColor(v)
		}

		return v
	case colorRgb:
		if h, err = colorful.Hex(string(v)); err != nil {
			return nil
		}
		if tp < terminalTrueColor {
			ac = hexToANSI256Color(h)
			if tp == terminalANSI {
				return ansi256ToANSIColor(ac)
			}

			return ac
		}

		return v
	}

	return c
}

// Color Создание цвета из строки с учётом профиля терминала.
func (tp terminalProfile) Color(s string) (ret Color) {
	var (
		c   Color
		i   int
		err error
	)

	if len(s) == 0 {
		return nil
	}
	if strings.HasPrefix(s, keyHash) {
		c = colorRgb(s)
	} else {
		if i, err = strconv.Atoi(s); err != nil {
			return nil
		}
		if i < 16 {
			c = colorAnsi(i)
		} else {
			c = colorAnsi256(i)
		}
	}

	return tp.Convert(c)
}

// FromColor Конвертация цвета стандартной библиотеки в цвет, поддерживаемый профилем терминала.
func (tp terminalProfile) FromColor(c color.Color) (ret Color) {
	var (
		col colorful.Color
		ok  bool
	)

	if col, ok = colorful.MakeColor(c); !ok {
		return nil
	}

	return tp.Color(col.Hex())
}

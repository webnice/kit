package dye

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/webnice/kit/v5/module/dye/colorful"
	"github.com/webnice/kit/v5/module/dye/isatty"
)

// Проверка терминала по файловому дескриптору.
func isTTY(fd uintptr) bool {
	if len(os.Getenv(envCi)) > 0 {
		return false
	}

	return isatty.IsTerminal(fd)
}

// Определение поддерживаемого цвета терминала.
func getColorProfile() terminalProfile {
	return getColorProfileFd(os.Stdout.Fd())
}

// Определение поддерживаемого цвета терминала по файловому дескриптору.
func getColorProfileFd(fd uintptr) terminalProfile {
	if !isTTY(fd) {
		return terminalAscii
	}

	return colorProfile()
}

// Проверка принудительного включения цвета через переменную окружения.
func cliColorForced() bool {
	var forced string

	if forced = os.Getenv(envCliColorForce); forced != "" {
		return forced != "0"
	}

	return false
}

// Разбор ответа терминала на запрос цвета в формате RGB.
func xTermColor(s string) (ret colorRgb, err error) {
	const tplHexColor = "#%s%s%s"
	var (
		prefix string
		h      []string
	)

	if len(s) < 24 || len(s) > 25 {
		return colorRgb(""), errInvalidColor
	}
	switch {
	case strings.HasSuffix(s, "\a"):
		s = strings.TrimSuffix(s, "\a")
	case strings.HasSuffix(s, "\033"):
		s = strings.TrimSuffix(s, "\033")
	case strings.HasSuffix(s, "\033\\"):
		s = strings.TrimSuffix(s, "\033\\")
	default:
		return colorRgb(""), errInvalidColor
	}
	s = s[4:]
	prefix = prefixRgb
	if !strings.HasPrefix(s, prefix) {
		return colorRgb(""), errInvalidColor
	}
	s = strings.TrimPrefix(s, prefix)
	h = strings.Split(s, "/")
	ret = colorRgb(fmt.Sprintf(tplHexColor, h[0][:2], h[1][:2], h[2][:2]))

	return
}

// Конвертация цвета ANSI256 в ближайший цвет ANSI.
func ansi256ToANSIColor(c colorAnsi256) (ret colorAnsi) {
	var (
		r   int
		md  float64
		h   colorful.Color
		hb  colorful.Color
		d   float64
		err error
		i   int
	)

	md = math.MaxFloat64
	h, err = colorful.Hex(colorAnsiHex[c])
	if err != nil {
		return 0
	}
	for i = 0; i <= 15; i++ {
		if hb, err = colorful.Hex(colorAnsiHex[i]); err != nil {
			continue
		}
		d = h.DistanceHSLuv(hb)
		if d < md {
			md = d
			r = i
		}
	}
	ret = colorAnsi(r)

	return
}

// Конвертация HEX цвета в ближайший цвет ANSI256.
func hexToANSI256Color(c colorful.Color) (ret colorAnsi256) {
	const (
		colorShift      = 16
		grayShift       = 232
		grayLimit       = 238
		grayBase        = 8
		grayStep        = 10
		grayOffset      = 3
		colorCubeSize   = 6
		colorMultiplier = 36
	)
	// Вычисление ближайшего индекса цвета в диапазоне 0..5 по компоненте цвета.
	v2ci := func(v float64) int {
		if v < 48 {
			return 0
		}
		if v < 115 {
			return 1
		}

		return int((v - 35) / 40)
	}
	var (
		r, g, b    int
		ci         int
		i2cv       [6]int
		cr, cg, cb int
		grayIdx    int
		average    int
		gv         int
		c2         colorful.Color
		g2         colorful.Color
		colorDist  float64
		grayDist   float64
	)

	// Вычисление ближайшего индекса цвета с нуля в диапазоне 16..231.
	r = v2ci(c.R * 255.0) // 0..5 каждый компонент.
	g = v2ci(c.G * 255.0)
	b = v2ci(c.B * 255.0)
	ci = colorMultiplier*r + colorCubeSize*g + b /* 0..215 */
	// Вычисление представленных цветов обратно из индекса.
	i2cv = [6]int{0, 0x5f, 0x87, 0xaf, 0xd7, 0xff}
	cr = i2cv[r] // r/g/b, 0..255 каждый компонент.
	cg = i2cv[g]
	cb = i2cv[b]
	// Вычисление ближайшего индекса серого цвета с нуля в диапазоне 232..255.
	average = (r + g + b) / 3
	if average > grayLimit {
		grayIdx = 23
	} else {
		grayIdx = (average - grayOffset) / grayStep // 0..23
	}
	gv = grayBase + grayStep*grayIdx // одинаковое значение для r/g/b, 0..255.
	// Возврат того, который ближе к исходному RGB значению.
	c2 = colorful.Color{R: float64(cr) / 255.0, G: float64(cg) / 255.0, B: float64(cb) / 255.0}
	g2 = colorful.Color{R: float64(gv) / 255.0, G: float64(gv) / 255.0, B: float64(gv) / 255.0}
	colorDist = c.DistanceHSLuv(c2)
	grayDist = c.DistanceHSLuv(g2)
	if colorDist <= grayDist {
		return colorAnsi256(colorShift + ci)
	}

	return colorAnsi256(grayShift + grayIdx)
}

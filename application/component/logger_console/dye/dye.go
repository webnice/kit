// Package dye
package dye

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
)

// New Конструктор объекта пакета, возвращается интерфейс пакета.
// Если переданы последовательности, тогда они добавляются "до".
func New(seq ...string) Interface {
	var dye = &impl{profile: termenv.ColorProfile()}
	dye.Add(seq...)

	return dye
}

// Add Добавление последовательности в конец текущей последовательности.
func (dye *impl) Add(seq ...string) Interface {
	dye.seq = append(dye.seq, seq...)

	return dye
}

// Source Возвращает текущую последовательность в исходном виде как срез строк.
func (dye *impl) Source() []string { return dye.seq }

// Done Завершение и возврат последовательности.
func (dye *impl) Done() Sequence {
	var ret string

	ret = strings.Join(dye.seq, ";")
	ret = fmt.Sprintf("%s%sm", seqCSI, ret)

	return sequence(ret)
}

// Background Флаг, говорящий что следующий цвет будет цветом фона, только для ANSI цветов.
func (dye *impl) Background() Interface { dye.isBackground = true; return dye }

// Foreground Флаг, говорящий что следующий цвет будет цветом текста, только для ANSI цветов, по умолчанию.
func (dye *impl) Foreground() Interface { dye.isBackground = false; return dye }

// Bright Флаг, указывающий что следующий цвет будет ярким, только для ANSI цветов.
func (dye *impl) Bright() Interface { dye.isBright = true; return dye }

// Normal Флаг, указывающий что следующий цвет будет нормальной яркости, только для ANSI цветов.
func (dye *impl) Normal() Interface { dye.isBright = false; return dye }

// Reset Добавление в последовательность "Сброс", переход в нормальный режим.
func (dye *impl) Reset() Interface { dye.seq = append(dye.seq, seqReset); return dye }

// Black Добавление чёрного ANSI цвета в последовательность.
func (dye *impl) Black() Interface {
	switch dye.isBright {
	case true:
		dye.seq = append(dye.seq, termenv.ANSIBrightBlack.Sequence(dye.isBackground))
	default:
		dye.seq = append(dye.seq, termenv.ANSIBlack.Sequence(dye.isBackground))
	}
	dye.isBright, dye.isBackground = false, false

	return dye
}

// Red Добавление красного ANSI цвета в последовательность.
func (dye *impl) Red() Interface {
	switch dye.isBright {
	case true:
		dye.seq = append(dye.seq, termenv.ANSIBrightRed.Sequence(dye.isBackground))
	default:
		dye.seq = append(dye.seq, termenv.ANSIRed.Sequence(dye.isBackground))
	}
	dye.isBright, dye.isBackground = false, false

	return dye
}

// Green Добавление зелёного ANSI цвета в последовательность.
func (dye *impl) Green() Interface {
	switch dye.isBright {
	case true:
		dye.seq = append(dye.seq, termenv.ANSIBrightGreen.Sequence(dye.isBackground))
	default:
		dye.seq = append(dye.seq, termenv.ANSIGreen.Sequence(dye.isBackground))
	}
	dye.isBright, dye.isBackground = false, false

	return dye
}

// Yellow Добавление жёлтого ANSI цвета в последовательность.
func (dye *impl) Yellow() Interface {
	switch dye.isBright {
	case true:
		dye.seq = append(dye.seq, termenv.ANSIBrightYellow.Sequence(dye.isBackground))
	default:
		dye.seq = append(dye.seq, termenv.ANSIYellow.Sequence(dye.isBackground))
	}
	dye.isBright, dye.isBackground = false, false

	return dye
}

// Blue Добавление синего ANSI цвета в последовательность.
func (dye *impl) Blue() Interface {
	switch dye.isBright {
	case true:
		dye.seq = append(dye.seq, termenv.ANSIBrightBlue.Sequence(dye.isBackground))
	default:
		dye.seq = append(dye.seq, termenv.ANSIBlue.Sequence(dye.isBackground))
	}
	dye.isBright, dye.isBackground = false, false

	return dye
}

// Magenta Добавление пурпурного ANSI цвета в последовательность.
func (dye *impl) Magenta() Interface {
	switch dye.isBright {
	case true:
		dye.seq = append(dye.seq, termenv.ANSIBrightMagenta.Sequence(dye.isBackground))
	default:
		dye.seq = append(dye.seq, termenv.ANSIMagenta.Sequence(dye.isBackground))
	}
	dye.isBright, dye.isBackground = false, false

	return dye
}

// Cyan Добавление бирюзового ANSI цвета в последовательность.
func (dye *impl) Cyan() Interface {
	switch dye.isBright {
	case true:
		dye.seq = append(dye.seq, termenv.ANSIBrightCyan.Sequence(dye.isBackground))
	default:
		dye.seq = append(dye.seq, termenv.ANSICyan.Sequence(dye.isBackground))
	}
	dye.isBright, dye.isBackground = false, false

	return dye
}

// White Добавление белого ANSI цвета в последовательность.
func (dye *impl) White() Interface {
	switch dye.isBright {
	case true:
		dye.seq = append(dye.seq, termenv.ANSIBrightWhite.Sequence(dye.isBackground))
	default:
		dye.seq = append(dye.seq, termenv.ANSIWhite.Sequence(dye.isBackground))
	}
	dye.isBright, dye.isBackground = false, false

	return dye
}

// HexRgb Конвертация HEX RGB цвета в формате #000000 в ANSI цвет и добавление в последовательность.
func (dye *impl) HexRgb(hexRgb string) (err error) {
	var arr []string

	if arr = rexHexColor.FindStringSubmatch(hexRgb); len(arr) != 2 {
		err = fmt.Errorf(tplBadParam, hexRgb)
		return
	}
	dye.seq = append(dye.seq, dye.profile.Color(arr[1]).Sequence(dye.isBackground))

	return
}

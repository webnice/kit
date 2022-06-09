// Package tpl
package tpl

import (
	"bytes"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Разделение строки с использованием разделителя и учётом возможности экранирования разделителя.
func splitBySeparator(s string, separator string) (ret []string) {
	var (
		sep string
		buf []byte
		arr [][]byte
		n   int
	)

	sep = keyQuote + separator
	buf = bytes.Replace([]byte(s), []byte(sep), []byte{0}, -1)
	arr = bytes.SplitN(buf, []byte(separator), 3)
	ret = make([]string, 0, len(arr))
	for n = range arr {
		buf = bytes.Replace(arr[n], []byte{0}, []byte(separator), -1)
		ret = append(ret, string(buf))
	}

	return
}

func fixSlice(a []string, c int, separator string) (ret []string) {
	var n int

	ret = make([]string, c)
	for n = range a {
		if len(ret)-1 == n && len(a) >= len(ret) {
			ret[n] = strings.Join(a[n:], separator)
		} else if len(ret) > n {
			ret[n] = a[n]
		}
	}

	return
}

func stringToRune(s string) (ret []rune) {
	var (
		r    rune
		size int
	)

	for len(s) > 0 {
		r, size = utf8.DecodeRuneInString(s)
		ret = append(ret, r)
		s = s[size:]
	}

	return
}

func stringToInt(s string) (ret int) {
	var tmp uint64

	if tmp, _ = strconv.ParseUint(s, 10, 64); tmp > 0 {
		ret = int(tmp)
	}

	return
}

func runeMax(msg []rune, max int, suffix string) (ret []rune) {
	ret = make([]rune, 0, max)
	switch {
	case max > 0 && len(msg) > max:
		if max > len(suffix) {
			max -= len(suffix)
		}
		ret = append(msg[:max], stringToRune(suffix)...)
	default:
		ret = append(ret, msg...)
	}

	return
}

func runeMin(msg []rune, min int) (ret []rune) {
	const space = rune(' ')
	var n int

	ret = append(ret, msg...)
	if min > 0 && len(msg) < min {
		for n = min - len(msg); n > 0; n-- {
			ret = append(ret, space)
		}
	}

	return
}

func runeToString(msg []rune) (ret string) {
	var (
		n   int
		buf strings.Builder
	)

	for n = range msg {
		buf.WriteRune(msg[n])
	}
	ret = buf.String()

	return
}

package fmt

import "strings"

// stripNumbers оставляет в строке только цифры.
func stripNumbers(s string) string {
	var (
		result strings.Builder
		n      int
		bt     uint8
	)

	for n = 0; n < len(s); n++ {
		bt = s[n]
		if '0' <= bt && bt <= '9' {
			result.WriteByte(bt)
		}
	}

	return result.String()
}

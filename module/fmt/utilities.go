package fmt

import "strings"

// Удаление из строки всех символов кроме числовых.
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

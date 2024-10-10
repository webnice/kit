package fmt

import "fmt"

// PhoneNumberFormat Форматирование номера телефона, все форматы.
func PhoneNumberFormat(src string) (ret string) {
	const (
		fmt23  = "%s-%s"
		fmt45  = "%s-%s-%s"
		fmt6   = "%s-%s-%s"
		fmt7   = "%s-%s-%s"
		fmt8   = "(%s) %s-%s"
		fmt910 = "(%s) %s-%s-%s"
		fmt11  = "+%s (%s) %s-%s-%s"
		fmt12  = "+%s (%s) %s-%s-%s"
		fmt13  = "+%s (%s) %s-%s-%s"
		fmt14  = "+%s (%s) %s-%s-%s"
		fmtDt  = "+%s (%s) %s-%s-%s-%s"
	)

	src = stripNumbers(src)
	switch len(src) {
	case 0, 1: // 1
		ret = src
	case 2, 3: // 7-1, 7-12
		ret = fmt.Sprintf(fmt23, src[0:1], src[1:])
	case 4, 5: // 7-12-3, 7-12-34
		ret = fmt.Sprintf(fmt45, src[0:1], src[1:3], src[3:])
	case 6: // 71-23-45
		ret = fmt.Sprintf(fmt6, src[0:2], src[2:4], src[4:])
	case 7: // 712-34-56
		ret = fmt.Sprintf(fmt7, src[0:3], src[3:5], src[5:])
	case 8: // (712) 345-67
		ret = fmt.Sprintf(fmt8, src[0:3], src[3:6], src[6:])
	case 9, 10: // (712) 345-67-8, (712) 345-67-89
		ret = fmt.Sprintf(fmt910, src[0:3], src[3:6], src[6:8], src[8:])
	case 11: // +7 (123) 456-78-90
		ret = fmt.Sprintf(fmt11, src[0:1], src[1:4], src[4:7], src[7:9], src[9:])
	case 12: // +71 (234) 567-89-01
		ret = fmt.Sprintf(fmt12, src[0:2], src[2:5], src[5:8], src[8:10], src[10:])
	case 13: // +712 (345) 678-90-12
		ret = fmt.Sprintf(fmt13, src[0:3], src[3:6], src[6:9], src[9:11], src[11:])
	case 14: // +7123 (456) 789-01-23
		ret = fmt.Sprintf(fmt14, src[0:4], src[4:7], src[7:10], src[10:12], src[12:])
	default: // +7123 (456) 789-01-23-4...
		ret = fmt.Sprintf(fmtDt, src[0:4], src[4:7], src[7:10], src[10:12], src[12:14], src[14:])
	}

	return
}

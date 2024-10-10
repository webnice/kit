package fmt

import "testing"

func TestPhoneNumberFormat(t *testing.T) {
	var result string

	tests := map[string]struct {
		Source   string
		Expected string
	}{
		"01": {Source: "", Expected: ""},
		"02": {Source: "7", Expected: "7"},
		"03": {Source: "71", Expected: "7-1"},
		"04": {Source: "712", Expected: "7-12"},
		"05": {Source: "7123", Expected: "7-12-3"},
		"06": {Source: "71234", Expected: "7-12-34"},
		"07": {Source: "712345", Expected: "71-23-45"},
		"08": {Source: "7123456", Expected: "712-34-56"},
		"09": {Source: "71234567", Expected: "(712) 345-67"},
		"10": {Source: "712345678", Expected: "(712) 345-67-8"},
		"11": {Source: "7123456789", Expected: "(712) 345-67-89"},
		"12": {Source: "71234567890", Expected: "+7 (123) 456-78-90"},
		"13": {Source: "712345678901", Expected: "+71 (234) 567-89-01"},
		"14": {Source: "7123456789012", Expected: "+712 (345) 678-90-12"},
		"15": {Source: "71234567890123", Expected: "+7123 (456) 789-01-23"},
		"16": {Source: "712345678901234", Expected: "+7123 (456) 789-01-23-4"},
		"17": {Source: "7123456789012345", Expected: "+7123 (456) 789-01-23-45"},
		"18": {Source: "71234567890123456", Expected: "+7123 (456) 789-01-23-456"},
		"19": {Source: "712345678901234567", Expected: "+7123 (456) 789-01-23-4567"},
		"20": {Source: "7123456789012345678", Expected: "+7123 (456) 789-01-23-45678"},
		"21": {Source: "71234567890123456789", Expected: "+7123 (456) 789-01-23-456789"},
	}
	for key := range tests {
		result = PhoneNumberFormat(tests[key].Source)
		if result != tests[key].Expected {
			t.Log(tests[key].Expected)
			t.Log(result)
			t.Fatalf("Tест %q провален.", key)
		}
	}
}

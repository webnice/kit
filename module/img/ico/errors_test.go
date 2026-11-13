package ico

import "testing"

// TestFormatErrorMessage Сообщение ошибки формата ICO.
func TestFormatErrorMessage(t *testing.T) {
	if got := FormatError("test").Error(); got != "invalid ICO format: test" {
		t.Fatalf("FormatError.Error() = %q", got)
	}
}

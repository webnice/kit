package ico

import "testing"

// TestFormatErrorMessage Сообщение ошибки формата ICO.
func TestFormatErrorMessage(t *testing.T) {
	const (
		tplErrorMismatch = "FormatError.Error() = %q."
		errExpected      = "invalid ICO format: test"
	)

	if got := FormatError("test").Error(); got != errExpected {
		t.Fatalf(tplErrorMismatch, got)
	}
}

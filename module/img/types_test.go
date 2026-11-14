package img

import "testing"

// TestTypeString Строковое представление типа изображения.
func TestTypeString(t *testing.T) {
	const errStringWrong = "Type.String() некорректен."

	if Type("abc").String() != "abc" {
		t.Fatal(errStringWrong)
	}
}

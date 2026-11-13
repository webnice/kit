package img

import "testing"

// TestTypeString Строковое представление типа изображения.
func TestTypeString(t *testing.T) {
	if Type("abc").String() != "abc" {
		t.Fatal("Type.String() некорректен")
	}
}

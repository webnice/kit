package uuid

import "testing"

// TestTypeString Строковые представления типов версии и пространства имён.
func TestTypeString(t *testing.T) {
	if Namespace.DNS.String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Fatalf("Namespace.DNS.String() = %q", Namespace.DNS.String())
	}
	if Version.V7.String() != "V7" || Version.V8.String() != "V8" {
		t.Fatal("String() версии UUID некорректен")
	}
}

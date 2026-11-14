package cli

import "testing"

// TestConstantEnvironmentNameMustFindByAnchor Поиск окружения константы по якорю.
func TestConstantEnvironmentNameMustFindByAnchor(t *testing.T) {
	const (
		errAnchorNotFound      = "MustFindByAnchor() не нашёл существующий якорь."
		errMissingNotEmpty     = "MustFindByAnchor() для отсутствующего якоря вернул не пустой объект."
		errDestinationNotFound = "Destination() не нашёл существующий якорь."
	)
	var env ConstantEnvironmentName

	env = ConstantEnvironmentName{{Anchor: "one", Destination: "FIRST"}}
	if env.MustFindByAnchor("one").Destination != "FIRST" {
		t.Fatal(errAnchorNotFound)
	}
	if env.MustFindByAnchor("missing").Anchor != "" {
		t.Fatal(errMissingNotEmpty)
	}
	if env.Destination("one") != "FIRST" {
		t.Fatal(errDestinationNotFound)
	}
}

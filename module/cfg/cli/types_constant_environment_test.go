package cli

import "testing"

// TestConstantEnvironmentNameMustFindByAnchor Поиск окружения константы по якорю.
func TestConstantEnvironmentNameMustFindByAnchor(t *testing.T) {
	var env = ConstantEnvironmentName{{Anchor: "one", Destination: "FIRST"}}
	if env.MustFindByAnchor("one").Destination != "FIRST" {
		t.Fatal("MustFindByAnchor() не нашёл существующий якорь")
	}
	if env.MustFindByAnchor("missing").Anchor != "" {
		t.Fatal("MustFindByAnchor() для отсутствующего якоря вернул не пустой объект")
	}
	if env.Destination("one") != "FIRST" {
		t.Fatal("Destination() не нашёл существующий якорь")
	}
}

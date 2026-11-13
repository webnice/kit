package ans

import "testing"

// TestImpl_VerifyExtended Верификация объекта по тегам валидации.
func TestImpl_VerifyExtended(t *testing.T) {
	var ans = New(nil)

	type valid struct {
		Name string `validate:"required"`
	}
	if reo := ans.(*impl).Verify(&valid{Name: "ok"}); reo != nil {
		t.Fatalf("Verify() для валидного объекта вернул %+v", reo)
	}

	type invalid struct {
		Name string `validate:"required"`
	}
	if reo := ans.(*impl).Verify(&invalid{}); reo == nil || len(reo.Field()) != 1 || reo.Field()[0].Field != "Name" {
		t.Fatalf("Verify() для невалидного объекта вернул %+v", reo)
	}

	var slice = []invalid{{Name: "ok"}, {}}
	if reo := ans.(*impl).Verify(&slice); reo == nil || len(reo.Field()) != 1 {
		t.Fatalf("Verify() для невалидного среза вернул %+v", reo)
	}

	if reo := ans.(*impl).Verify(1); reo == nil || reo.AsError() == nil {
		t.Fatal("Verify() для неструктурного значения не вернул ошибку")
	}

	type invalidTag struct {
		Name string `validate:"unknown-tag"`
	}
	if reo := ans.(*impl).Verify(&invalidTag{Name: "ok"}); reo == nil || reo.AsError() == nil {
		t.Fatal("Verify() с невалидным тегом валидации не вернул ошибку")
	}

	if reo := ans.(*impl).Verify(nil); reo == nil || reo.AsError() == nil {
		t.Fatal("Verify(nil) не вернул ошибку")
	}
}

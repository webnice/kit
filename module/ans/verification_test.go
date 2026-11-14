package ans

import "testing"

// TestImpl_VerifyExtended Верификация объекта по тегам валидации.
func TestImpl_VerifyExtended(t *testing.T) {
	const (
		tplValidObject   = "Verify() для валидного объекта вернул %+v."
		tplInvalidObject = "Verify() для невалидного объекта вернул %+v."
		tplInvalidSlice  = "Verify() для невалидного среза вернул %+v."
		errNoStructNoErr = "Verify() для неструктурного значения не вернул ошибку."
		errBadTagNoErr   = "Verify() с невалидным тегом валидации не вернул ошибку."
		errNilNoErr      = "Verify(nil) не вернул ошибку."
	)
	type valid struct {
		Name string `validate:"required"`
	}
	type invalid struct {
		Name string `validate:"required"`
	}
	type invalidTag struct {
		Name string `validate:"unknown-tag"`
	}
	var (
		ans   Interface
		slice []invalid
		reo   RestErrorInterface
	)

	ans = New(nil)
	if reo = ans.(*impl).Verify(&valid{Name: "ok"}); reo != nil {
		t.Fatalf(tplValidObject, reo)
	}
	if reo = ans.(*impl).Verify(&invalid{}); reo == nil || len(reo.Field()) != 1 || reo.Field()[0].Field != "Name" {
		t.Fatalf(tplInvalidObject, reo)
	}
	slice = []invalid{{Name: "ok"}, {}}
	if reo = ans.(*impl).Verify(&slice); reo == nil || len(reo.Field()) != 1 {
		t.Fatalf(tplInvalidSlice, reo)
	}
	if reo = ans.(*impl).Verify(1); reo == nil || reo.AsError() == nil {
		t.Fatal(errNoStructNoErr)
	}
	if reo = ans.(*impl).Verify(&invalidTag{Name: "ok"}); reo == nil || reo.AsError() == nil {
		t.Fatal(errBadTagNoErr)
	}
	if reo = ans.(*impl).Verify(nil); reo == nil || reo.AsError() == nil {
		t.Fatal(errNilNoErr)
	}
}

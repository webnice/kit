package bus

import (
	"reflect"
	"testing"
)

// TestSubscribers Реестр подписчиков.
func TestSubscribers(t *testing.T) {
	const (
		errNotEmpty           = "newSubscribers() вернул не пустой объект."
		badElement            = "bad element"
		tplLenAfterPush       = "Len() после добавления служебного элемента = %d, ожидался 1."
		tplLenAfterStore      = "Len() = %d, ожидался 2."
		errTypeNotFound       = "IsExistSubscriber() не нашёл подписчика типа."
		errOtherTypeFound     = "IsExistSubscriber() нашёл подписчика для незарегистрированного типа."
		tplGetTyped           = "GetSubscriber(dataType) = %+v."
		tplGetOther           = "GetSubscriber(otherType) = %+v."
		errAnyNotCounted      = "IsExistSubscriber() не учёл any-подписчика."
		tplGetTypedAfterAny   = "GetSubscriber(dataType) после any = %+v."
		tplGetOtherAfterAny   = "GetSubscriber(otherType) после any = %+v."
		errAliasSlice         = "GetSubscriber() вернул срез-алиас внутренней карты."
		errDeleteUnknownNoErr = "Delete() для отсутствующего подписчика вернул nil ошибку."
		tplDeleteError        = "Delete() вернул ошибку: %v."
		tplGetAfterDelete     = "GetSubscriber(dataType) после удаления typed = %+v."
		tplDeleteAnyError     = "Delete() any вернул ошибку: %v."
		errFoundDeleted       = "IsExistSubscriber() нашёл удалённых подписчиков."
		errFoundDeletedAny    = "IsExistSubscriber() нашёл удалённого any-подписчика."
	)
	var (
		sss           *subscribers
		dataType      reflect.Type
		otherType     reflect.Type
		typed         *subscriber
		anySubscriber *subscriber
		got           []*subscriber
		err           error
	)

	sss = newSubscribers()
	dataType = reflect.TypeOf(busTestData{})
	otherType = reflect.TypeOf(busTestOtherData{})
	typed = &subscriber{
		Name: "typed",
		Item: &testSubscriber{},
		Types: []*subscriberTypeInfo{{
			Original:     busTestData{},
			OriginalType: dataType,
		}},
	}
	anySubscriber = &subscriber{Name: "any", Item: &testAnySubscriber{}}
	if sss.Len() != 0 {
		t.Fatal(errNotEmpty)
	}
	sss.items.PushBack(badElement)
	if sss.Len() != 1 {
		t.Fatalf(tplLenAfterPush, sss.Len())
	}
	sss.Store(typed)
	if sss.Len() != 2 {
		t.Fatalf(tplLenAfterStore, sss.Len())
	}
	if !sss.IsExistSubscriber(dataType) {
		t.Fatal(errTypeNotFound)
	}
	if sss.IsExistSubscriber(otherType) {
		t.Fatal(errOtherTypeFound)
	}
	got = sss.GetSubscriber(dataType)
	if len(got) != 1 || got[0] != typed {
		t.Fatalf(tplGetTyped, got)
	}
	got = sss.GetSubscriber(otherType)
	if len(got) != 0 {
		t.Fatalf(tplGetOther, got)
	}
	sss.Store(anySubscriber)
	if !sss.IsExistSubscriber(otherType) {
		t.Fatal(errAnyNotCounted)
	}
	got = sss.GetSubscriber(dataType)
	if len(got) != 2 {
		t.Fatalf(tplGetTypedAfterAny, got)
	}
	got = sss.GetSubscriber(otherType)
	if len(got) != 1 || got[0] != anySubscriber {
		t.Fatalf(tplGetOtherAfterAny, got)
	}
	got = sss.GetSubscriber(dataType)
	got = append(got, anySubscriber)
	if len(sss.GetSubscriber(dataType)) != 2 {
		t.Fatal(errAliasSlice)
	}
	if err = sss.Delete("unknown"); err == nil {
		t.Fatal(errDeleteUnknownNoErr)
	}
	if err = sss.Delete("TYPED"); err != nil {
		t.Fatalf(tplDeleteError, err)
	}
	got = sss.GetSubscriber(dataType)
	if len(got) != 1 || got[0] != anySubscriber {
		t.Fatalf(tplGetAfterDelete, got)
	}
	if err = sss.Delete("any"); err != nil {
		t.Fatalf(tplDeleteAnyError, err)
	}
	if sss.IsExistSubscriber(dataType) {
		t.Fatal(errFoundDeleted)
	}
	if sss.IsExistSubscriber(otherType) {
		t.Fatal(errFoundDeletedAny)
	}
}

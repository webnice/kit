package bus

import (
	"reflect"
	"testing"
)

// TestSubscribers Реестр подписчиков.
func TestSubscribers(t *testing.T) {
	var (
		sss       = newSubscribers()
		bad       = "bad element"
		dataType  = reflect.TypeOf(busTestData{})
		otherType = reflect.TypeOf(busTestOtherData{})
		typed     = &subscriber{
			Name: "typed",
			Item: &testSubscriber{},
			Types: []*subscriberTypeInfo{{
				Original:     busTestData{},
				OriginalType: dataType,
			}},
		}
		anySubscriber = &subscriber{Name: "any", Item: &testAnySubscriber{}}
	)

	if sss.Len() != 0 {
		t.Fatal("newSubscribers() вернул не пустой объект")
	}
	sss.items.PushBack(bad)
	if sss.Len() != 1 {
		t.Fatalf("Len() после добавления служебного элемента = %d, ожидался 1", sss.Len())
	}
	sss.Store(typed)
	if sss.Len() != 2 {
		t.Fatalf("Len() = %d, ожидался 2", sss.Len())
	}
	if !sss.IsExistSubscriber(dataType) {
		t.Fatal("IsExistSubscriber() не нашёл подписчика типа")
	}
	if sss.IsExistSubscriber(otherType) {
		t.Fatal("IsExistSubscriber() нашёл подписчика для незарегистрированного типа")
	}
	if got := sss.GetSubscriber(dataType); len(got) != 1 || got[0] != typed {
		t.Fatalf("GetSubscriber(dataType) = %+v", got)
	}
	if got := sss.GetSubscriber(otherType); len(got) != 0 {
		t.Fatalf("GetSubscriber(otherType) = %+v", got)
	}

	sss.Store(anySubscriber)
	if !sss.IsExistSubscriber(otherType) {
		t.Fatal("IsExistSubscriber() не учёл any-подписчика")
	}
	if got := sss.GetSubscriber(dataType); len(got) != 2 {
		t.Fatalf("GetSubscriber(dataType) после any = %+v", got)
	}
	if got := sss.GetSubscriber(otherType); len(got) != 1 || got[0] != anySubscriber {
		t.Fatalf("GetSubscriber(otherType) после any = %+v", got)
	}

	got := sss.GetSubscriber(dataType)
	got = append(got, anySubscriber)
	if len(sss.GetSubscriber(dataType)) != 2 {
		t.Fatal("GetSubscriber() вернул срез-алиас внутренней карты")
	}

	if err := sss.Delete("unknown"); err == nil {
		t.Fatal("Delete() для отсутствующего подписчика вернул nil ошибку")
	}
	if err := sss.Delete("TYPED"); err != nil {
		t.Fatalf("Delete() вернул ошибку: %v", err)
	}
	if got := sss.GetSubscriber(dataType); len(got) != 1 || got[0] != anySubscriber {
		t.Fatalf("GetSubscriber(dataType) после удаления typed = %+v", got)
	}
	if err := sss.Delete("any"); err != nil {
		t.Fatalf("Delete() any вернул ошибку: %v", err)
	}
	if sss.IsExistSubscriber(dataType) {
		t.Fatal("IsExistSubscriber() нашёл удалённых подписчиков")
	}
	if sss.IsExistSubscriber(otherType) {
		t.Fatal("IsExistSubscriber() нашёл удалённого any-подписчика")
	}
}

package pdw

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

// foreignData Объект реализующий интерфейс Data, но не являющийся *data.
type foreignData struct{}

func (foreignData) DataPut(any, bool, context.Context) {}
func (foreignData) DataGet() any                       { return nil }
func (foreignData) Type() reflect.Type                 { return nil }
func (foreignData) IsSync() bool                       { return false }
func (foreignData) DoneSet()                           {}
func (foreignData) Done() <-chan struct{}              { return nil }
func (foreignData) Result() Result                     { return nil }
func (foreignData) Context() context.Context           { return nil }
func (foreignData) Reset()                             {}

// TestNew Конструктор бассейна.
func TestNew(t *testing.T) {
	if New(false).Statistic() != nil {
		t.Error("Statistic(): для бассейна без отладки ожидалось nil")
	}
	if New(true).Statistic() == nil {
		t.Error("Statistic(): для бассейна с отладкой ожидалась статистика")
	}
}

// TestDebug Переключение режима отладки.
func TestDebug(t *testing.T) {
	p := New(false)
	if p.Statistic() != nil {
		t.Fatal("Statistic(): для бассейна без отладки ожидалось nil")
	}
	p.Debug(true)
	if p.Statistic() == nil {
		t.Error("Statistic(): после включения отладки ожидалась статистика")
	}
	p.Debug(false)
	if p.Statistic() != nil {
		t.Error("Statistic(): после выключения отладки ожидалось nil")
	}
}

// TestGetPut Получение и возвращение объекта в бассейн.
func TestGetPut(t *testing.T) {
	var (
		p  = New(true)
		d1 Data
		d2 Data
		ok bool
	)

	// Объект из бассейна инициализирован конструктором.
	if d1 = p.Get(); d1 == nil {
		t.Fatal("Get(): ожидался объект, получено nil")
	}
	if d1.Context() == nil {
		t.Error("Context(): ожидался контекст по умолчанию")
	}
	if d1.IsSync() {
		t.Error("IsSync(): по умолчанию ожидался асинхронный режим")
	}
	if d1.DataGet() != nil {
		t.Error("DataGet(): по умолчанию ожидались пустые данные")
	}
	if d1.Type() != nil {
		t.Error("Type(): по умолчанию ожидался пустой тип")
	}
	if d1.Result() == nil {
		t.Error("Result(): ожидался объект результата")
	}

	// Возвращение объекта в бассейн и повторное получение.
	p.Put(d1)
	if d2 = p.Get(); d2 == nil {
		t.Fatal("Get(): ожидался объект, получено nil")
	}
	if _, ok = d2.(*data); !ok {
		t.Fatalf("Get(): ожидался тип *data, получено %T", d2)
	}
	// После возвращения в бассейн объект очищен.
	if d2.DataGet() != nil || d2.IsSync() || d2.Type() != nil {
		t.Error("Reset(): после возвращения в бассейн объект не очищен")
	}
	if d2.Context() == nil {
		t.Error("Reset(): после возвращения в бассейн контекст не сброшен")
	}
	p.Put(d2)
}

// TestStatistic Статистика работы бассейна.
func TestStatistic(t *testing.T) {
	var (
		p = New(true)
		n = 3
		s *Statistic
	)
	for i := 0; i < n; i++ {
		d := p.Get()
		p.Put(d)
	}
	if s = p.Statistic(); s == nil {
		t.Fatal("Statistic(): ожидалась статистика")
	}
	if s.Constructor < 1 {
		t.Errorf("Statistic(): ожидалось создание объектов, получено %d", s.Constructor)
	}
	if s.GetObject != int64(n) {
		t.Errorf("Statistic(): ожидалось %d получений, получено %d", n, s.GetObject)
	}
	if s.PutObject != int64(n) {
		t.Errorf("Statistic(): ожидалось %d возвращений, получено %d", n, s.PutObject)
	}
}

// TestGetWithoutConstructor Получение объекта из бассейна без конструктора.
func TestGetWithoutConstructor(t *testing.T) {
	// Объект без инициализации конструктора, имитация состояния "пустого" бассейна.
	p := &impl{Pool: sync.Pool{}, debug: false, statistic: new(Statistic)}
	if got := p.Get(); got != nil {
		t.Errorf("Get(): ожидалось nil, получено %T", got)
	}
}

// TestPutType Проверка типа возвращаемого объекта.
func TestPutType(t *testing.T) {
	p := New(false)
	defer func() {
		if e := recover(); e == nil {
			t.Error("Put(): ожидалась паника для объекта чужого типа")
		}
	}()
	// Передача объекта не являющегося *data.
	p.Put(foreignData{})
}

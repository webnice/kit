package pdw

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

// foreignData Объект реализующий интерфейс Data, но не являющийся *data.
type foreignData struct{}

// DataPut Реализация интерфейса Data.
func (foreignData) DataPut(any, bool, context.Context) {}

// DataGet Реализация интерфейса Data.
func (foreignData) DataGet() any { return nil }

// Type Реализация интерфейса Data.
func (foreignData) Type() reflect.Type { return nil }

// IsSync Реализация интерфейса Data.
func (foreignData) IsSync() bool { return false }

// DoneSet Реализация интерфейса Data.
func (foreignData) DoneSet() {}

// Done Реализация интерфейса Data.
func (foreignData) Done() <-chan struct{} { return nil }

// Result Реализация интерфейса Data.
func (foreignData) Result() Result { return nil }

// Context Реализация интерфейса Data.
func (foreignData) Context() context.Context { return nil }

// Reset Реализация интерфейса Data.
func (foreignData) Reset() {}

// TestNew Конструктор бассейна.
func TestNew(t *testing.T) {
	const (
		errNoDebugStatNotNil = "Statistic(): для бассейна без отладки ожидалось nil."
		errDebugStatNil      = "Statistic(): для бассейна с отладкой ожидалась статистика."
	)

	if New(false).Statistic() != nil {
		t.Error(errNoDebugStatNotNil)
	}
	if New(true).Statistic() == nil {
		t.Error(errDebugStatNil)
	}
}

// TestDebug Переключение режима отладки.
func TestDebug(t *testing.T) {
	const (
		errNoDebugStatNotNil = "Statistic(): для бассейна без отладки ожидалось nil."
		errDebugStatNil      = "Statistic(): после включения отладки ожидалась статистика."
		errNoDebugStatAfter  = "Statistic(): после выключения отладки ожидалось nil."
	)
	var p Interface

	p = New(false)
	if p.Statistic() != nil {
		t.Fatal(errNoDebugStatNotNil)
	}
	p.Debug(true)
	if p.Statistic() == nil {
		t.Error(errDebugStatNil)
	}
	p.Debug(false)
	if p.Statistic() != nil {
		t.Error(errNoDebugStatAfter)
	}
}

// TestGetPut Получение и возвращение объекта в бассейн.
func TestGetPut(t *testing.T) {
	const (
		errGetNil          = "Get(): ожидался объект, получено nil."
		errContextNil      = "Context(): ожидался контекст по умолчанию."
		errIsSyncDefault   = "IsSync(): по умолчанию ожидался асинхронный режим."
		errDataNotEmpty    = "DataGet(): по умолчанию ожидались пустые данные."
		errTypeNotEmpty    = "Type(): по умолчанию ожидался пустой тип."
		errResultNil       = "Result(): ожидался объект результата."
		errTypeWrong       = "Get(): ожидался тип *data, получено %T."
		errNotReset        = "Reset(): после возвращения в бассейн объект не очищен."
		errContextNotReset = "Reset(): после возвращения в бассейн контекст не сброшен."
	)
	var (
		p  Interface
		d1 Data
		d2 Data
		ok bool
	)

	p = New(true)
	// Объект из бассейна инициализирован конструктором.
	if d1 = p.Get(); d1 == nil {
		t.Fatal(errGetNil)
	}
	if d1.Context() == nil {
		t.Error(errContextNil)
	}
	if d1.IsSync() {
		t.Error(errIsSyncDefault)
	}
	if d1.DataGet() != nil {
		t.Error(errDataNotEmpty)
	}
	if d1.Type() != nil {
		t.Error(errTypeNotEmpty)
	}
	if d1.Result() == nil {
		t.Error(errResultNil)
	}
	// Возвращение объекта в бассейн и повторное получение.
	p.Put(d1)
	if d2 = p.Get(); d2 == nil {
		t.Fatal(errGetNil)
	}
	if _, ok = d2.(*data); !ok {
		t.Fatalf(errTypeWrong, d2)
	}
	// После возвращения в бассейн объект очищен.
	if d2.DataGet() != nil || d2.IsSync() || d2.Type() != nil {
		t.Error(errNotReset)
	}
	if d2.Context() == nil {
		t.Error(errContextNotReset)
	}
	p.Put(d2)
}

// TestStatistic Статистика работы бассейна.
func TestStatistic(t *testing.T) {
	const (
		errStatNil     = "Statistic(): ожидалась статистика."
		tplConstructor = "Statistic(): ожидалось создание объектов, получено %d."
		tplGetObject   = "Statistic(): ожидалось %d получений, получено %d."
		tplPutObject   = "Statistic(): ожидалось %d возвращений, получено %d."
	)
	var (
		p    Interface
		s    *Statistic
		d    Data
		i, n int
	)

	p = New(true)
	n = 3
	for i = 0; i < n; i++ {
		d = p.Get()
		p.Put(d)
	}
	if s = p.Statistic(); s == nil {
		t.Fatal(errStatNil)
	}
	if s.Constructor < 1 {
		t.Errorf(tplConstructor, s.Constructor)
	}
	if s.GetObject != int64(n) {
		t.Errorf(tplGetObject, n, s.GetObject)
	}
	if s.PutObject != int64(n) {
		t.Errorf(tplPutObject, n, s.PutObject)
	}
}

// TestGetWithoutConstructor Получение объекта из бассейна без конструктора.
func TestGetWithoutConstructor(t *testing.T) {
	const (
		tplGetNotNil = "Get(): ожидалось nil, получено %T."
	)
	// Объект без инициализации конструктора, имитация состояния "пустого" бассейна.
	var p *impl

	p = &impl{Pool: sync.Pool{}, debug: false, statistic: new(Statistic)}
	if got := p.Get(); got != nil {
		t.Errorf(tplGetNotNil, got)
	}
}

// TestPutType Проверка типа возвращаемого объекта.
func TestPutType(t *testing.T) {
	const (
		errNoPanic = "Put(): ожидалась паника для объекта чужого типа."
	)
	var p Interface

	p = New(false)
	defer func() {
		if e := recover(); e == nil {
			t.Error(errNoPanic)
		}
	}()
	// Передача объекта не являющегося *data.
	p.Put(foreignData{})
}

package pdw

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

// Тестовая ошибка результата.
var errTestResult = errors.New("тестовая ошибка результата")

// TestDataMethods Методы объекта обёртки данных.
func TestDataMethods(t *testing.T) {
	var (
		d   = New(false).Get()
		ctx = context.WithValue(context.Background(), struct{ k string }{"k"}, "v")
	)

	// Загрузка данных с контекстом.
	d.DataPut("тестовые данные", true, ctx)
	if d.DataGet() != "тестовые данные" {
		t.Errorf("DataGet(): ожидались данные %q, получено %v", "тестовые данные", d.DataGet())
	}
	if d.Type() != reflect.TypeOf("тестовые данные") {
		t.Errorf("Type(): ожидался тип %v, получено %v", reflect.TypeOf(""), d.Type())
	}
	if !d.IsSync() {
		t.Error("IsSync(): ожидался синхронный режим")
	}
	if d.Context() != ctx {
		t.Error("Context(): ожидался переданный контекст")
	}
}

// TestDataPutNilContext Загрузка данных с пустым контекстом.
func TestDataPutNilContext(t *testing.T) {
	d := New(false).Get()
	d.DataPut("данные", false, nil)
	if d.Context() == nil {
		t.Error("Context(): при nil контексте ожидался контекст по умолчанию")
	}
}

// TestDoneSetDone Сигнал завершения обработки данных.
func TestDoneSetDone(t *testing.T) {
	d := New(false).Get()
	d.DataPut("данные", true, nil)
	d.DoneSet()
	select {
	case <-d.Done():
	case <-time.After(time.Second):
		t.Error("Done(): ожидался сигнал завершения обработки данных")
	}
	// Повторная установка сигнала не вызывает панику.
	d.DoneSet()
	// После сброса объект снова готов к работе.
	d.Reset()
	select {
	case <-d.Done():
		t.Error("Done(): после Reset() сигнала завершения быть не должно")
	case <-time.After(50 * time.Millisecond):
	}
}

// TestDoneSetWithoutReceiver Сигнал завершения без ожидающего получателя не блокируется.
func TestDoneSetWithoutReceiver(t *testing.T) {
	d := New(false).Get()
	d.DataPut("данные", false, nil) // Асинхронный режим, получателя нет.
	done := make(chan struct{})
	go func() { d.DoneSet(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("DoneSet(): без ожидающего получателя вызов блокируется")
	}
}

// TestResult Работа с результатом обработки данных.
func TestResult(t *testing.T) {
	d := New(false).Get()
	r := d.Result()
	r.DataPut("результат 1", "результат 2")
	r.ErrPut(errTestResult)
	if got := r.DataGet(); len(got) != 2 || got[0] != "результат 1" || got[1] != "результат 2" {
		t.Errorf("DataGet(): неожиданные данные результата: %v", got)
	}
	if got := r.ErrGet(); len(got) != 1 || got[0] != errTestResult {
		t.Errorf("ErrGet(): неожиданные ошибки результата: %v", got)
	}
	// Очистка результата при сбросе объекта.
	d.Reset()
	if got := r.DataGet(); len(got) != 0 {
		t.Errorf("Reset(): ожидался пустой результат, получено %v", got)
	}
	if got := r.ErrGet(); len(got) != 0 {
		t.Errorf("Reset(): ожидался пустой список ошибок, получено %v", got)
	}
}

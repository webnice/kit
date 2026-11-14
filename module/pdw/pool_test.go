package pdw

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

// Тестовая ошибка результата.
var errTestResult = errors.New("тестовая ошибка результата.")

// TestDataMethods Методы объекта обёртки данных.
func TestDataMethods(t *testing.T) {
	const (
		tplDataMismatch = "DataGet(): ожидались данные %q, получено %v."
		tplTypeMismatch = "Type(): ожидался тип %v, получено %v."
		errNotSync      = "IsSync(): ожидался синхронный режим."
		errContextWrong = "Context(): ожидался переданный контекст."
		testData        = "тестовые данные."
	)
	var (
		d   Data
		ctx context.Context
	)

	d = New(false).Get()
	ctx = context.WithValue(context.Background(), struct{ k string }{"k"}, "v")
	// Загрузка данных с контекстом.
	d.DataPut(testData, true, ctx)
	if d.DataGet() != testData {
		t.Errorf(tplDataMismatch, testData, d.DataGet())
	}
	if d.Type() != reflect.TypeOf(testData) {
		t.Errorf(tplTypeMismatch, reflect.TypeOf(""), d.Type())
	}
	if !d.IsSync() {
		t.Error(errNotSync)
	}
	if d.Context() != ctx {
		t.Error(errContextWrong)
	}
}

// TestDataPutNilContext Загрузка данных с пустым контекстом.
func TestDataPutNilContext(t *testing.T) {
	const (
		errContextNil = "Context(): при nil контексте ожидался контекст по умолчанию."
	)
	var d Data

	d = New(false).Get()
	d.DataPut("данные", false, nil)
	if d.Context() == nil {
		t.Error(errContextNil)
	}
}

// TestDoneSetDone Сигнал завершения обработки данных.
func TestDoneSetDone(t *testing.T) {
	const (
		errNoDoneSignal   = "Done(): ожидался сигнал завершения обработки данных."
		errDoneAfterReset = "Done(): после Reset() сигнала завершения быть не должно."
	)
	var d Data

	d = New(false).Get()
	d.DataPut("данные", true, nil)
	d.DoneSet()
	select {
	case <-d.Done():
	case <-time.After(time.Second):
		t.Error(errNoDoneSignal)
	}
	// Повторная установка сигнала не вызывает панику.
	d.DoneSet()
	// После сброса объект снова готов к работе.
	d.Reset()
	select {
	case <-d.Done():
		t.Error(errDoneAfterReset)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestDoneSetWithoutReceiver Сигнал завершения без ожидающего получателя не блокируется.
func TestDoneSetWithoutReceiver(t *testing.T) {
	const (
		errDoneBlocked = "DoneSet(): без ожидающего получателя вызов блокируется."
	)
	var (
		d    Data
		done chan struct{}
	)

	d = New(false).Get()
	d.DataPut("данные", false, nil) // Асинхронный режим, получателя нет.
	done = make(chan struct{})
	go func() { d.DoneSet(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal(errDoneBlocked)
	}
}

// TestResult Работа с результатом обработки данных.
func TestResult(t *testing.T) {
	const (
		tplResultDataWrong   = "DataGet(): неожиданные данные результата: %v."
		tplResultErrWrong    = "ErrGet(): неожиданные ошибки результата: %v."
		tplResetDataNotEmpty = "Reset(): ожидался пустой результат, получено %v."
		tplResetErrNotEmpty  = "Reset(): ожидался пустой список ошибок, получено %v."
	)
	var (
		d Data
		r Result
	)

	d = New(false).Get()
	r = d.Result()
	r.DataPut("результат 1", "результат 2")
	r.ErrPut(errTestResult)
	if got := r.DataGet(); len(got) != 2 || got[0] != "результат 1" || got[1] != "результат 2" {
		t.Errorf(tplResultDataWrong, got)
	}
	if got := r.ErrGet(); len(got) != 1 || got[0] != errTestResult {
		t.Errorf(tplResultErrWrong, got)
	}
	// Очистка результата при сбросе объекта.
	d.Reset()
	if got := r.DataGet(); len(got) != 0 {
		t.Errorf(tplResetDataNotEmpty, got)
	}
	if got := r.ErrGet(); len(got) != 0 {
		t.Errorf(tplResetErrNotEmpty, got)
	}
}

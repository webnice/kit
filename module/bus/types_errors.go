package bus

import "fmt"

type (
	// Error Объект-одиночка со списком ошибок которые можно сравнивать по якорю через '=='.
	Error struct{}

	// Внутренняя структура объекта ошибки с кодом, шаблоном, якорем и интерфейсом error.
	err struct {
		tpl    string        // Шаблон ошибки.
		code   uint8         // Код ошибки.
		args   []any         // Иные аргументы ошибки.
		anchor error         // Константа ошибки с фиксированным адресом.
		errFn  func() string // Функция интерфейса error.
	}

	// Err Интерфейс ошибки приложения.
	Err interface {
		Anchor() error // Якорь, по которому можно сравнивать две ошибки между собой.
		Code() uint8   // Код ошибки.
		Error() string // Сообщение об ошибке или шаблон сообщения об ошибке.
	}
)

// Anchor Реализация интерфейса error.
func (err err) Anchor() error { return err.anchor }

// Code Возврат кода ошибки.
func (err err) Code() uint8 { return err.code }

// Error Реализация интерфейса error.
func (err err) Error() string { return err.errFn() }

// Errors Справочник всех ошибок пакета.
func Errors() *Error { return errSingleton }

// Конструктор объекта ошибки.
func newErr(obj *err, code uint8, arg ...any) Err {
	if code == 0 {
		code = obj.code // Если код ошибки не изменён, используется код ошибки из шаблона.
	}
	return &err{
		anchor: obj,
		code:   code,
		args:   arg,
		errFn:  func() string { return fmt.Sprintf(obj.tpl, arg...) },
	}
}

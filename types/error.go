// Package types
package types

import "fmt"

// ErrorWithCode Описание интерфейса ошибки приложения
type ErrorWithCode interface {
	// Code Возврат кода ошибки
	Code() uint8

	// Error Реализация интерфейса error
	Error() string
}

// ErrorWithCode Описание структуры ошибки приложения
type errorWithCode struct {
	c uint8
	s string
}

// NewErrorWithCode Создание объекта ошибки приложения и возврат интерфейса error
func NewErrorWithCode(code uint8, format string, arg ...interface{}) error {
	return &errorWithCode{code, fmt.Sprintf(format, arg...)}
}

// Code Возврат кода ошибки
func (err *errorWithCode) Code() uint8 { return err.c }

// Error Реализация интерфейса error
func (err *errorWithCode) Error() string { return err.s }

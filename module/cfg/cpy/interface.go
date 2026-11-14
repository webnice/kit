package cpy

// Gist Интерфейс к служебным методам.
func Gist() *Cpy { return singleton }

// All Копирование всех полей из одного объекта в другой.
func All(toObj any, fromObj any) error {
	return singleton.Copy(toObj, fromObj, nil, nil, nil)
}

// Select Копирование только выбранных полей.
// Применяется только для структур.
func Select(toObj any, fromObj any, fields ...string) error {
	return singleton.Copy(toObj, fromObj, fields, nil, nil)
}

// Omit Копирование всех полей, кроме перечисленных.
// Применяется только для структур.
func Omit(toObj any, fromObj any, fields ...string) error {
	return singleton.Copy(toObj, fromObj, nil, fields, nil)
}

// Filter Копирование данных с фильтрацией элементов, применяется для массивов, срезов и карт.
func Filter(toObj any, fromObj any, filter FilterFn) error {
	return singleton.Copy(toObj, fromObj, nil, nil, filter)
}

// Errors Справочник ошибок.
func Errors() *Error { return errSingleton }

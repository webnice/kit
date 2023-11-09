package cpy

const tagName = `cpy`

var singleton = &Cpy{}

// FilterFn Тип функции фильтрации.
// Вернётся "истина", для пропуска данных.
type FilterFn func(key interface{}, object interface{}) (skip bool)

// Cpy Объект сущности пакета.
type Cpy struct{}

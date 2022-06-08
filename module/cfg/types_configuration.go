// Package cfg
package cfg

import "reflect"

// Структура регистра хранения объектов конфигурации.
type mainConfiguration struct {
	Items []*configurationItem // Зарегистрированные и разобранные объекты конфигураций.
	Union interface{}          // Объединённая структура, содержащая в себе все зарегистрированные структуры.
}

// Структура регистра хранения одного объекта одной конфигурации.
type configurationItem struct {
	Original interface{}           // Ссылка на оригинальный объект конфигурации
	Fields   []reflect.StructField // Все найденные экспортируемые поля оригинальной структуры
	Type     reflect.Type          // Тип reflect.Type структуры
	Value    reflect.Value         // Тип reflect.Value оригинального объекта
}

// IsName Проверка существования поля с указанным именем во всех уже добавленных объектах конфигурации.
// Вернётся истина, если будет найдено совпадение.
func (mcn mainConfiguration) IsName(name string) (ret bool) {
	var n, i int

	for n = range mcn.Items {
		for i = range mcn.Items[n].Fields {
			if mcn.Items[n].Fields[i].Name == name {
				ret = true
				return
			}
		}
	}

	return
}

// IsTagValue Проверка существования поля с указанным тегом имеющим указанное значение во всех полях всех уже
// добавленных объектах конфигурации.
// Вернётся истина, если будет найдено совпадение.
func (mcn mainConfiguration) IsTagValue(name string, value string) (ret bool) {
	var (
		n, i     int
		ok       bool
		tagValue string
	)

	for n = range mcn.Items {
		for i = range mcn.Items[n].Fields {
			if tagValue, ok = mcn.Items[n].Fields[i].Tag.Lookup(name); ok {
				if tagValue == value {
					ret = true
					return
				}
			}
		}
	}

	return
}

// StructField Функция собирает и возвращает все поля всех зарегистрированных объектов конфигурации в один срез.
func (mcn mainConfiguration) StructField() (ret []reflect.StructField) {
	var n int

	for n = range mcn.Items {
		ret = append(ret, mcn.Items[n].Fields...)
	}

	return
}

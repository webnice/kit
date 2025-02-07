package sql

import "gorm.io/gorm"

// ModelGorm Минимальный обязательный интерфейс объекта модели.
type ModelGorm interface {
	// TableName Явное указание названия таблицы.
	TableName() string

	// BeforeCreate Функция вызываемая до создания нового объекта в базе данных.
	BeforeCreate(tx *gorm.DB) (err error)

	// BeforeUpdate Функция вызываемая до обновления объекта в базе данных.
	BeforeUpdate(tx *gorm.DB) (err error)
}

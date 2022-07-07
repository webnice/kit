// Package sql
package sql

import (
	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

// Gist Возвращается настроенный и готовый к работе интерфейс подключения к базе данных.
func (db *Implementation) Gist() Interface { return db.getParent() }

// Gorm Возвращается настроенный и готовый к работе объект ORM gorm.io/gorm.
func (db *Implementation) Gorm() *gorm.DB { return db.getParent().GormDB() }

// Sqlx Настроенный и готовый к работе объект обёртки над соединением с БД github.com/jmoiron/sqlx.
func (db *Implementation) Sqlx() *sqlx.DB { return db.getParent().SqlxDB() }

// Возвращает объект родителя, с запоминанием объекта.
func (db *Implementation) getParent() Interface {
	if db.parent != nil {
		return db.parent
	}
	db.parent = Get()

	return db.parent
}

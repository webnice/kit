// Package sql
package sql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

// Gist Возвращается настроенный и готовый к работе интерфейс подключения к базе данных.
func (db *Implementation) Gist() Interface { return db.getParent() }

// Gorm Возвращается настроенный и готовый к работе объект ORM gorm.io/gorm.
func (db *Implementation) Gorm(opts ...*Option) (ret *gorm.DB) {
	var n int

	ret = db.getParent().GormDB()
	for n = range opts {
		if opts[n] != nil && opts[n].ctx != nil {
			ret = ret.Session(&gorm.Session{Context: opts[n].ctx})
		}
	}

	return
}

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

// OptionSilent Полное отключение логирования запросов к базе данных.
func (db *Implementation) OptionSilent() *Option {
	return &Option{ctx: context.WithValue(context.Background(), keyContextLogLevel, keyLogSilent)}
}

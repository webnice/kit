// Package sql
package sql

import (
	"context"
	"database/sql"
	"sync"

	kitModuleDbSqlTypes "github.com/webnice/kit/v3/module/db/sql/types"
	kmll "github.com/webnice/kit/v3/module/log/level"
	kitTypesDb "github.com/webnice/kit/v3/types/db"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

const (
	driverMySQL         = `mysql`
	driverPostgreSQL    = `postgres`
	driverSqlite        = `sqlite`
	dsnTimeSettings     = `?parseTime=%t`
	dsnLocationSettings = `&loc=%s`
	dsnUnixTpl          = `@unix(%s)`
	dsnCharsetTpl       = `&charset=%s`
)

const (
	keyContextLogLevel = `log_level`
	keyLogSilent       = `silent`
)

var (
	singleton      *impl
	supportDrivers = []string{driverMySQL, driverPostgreSQL, driverSqlite}
)

// Interface Интерфейс пакета.
type Interface interface {
	// Close Закрытие соединения с базой данных.
	Close() (err error)

	// E Ошибка соединения с базой данных.
	// Если err==nil - база данных доступна, соединение активно, ошибок нет.
	// Если err!=nil - есть проблема с соединением с базой данных.
	E() (err error)

	// Status Возвращает состояние подключения к базе данных.
	Status() (ret *sql.DBStats)

	// SqlDB Настроенный и готовый к работе бассейн соединений database/sql.
	// Если возвращается nil - есть ошибки, ошибка доступна в функции E()
	SqlDB() (ret *sql.DB)

	// GormDB Настроенный и готовый к работе объект ORM gorm.io/gorm.
	// Если возвращается nil - есть ошибки, ошибка доступна в функции E()
	GormDB() (ret *gorm.DB)

	// SqlxDB Настроенный и готовый к работе объект обёртки над соединением с БД github.com/jmoiron/sqlx.
	// Если возвращается nil - есть ошибки, ошибка доступна в функции E()
	SqlxDB() (ret *sqlx.DB)

	// MigrationUp Применение миграций базы данных.
	MigrationUp() (err error)

	// ОШИБКИ

	// Errors Справочник всех ошибок пакета.
	Errors() *Error
}

// Объект сущности, реализующий интерфейс Interface.
type impl struct {
	databaseSql *kitTypesDb.DatabaseSqlConfiguration // Конфигурация для подключения к базе данных.
	cfg         *kitModuleDbSqlTypes.Configuration   // Сегмент конфигурации для подключения к базе данных.
	dsn         string                               // Строка в формате DSN для подключения к базе данных.
	error       error                                // Последняя ошибка, возникшая при работе с соединением или драйвером базы данных.
	connect     *sql.DB                              // Установленное соединение с базой данных.
	connectMux  *sync.RWMutex                        // Блокировка доступа на время установки соединения или переподключения.
}

// Implementation Встраиваемая структура в модель базы данных, для лёгкого подключения "по требованию" к базе данных.
type Implementation struct {
	parent Interface // Временная копия родительского объекта подключения к базе данных.
}

// Option Опциональные настройки работы библиотеки.
type Option struct {
	ctx context.Context // Контекст со значениями опциональных настроек.
}

// Объект сессии Gorm с настройками логирования и возможностью переопределения в пределах сессии уровня
// логирования.
type logGorm struct {
	parent   *impl
	Loglevel kmll.Level
}

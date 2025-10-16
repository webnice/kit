package migration

import (
	"database/sql"

	"gorm.io/gorm"
)

// New Конструктор объекта сущности пакета, возвращается интерфейс пакета.
//
//goland:noinspection GoUnusedExportedFunction
func New() Interface {
	var mmn = new(impl)
	return mmn
}

// Errors Справочник всех ошибок пакета.
func (mmn *impl) Errors() *Error { return Errors() }

// CurrentVersion Возвращается текущая версия схемы базы данных.
func (mmn *impl) CurrentVersion() (ver *DbVersion, err error) {
	var (
		dbi *sql.DB
		orm *gorm.DB
	)

	if orm = mmn.Gist().GormDB(); orm == nil {
		err = mmn.Errors().DatabaseIsNotInUse.Bind()
		return
	}
	if dbi, err = mmn.Gist().GormDB().DB(); err != nil {
		err = mmn.Errors().DatabaseIsNotInUse.Bind()
		return
	}
	if err = dbi.Ping(); err != nil {
		err = mmn.Errors().DatabaseUnexpectedError.Bind(err)
		return
	}
	ver = new(DbVersion)
	if err = mmn.Gorm().
		Where("`is_applied` = 1").
		Order("`tstamp` DESC, `id` DESC").
		Take(ver).
		Error; err != nil {
		err = mmn.QErr(err)
		return
	}

	return
}

package migration

import (
	"database/sql"

	kitModuleDbSql "github.com/webnice/kit/v4/module/db/sql"

	"gorm.io/gorm"
)

// New Конструктор объекта сущности пакета, возвращается интерфейс пакета.
//
//goland:noinspection GoUnusedExportedFunction
func New() Interface {
	var mmn = new(impl)
	return mmn
}

func newWithDB(dbi kitModuleDbSql.Interface) Interface { return &impl{db: dbi} }

// Errors Справочник всех ошибок пакета.
func (mmn *impl) Errors() *Error { return Errors() }

func (mmn *impl) gist() kitModuleDbSql.Interface {
	if mmn.db != nil {
		return mmn.db
	}

	return mmn.Gist()
}

// CurrentVersion Возвращается текущая версия схемы базы данных.
func (mmn *impl) CurrentVersion() (ver *DbVersion, err error) {
	var (
		dbi *sql.DB
		orm *gorm.DB
		db  kitModuleDbSql.Interface
	)

	db = mmn.gist()
	if orm = db.GormDB(); orm == nil {
		err = mmn.Errors().DatabaseIsNotInUse.Bind()
		return
	}
	if dbi, err = orm.DB(); err != nil {
		err = mmn.Errors().DatabaseIsNotInUse.Bind()
		return
	}
	if err = dbi.Ping(); err != nil {
		err = mmn.Errors().DatabaseUnexpectedError.Bind(err)
		return
	}
	ver = new(DbVersion)
	if err = orm.
		Where("`is_applied` = 1").
		Order("`tstamp` DESC, `id` DESC").
		Take(ver).
		Error; err != nil {
		err = mmn.QErr(err)
		return
	}

	return
}

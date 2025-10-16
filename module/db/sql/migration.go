package sql

import (
	"errors"
	"math"

	//kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	"github.com/webnice/migrate/goose"

	// Обязательно наличие зарегистрированных драйверов баз данных.
	_ "github.com/go-sql-driver/mysql" // mysql.
	_ "github.com/jackc/pgx/v4"        // postgre, cockroach, redshift.
)

// MigrationUp Применение миграций базы данных.
func (mys *impl) MigrationUp() (err error) {
	const (
		tplNewMigration   = "Найдены новые миграции базы данных %q, новых миграций: %d."
		tplNoNewMigration = "Нет новых миграций базы данных %q, версия базы данных: %d."
		tplNewApply       = "Применение миграций базы данных."
		tplNewApplied     = "Новые миграции базы данных %q успешно применены."
	)
	var (
		migration goose.Migrations
		next      *goose.Migration
		current   int64
		n         int
		count     int
		end       bool
	)

	// Соединение заблокировано, если оно находится в состоянии подключения или переподключения, необходимо подождать.
	mys.connectMux.RLock()
	mys.connectMux.RUnlock()
	// Отключение таймаута на время применения миграций.
	if mys.connect != nil {
		mys.connect.SetConnMaxLifetime(0)
	}
	defer func() { mys.connect.SetConnMaxLifetime(mys.cfg.MaxLifetimeConn) }()
	// Настройка диалекта библиотеки применения миграций.
	if err = goose.SetDialect(mys.cfg.Driver); err != nil {
		err = mys.Errors().UnknownDialect.Bind(mys.cfg.Driver, err)
		return
	}
	//kitModuleCfg.Get().Gist().AbsolutePathAndUpdate(&mys.cfg.Migration)
	// Получение текущей версии базы данных и подсчёт количества новых миграций.
	switch current, err = goose.EnsureDBVersion(mys.conn()); {
	case err == nil:
	case errors.Is(err, goose.ErrNoNextVersion):
		err = nil
	default:
		return
	}
	// Поиск миграций в папке миграций.
	if migration, err = goose.
		CollectMigrations(mys.cfg.Migration, 0, math.MaxInt64); err != nil {
		return
	}
	for n = range migration {
		if migration[n].Version > current {
			count++
		}
	}
	if count <= 0 {
		mys.info(tplNoNewMigration, mys.cfg.Driver, current)
		return
	}
	mys.info(tplNewMigration, mys.cfg.Driver, count)
	// Применение миграций.
	mys.info(tplNewApply)
	for {
		if end {
			break
		}
		if current, err = goose.GetDBVersion(mys.conn()); err != nil {
			end, err = true, nil
			continue
		}
		switch next, err = migration.Next(current); {
		case err == nil:
		case errors.Is(err, goose.ErrNoNextVersion):
			end, err = true, nil
			continue
		}
		if err = next.Up(mys.conn()); err != nil {
			end, err = true, mys.Errors().ApplyMigration.Bind(next.Source, err)
			continue
		}
	}
	if err == nil {
		mys.info(tplNewApplied, mys.cfg.Driver)
	}

	return
}

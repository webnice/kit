// Package sql
package sql

import (
	"context"
	"time"

	kitModuleDye "github.com/webnice/kit/v3/module/dye"
	kitTypes "github.com/webnice/kit/v3/types"

	gormLogger "gorm.io/gorm/logger"
)

func (mys *impl) LogMode(l gormLogger.LogLevel) gormLogger.Interface {
	mys.log().Noticef("gorm уровень логирования: %d", int(l))
	return mys
}

func (mys *impl) Info(_ context.Context, s string, i ...interface{}) { mys.log().Infof(s, i...) }

func (mys *impl) Warn(_ context.Context, s string, i ...interface{}) { mys.log().Warningf(s, i...) }

func (mys *impl) Error(_ context.Context, s string, i ...interface{}) { mys.log().Errorf(s, i...) }

func (mys *impl) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	const (
		keyQuery, keySql               = `query`, `sql`
		keyDriver, keyElapsed, keyRows = `driver`, `elapsed`, `rows`
		tplTracef, tplErrorf           = `sql:"%s"`, `sql:"%s", ошибка: %s`
	)
	var (
		elapsed time.Duration
		sql     string
		rows    int64
		keys    kitTypes.LoggerKey
	)

	elapsed = time.Since(begin)
	sql, rows = fc()
	keys = kitTypes.LoggerKey{
		keyQuery:   keySql,
		keyDriver:  mys.cfg.Driver,
		keyElapsed: elapsed,
		keyRows:    rows,
	}
	switch err {
	case nil:
		mys.log().Key(keys).Tracef(
			tplTracef,
			kitModuleDye.New().Yellow().Done().String()+sql+kitModuleDye.New().Normal().Done().String(),
		)
	default:
		mys.log().Key(keys).Errorf(
			tplErrorf,
			kitModuleDye.New().Yellow().Done().String()+sql+kitModuleDye.New().Reset().Done().String(),
			kitModuleDye.New().Red().Done().String()+err.Error()+kitModuleDye.New().Reset().Done().String(),
		)
	}
}

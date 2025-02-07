package sql

import (
	"context"
	"strings"
	"time"

	kitModuleDye "github.com/webnice/kit/v4/module/dye"
	kmll "github.com/webnice/kit/v4/module/log/level"
	kitTypes "github.com/webnice/kit/v4/types"

	gormLogger "gorm.io/gorm/logger"
)

// NewLoggerGorm Создание объекта с интерфейсом gorm.logger.Interface.
func NewLoggerGorm(parent *impl) gormLogger.Interface {
	var lgm = &logGorm{
		parent:   parent,
		Loglevel: parent.cfg.Loglevel,
	}

	return lgm
}

// LogMode Переключение уровня логирования из ОРМ библиотеки.
func (lgm *logGorm) LogMode(l gormLogger.LogLevel) gormLogger.Interface {
	switch l {
	case gormLogger.Silent:
		lgm.Loglevel = kmll.Off
	case gormLogger.Error:
		lgm.Loglevel = kmll.Error
	case gormLogger.Warn:
		lgm.Loglevel = kmll.Warning
	case gormLogger.Info:
		lgm.Loglevel = kmll.Info
	default:
		lgm.parent.log().
			Errorf("получен не поддерживаемый уровень логирования: %d", int(l))
	}

	return lgm
}

// Info Все без исключения запросы к базе данных.
func (lgm *logGorm) Info(_ context.Context, s string, i ...any) {
	lgm.parent.log().Infof(s, i...)
}

// Warn Запросы с ошибками, а так же требующие повышенного внимания, но не являющиеся ошибкой.
func (lgm *logGorm) Warn(_ context.Context, s string, i ...any) {
	lgm.parent.log().Warningf(s, i...)
}

// Error Запросы, выполнение которых завершилось ошибкой.
func (lgm *logGorm) Error(_ context.Context, s string, i ...any) {
	lgm.parent.log().Errorf(s, i...)
}

// Trace Трассировка запросов к базе данных.
func (lgm *logGorm) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	const (
		keyQuery, keySql               = "query", "sql"
		keyDriver, keyElapsed, keyRows = "driver", "elapsed", "rows"
		tplTracef, tplErrorf           = "sql:\"%s\"", "sql:\"%s\", ошибка: %s"
		// Увеличение глубины поиска в стеке вызовов, функции вызвавшей логирование, так как используется ORM.
		stackBackCorrect = 5
	)
	var (
		logLevel string
		elapsed  time.Duration
		sql      string
		qSql     string
		rows     int64
		keys     kitTypes.LoggerKey
		ok       bool
		msgFn    func(error, string)
	)

	// Отключение логирования из контекста.
	if logLevel, ok = ctx.Value(keyContextLogLevel).(string); ok && logLevel == keyLogSilent {
		return
	}
	elapsed = time.Since(begin)
	sql, rows = fc()
	keys = kitTypes.LoggerKey{
		keyQuery:   keySql,
		keyDriver:  lgm.parent.cfg.Driver,
		keyElapsed: elapsed,
		keyRows:    rows,
	}
	qSql = strings.ReplaceAll(sql, "\"", "\\\"")
	msgFn = func(e error, color string) {
		if err == nil {
			return
		}
		lgm.parent.log().
			Key(keys).
			StackBackCorrect(stackBackCorrect).
			Errorf(
				tplErrorf,
				kitModuleDye.New().Yellow().Done().String()+
					qSql+
					kitModuleDye.New().Reset().Done().String(),
				color+
					e.Error()+
					kitModuleDye.New().Reset().Done().String(),
			)
	}
	switch lgm.Loglevel {
	case kmll.Off:
		return
	case kmll.Error:
		msgFn(err, kitModuleDye.New().Red().Done().String())
	case kmll.Warning:
		msgFn(err, kitModuleDye.New().Magenta().Done().String())
	case kmll.Info:
		lgm.parent.log().
			Key(keys).
			StackBackCorrect(stackBackCorrect).
			Tracef(
				tplTracef,
				kitModuleDye.New().Yellow().Done().String()+qSql+kitModuleDye.New().Normal().Done().String(),
			)
	default:
		//lgm.parent.log().
		//	Errorf("не поддерживаемый уровень логирования %q", lgm.Loglevel.String())
		return
	}
}

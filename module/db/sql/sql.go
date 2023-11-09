package sql

import (
	"database/sql"
	"runtime"
	"sync"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitTypes "github.com/webnice/kit/v4/types"
	kitTypesDb "github.com/webnice/kit/v4/types/db"

	// Библиотеки работы с базой данных.
	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"

	// Драйверы базы данных.
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
)

// Get Возвращается интерфейс для работы с базой данных.
// Если база данных доступна, тогда возвращается полностью настроенное и готовое к работе соединение с базой данных.
// Если база данных не доступна, тогда возвращается объект, методы которого заблокированы до момента установки
// соединения с базой данных. Параллельно запущен процесс подключения к базе данных, по окончании которого,
// блокировка методов объекта снимается.
func Get() Interface {
	if singleton != nil {
		return singleton
	}
	singleton = constructor()

	return singleton
}

// Free Освобождает соединение работы с базой данных.
// Объект работы с базой данных полностью удаляется из памяти.
func Free() { singleton = nil }

// Создание нового объекта подключения к базе данных.
func constructor() (mys *impl) {
	var onDone chan struct{}

	mys = &impl{
		databaseSql: new(kitTypesDb.DatabaseSqlConfiguration),
		connectMux:  new(sync.RWMutex),
	}
	if mys.error = kitModuleCfg.Get().
		ConfigurationCopyByObject(mys.databaseSql); mys.error != nil {
		return mys
	}
	mys.cfg = &mys.databaseSql.SqlDB
	// Запуск подключения в отдельном процессе.
	onDone = make(chan struct{})
	go mys.makeConnect(onDone)
	// Ожидание запуска процесса подключения и настройки соединения.
	<-onDone
	close(onDone)
	runtime.SetFinalizer(mys, destructor)

	return
}

// Деструктор объекта. Закрывает соединение с базой данных.
func destructor(mys *impl) {
	if mys.connect == nil {
		return
	}
	_ = mys.Close()
}

// Выполнение подключения к базе данных и настройки соединения.
func (mys *impl) makeConnect(onDone chan<- struct{}) {
	var err error

	mys.connectMux.Lock()
	defer mys.connectMux.Unlock()
	onDone <- struct{}{}
	// Создание DSN для подключения к базе данных.
	if mys.error = mys.makeDsn(); mys.error != nil {
		return
	}
	// Выполнение подключения к базе данных.
	if mys.connect, err = sql.Open(mys.cfg.Driver, mys.dsn); err != nil {
		mys.error = mys.Errors().ConnectError(0, err)
		return
	}
	// Настройка подключения к базе данных.
	mys.connect.SetConnMaxIdleTime(mys.cfg.MaxIdleTimeConn)
	mys.connect.SetConnMaxLifetime(mys.cfg.MaxLifetimeConn)
	mys.connect.SetMaxIdleConns(mys.cfg.MaxIdleConn)
	mys.connect.SetMaxOpenConns(mys.cfg.MaxOpenConn)
}

// Close Закрытие соединения с базой данных.
func (mys *impl) Close() (err error) {
	mys.connectMux.Lock()
	defer mys.connectMux.Unlock()
	if mys.connect == nil {
		return
	}
	err = mys.connect.Close()
	mys.connect = nil

	return
}

// Возвращает соединение с базой данных, если соединение отсутствовало, восстанавливает его.
func (mys *impl) conn() *sql.DB {
	var onDone chan struct{}

	if mys.connect != nil {
		return mys.connect
	}
	// Запуск подключения в отдельном процессе.
	onDone = make(chan struct{})
	go mys.makeConnect(onDone)
	// Ожидание запуска процесса подключения и настройки соединения.
	<-onDone
	close(onDone)
	mys.connectMux.RLock()
	mys.connectMux.RUnlock()

	return mys.connect
}

// Возвращает последнюю ошибку работы с соединением базы данных с проверкой блокировки.
func (mys *impl) err() error {
	mys.connectMux.RLock()
	defer mys.connectMux.RUnlock()
	return mys.error
}

// Ссылка на менеджер логирования, для удобного использования внутри компоненты или модуля.
func (mys *impl) log() kitTypes.Logger { return kitModuleCfg.Get().Log() }

// Errors Справочник всех ошибок пакета.
func (mys *impl) Errors() *Error { return Errors() }

// E Ошибка соединения с базой данных.
// Если err==nil - база данных доступна, соединение активно, ошибок нет.
// Если err!=nil - есть проблема с соединением с базой данных.
func (mys *impl) E() (err error) {
	if err = mys.err(); err != nil {
		return
	}
	if err = mys.conn().Ping(); err != nil {
		return
	}

	return
}

// Status Возвращает состояние подключения к базе данных.
func (mys *impl) Status() (ret *sql.DBStats) {
	var stats sql.DBStats

	stats = mys.conn().Stats()
	ret = &stats

	return
}

// SqlDB Настроенный и готовый к работе бассейн соединений database/sql.
// Если возвращается nil - есть ошибки, ошибка доступна в функции E()
func (mys *impl) SqlDB() (ret *sql.DB) { ret = mys.conn(); return }

// GormDB Настроенный и готовый к работе объект ORM gorm.io/gorm.
// Если возвращается nil - есть ошибки, ошибка доступна в функции E()
func (mys *impl) GormDB() (ret *gorm.DB) {
	if mys.err() != nil {
		return
	}
	switch mys.cfg.Driver {
	case driverMySQL:
		ret, mys.error = gorm.Open(mysql.New(mysql.Config{
			DriverName:               mys.cfg.Driver,
			DSN:                      mys.dsn,
			Conn:                     mys.conn(),
			DefaultStringSize:        mys.cfg.DefaultStringSize,
			DisableDatetimePrecision: mys.cfg.DisableDatetimePrecision,
		}), &gorm.Config{
			SkipDefaultTransaction: mys.cfg.SkipDefaultTransaction,
			DisableAutomaticPing:   mys.cfg.DisableAutomaticPing,
			PrepareStmt:            mys.cfg.PrepareStmt,
			CreateBatchSize:        mys.cfg.CreateBatchSize,
			Logger:                 NewLoggerGorm(mys),
		})
	case driverPostgreSQL:
		ret, mys.error = gorm.Open(postgres.New(postgres.Config{
			DriverName:           mys.cfg.Driver,
			DSN:                  mys.dsn,
			Conn:                 mys.conn(),
			PreferSimpleProtocol: mys.cfg.PostgreSQLPreferSimpleProtocol,
		}), &gorm.Config{
			SkipDefaultTransaction: mys.cfg.SkipDefaultTransaction,
			DisableAutomaticPing:   mys.cfg.DisableAutomaticPing,
			PrepareStmt:            mys.cfg.PrepareStmt,
			CreateBatchSize:        mys.cfg.CreateBatchSize,
			Logger:                 NewLoggerGorm(mys),
		})
	//case driverSqlite:
	//	ret, mys.error = gorm.Open(sqlite.Open(mys.dsn), &gorm.Config{
	//		SkipDefaultTransaction: mys.cfg.SkipDefaultTransaction,
	//		DisableAutomaticPing:   mys.cfg.DisableAutomaticPing,
	//		PrepareStmt:            mys.cfg.PrepareStmt,
	//		CreateBatchSize:        mys.cfg.CreateBatchSize,
	//		Logger:                 NewLoggerGorm(mys),
	//	})
	default:
		mys.error = mys.Errors().DriverUnImplemented(0, mys.cfg.Driver)
		return
	}

	return
}

// SqlxDB Настроенный и готовый к работе объект обёртки над соединением с БД github.com/jmoiron/sqlx.
// Если возвращается nil - есть ошибки, ошибка доступна в функции E()
func (mys *impl) SqlxDB() (ret *sqlx.DB) {
	if mys.err() != nil {
		return
	}
	ret = sqlx.NewDb(mys.conn(), mys.cfg.Driver)

	return
}

package connector

import (
	"database/sql"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	log "github.com/webnice/lv2"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// gorm dependencies
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlserver"
)

var singleton *impl

// Interface is an interface of connector
type Interface interface {
	Open(string, string) error
	IsOpened() bool
	Close() error
	Gist() *gorm.DB
}

// impl is an implementation of connector
type impl struct {
	sync.RWMutex
	Gorm    *gorm.DB // Само соединение
	Counter int64    // Счётчик для подсчёта Open/Close. Open +1, Close -1. Если <= 0 то делается закрытие соединения
	Debug   bool     // Для отладки
	Dialect string   // Database dialect identificator
	Dsn     string   // Database DSN
}

// New Create new object
func New() Interface {
	if singleton == nil {
		singleton = new(impl)
		runtime.SetFinalizer(singleton, destructor)
	}
	// singleton.Debug = true
	return singleton
}

func destructor(conn *impl) {
	if conn.Debug {
		log.Notice(" --- Destroy ImplementationConnection object")
	}
	if conn.Gorm == nil {
		return
	}
	conn.Gorm = nil
}

// Open database connection
func (conn *impl) Open(dialect string, dsn string) (err error) {
	const (
		defaultStringSize      = 256
		defaultCreateBatchSize = 100
	)
	var (
		obj       *gorm.DB
		dialector gorm.Dialector
		config    *gorm.Config
		sqlDB     *sql.DB
	)

	conn.RLock()
	defer conn.RUnlock()

	conn.Dialect, conn.Dsn = dialect, dsn
	if conn.Gorm != nil {
		v := atomic.AddInt64(&conn.Counter, 1)
		if conn.Debug {
			log.Noticef(" * Already open (%v)", v)
		}
		return
	}

	dialector = mysql.New(mysql.Config{
		DriverName:        conn.Dialect,
		DSN:               conn.Dsn,
		DefaultStringSize: defaultStringSize,
	})
	_ = sqlDB
	config = &gorm.Config{
		Logger:          logger.Default.LogMode(logger.Silent),
		CreateBatchSize: defaultCreateBatchSize,
	}
	if conn.Debug {
		config.Logger = logger.Default.LogMode(logger.Info)
	}
	if obj, err = gorm.Open(dialector, config); err != nil {
		return
	}
	if obj == nil {
		err = fmt.Errorf("db connection object is nil")
		return
	}
	if obj.Logger = obj.Logger.LogMode(logger.Silent); conn.Debug {
		config.Logger = logger.Default.LogMode(logger.Info)
		obj.Logger = obj.Logger.LogMode(logger.Info)
	}
	//if db, err = obj.DB(); err == nil {
	//	db.SetMaxIdleConns(10)
	//	db.SetMaxOpenConns(100)
	//	db.SetConnMaxIdleTime(time.Minute / 4)
	//	db.SetConnMaxLifetime(time.Minute / 2)
	//}
	conn.Gorm = obj
	atomic.AddInt64(&conn.Counter, 1)
	if conn.Debug {
		log.Noticef(" + Real open (%v)", int64(1))
	}

	return
}

// IsOpened return true if connection already open
func (conn *impl) IsOpened() (ret bool) {
	if conn.Gorm == nil {
		return
	}
	if v := atomic.LoadInt64(&conn.Counter); v > 0 {
		ret = true
	}
	return
}

// Close database connection
func (conn *impl) Close() (err error) {
	conn.RLock()
	defer conn.RUnlock()

	atomic.AddInt64(&conn.Counter, -1)

	if conn.Gorm == nil {
		atomic.StoreInt64(&conn.Counter, 0)
		if conn.Debug {
			log.Noticef(" * Already close (%v)", 0)
		}
		return
	}

	if conn.Gorm == nil {
		atomic.StoreInt64(&conn.Counter, 0)
		if conn.Debug {
			log.Noticef(" * Already close (%v)", 0)
		}
		return
	}

	if v := atomic.LoadInt64(&conn.Counter); v <= 0 {
		//		err = conn.Gorm.DB().Close()
		if conn.Debug {
			log.Noticef(" - Real close (%v)", v)
		}
		//		conn.Gorm = nil
	} else {
		if conn.Debug {
			log.Noticef(" - Fake close (%v)", v)
		}
	}

	return
}

// Gist return database object
func (conn *impl) Gist() *gorm.DB {
	var ok bool

	if conn.IsOpened() {
		ok = true
	}
	if !ok {
		if err := conn.Open(conn.Dialect, conn.Dsn); err != nil {
			if conn.Debug {
				log.Errorf("Gist() Open(): %s", err.Error())
			}
		}
	}
	if conn.Debug && conn.Gorm == nil {
		log.Alertf("Gist(%v) == nil: %v", conn.Counter, conn.Gorm != nil)
	}

	return conn.Gorm
}

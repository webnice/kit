package connector

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	log "github.com/webnice/lv2"

	clickhouse "github.com/ClickHouse/clickhouse-go"
	"github.com/jmoiron/sqlx"
)

var singleton *impl

// Interface is an interface of connector
type Interface interface {
	// Open database connection
	Open(dsn string) error

	// IsOpened return true if connection already open
	IsOpened() bool

	// Close database connection
	Close() error

	// Gist return database object
	Gist() *sqlx.DB
}

// impl is an implementation of connector
type impl struct {
	sync.RWMutex
	Db      *sqlx.DB // Соединение с базой данных через sqlx
	Counter int64    // Счётчик для подсчёта Open/Close. Open +1, Close -1. Если <= 0 то делается закрытие соединения
	Dsn     string   // Database data source name (DSN)
	Debug   bool     // Для отладки
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
	if conn.Db == nil {
		return
	}
	if conn.Db.Ping() == nil {
		if err := conn.Db.Close(); err != nil {
			log.Errorf("Error close database connection: %s", err.Error())
		}
	}
	if err := conn.Db.Close(); err != nil {
		log.Errorf("Error close gorm object: %s", err.Error())
	}
	conn.Db = nil
}

// Open database connection
func (conn *impl) Open(dsn string) (err error) {
	const dialect = `clickhouse`
	var obj *sqlx.DB

	conn.RLock()
	defer conn.RUnlock()
	conn.Dsn = dsn

	if conn.Db != nil {
		v := atomic.AddInt64(&conn.Counter, 1)
		if conn.Debug {
			log.Noticef(" * Already open (%v)", v)
		}
		return
	}
	if obj, err = sqlx.Open(dialect, conn.Dsn); err != nil {
		return
	} else if obj == nil {
		err = fmt.Errorf("db connection object is nil")
		return
	}
	conn.Db = obj
	atomic.AddInt64(&conn.Counter, 1)
	if conn.Debug {
		log.Noticef(" + Real open (%v)", int64(1))
	}

	return
}

// IsOpened return true if connection already open
func (conn *impl) IsOpened() (ret bool) {
	if conn.Db == nil {
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

	if conn.Db == nil {
		atomic.StoreInt64(&conn.Counter, 0)
		if conn.Debug {
			log.Noticef(" * Already close (%v)", 0)
		}
		return
	}

	if conn.Db == nil {
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
func (conn *impl) Gist() *sqlx.DB {
	const dialect = `clickhouse`
	var ok bool

	if conn.IsOpened() {
		if err := conn.Db.Ping(); err != nil {
			if exception, ok := err.(*clickhouse.Exception); ok {
				if conn.Debug {
					log.Errorf("Gist() Ping(): [%d] %s:\n%s", exception.Code, exception.Message, exception.StackTrace)
				}
			} else {
				if conn.Debug {
					log.Errorf("Gist() Ping(): %s", err.Error())
				}
			}
		} else {
			ok = true
		}

	}
	if !ok {
		if err := conn.Open(conn.Dsn); err != nil {
			if conn.Debug {
				log.Errorf("Gist() Open(): %s", err.Error())
			}
		}
	}
	if conn.Debug && conn.Db == nil {
		log.Alertf("Gist(%v) == nil: %v", conn.Counter, conn.Db != nil)
	}

	return conn.Db
}

package connector

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/webnice/log/v2"

	"github.com/jinzhu/gorm"

	// gorm dependencies
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var singleton *impl

// Interface is an interface of connector
type Interface interface {
	Open(string, ...interface{}) error
	IsOpened() bool
	Close() error
	Gist() *gorm.DB
}

// impl is an implementation of connector
type impl struct {
	sync.RWMutex
	Gorm    *gorm.DB      // Само соединение
	Counter int64         // Счётчик для подсчёта Open/Close. Open +1, Close -1. Если <= 0 то делается закрытие соединения
	Debug   bool          // Для отладки
	Dialect string        // Database dialect identificator
	Args    []interface{} // Database arguments
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
	if conn.Gorm.DB().Ping() == nil {
		if err := conn.Gorm.DB().Close(); err != nil {
			log.Errorf("Error close database connection: %s", err.Error())
		}
	}
	if err := conn.Gorm.Close(); err != nil {
		log.Errorf("Error close gorm object: %s", err.Error())
	}
	conn.Gorm = nil
}

// Open database connection
func (conn *impl) Open(dialect string, args ...interface{}) (err error) {
	var obj *gorm.DB

	conn.RLock()
	defer conn.RUnlock()

	conn.Dialect = dialect
	conn.Args = args

	if conn.Gorm != nil {
		v := atomic.AddInt64(&conn.Counter, 1)
		if conn.Debug {
			log.Noticef(" * Already open (%v)", v)
		}
		return
	}

	if obj, err = gorm.Open(conn.Dialect, conn.Args...); err != nil {
		return
	} else if obj == nil {
		err = fmt.Errorf("db connection object is nil")
		return
	}

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

	if conn.Gorm.DB() == nil {
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
		err := conn.Gorm.DB().Ping()
		if err != nil {
			if conn.Debug {
				log.Errorf("Gist() Ping(): %s", err.Error())
			}
		} else {
			ok = true
		}
	}
	if !ok {
		if err := conn.Open(conn.Dialect, conn.Args...); err != nil {
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

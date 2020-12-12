package connector

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/webnice/kit/modules/dbtarantool/tarantool"
	log "github.com/webnice/lv2"
)

var singleton *impl

// Interface is an interface of connector
type Interface interface {
	// Open database connection
	Open(network string, host string, port uint16, opt *tarantool.Options) error

	// IsOpened return true if connection already open
	IsOpened() bool

	// Close database connection
	Close() error

	// Gist return database object
	Gist() *tarantool.Connection
}

// impl is an implementation of connector
type impl struct {
	sync.RWMutex
	Tarantool *tarantool.Connection // Само соединение
	Counter   int64                 // Счётчик для подсчёта Open/Close. Open +1, Close -1. Если <= 0 то делается закрытие соединения
	Debug     bool                  // Для отладки
	Network   string                // Сеть подключения. tcp, socket и т.п. Как описано https://golang.org/pkg/net/#Dial
	Host      string                // Хост подключения
	Port      uint16                // Порт подключения
	Opt       *tarantool.Options    // Опции подключения
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
	var err error
	var ok bool

	if conn.Debug {
		log.Notice(" --- Destroy connector object")
	}
	if conn.Tarantool == nil {
		return
	}
	if ok, err = conn.Tarantool.IsClosed(); !ok && err == nil {
		conn.Tarantool.Close()
	}
	conn.Tarantool = nil
}

// Open database connection
func (conn *impl) Open(network string, host string, port uint16, opt *tarantool.Options) (err error) {
	var obj *tarantool.Connection
	var hps string
	var ok bool

	conn.RLock()
	defer conn.RUnlock()

	conn.Network, conn.Host, conn.Port, conn.Opt = network, host, port, opt
	if conn.Tarantool != nil {
		if ok, err = conn.Tarantool.IsClosed(); err == nil && !ok {
			v := atomic.AddInt64(&conn.Counter, 1)
			if conn.Debug {
				log.Noticef(" * Already open (%v)", v)
			}
			return
		}
	}
	if port != 0 {
		hps = fmt.Sprintf("%s:%d", host, port)
	} else {
		hps = host
	}
	if obj, err = tarantool.Connect(conn.Network, hps, conn.Opt); err != nil {
		return
	} else if obj == nil {
		err = fmt.Errorf("db connection object is nil")
		return
	}
	conn.Tarantool = obj
	atomic.AddInt64(&conn.Counter, 1)
	if conn.Debug {
		log.Noticef(" + Real open (%v)", int64(1))
	}

	return
}

// IsOpened return true if connection already open
func (conn *impl) IsOpened() (ret bool) {
	if conn.Tarantool == nil {
		return
	}
	if ok, err := conn.Tarantool.IsClosed(); err != nil || ok {
		return
	}
	if v := atomic.LoadInt64(&conn.Counter); v > 0 {
		ret = true
	}
	return
}

// Close database connection
func (conn *impl) Close() (err error) {
	var ok bool

	conn.RLock()
	defer conn.RUnlock()

	atomic.AddInt64(&conn.Counter, -1)
	if conn.Tarantool == nil {
		atomic.StoreInt64(&conn.Counter, 0)
		if conn.Debug {
			log.Noticef(" * Already close (%v)", 0)
		}
		return
	}
	if ok, err = conn.Tarantool.IsClosed(); ok || err != nil {
		atomic.StoreInt64(&conn.Counter, 0)
		if conn.Debug {
			log.Noticef(" * Already close (%v)", 0)
		}
		return
	}
	if v := atomic.LoadInt64(&conn.Counter); v <= 0 {
		conn.Tarantool.Close()
		if conn.Debug {
			log.Noticef(" - Real close (%v)", v)
		}
		conn.Tarantool = nil
	} else {
		if conn.Debug {
			log.Noticef(" - Fake close (%v)", v)
		}
	}

	return
}

// Gist return database object
func (conn *impl) Gist() *tarantool.Connection {
	var err error
	var ok bool

	if conn.IsOpened() {
		ok, err = conn.Tarantool.IsClosed()
		ok = !ok
		if err != nil {
			if conn.Debug {
				log.Errorf("Gist() IsClosed(): %s", err.Error())
			}
			ok = false
		}
	}
	if !ok {
		if err := conn.Open(conn.Network, conn.Host, conn.Port, conn.Opt); err != nil {
			if conn.Debug {
				log.Errorf("Gist() Open(): %s", err.Error())
			}
		}
	}
	if conn.Debug && conn.Tarantool == nil {
		log.Alertf("Gist(%v) == nil: %v", conn.Counter, conn.Tarantool != nil)
	}

	return conn.Tarantool
}

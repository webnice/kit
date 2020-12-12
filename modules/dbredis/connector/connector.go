package connector

import (
	"runtime"
	"sync"
	"sync/atomic"

	log "github.com/webnice/lv2"

	redis "github.com/go-redis/redis/v7"
)

var singleton *impl

// Interface is an interface of connector
type Interface interface {
	// Open database connection
	Open(opt *redis.Options) error

	// IsOpened return true if connection already open
	IsOpened() bool

	// Close database connection
	Close() error

	// Gist return database object
	Gist() *redis.Client
}

// impl is an implementation of connector
type impl struct {
	sync.RWMutex
	Redis   *redis.Client  // Само соединение
	Counter int64          // Счётчик для подсчёта Open/Close. Open +1, Close -1. Если <= 0 то делается закрытие соединения
	Debug   bool           // Для отладки
	Opt     *redis.Options // Опции подключения
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

	if conn.Debug {
		log.Notice("--- Destroy connector object")
	}
	if conn.Redis == nil {
		return
	}
	if _, err = conn.Redis.Ping().Result(); err == nil {
		_ = conn.Redis.Close()
	}
	conn.Redis = nil
}

// Open database connection
func (conn *impl) Open(opt *redis.Options) (err error) {
	var obj *redis.Client

	conn.RLock()
	defer conn.RUnlock()
	conn.Opt = opt
	if conn.Redis != nil {
		if _, err = conn.Redis.Ping().Result(); err == nil {
			v := atomic.AddInt64(&conn.Counter, 1)
			if conn.Debug {
				log.Noticef("* already open (%v)", v)
			}
			return
		}
	}
	obj = redis.NewClient(conn.Opt)
	conn.Redis = obj
	atomic.AddInt64(&conn.Counter, 1)
	if conn.Debug {
		log.Noticef(" + Real open (%v)", int64(1))
	}

	return
}

// IsOpened return true if connection already open
func (conn *impl) IsOpened() (ret bool) {
	if conn.Redis == nil {
		return
	}
	if _, err := conn.Redis.Ping().Result(); err != nil {
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
	if conn.Redis == nil {
		atomic.StoreInt64(&conn.Counter, 0)
		if conn.Debug {
			log.Noticef("* Already close (%v)", 0)
		}
		return
	}
	if _, err = conn.Redis.Ping().Result(); err != nil {
		atomic.StoreInt64(&conn.Counter, 0)
		if conn.Debug {
			log.Noticef("* Already close (%v)", 0)
		}
		return
	}
	if v := atomic.LoadInt64(&conn.Counter); v <= 0 {
		_ = conn.Redis.Close()
		if conn.Debug {
			log.Noticef("- Real close (%v)", v)
		}
		conn.Redis = nil
	} else {
		if conn.Debug {
			log.Noticef("- Fake close (%v)", v)
		}
	}

	return
}

// Gist return database object
func (conn *impl) Gist() *redis.Client {
	var (
		err error
		ok  bool
	)

	if conn.IsOpened() {
		ok = true
		if _, err = conn.Redis.Ping().Result(); err != nil {
			if conn.Debug {
				log.Errorf("Gist() IsClosed(): %s", err.Error())
			}
			ok = false
		}
	}
	if !ok {
		if err := conn.Open(conn.Opt); err != nil {
			if conn.Debug {
				log.Errorf("Gist() Open(): %s", err.Error())
			}
		}
	}
	if conn.Debug && conn.Redis == nil {
		log.Alertf("Gist(%v) == nil: %v", conn.Counter, conn.Redis != nil)
	}

	return conn.Redis
}

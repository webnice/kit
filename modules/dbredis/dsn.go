package dbredis // import "gopkg.in/webnice/kit.v1/modules/dbredis"

//import "gopkg.in/webnice/debug.v1"
import "gopkg.in/webnice/log.v2"
import (
	"fmt"

	redis "github.com/go-redis/redis/v7"
)

// MakeConnectArgs Создание параметров подключения к базе данных
func (tt *Implementation) MakeConnectArgs() (opt *redis.Options) {
	const (
		unix = `unix`
		tcp  = `tcp`
	)
	var addr string

	if tt.cnf == nil {
		tt.cnf = defaultConfiguration
	}
	if tt.cnf == nil {
		log.Fatalf("redis configuration is empty")
		return
	}
	Defaults(tt.cnf)
	switch tt.cnf.Type {
	case unix:
		addr = tt.cnf.Socket
	case tcp:
		addr = fmt.Sprintf("%s:%d", tt.cnf.Host, tt.cnf.Port)
	}
	tt.opt = &redis.Options{
		Network:            tt.cnf.Type,
		Addr:               addr,
		Password:           tt.cnf.Password,
		DB:                 int(tt.cnf.Database),
		MaxRetries:         int(tt.cnf.MaxRetries),
		MinRetryBackoff:    tt.cnf.MinRetryBackoff,
		MaxRetryBackoff:    tt.cnf.MaxRetryBackoff,
		DialTimeout:        tt.cnf.DialTimeout,
		ReadTimeout:        tt.cnf.ReadTimeout,
		WriteTimeout:       tt.cnf.WriteTimeout,
		PoolSize:           int(tt.cnf.PoolSize),
		MinIdleConns:       int(tt.cnf.MinIdleConns),
		MaxConnAge:         tt.cnf.MaxConnAge,
		PoolTimeout:        tt.cnf.PoolTimeout,
		IdleTimeout:        tt.cnf.IdleTimeout,
		IdleCheckFrequency: tt.cnf.IdleCheckFrequency,
	}
	opt = tt.opt

	return
}

package dbtt // import "gopkg.in/webnice/kit.v1/modules/dbtt"

//import "gopkg.in/webnice/debug.v1"
import "gopkg.in/webnice/log.v2"
import (
	"gopkg.in/webnice/kit.v1/modules/dbtt/tarantool"
)

// MakeConnectArgs Создание параметров подключения к базе данных
func (tt *Implementation) MakeConnectArgs() (network string, host string, port uint16, opt *tarantool.Options) {
	if tt.cnf == nil {
		tt.cnf = defaultConfiguration
	}
	if tt.cnf == nil {
		log.Fatalf("Tarantool configuration is empty")
		return
	}
	tt.opt = &tarantool.Options{
		ConnectTimeout: tt.cnf.ConnectTimeout,
		QueryTimeout:   tt.cnf.QueryTimeout,
		DefaultSpace:   tt.cnf.DefaultSpace,
		User:           tt.cnf.Login,
		Password:       tt.cnf.Password,
	}
	network, host, port, opt = tt.cnf.Type, tt.cnf.Host, tt.cnf.Port, tt.opt

	return
}

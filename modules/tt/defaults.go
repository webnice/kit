package tt

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"strings"
	"time"
)

// Defaults Проверка конфигурации и установка значений по умолчанию
func Defaults(cnf *Configuration) {
	const (
		socket = `socket`
		tcp    = `tcp`
	)
	if cnf.Host == "" {
		cnf.Host = "localhost"
	}
	if cnf.Port == 0 {
		cnf.Port = 3301
	}
	switch strings.ToLower(cnf.Type) {
	case socket:
		cnf.Type = socket
	default:
		cnf.Type = tcp
	}
	if cnf.ConnectTimeout == 0 {
		cnf.ConnectTimeout = time.Second
	}
}

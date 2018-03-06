package db // import "gopkg.in/webnice/kit.v1/modules/db"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"strings"
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
	if cnf.Driver == "" {
		cnf.Driver = `mysql`
	}
	if cnf.Port == 0 {
		cnf.Port = 3306
	}
	switch strings.ToLower(cnf.Type) {
	case socket:
		cnf.Type = socket
	default:
		cnf.Type = tcp
	}
	if cnf.Name == "" {
		cnf.Name = `test`
	}
	if cnf.Login == "" {
		cnf.Login = `root`
	}
	if cnf.Charset == "" {
		cnf.Charset = `utf8`
	}
}

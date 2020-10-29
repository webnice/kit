package dbredis

import "strings"

// Defaults Проверка конфигурации и установка значений по умолчанию
func Defaults(cnf *Configuration) {
	const (
		socket = `socket`
		unix   = `unix`
		tcp    = `tcp`
	)
	if cnf.Host == "" {
		cnf.Host = "localhost"
	}
	if cnf.Port == 0 {
		cnf.Port = 6379
	}
	switch strings.ToLower(cnf.Type) {
	case socket, unix:
		cnf.Type = unix
	default:
		cnf.Type = tcp
	}
}

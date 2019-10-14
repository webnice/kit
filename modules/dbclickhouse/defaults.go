package dbclickhouse // import "gopkg.in/webnice/kit.v1/modules/dbclickhouse"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"strings"
)

// Defaults Проверка конфигурации и установка значений по умолчанию
func Defaults(cnf *Configuration) {
	// Логин
	if cnf.Login == "" {
		cnf.Login = `default`
	}
	// База данных
	if cnf.Database == "" {
		cnf.Database = `default`
	}
	// Стратегия подключения
	switch strings.ToLower(cnf.OpenStrategy) {
	case `random`, `in_order`:
		cnf.OpenStrategy = strings.ToLower(cnf.OpenStrategy)
	default:
		cnf.OpenStrategy = `random`
	}
	// Максимальное количество строк в блоке
	if cnf.BlockSize == 0 {
		cnf.BlockSize = 1000000
	}
}

package db // import "gopkg.in/webnice/kit.v1/modules/db"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"fmt"
	"strings"
)

// MakeDsn Создание строки подключения к базе данных для database/sql подобных драйверов
func (db *Implementation) MakeDsn() (err error) {
	var cnf *Configuration = db.cnf

	if cnf == nil {
		err = fmt.Errorf("Configuration is empty")
		return
	}

	// Driver
	db.drv = strings.ToLower(cnf.Driver)
	if !strings.EqualFold(db.drv, "mysql") &&
		!strings.EqualFold(db.drv, "postgres") &&
		!strings.EqualFold(db.drv, "sqlite3") &&
		!strings.EqualFold(db.drv, "mssql") {
		err = fmt.Errorf("Unknown database driver '%s'", cnf.Driver)
		return
	}

	// Login and password
	if cnf.Login == "" {
		err = fmt.Errorf("Database username(login) can not be empty")
		return
	}
	db.dsn = fmt.Sprintf("%s:%s",
		cnf.Login,
		cnf.Password,
	)

	// Connection type and host, port, socket
	if strings.EqualFold(cnf.Type, "tcp") {
		db.dsn += fmt.Sprintf("@tcp(%s:%d)", cnf.Host, cnf.Port)
	} else if strings.EqualFold(cnf.Type, "socket") {
		db.dsn += fmt.Sprintf("@unix(%s)", cnf.Socket)
	} else {
		err = fmt.Errorf("Wrong connection type '%s'", cnf.Type)
		return
	}

	// Database name
	db.dsn += fmt.Sprintf("/%s?parseTime=True&loc=Local", cnf.Name)

	// Charset
	if cnf.Charset != "" {
		db.dsn += fmt.Sprintf("&charset=%s", cnf.Charset)
	}

	return
}

// Dsn Return DSN string
func (db *Implementation) Dsn() (ret string, err error) {
	if db.dsn != "" {
		ret = db.dsn
		return
	}
	if err = db.MakeDsn(); err != nil {
		return
	}
	ret = db.dsn
	return
}

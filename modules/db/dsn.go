package db // import "github.com/webnice/kit/v1/modules/db"

import (
	"fmt"
	"strings"
)

// MakeDsn Создание строки подключения к базе данных для database/sql подобных драйверов
func (db *Implementation) MakeDsn() (err error) {
	const (
		errNoConfiguration       = `configuration is empty`
		errUnknownDatabaseDriver = `unknown database driver '%s'`
		errNoUsername            = `database username(login) can not be empty`
		errWrongConnectionType   = `wrong connection type '%s'`
		keyMysql                 = `mysql`
		keyPostgres              = `postgres`
		keySqlite3               = `sqlite3`
		keyMssql                 = `mssql`
		keyTimeSettings          = `parseTime=True&loc=Local`
		keyCharsetSettings       = `&charset=%s`
		keyTcp                   = `tcp`
		keySocket                = `socket`
		keyUnix                  = `@unix(%s)`
	)
	var cnf *Configuration = db.cnf

	if cnf == nil {
		cnf = defaultConfiguration
	}
	if cnf == nil {
		err = fmt.Errorf(errNoConfiguration)
		return
	}
	// Driver
	db.drv = strings.ToLower(cnf.Driver)
	if !strings.EqualFold(db.drv, keyMysql) &&
		!strings.EqualFold(db.drv, keyPostgres) &&
		!strings.EqualFold(db.drv, keySqlite3) &&
		!strings.EqualFold(db.drv, keyMssql) {
		err = fmt.Errorf(errUnknownDatabaseDriver, cnf.Driver)
		return
	}
	// sqlite3
	if strings.EqualFold(db.drv, keySqlite3) {
		db.dsn += fmt.Sprintf("%s?%s", cnf.Name, keyTimeSettings)
		return
	}
	// Login and password
	if cnf.Login == "" {
		err = fmt.Errorf(errNoUsername)
		return
	}
	db.dsn = fmt.Sprintf("%s:%s",
		cnf.Login,
		cnf.Password,
	)
	// Connection type and host, port, socket
	switch strings.ToLower(cnf.Type) {
	case keyTcp:
		db.dsn += fmt.Sprintf("@%s(%s:%d)", keyTcp, cnf.Host, cnf.Port)
	case keySocket:
		db.dsn += fmt.Sprintf(keyUnix, cnf.Socket)
	default:
		err = fmt.Errorf(errWrongConnectionType, cnf.Type)
		return
	}
	// Database name
	db.dsn += fmt.Sprintf("/%s?%s", cnf.Name, keyTimeSettings)
	// Charset
	if cnf.Charset != "" {
		db.dsn += fmt.Sprintf(keyCharsetSettings, cnf.Charset)
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

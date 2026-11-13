package sql

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Создание DSN для подключения к базе данных.
func (mys *impl) makeDsn() (err error) {
	const keyTcp, keySocket = "tcp", "socket"
	var (
		n     int
		found bool
	)

	// Проверка конфигурации.
	if mys.cfg == nil {
		err = mys.Errors().ConfigurationIsEmpty.Bind()
		return
	}
	// Проверка драйвера базы данных.
	for n = range supportDrivers {
		if strings.EqualFold(mys.cfg.Driver, supportDrivers[n]) {
			mys.cfg.Driver, found = supportDrivers[n], true
			break
		}
	}
	if !found {
		err = mys.Errors().UnknownDatabaseDriver.Bind(mys.cfg.Driver)
		return
	}

	switch mys.cfg.Driver {
	case driverMySQL:
		err = mys.makeDsnMySQL(keyTcp, keySocket)
	case driverPostgreSQL:
		err = mys.makeDsnPostgreSQL(keyTcp, keySocket)
	}

	return
}

// Создание DSN для подключения к серверу MySQL.
func (mys *impl) makeDsnMySQL(keyTcp, keySocket string) (err error) {
	var (
		cfgMySQL = mysql.NewConfig()
		loc      *time.Location
	)

	if mys.cfg.Login == "" {
		err = mys.Errors().UsernameIsEmpty.Bind()
		return
	}
	cfgMySQL.User = mys.cfg.Login
	cfgMySQL.Passwd = mys.cfg.Password
	cfgMySQL.DBName = mys.cfg.Name
	cfgMySQL.ParseTime = mys.cfg.ParseTime
	cfgMySQL.Params = make(map[string]string)
	if mys.cfg.Charset != "" {
		cfgMySQL.Params["charset"] = mys.cfg.Charset
	}
	if mys.cfg.TimezoneLocation != "" {
		if loc, err = time.LoadLocation(mys.cfg.TimezoneLocation); err != nil {
			err = mys.Errors().ConnectError.Bind(err)
			return
		}
		cfgMySQL.Loc = loc
	}
	switch strings.ToLower(mys.cfg.Type) {
	case keyTcp:
		cfgMySQL.Net = keyTcp
		cfgMySQL.Addr = fmt.Sprintf("%s:%d", mys.cfg.Host, mys.cfg.Port)
	case keySocket:
		cfgMySQL.Net = "unix"
		cfgMySQL.Addr = mys.cfg.Socket
	default:
		err = mys.Errors().WrongConnectionType.Bind(mys.cfg.Type)
		return
	}
	mys.cfg.Type = strings.ToLower(mys.cfg.Type)
	mys.dsn = cfgMySQL.FormatDSN()

	return
}

// Создание DSN для подключения к серверу PostgreSQL.
func (mys *impl) makeDsnPostgreSQL(keyTcp, keySocket string) (err error) {
	var (
		host string
		keys = make([]string, 0, 6)
	)

	switch strings.ToLower(mys.cfg.Type) {
	case keyTcp:
		host = mys.cfg.Host
	case keySocket:
		host = mys.cfg.Socket
	default:
		err = mys.Errors().WrongConnectionType.Bind(mys.cfg.Type)
		return
	}
	mys.cfg.Type = strings.ToLower(mys.cfg.Type)
	keys = append(keys,
		fmt.Sprintf("host=%s", postgresQuoteValue(host)),
		fmt.Sprintf("port=%d", mys.cfg.Port),
		fmt.Sprintf("user=%s", postgresQuoteValue(mys.cfg.Login)),
		fmt.Sprintf("password=%s", postgresQuoteValue(mys.cfg.Password)),
		fmt.Sprintf("dbname=%s", postgresQuoteValue(mys.cfg.Name)),
	)
	mys.dsn = strings.Join(keys, " ")

	return
}

// Экранирование значения keyword DSN PostgreSQL.
func postgresQuoteValue(value string) (ret string) {
	ret = strings.ReplaceAll(value, `\`, `\\`)
	ret = strings.ReplaceAll(ret, `'`, `\'`)

	return
}

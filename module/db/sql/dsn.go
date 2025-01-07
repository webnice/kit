package sql

import (
	"fmt"
	"strings"
)

// Создание DSN для подключения к базе данных.
func (mys *impl) makeDsn() (err error) {
	const keyTcp, keySocket = `tcp`, `socket`
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
	// Самая простая конфигурация: sqlite
	if mys.cfg.Driver == driverSqlite {
		mys.dsn = fmt.Sprintf("%s?%s", mys.cfg.Name, dsnTimeSettings)
		return
	}
	// Имя пользователя.
	if mys.cfg.Login == "" {
		err = mys.Errors().UsernameIsEmpty.Bind()
		return
	}
	// Имя пользователя и пароль можно добавлять в DSN.
	mys.dsn = fmt.Sprintf("%s:%s", mys.cfg.Login, mys.cfg.Password)
	// Тип подключения.
	switch strings.ToLower(mys.cfg.Type) {
	case keyTcp:
		mys.dsn += fmt.Sprintf("@%s(%s:%d)", keyTcp, mys.cfg.Host, mys.cfg.Port)
	case keySocket:
		mys.dsn += fmt.Sprintf(dsnUnixTpl, mys.cfg.Socket)
	default:
		err = mys.Errors().WrongConnectionType.Bind(mys.cfg.Type)
		return
	}
	mys.cfg.Type = strings.ToLower(mys.cfg.Type)
	// Название базы данных.
	mys.dsn += fmt.Sprintf("/%s", mys.cfg.Name)
	// Парсинг времени.
	mys.dsn += fmt.Sprintf(dsnTimeSettings, mys.cfg.ParseTime)
	// Зона времени.
	if mys.cfg.TimezoneLocation != "" {
		mys.dsn += fmt.Sprintf(dsnLocationSettings, mys.cfg.TimezoneLocation)
	}
	// Кодировка соединения с базой данных.
	if mys.cfg.Charset != "" {
		mys.dsn += fmt.Sprintf(dsnCharsetTpl, mys.cfg.Charset)
	}

	return
}

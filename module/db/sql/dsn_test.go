package sql

import (
	"strings"
	"testing"
)

func TestImpl_MakeDsnErrors(t *testing.T) {
	const (
		tplFatalNilCfg        = "makeDsn() с nil cfg вернул %v."
		tplFatalUnknownDriver = "makeDsn() с unknown driver вернул %v."
	)
	var mys impl

	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "пустая") {
		t.Fatalf(tplFatalNilCfg, err)
	}
	mys.cfg = testConfig("unknown-driver")
	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "драйвер") {
		t.Fatalf(tplFatalUnknownDriver, err)
	}
}

func TestImpl_MakeDsnMySQL(t *testing.T) {
	const (
		tplFatalMySQLError       = "makeDsn(MySQL) вернул ошибку: %v."
		tplDsnPartMissing        = "makeDsn(MySQL) = %q, ожидалась часть %q."
		tplConfigNotNormalized   = "makeDsn(MySQL) не нормализовал конфигурацию: %+v."
		tplFatalMySQLSocketError = "makeDsn(MySQL socket) вернул ошибку: %v."
		tplDsnUnexpected         = "makeDsn(MySQL socket) = %q."
		tplFatalEmptyLogin       = "makeDsn(MySQL empty login) вернул %v."
		tplFatalWrongType        = "makeDsn(MySQL wrong type) вернул %v."
		errBadLocationNoError    = "makeDsn(MySQL bad location) не вернул ошибку."
	)
	var (
		mys  impl
		part string
	)

	mys.cfg = testConfig("MYSQL")
	mys.cfg.TimezoneLocation = "Europe/Moscow"
	if err := mys.makeDsn(); err != nil {
		t.Fatalf(tplFatalMySQLError, err)
	}
	for _, part = range []string{"user:pass@tcp(localhost:3306)/appdb", "parseTime=true", "loc=Europe%2FMoscow", "charset=utf8mb4"} {
		if !strings.Contains(mys.dsn, part) {
			t.Fatalf(tplDsnPartMissing, mys.dsn, part)
		}
	}
	if mys.cfg.Driver != driverMySQL || mys.cfg.Type != "tcp" {
		t.Fatalf(tplConfigNotNormalized, mys.cfg)
	}
	mys = impl{cfg: testConfig(driverMySQL)}
	mys.cfg.Type = "socket"
	mys.cfg.Socket = "/tmp/mysql.sock"
	if err := mys.makeDsn(); err != nil {
		t.Fatalf(tplFatalMySQLSocketError, err)
	}
	if !strings.Contains(mys.dsn, "user:pass@unix(/tmp/mysql.sock)/appdb") {
		t.Fatalf(tplDsnUnexpected, mys.dsn)
	}
	mys = impl{cfg: testConfig(driverMySQL)}
	mys.cfg.Login = ""
	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "имя пользователя") {
		t.Fatalf(tplFatalEmptyLogin, err)
	}
	mys = impl{cfg: testConfig(driverMySQL)}
	mys.cfg.Type = "udp"
	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "способ подключения") {
		t.Fatalf(tplFatalWrongType, err)
	}
	mys = impl{cfg: testConfig(driverMySQL)}
	mys.cfg.TimezoneLocation = "bad/location"
	if err := mys.makeDsn(); err == nil {
		t.Fatal(errBadLocationNoError)
	}
}

func TestImpl_MakeDsnPostgreSQL(t *testing.T) {
	const (
		tplFatalPgError        = "makeDsn(PostgreSQL) вернул ошибку: %v."
		tplDsnMismatch         = "makeDsn(PostgreSQL) = %q, ожидалось %q."
		tplConfigNotNormalized = "makeDsn(PostgreSQL) не нормализовал конфигурацию: %+v."
		tplFatalSocketError    = "makeDsn(PostgreSQL socket) вернул ошибку: %v."
		tplDsnUnexpected       = "makeDsn(PostgreSQL socket) = %q."
		tplFatalQuoteError     = "makeDsn(PostgreSQL quote) вернул ошибку: %v."
		tplQuoteUnexpected     = "makeDsn(PostgreSQL quote) = %q."
		tplFatalWrongType      = "makeDsn(PostgreSQL wrong type) вернул %v."
	)
	var (
		mys      impl
		expected string
	)

	mys.cfg = testConfig("POSTGRES")
	if err := mys.makeDsn(); err != nil {
		t.Fatalf(tplFatalPgError, err)
	}
	expected = "host=localhost port=5432 user=user password=pass dbname=appdb"
	if mys.dsn != expected {
		t.Fatalf(tplDsnMismatch, mys.dsn, expected)
	}
	if mys.cfg.Driver != driverPostgreSQL || mys.cfg.Type != "tcp" {
		t.Fatalf(tplConfigNotNormalized, mys.cfg)
	}
	mys = impl{cfg: testConfig(driverPostgreSQL)}
	mys.cfg.Type = "socket"
	mys.cfg.Socket = "/tmp"
	if err := mys.makeDsn(); err != nil {
		t.Fatalf(tplFatalSocketError, err)
	}
	if !strings.Contains(mys.dsn, "host=/tmp") {
		t.Fatalf(tplDsnUnexpected, mys.dsn)
	}
	mys = impl{cfg: testConfig(driverPostgreSQL)}
	mys.cfg.Host = "local'host"
	if err := mys.makeDsn(); err != nil {
		t.Fatalf(tplFatalQuoteError, err)
	}
	if !strings.Contains(mys.dsn, `host=local\'host`) {
		t.Fatalf(tplQuoteUnexpected, mys.dsn)
	}
	mys = impl{cfg: testConfig(driverPostgreSQL)}
	mys.cfg.Type = "udp"
	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "способ подключения") {
		t.Fatalf(tplFatalWrongType, err)
	}
}

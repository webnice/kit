package sql

import (
	"strings"
	"testing"
)

func TestImpl_MakeDsnErrors(t *testing.T) {
	var mys impl

	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "пустая") {
		t.Fatalf("makeDsn() с nil cfg вернул %v", err)
	}
	mys.cfg = testConfig("unknown-driver")
	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "драйвер") {
		t.Fatalf("makeDsn() с unknown driver вернул %v", err)
	}
}

func TestImpl_MakeDsnMySQL(t *testing.T) {
	var mys impl

	mys.cfg = testConfig("MYSQL")
	mys.cfg.TimezoneLocation = "Europe/Moscow"
	if err := mys.makeDsn(); err != nil {
		t.Fatalf("makeDsn(MySQL) вернул ошибку: %v", err)
	}
	for _, part := range []string{"user:pass@tcp(localhost:3306)/appdb", "parseTime=true", "loc=Europe%2FMoscow", "charset=utf8mb4"} {
		if !strings.Contains(mys.dsn, part) {
			t.Fatalf("makeDsn(MySQL) = %q, ожидалась часть %q", mys.dsn, part)
		}
	}
	if mys.cfg.Driver != driverMySQL || mys.cfg.Type != "tcp" {
		t.Fatalf("makeDsn(MySQL) не нормализовал конфигурацию: %+v", mys.cfg)
	}

	mys = impl{cfg: testConfig(driverMySQL)}
	mys.cfg.Type = "socket"
	mys.cfg.Socket = "/tmp/mysql.sock"
	if err := mys.makeDsn(); err != nil {
		t.Fatalf("makeDsn(MySQL socket) вернул ошибку: %v", err)
	}
	if !strings.Contains(mys.dsn, "user:pass@unix(/tmp/mysql.sock)/appdb") {
		t.Fatalf("makeDsn(MySQL socket) = %q", mys.dsn)
	}

	mys = impl{cfg: testConfig(driverMySQL)}
	mys.cfg.Login = ""
	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "имя пользователя") {
		t.Fatalf("makeDsn(MySQL empty login) вернул %v", err)
	}

	mys = impl{cfg: testConfig(driverMySQL)}
	mys.cfg.Type = "udp"
	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "способ подключения") {
		t.Fatalf("makeDsn(MySQL wrong type) вернул %v", err)
	}

	mys = impl{cfg: testConfig(driverMySQL)}
	mys.cfg.TimezoneLocation = "bad/location"
	if err := mys.makeDsn(); err == nil {
		t.Fatal("makeDsn(MySQL bad location) не вернул ошибку")
	}
}

func TestImpl_MakeDsnPostgreSQL(t *testing.T) {
	var mys impl

	mys.cfg = testConfig("POSTGRES")
	if err := mys.makeDsn(); err != nil {
		t.Fatalf("makeDsn(PostgreSQL) вернул ошибку: %v", err)
	}
	expected := "host=localhost port=5432 user=user password=pass dbname=appdb"
	if mys.dsn != expected {
		t.Fatalf("makeDsn(PostgreSQL) = %q, ожидалось %q", mys.dsn, expected)
	}
	if mys.cfg.Driver != driverPostgreSQL || mys.cfg.Type != "tcp" {
		t.Fatalf("makeDsn(PostgreSQL) не нормализовал конфигурацию: %+v", mys.cfg)
	}

	mys = impl{cfg: testConfig(driverPostgreSQL)}
	mys.cfg.Type = "socket"
	mys.cfg.Socket = "/tmp"
	if err := mys.makeDsn(); err != nil {
		t.Fatalf("makeDsn(PostgreSQL socket) вернул ошибку: %v", err)
	}
	if !strings.Contains(mys.dsn, "host=/tmp") {
		t.Fatalf("makeDsn(PostgreSQL socket) = %q", mys.dsn)
	}

	mys = impl{cfg: testConfig(driverPostgreSQL)}
	mys.cfg.Host = "local'host"
	if err := mys.makeDsn(); err != nil {
		t.Fatalf("makeDsn(PostgreSQL quote) вернул ошибку: %v", err)
	}
	if !strings.Contains(mys.dsn, `host=local\'host`) {
		t.Fatalf("makeDsn(PostgreSQL quote) = %q", mys.dsn)
	}

	mys = impl{cfg: testConfig(driverPostgreSQL)}
	mys.cfg.Type = "udp"
	if err := mys.makeDsn(); err == nil || !strings.Contains(err.Error(), "способ подключения") {
		t.Fatalf("makeDsn(PostgreSQL wrong type) вернул %v", err)
	}
}

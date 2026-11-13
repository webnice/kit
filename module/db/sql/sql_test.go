package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitModuleDbSqlTypes "github.com/webnice/kit/v4/module/db/sql/types"
	kmll "github.com/webnice/kit/v4/module/log/level"
	"gorm.io/gorm"
)

const testDriverName = "kit_db_test_driver"

var (
	testDriverPingErr   error
	testDriverQueryErr  error
	testDriverExecErr   error
	testDriverQueryRows [][]driver.Value
)

func init() {
	sql.Register(testDriverName, testDriver{})
	supportDrivers = append(supportDrivers, testDriverName)
}

type testDriver struct{}

func (testDriver) Open(string) (driver.Conn, error) { return &testConn{}, nil }

type testConn struct{}

func (testConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (testConn) Close() error                        { return nil }
func (testConn) Begin() (driver.Tx, error)           { return testTx{}, nil }
func (testConn) Ping(context.Context) error          { return testDriverPingErr }

func (testConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	if testDriverQueryErr != nil {
		return nil, testDriverQueryErr
	}
	return &testRows{data: testDriverQueryRows}, nil
}

func (testConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	if testDriverExecErr != nil {
		return nil, testDriverExecErr
	}
	testDriverQueryRows = [][]driver.Value{{int64(1), true}}
	return testResult{}, nil
}

type testTx struct{}

func (testTx) Commit() error   { return nil }
func (testTx) Rollback() error { return nil }

type testResult struct{}

func (testResult) LastInsertId() (int64, error) { return 0, nil }
func (testResult) RowsAffected() (int64, error) { return 0, nil }

type testRows struct {
	data [][]driver.Value
	n    int
}

func (testRows) Columns() []string { return []string{"version_id", "is_applied"} }
func (testRows) Close() error      { return nil }

func (rows *testRows) Next(dest []driver.Value) error {
	if rows.n >= len(rows.data) {
		return io.EOF
	}
	copy(dest, rows.data[rows.n])
	rows.n++

	return nil
}

func testConfig(driver string) *kitModuleDbSqlTypes.Configuration {
	var port uint16 = 3306
	if strings.EqualFold(driver, driverPostgreSQL) {
		port = 5432
	}

	return &kitModuleDbSqlTypes.Configuration{
		Driver:            driver,
		Host:              "localhost",
		Port:              port,
		Type:              "tcp",
		Name:              "appdb",
		Login:             "user",
		Password:          "pass",
		Migration:         "./testdata/migrations",
		Charset:           "utf8mb4",
		ParseTime:         true,
		TimezoneLocation:  "UTC",
		DefaultStringSize: 256,
		CreateBatchSize:   100,
		MaxIdleConn:       2,
		MaxOpenConn:       4,
		MaxIdleTimeConn:   time.Minute,
		MaxLifetimeConn:   time.Hour,
		Loglevel:          kmll.Off,
	}
}

func startTestImpl(t *testing.T, cfg *kitModuleDbSqlTypes.Configuration) *impl {
	t.Helper()
	var (
		mys    = new(impl)
		onDone = make(chan struct{})
	)
	if err := mys.newConfigurationSet(onDone, cfg); err != nil {
		t.Fatalf("newConfigurationSet() вернул ошибку: %v", err)
	}
	<-onDone
	close(onDone)
	// onDone сигнализирует только о старте makeConnect; RLock дожидается завершения настройки соединения.
	mys.connectMux.RLock()
	mys.connectMux.RUnlock()
	t.Cleanup(func() { _ = mys.Close() })

	return mys
}

func newTestSQLDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open(testDriverName, "")
	if err != nil {
		t.Fatalf("sql.Open() вернул ошибку: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return db
}

func newGormReadyImpl(t *testing.T, cfg *kitModuleDbSqlTypes.Configuration) *impl {
	t.Helper()
	mys := &impl{
		cfg:        cfg,
		dsn:        "test-dsn",
		connect:    newTestSQLDB(t),
		connectMux: new(sync.RWMutex),
	}
	t.Cleanup(func() { _ = mys.Close() })

	return mys
}

type testModelGorm struct{}

func (testModelGorm) TableName() string           { return "test_table" }
func (testModelGorm) BeforeCreate(*gorm.DB) error { return nil }
func (testModelGorm) BeforeUpdate(*gorm.DB) error { return nil }

func TestGetFree(t *testing.T) {
	singleton = nil
	t.Cleanup(func() { singleton = nil })

	first := Get()
	if first == nil || first != singleton {
		t.Fatal("Get() вернул некорректный singleton")
	}
	if Get() != first {
		t.Fatal("Get() вернул разные объекты")
	}
	Free()
	if singleton != nil {
		t.Fatal("Free() не сбросил singleton")
	}
}

func TestImpl_ConnectionLifecycle(t *testing.T) {
	mys := startTestImpl(t, testConfig(testDriverName))
	if mys.err() != nil {
		t.Fatalf("startTestImpl() error = %v", mys.err())
	}
	if mys.SqlDB() == nil {
		t.Fatal("SqlDB() вернул nil")
	}
	if mys.Status() == nil {
		t.Fatal("Status() вернул nil")
	}
	if mys.SqlxDB() == nil {
		t.Fatal("SqlxDB() вернул nil")
	}
	if err := mys.E(); err != nil {
		t.Fatalf("E() вернул ошибку: %v", err)
	}
	if err := mys.Close(); err != nil {
		t.Fatalf("Close() вернул ошибку: %v", err)
	}
	if mys.connect != nil {
		t.Fatal("Close() не освободил connect")
	}
	if err := mys.Close(); err != nil {
		t.Fatalf("повторный Close() вернул ошибку: %v", err)
	}
	if mys.conn() == nil {
		t.Fatal("conn() не восстановил соединение")
	}
}

func TestImpl_ConnErrorBranches(t *testing.T) {
	bad := &impl{cfg: testConfig("unknown-driver"), connectMux: new(sync.RWMutex)}
	if bad.conn() != nil {
		t.Fatal("conn() с неверным драйвером вернул не nil")
	}
	if bad.err() == nil {
		t.Fatal("err() с неверным драйвером вернул nil")
	}
	if err := bad.E(); err == nil {
		t.Fatal("E() с неверным драйвером вернул nil")
	}
	if bad.Status() != nil {
		t.Fatal("Status() с неверным драйвером вернул не nil")
	}
	if bad.SqlDB() != nil {
		t.Fatal("SqlDB() с ошибкой вернул не nil")
	}
	if bad.SqlxDB() != nil {
		t.Fatal("SqlxDB() с ошибкой вернул не nil")
	}
}

func TestImpl_E_PingError(t *testing.T) {
	testDriverPingErr = errors.New("ping error")
	t.Cleanup(func() { testDriverPingErr = nil })
	mys := startTestImpl(t, testConfig(testDriverName))
	if err := mys.E(); err == nil || !strings.Contains(err.Error(), "ping error") {
		t.Fatalf("E() с ping error вернул %v", err)
	}
}

func TestImpl_GormDB(t *testing.T) {
	cfg := testConfig(driverMySQL)
	cfg.DisableAutomaticPing = true
	mys := newGormReadyImpl(t, cfg)
	gormDB := mys.GormDB()
	if gormDB == nil || mys.err() != nil {
		t.Fatalf("GormDB(MySQL) = %v, err=%v", gormDB, mys.err())
	}

	cfg = testConfig(driverPostgreSQL)
	cfg.DisableAutomaticPing = true
	mys = newGormReadyImpl(t, cfg)
	gormDB = mys.GormDB()
	if gormDB == nil || mys.err() != nil {
		t.Fatalf("GormDB(PostgreSQL) = %v, err=%v", gormDB, mys.err())
	}

	mys = newGormReadyImpl(t, testConfig(testDriverName))
	if gormDB = mys.GormDB(); gormDB != nil || mys.err() == nil {
		t.Fatalf("GormDB(unknown implemented driver) = %v, err=%v", gormDB, mys.err())
	}

	mys = newGormReadyImpl(t, cfg)
	mys.error = errors.New("previous error")
	if gormDB = mys.GormDB(); gormDB != nil {
		t.Fatal("GormDB() с предыдущей ошибкой вернул не nil")
	}
}

func TestImpl_GormDB_PingError(t *testing.T) {
	testDriverPingErr = errors.New("gorm ping error")
	t.Cleanup(func() { testDriverPingErr = nil })
	cfg := testConfig(driverMySQL)
	mys := newGormReadyImpl(t, cfg)
	gormDB := mys.GormDB()
	if mys.err() == nil || !strings.Contains(mys.err().Error(), "gorm ping error") {
		t.Fatalf("GormDB() с ping error = %v, err=%v", gormDB, mys.err())
	}
}

func TestImplementation(t *testing.T) {
	var implementation Implementation
	mysqlCfg := testConfig(driverMySQL)
	mysqlCfg.DisableAutomaticPing = true
	implementation.NewConfigurationSet(testConfig(testDriverName))
	if implementation.Gist() == nil {
		t.Fatal("Gist() вернул nil")
	}
	if implementation.Sqlx() == nil {
		t.Fatal("Sqlx() вернул nil")
	}
	if gormDB := implementation.Gorm(implementation.OptionSilent()); gormDB != nil {
		t.Fatalf("Gorm() для тестового драйвера вернул %v", gormDB)
	}

	implementation.NewConfigurationSet(mysqlCfg)
	if gormDB := implementation.Gorm(implementation.OptionSilent()); gormDB == nil {
		t.Fatal("Gorm(MySQL) вернул nil")
	}

	empty := Implementation{}
	if empty.Gist() == nil {
		t.Fatal("getParent() не установил глобальный объект")
	}
	singleton = nil
}

func TestImpl_NilMuxBranches(t *testing.T) {
	if err := (&impl{}).Close(); err != nil {
		t.Fatalf("Close() с nil mux вернул ошибку: %v", err)
	}
	if (&impl{}).conn() != nil {
		t.Fatal("conn() с nil mux вернул не nil")
	}
	if err := (&impl{error: errors.New("stored error")}).err(); err == nil {
		t.Fatal("err() с nil mux не вернул сохранённую ошибку")
	}
	if err := (&impl{connect: newTestSQLDB(t)}).Close(); err != nil {
		t.Fatalf("Close() с nil mux и connect вернул ошибку: %v", err)
	}

	mys := &impl{cfg: testConfig(testDriverName)}
	if err := mys.E(); err == nil || !strings.Contains(err.Error(), "database connection is nil") {
		t.Fatalf("E() с nil mux вернул %v", err)
	}
	mys = &impl{cfg: testConfig(driverMySQL)}
	if gormDB := mys.GormDB(); gormDB != nil || mys.err() == nil {
		t.Fatalf("GormDB() с nil mux = %v, err=%v", gormDB, mys.err())
	}
	mys = &impl{cfg: testConfig(testDriverName)}
	if sqlxDB := mys.SqlxDB(); sqlxDB != nil {
		t.Fatalf("SqlxDB() с nil mux = %v", sqlxDB)
	}
}

func TestImpl_MakeConnectOpenError(t *testing.T) {
	oldSupportDrivers := supportDrivers
	supportDrivers = append(append([]string{}, supportDrivers...), "not_registered_driver")
	t.Cleanup(func() { supportDrivers = oldSupportDrivers })

	mys := startTestImpl(t, testConfig("not_registered_driver"))
	if mys.err() == nil || !strings.Contains(mys.err().Error(), "Подключение") {
		t.Fatalf("makeConnect() с незарегистрированным драйвером err=%v", mys.err())
	}
}

func TestImpl_InfoDebug(t *testing.T) {
	oldDebug := kitModuleCfg.Get().Debug()
	kitModuleCfg.Get().Gist().Debug(!oldDebug)
	t.Cleanup(func() { kitModuleCfg.Get().Gist().Debug(oldDebug) })

	mys := &impl{}
	mys.info("debug message %s", "ok")
}

func writeTestMigration(t *testing.T, dir string) {
	t.Helper()
	content := []byte(`-- +goose Up
CREATE TABLE test_table (id INTEGER);

-- +goose Down
DROP TABLE test_table;
`)
	if err := os.WriteFile(filepath.Join(dir, "00001_test.sql"), content, 0644); err != nil {
		t.Fatalf("os.WriteFile() вернул ошибку: %v", err)
	}
}

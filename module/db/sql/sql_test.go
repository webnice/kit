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

	"github.com/jmoiron/sqlx"
	kitModuleCfg "github.com/webnice/kit/v4/module/cfg"
	kitModuleDbSqlTypes "github.com/webnice/kit/v4/module/db/sql/types"
	kmll "github.com/webnice/kit/v4/module/log/level"
	"gorm.io/gorm"
)

// Имя тестового драйвера базы данных.
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

func testConfig(driver string) (cfg *kitModuleDbSqlTypes.Configuration) {
	var port uint16

	port = 3306
	if strings.EqualFold(driver, driverPostgreSQL) {
		port = 5432
	}
	cfg = &kitModuleDbSqlTypes.Configuration{
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

	return
}

func startTestImpl(t *testing.T, cfg *kitModuleDbSqlTypes.Configuration) (mys *impl) {
	const tplFatalNewConfig = "newConfigurationSet() вернул ошибку: %v."
	var onDone chan struct{}

	t.Helper()
	mys = new(impl)
	onDone = make(chan struct{})
	if err := mys.newConfigurationSet(onDone, cfg); err != nil {
		t.Fatalf(tplFatalNewConfig, err)
	}
	<-onDone
	close(onDone)
	// onDone сигнализирует только о старте makeConnect; RLock дожидается завершения настройки соединения.
	mys.connectMux.RLock()
	mys.connectMux.RUnlock()
	t.Cleanup(func() { _ = mys.Close() })

	return
}

func newTestSQLDB(t *testing.T) (db *sql.DB) {
	const tplFatalOpen = "sql.Open() вернул ошибку: %v."
	var err error

	t.Helper()
	db, err = sql.Open(testDriverName, "")
	if err != nil {
		t.Fatalf(tplFatalOpen, err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return
}

func newGormReadyImpl(t *testing.T, cfg *kitModuleDbSqlTypes.Configuration) (mys *impl) {
	t.Helper()
	mys = &impl{
		cfg:        cfg,
		dsn:        "test-dsn",
		connect:    newTestSQLDB(t),
		connectMux: new(sync.RWMutex),
	}
	t.Cleanup(func() { _ = mys.Close() })

	return
}

type testModelGorm struct{}

func (testModelGorm) TableName() string           { return "test_table" }
func (testModelGorm) BeforeCreate(*gorm.DB) error { return nil }
func (testModelGorm) BeforeUpdate(*gorm.DB) error { return nil }

func TestGetFree(t *testing.T) {
	const (
		errInvalidSingleton = "Get() вернул некорректный singleton."
		errDifferentObjects = "Get() вернул разные объекты."
		errFreeNotReset     = "Free() не сбросил singleton."
	)
	var first Interface

	singleton = nil
	t.Cleanup(func() { singleton = nil })
	first = Get()
	if first == nil || first != singleton {
		t.Fatal(errInvalidSingleton)
	}
	if Get() != first {
		t.Fatal(errDifferentObjects)
	}
	Free()
	if singleton != nil {
		t.Fatal(errFreeNotReset)
	}
}

func TestImpl_ConnectionLifecycle(t *testing.T) {
	const (
		tplFatalStart      = "startTestImpl() error = %v."
		errSqlDBNil        = "SqlDB() вернул nil."
		errStatusNil       = "Status() вернул nil."
		errSqlxDBNil       = "SqlxDB() вернул nil."
		tplFatalE          = "E() вернул ошибку: %v."
		tplFatalClose      = "Close() вернул ошибку: %v."
		errConnectNotFreed = "Close() не освободил connect."
		tplFatalCloseAgain = "повторный Close() вернул ошибку: %v."
		errConnNotRestored = "conn() не восстановил соединение."
	)
	var mys *impl

	mys = startTestImpl(t, testConfig(testDriverName))
	if mys.err() != nil {
		t.Fatalf(tplFatalStart, mys.err())
	}
	if mys.SqlDB() == nil {
		t.Fatal(errSqlDBNil)
	}
	if mys.Status() == nil {
		t.Fatal(errStatusNil)
	}
	if mys.SqlxDB() == nil {
		t.Fatal(errSqlxDBNil)
	}
	if err := mys.E(); err != nil {
		t.Fatalf(tplFatalE, err)
	}
	if err := mys.Close(); err != nil {
		t.Fatalf(tplFatalClose, err)
	}
	if mys.connect != nil {
		t.Fatal(errConnectNotFreed)
	}
	if err := mys.Close(); err != nil {
		t.Fatalf(tplFatalCloseAgain, err)
	}
	if mys.conn() == nil {
		t.Fatal(errConnNotRestored)
	}
}

func TestImpl_ConnErrorBranches(t *testing.T) {
	const (
		errConnNotNil   = "conn() с неверным драйвером вернул не nil."
		errErrNil       = "err() с неверным драйвером вернул nil."
		errENil         = "E() с неверным драйвером вернул nil."
		errStatusNotNil = "Status() с неверным драйвером вернул не nil."
		errSqlDBNotNil  = "SqlDB() с ошибкой вернул не nil."
		errSqlxDBNotNil = "SqlxDB() с ошибкой вернул не nil."
	)
	var bad *impl

	bad = &impl{cfg: testConfig("unknown-driver"), connectMux: new(sync.RWMutex)}
	if bad.conn() != nil {
		t.Fatal(errConnNotNil)
	}
	if bad.err() == nil {
		t.Fatal(errErrNil)
	}
	if err := bad.E(); err == nil {
		t.Fatal(errENil)
	}
	if bad.Status() != nil {
		t.Fatal(errStatusNotNil)
	}
	if bad.SqlDB() != nil {
		t.Fatal(errSqlDBNotNil)
	}
	if bad.SqlxDB() != nil {
		t.Fatal(errSqlxDBNotNil)
	}
}

func TestImpl_E_PingError(t *testing.T) {
	const (
		errTextPingError = "ping error"
		tplFatalEPing    = "E() с ping error вернул %v."
	)
	var mys *impl

	testDriverPingErr = errors.New(errTextPingError)
	t.Cleanup(func() { testDriverPingErr = nil })
	mys = startTestImpl(t, testConfig(testDriverName))
	if err := mys.E(); err == nil || !strings.Contains(err.Error(), errTextPingError) {
		t.Fatalf(tplFatalEPing, err)
	}
}

func TestImpl_GormDB(t *testing.T) {
	const (
		errTextPreviousError = "previous error"
		tplFatalGormMySQL    = "GormDB(MySQL) = %v, err=%v."
		tplFatalGormPgSQL    = "GormDB(PostgreSQL) = %v, err=%v."
		tplFatalGormUnknown  = "GormDB(unknown implemented driver) = %v, err=%v."
		errGormNotNil        = "GormDB() с предыдущей ошибкой вернул не nil."
	)
	var (
		cfg    *kitModuleDbSqlTypes.Configuration
		mys    *impl
		gormDB *gorm.DB
	)

	cfg = testConfig(driverMySQL)
	cfg.DisableAutomaticPing = true
	mys = newGormReadyImpl(t, cfg)
	gormDB = mys.GormDB()
	if gormDB == nil || mys.err() != nil {
		t.Fatalf(tplFatalGormMySQL, gormDB, mys.err())
	}
	cfg = testConfig(driverPostgreSQL)
	cfg.DisableAutomaticPing = true
	mys = newGormReadyImpl(t, cfg)
	gormDB = mys.GormDB()
	if gormDB == nil || mys.err() != nil {
		t.Fatalf(tplFatalGormPgSQL, gormDB, mys.err())
	}
	mys = newGormReadyImpl(t, testConfig(testDriverName))
	if gormDB = mys.GormDB(); gormDB != nil || mys.err() == nil {
		t.Fatalf(tplFatalGormUnknown, gormDB, mys.err())
	}
	mys = newGormReadyImpl(t, cfg)
	mys.error = errors.New(errTextPreviousError)
	if gormDB = mys.GormDB(); gormDB != nil {
		t.Fatal(errGormNotNil)
	}
}

func TestImpl_GormDB_PingError(t *testing.T) {
	const (
		errTextGormPingError = "gorm ping error"
		tplFatalGormPing     = "GormDB() с ping error = %v, err=%v."
	)
	var (
		cfg    *kitModuleDbSqlTypes.Configuration
		mys    *impl
		gormDB *gorm.DB
	)

	testDriverPingErr = errors.New(errTextGormPingError)
	t.Cleanup(func() { testDriverPingErr = nil })
	cfg = testConfig(driverMySQL)
	mys = newGormReadyImpl(t, cfg)
	gormDB = mys.GormDB()
	if mys.err() == nil || !strings.Contains(mys.err().Error(), errTextGormPingError) {
		t.Fatalf(tplFatalGormPing, gormDB, mys.err())
	}
}

func TestImplementation(t *testing.T) {
	const (
		errGistNil       = "Gist() вернул nil."
		errSqlxNil       = "Sqlx() вернул nil."
		tplFatalGormTest = "Gorm() для тестового драйвера вернул %v."
		errGormMySQLNil  = "Gorm(MySQL) вернул nil."
		errGlobalNotSet  = "getParent() не установил глобальный объект."
	)
	var (
		implementation Implementation
		mysqlCfg       *kitModuleDbSqlTypes.Configuration
		gormDB         *gorm.DB
		empty          Implementation
	)

	mysqlCfg = testConfig(driverMySQL)
	mysqlCfg.DisableAutomaticPing = true
	implementation.NewConfigurationSet(testConfig(testDriverName))
	if implementation.Gist() == nil {
		t.Fatal(errGistNil)
	}
	if implementation.Sqlx() == nil {
		t.Fatal(errSqlxNil)
	}
	if gormDB = implementation.Gorm(implementation.OptionSilent()); gormDB != nil {
		t.Fatalf(tplFatalGormTest, gormDB)
	}
	implementation.NewConfigurationSet(mysqlCfg)
	if gormDB = implementation.Gorm(implementation.OptionSilent()); gormDB == nil {
		t.Fatal(errGormMySQLNil)
	}
	empty = Implementation{}
	if empty.Gist() == nil {
		t.Fatal(errGlobalNotSet)
	}
	singleton = nil
}

func TestImpl_NilMuxBranches(t *testing.T) {
	const (
		errTextStoredError    = "stored error"
		errNilMuxClose        = "Close() с nil mux вернул ошибку: %v."
		errNilMuxConn         = "conn() с nil mux вернул не nil."
		errNilMuxErr          = "err() с nil mux не вернул сохранённую ошибку."
		errNilMuxConnectClose = "Close() с nil mux и connect вернул ошибку: %v."
		tplFatalENilMux       = "E() с nil mux вернул %v."
		tplFatalGormNilMux    = "GormDB() с nil mux = %v, err=%v."
		tplFatalSqlxNilMux    = "SqlxDB() с nil mux = %v."
	)
	var (
		mys    *impl
		gormDB *gorm.DB
		sqlxDB *sqlx.DB
		err    error
	)

	if err = (&impl{}).Close(); err != nil {
		t.Fatalf(errNilMuxClose, err)
	}
	if (&impl{}).conn() != nil {
		t.Fatal(errNilMuxConn)
	}
	if err = (&impl{error: errors.New(errTextStoredError)}).err(); err == nil {
		t.Fatal(errNilMuxErr)
	}
	if err = (&impl{connect: newTestSQLDB(t)}).Close(); err != nil {
		t.Fatalf(errNilMuxConnectClose, err)
	}
	mys = &impl{cfg: testConfig(testDriverName)}
	if err = mys.E(); err == nil || !strings.Contains(err.Error(), "database connection is nil") {
		t.Fatalf(tplFatalENilMux, err)
	}
	mys = &impl{cfg: testConfig(driverMySQL)}
	if gormDB = mys.GormDB(); gormDB != nil || mys.err() == nil {
		t.Fatalf(tplFatalGormNilMux, gormDB, mys.err())
	}
	mys = &impl{cfg: testConfig(testDriverName)}
	if sqlxDB = mys.SqlxDB(); sqlxDB != nil {
		t.Fatalf(tplFatalSqlxNilMux, sqlxDB)
	}
}

func TestImpl_MakeConnectOpenError(t *testing.T) {
	const (
		errTextNotRegisteredDriver = "not_registered_driver"
		tplFatalMakeConnect        = "makeConnect() с незарегистрированным драйвером err=%v."
	)
	var (
		oldSupportDrivers []string
		mys               *impl
	)

	oldSupportDrivers = supportDrivers
	supportDrivers = append(append([]string{}, supportDrivers...), errTextNotRegisteredDriver)
	t.Cleanup(func() { supportDrivers = oldSupportDrivers })
	mys = startTestImpl(t, testConfig(errTextNotRegisteredDriver))
	if mys.err() == nil || !strings.Contains(mys.err().Error(), "Подключение") {
		t.Fatalf(tplFatalMakeConnect, mys.err())
	}
}

func TestImpl_InfoDebug(t *testing.T) {
	const (
		tplDebugMessage = "debug message %s"
		testMessage     = "ok"
	)
	var (
		oldDebug bool
		mys      *impl
	)

	oldDebug = kitModuleCfg.Get().Debug()
	kitModuleCfg.Get().Gist().Debug(!oldDebug)
	t.Cleanup(func() { kitModuleCfg.Get().Gist().Debug(oldDebug) })
	mys = &impl{}
	mys.info(tplDebugMessage, testMessage)
}

func writeTestMigration(t *testing.T, dir string) {
	const (
		tplFatalWriteFile = "os.WriteFile() вернул ошибку: %v."
	)
	var content []byte

	t.Helper()
	content = []byte(`-- +goose Up
CREATE TABLE test_table (id INTEGER);

-- +goose Down
DROP TABLE test_table;
`)
	if err := os.WriteFile(filepath.Join(dir, "00001_test.sql"), content, 0644); err != nil {
		t.Fatalf(tplFatalWriteFile, err)
	}
}

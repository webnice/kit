package migration

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	kitModuleDbSql "github.com/webnice/kit/v5/module/db/sql"

	"github.com/jmoiron/sqlx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Имя тестового драйвера базы данных.
const migrationTestDriverName = "migration_test_driver"

// Строка тестовой таблицы версий миграций.
const migrationTestTableName = "goose_db_version"

// Строки ошибок тестового драйвера базы данных.
const (
	errTextTransactionsNotImplemented = "transactions not implemented"
	errTextExecNotImplemented         = "exec not implemented"
	errTextNoSuchTable                = "no such table: goose_db_version"
	errTextGetDbError                 = "get db error"
)

// Строка-шаблон поиска тестовой таблицы в запросе.
const migrationTestTablePattern = "`goose_db_version`"

type migrationTestRow struct {
	id        int64
	versionID int64
	isApplied bool
	tstamp    time.Time
}

var migrationTestData = struct {
	sync.RWMutex
	tableExists bool
	rows        []migrationTestRow
}{}

func init() { sql.Register(migrationTestDriverName, migrationTestDriver{}) }

func setMigrationTestData(tableExists bool, rows []migrationTestRow) {
	migrationTestData.Lock()
	defer migrationTestData.Unlock()
	migrationTestData.tableExists = tableExists
	migrationTestData.rows = append([]migrationTestRow(nil), rows...)
}

func resetMigrationTestData() { setMigrationTestData(false, nil) }

type migrationTestDriver struct{}

// Open Реализация интерфейса driver.Driver.
func (migrationTestDriver) Open(string) (driver.Conn, error) { return &migrationTestConn{}, nil }

type migrationTestConn struct{}

// Prepare Реализация интерфейса driver.Conn.
func (conn *migrationTestConn) Prepare(query string) (driver.Stmt, error) {
	return &migrationTestStmt{conn: conn, query: query}, nil
}

// Close Реализация интерфейса driver.Conn.
func (conn *migrationTestConn) Close() error { return nil }

// Begin Реализация интерфейса driver.Conn.
func (conn *migrationTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New(errTextTransactionsNotImplemented)
}

// QueryContext Реализация интерфейса driver.QueryerContext.
func (conn *migrationTestConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	return migrationTestQuery(query)
}

// Query Реализация интерфейса driver.Conn.
func (conn *migrationTestConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	var (
		named []driver.NamedValue
		n     int
	)

	named = make([]driver.NamedValue, len(args))
	for n = range args {
		named[n] = driver.NamedValue{Ordinal: n + 1, Value: args[n]}
	}

	return conn.QueryContext(context.Background(), query, named)
}

type migrationTestStmt struct {
	conn  *migrationTestConn
	query string
}

// Close Реализация интерфейса driver.Conn.
func (stmt *migrationTestStmt) Close() error { return nil }

// NumInput Реализация интерфейса driver.Stmt.
func (stmt *migrationTestStmt) NumInput() int { return 0 }

// Exec Реализация интерфейса driver.Stmt.
func (stmt *migrationTestStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, errors.New(errTextExecNotImplemented)
}

// Query Реализация интерфейса driver.Conn.
func (stmt *migrationTestStmt) Query(args []driver.Value) (driver.Rows, error) {
	var (
		named []driver.NamedValue
		n     int
	)

	named = make([]driver.NamedValue, len(args))
	for n = range args {
		named[n] = driver.NamedValue{Ordinal: n + 1, Value: args[n]}
	}

	return stmt.conn.QueryContext(context.Background(), stmt.query, named)
}

func migrationTestQuery(query string) (driver.Rows, error) {
	var (
		rows   []migrationTestRow
		values [][]driver.Value
		n      int
	)

	migrationTestData.RLock()
	defer migrationTestData.RUnlock()
	if !migrationTestData.tableExists || !strings.Contains(query, migrationTestTablePattern) {
		return nil, errors.New(errTextNoSuchTable)
	}
	// Отбор применённых версий миграций.
	rows = make([]migrationTestRow, 0, len(migrationTestData.rows))
	for n = range migrationTestData.rows {
		if migrationTestData.rows[n].isApplied {
			rows = append(rows, migrationTestData.rows[n])
		}
	}
	// Сортировка версий по времени применения, от новых к старым.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].tstamp.Equal(rows[j].tstamp) {
			return rows[i].id > rows[j].id
		}

		return rows[i].tstamp.After(rows[j].tstamp)
	})
	// Формирование значений строк результата запроса.
	values = make([][]driver.Value, len(rows))
	for n = range rows {
		values[n] = []driver.Value{rows[n].id, rows[n].versionID, rows[n].isApplied, rows[n].tstamp}
	}

	return &migrationTestRows{
		columns: []string{"id", "version_id", "is_applied", "tstamp"},
		rows:    values,
		index:   -1,
	}, nil
}

type migrationTestRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

// Columns Реализация интерфейса driver.Rows.
func (rows *migrationTestRows) Columns() []string { return rows.columns }

// Close Реализация интерфейса driver.Conn.
func (rows *migrationTestRows) Close() error { return nil }

// Next Реализация интерфейса driver.Rows.
func (rows *migrationTestRows) Next(dest []driver.Value) error {
	rows.index++
	if rows.index >= len(rows.rows) {
		return io.EOF
	}
	copy(dest, rows.rows[rows.index])

	return nil
}

type testDatabase struct {
	gormDB *gorm.DB
	sqlDB  *sql.DB
	err    error
}

// Close Реализация интерфейса driver.Conn.
func (tdb *testDatabase) Close() (err error) {
	if tdb.sqlDB != nil {
		err = tdb.sqlDB.Close()
	}

	return
}

// E Реализация интерфейса объекта базы данных.
func (tdb *testDatabase) E() error { return tdb.err }

// Status Реализация интерфейса объекта базы данных.
func (tdb *testDatabase) Status() (ret *sql.DBStats) {
	var stats sql.DBStats

	if tdb.sqlDB == nil {
		return
	}
	stats = tdb.sqlDB.Stats()
	ret = &stats

	return
}

// SqlDB Реализация интерфейса объекта базы данных.
func (tdb *testDatabase) SqlDB() *sql.DB { return tdb.sqlDB }

// GormDB Реализация интерфейса объекта базы данных.
func (tdb *testDatabase) GormDB() *gorm.DB { return tdb.gormDB }

// SqlxDB Реализация интерфейса объекта базы данных.
func (tdb *testDatabase) SqlxDB() *sqlx.DB { return nil }

// MigrationUp Реализация интерфейса объекта базы данных.
func (tdb *testDatabase) MigrationUp() error { return nil }

// Errors Реализация интерфейса объекта базы данных.
func (tdb *testDatabase) Errors() *kitModuleDbSql.Error { return kitModuleDbSql.Errors() }

type testErrConnPool struct{ err error }

// PrepareContext Реализация интерфейса объекта пула соединений.
func (pool testErrConnPool) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, pool.err
}

// ExecContext Реализация интерфейса объекта пула соединений.
func (pool testErrConnPool) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, pool.err
}

// QueryContext Реализация интерфейса driver.QueryerContext.
func (pool testErrConnPool) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, pool.err
}

// QueryRowContext Реализация интерфейса объекта пула соединений.
func (pool testErrConnPool) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return nil
}

// GetDBConn Реализация интерфейса объекта пула соединений.
func (pool testErrConnPool) GetDBConn() (*sql.DB, error) { return nil, pool.err }

func newTestGorm(t *testing.T) (sqlDB *sql.DB, orm *gorm.DB) {
	const (
		tplSqlOpenError  = "sql.Open() вернул ошибку: %v."
		tplGormOpenError = "gorm.Open() вернул ошибку: %v."
	)
	var err error

	resetMigrationTestData()
	if sqlDB, err = sql.Open(migrationTestDriverName, ""); err != nil {
		t.Fatalf(tplSqlOpenError, err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
		resetMigrationTestData()
	})
	if orm, err = gorm.Open(
		mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}),
		&gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)},
	); err != nil {
		t.Fatalf(tplGormOpenError, err)
	}

	return
}

// TestImpl_New Конструктор объекта миграций.
func TestImpl_New(t *testing.T) {
	const errNewNil = "New() вернул nil."
	var mmn Interface

	mmn = New()
	if mmn == nil {
		t.Fatal(errNewNil)
	}
}

func TestImpl_GistFallback(t *testing.T) {
	const errGistNil = "gist() вернул nil."

	if New().(*impl).gist() == nil {
		t.Fatal(errGistNil)
	}
}

func TestImpl_CurrentVersionOrmIsNil(t *testing.T) {
	const (
		errCurrentVersionNilError = "CurrentVersion() без ORM вернул nil ошибку."
		tplCurrentVersionNotNil   = "CurrentVersion() без ORM вернул версию %+v."
	)
	var (
		mmn Interface
		ver *DbVersion
		err error
	)

	mmn = newWithDB(&testDatabase{})
	ver, err = mmn.CurrentVersion()
	if err == nil {
		t.Fatal(errCurrentVersionNilError)
	}
	if ver != nil {
		t.Fatalf(tplCurrentVersionNotNil, ver)
	}
}

func TestImpl_CurrentVersionDBError(t *testing.T) {
	const (
		errCurrentVersionNilError = "CurrentVersion() при ошибке DB() вернул nil ошибку."
		tplCurrentVersionNotNil   = "CurrentVersion() при ошибке DB() вернул версию %+v."
	)
	var (
		expected error
		orm      *gorm.DB
		mmn      Interface
		ver      *DbVersion
		err      error
	)

	expected = errors.New(errTextGetDbError)
	orm = &gorm.DB{Config: &gorm.Config{ConnPool: testErrConnPool{err: expected}}}
	mmn = newWithDB(&testDatabase{gormDB: orm})
	ver, err = mmn.CurrentVersion()
	if err == nil {
		t.Fatal(errCurrentVersionNilError)
	}
	if ver != nil {
		t.Fatalf(tplCurrentVersionNotNil, ver)
	}
}

func TestImpl_CurrentVersionPingError(t *testing.T) {
	const (
		tplSqlOpenError           = "sql.Open() вернул ошибку: %v."
		tplGormOpenError          = "gorm.Open() вернул ошибку: %v."
		errCurrentVersionNilError = "CurrentVersion() при ошибке Ping() вернул nil ошибку."
		tplCurrentVersionNotNil   = "CurrentVersion() при ошибке Ping() вернул версию %+v."
	)
	var (
		sqlDB *sql.DB
		orm   *gorm.DB
		ver   *DbVersion
		err   error
	)

	if sqlDB, err = sql.Open(migrationTestDriverName, ""); err != nil {
		t.Fatalf(tplSqlOpenError, err)
	}
	_ = sqlDB.Close()
	if orm, err = gorm.Open(
		mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}),
		&gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)},
	); err != nil {
		t.Fatalf(tplGormOpenError, err)
	}
	ver, err = newWithDB(&testDatabase{gormDB: orm, sqlDB: sqlDB}).CurrentVersion()
	if err == nil {
		t.Fatal(errCurrentVersionNilError)
	}
	if ver != nil {
		t.Fatalf(tplCurrentVersionNotNil, ver)
	}
}

func TestImpl_CurrentVersionQueryError(t *testing.T) {
	const (
		errCurrentVersionNilError = "CurrentVersion() при отсутствующей таблице вернул nil ошибку."
		errCurrentVersionNilVer   = "CurrentVersion() при ошибке запроса вернул nil версию."
		tplQueryErrorNotWrapped   = "ошибка запроса не обёрнута через QErr: %v."
	)
	var (
		sqlDB *sql.DB
		orm   *gorm.DB
		mmn   Interface
		ver   *DbVersion
		err   error
	)

	sqlDB, orm = newTestGorm(t)
	mmn = newWithDB(&testDatabase{gormDB: orm, sqlDB: sqlDB})
	ver, err = mmn.CurrentVersion()
	if err == nil {
		t.Fatal(errCurrentVersionNilError)
	}
	if ver == nil {
		t.Fatal(errCurrentVersionNilVer)
	}
	if !strings.Contains(err.Error(), "ошибка базы данных") {
		t.Fatalf(tplQueryErrorNotWrapped, err)
	}
}

func TestImpl_CurrentVersionSuccess(t *testing.T) {
	const (
		tplCurrentVersionError = "CurrentVersion() вернул ошибку: %v."
		errCurrentVersionNil   = "CurrentVersion() вернул nil версию."
		tplIDUnexpected        = "ID = %d, ожидался 2."
		tplVersionIDUnexpected = "VersionID = %d, ожидался 11."
		errIsAppliedFalse      = "IsApplied = false, ожидался true."
		errTstampInvalid       = "Tstamp.Valid = false, ожидался true."
	)
	var (
		sqlDB *sql.DB
		orm   *gorm.DB
		mmn   Interface
		ver   *DbVersion
		err   error
	)

	sqlDB, orm = newTestGorm(t)
	setMigrationTestData(true, []migrationTestRow{
		{id: 1, versionID: 10, isApplied: true, tstamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{id: 2, versionID: 11, isApplied: true, tstamp: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		{id: 3, versionID: 12, isApplied: false, tstamp: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
	})
	mmn = newWithDB(&testDatabase{gormDB: orm, sqlDB: sqlDB})
	if ver, err = mmn.CurrentVersion(); err != nil {
		t.Fatalf(tplCurrentVersionError, err)
	}
	if ver == nil {
		t.Fatal(errCurrentVersionNil)
	}
	if ver.ID != 2 {
		t.Fatalf(tplIDUnexpected, ver.ID)
	}
	if ver.VersionID != 11 {
		t.Fatalf(tplVersionIDUnexpected, ver.VersionID)
	}
	if !ver.IsApplied {
		t.Fatal(errIsAppliedFalse)
	}
	if !ver.Tstamp.Valid {
		t.Fatal(errTstampInvalid)
	}
}

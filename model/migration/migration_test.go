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

	kitModuleDbSql "github.com/webnice/kit/v4/module/db/sql"

	"github.com/jmoiron/sqlx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const migrationTestDriverName = "migration_test_driver"

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

func (migrationTestDriver) Open(string) (driver.Conn, error) { return &migrationTestConn{}, nil }

type migrationTestConn struct{}

func (conn *migrationTestConn) Prepare(query string) (driver.Stmt, error) {
	return &migrationTestStmt{conn: conn, query: query}, nil
}

func (conn *migrationTestConn) Close() error { return nil }

func (conn *migrationTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not implemented")
}

func (conn *migrationTestConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	return migrationTestQuery(query)
}

func (conn *migrationTestConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for n := range args {
		named[n] = driver.NamedValue{Ordinal: n + 1, Value: args[n]}
	}

	return conn.QueryContext(context.Background(), query, named)
}

type migrationTestStmt struct {
	conn  *migrationTestConn
	query string
}

func (stmt *migrationTestStmt) Close() error { return nil }

func (stmt *migrationTestStmt) NumInput() int { return 0 }

func (stmt *migrationTestStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, errors.New("exec not implemented")
}

func (stmt *migrationTestStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for n := range args {
		named[n] = driver.NamedValue{Ordinal: n + 1, Value: args[n]}
	}

	return stmt.conn.QueryContext(context.Background(), stmt.query, named)
}

func migrationTestQuery(query string) (driver.Rows, error) {
	migrationTestData.RLock()
	defer migrationTestData.RUnlock()
	if !migrationTestData.tableExists || !strings.Contains(query, "`goose_db_version`") {
		return nil, errors.New("no such table: goose_db_version")
	}

	rows := make([]migrationTestRow, 0, len(migrationTestData.rows))
	for _, row := range migrationTestData.rows {
		if row.isApplied {
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].tstamp.Equal(rows[j].tstamp) {
			return rows[i].id > rows[j].id
		}
		return rows[i].tstamp.After(rows[j].tstamp)
	})

	values := make([][]driver.Value, len(rows))
	for n := range rows {
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

func (rows *migrationTestRows) Columns() []string { return rows.columns }

func (rows *migrationTestRows) Close() error { return nil }

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

func (tdb *testDatabase) Close() (err error) {
	if tdb.sqlDB != nil {
		err = tdb.sqlDB.Close()
	}
	return
}

func (tdb *testDatabase) E() error { return tdb.err }

func (tdb *testDatabase) Status() (ret *sql.DBStats) {
	if tdb.sqlDB == nil {
		return
	}
	stats := tdb.sqlDB.Stats()
	ret = &stats

	return
}

func (tdb *testDatabase) SqlDB() *sql.DB                { return tdb.sqlDB }
func (tdb *testDatabase) GormDB() *gorm.DB              { return tdb.gormDB }
func (tdb *testDatabase) SqlxDB() *sqlx.DB              { return nil }
func (tdb *testDatabase) MigrationUp() error            { return nil }
func (tdb *testDatabase) Errors() *kitModuleDbSql.Error { return kitModuleDbSql.Errors() }

type testErrConnPool struct{ err error }

func (pool testErrConnPool) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, pool.err
}

func (pool testErrConnPool) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, pool.err
}

func (pool testErrConnPool) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, pool.err
}

func (pool testErrConnPool) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return nil
}

func (pool testErrConnPool) GetDBConn() (*sql.DB, error) { return nil, pool.err }

func newTestGorm(t *testing.T) (sqlDB *sql.DB, orm *gorm.DB) {
	var err error

	resetMigrationTestData()
	if sqlDB, err = sql.Open(migrationTestDriverName, ""); err != nil {
		t.Fatalf("sql.Open() вернул ошибку: %v", err)
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
		t.Fatalf("gorm.Open() вернул ошибку: %v", err)
	}

	return
}

func TestImpl_NewAndErrors(t *testing.T) {
	var mmn = New()

	if mmn == nil {
		t.Fatal("New() вернул nil")
	}
	if mmn.Errors() != Errors() {
		t.Fatal("Errors() вернул не справочник ошибок пакета")
	}
	if mmn.Errors().DatabaseIsNotInUse == nil {
		t.Fatal("DatabaseIsNotInUse не определена")
	}
	if mmn.Errors().DatabaseUnexpectedError == nil {
		t.Fatal("DatabaseUnexpectedError не определена")
	}
}

func TestImpl_DbVersionTableName(t *testing.T) {
	if (&DbVersion{}).TableName() != "goose_db_version" {
		t.Fatal("TableName() != goose_db_version")
	}
}

func TestImpl_GistFallback(t *testing.T) {
	if New().(*impl).gist() == nil {
		t.Fatal("gist() вернул nil")
	}
}

func TestImpl_CurrentVersionOrmIsNil(t *testing.T) {
	var mmn = newWithDB(&testDatabase{})

	ver, err := mmn.CurrentVersion()
	if err == nil {
		t.Fatal("CurrentVersion() без ORM вернул nil ошибку")
	}
	if ver != nil {
		t.Fatalf("CurrentVersion() без ORM вернул версию %+v", ver)
	}
}

func TestImpl_CurrentVersionDBError(t *testing.T) {
	var (
		expected = errors.New("get db error")
		orm      = &gorm.DB{Config: &gorm.Config{ConnPool: testErrConnPool{err: expected}}}
		mmn      = newWithDB(&testDatabase{gormDB: orm})
	)

	ver, err := mmn.CurrentVersion()
	if err == nil {
		t.Fatal("CurrentVersion() при ошибке DB() вернул nil ошибку")
	}
	if ver != nil {
		t.Fatalf("CurrentVersion() при ошибке DB() вернул версию %+v", ver)
	}
}

func TestImpl_CurrentVersionPingError(t *testing.T) {
	var (
		sqlDB *sql.DB
		orm   *gorm.DB
		err   error
	)

	if sqlDB, err = sql.Open(migrationTestDriverName, ""); err != nil {
		t.Fatalf("sql.Open() вернул ошибку: %v", err)
	}
	_ = sqlDB.Close()
	if orm, err = gorm.Open(
		mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}),
		&gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)},
	); err != nil {
		t.Fatalf("gorm.Open() вернул ошибку: %v", err)
	}

	ver, err := newWithDB(&testDatabase{gormDB: orm, sqlDB: sqlDB}).CurrentVersion()
	if err == nil {
		t.Fatal("CurrentVersion() при ошибке Ping() вернул nil ошибку")
	}
	if ver != nil {
		t.Fatalf("CurrentVersion() при ошибке Ping() вернул версию %+v", ver)
	}
}

func TestImpl_CurrentVersionQueryError(t *testing.T) {
	var (
		sqlDB, orm = newTestGorm(t)
		mmn        = newWithDB(&testDatabase{gormDB: orm, sqlDB: sqlDB})
	)

	ver, err := mmn.CurrentVersion()
	if err == nil {
		t.Fatal("CurrentVersion() при отсутствующей таблице вернул nil ошибку")
	}
	if ver == nil {
		t.Fatal("CurrentVersion() при ошибке запроса вернул nil версию")
	}
	if !strings.Contains(err.Error(), "ошибка базы данных") {
		t.Fatalf("ошибка запроса не обёрнута через QErr: %v", err)
	}
}

func TestImpl_CurrentVersionSuccess(t *testing.T) {
	var (
		sqlDB, orm = newTestGorm(t)
		mmn        Interface
		ver        *DbVersion
		err        error
	)

	setMigrationTestData(true, []migrationTestRow{
		{id: 1, versionID: 10, isApplied: true, tstamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{id: 2, versionID: 11, isApplied: true, tstamp: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		{id: 3, versionID: 12, isApplied: false, tstamp: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
	})

	mmn = newWithDB(&testDatabase{gormDB: orm, sqlDB: sqlDB})
	if ver, err = mmn.CurrentVersion(); err != nil {
		t.Fatalf("CurrentVersion() вернул ошибку: %v", err)
	}
	if ver == nil {
		t.Fatal("CurrentVersion() вернул nil версию")
	}
	if ver.ID != 2 {
		t.Fatalf("ID = %d, ожидался 2", ver.ID)
	}
	if ver.VersionID != 11 {
		t.Fatalf("VersionID = %d, ожидался 11", ver.VersionID)
	}
	if !ver.IsApplied {
		t.Fatal("IsApplied = false, ожидался true")
	}
	if !ver.Tstamp.Valid {
		t.Fatal("Tstamp.Valid = false, ожидался true")
	}
}

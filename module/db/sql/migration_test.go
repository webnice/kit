package sql

import (
	"database/sql/driver"

	"errors"
	kitModuleDbSqlTypes "github.com/webnice/kit/v5/module/db/sql/types"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestImpl_MigrationUpErrors(t *testing.T) {
	const (
		tplFatalNilCfg     = "MigrationUp() с nil cfg вернул %v."
		tplFatalTestDriver = "MigrationUp() с тестовым драйвером вернул %v."
		errFakeConnNoError = "MigrationUp() с fake conn без поддержки goose queries не вернул ошибку."
	)
	var mys impl

	if err := mys.MigrationUp(); err == nil || !strings.Contains(err.Error(), "пустая") {
		t.Fatalf(tplFatalNilCfg, err)
	}
	mys = impl{cfg: testConfig(testDriverName)}
	if err := mys.MigrationUp(); err == nil || !strings.Contains(err.Error(), "диалекта") {
		t.Fatalf(tplFatalTestDriver, err)
	}
	mys = impl{cfg: testConfig(driverMySQL), connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err == nil {
		t.Fatal(errFakeConnNoError)
	}
}

func TestImpl_MigrationUpNoMigrations(t *testing.T) {
	const (
		tplFatalNoMigrations = "MigrationUp() без миграций вернул ошибку: %v."
	)
	var (
		cfg *kitModuleDbSqlTypes.Configuration
		mys *impl
	)

	testDriverQueryRows = [][]driver.Value{}
	t.Cleanup(func() { testDriverQueryRows = nil })
	cfg = testConfig(driverMySQL)
	cfg.Migration = t.TempDir()
	mys = &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err != nil {
		t.Fatalf(tplFatalNoMigrations, err)
	}
}

func TestImpl_MigrationUpCollectError(t *testing.T) {
	const (
		errMissingDirNoError = "MigrationUp() с отсутствующей директорией миграций не вернул ошибку."
	)
	var (
		cfg *kitModuleDbSqlTypes.Configuration
		mys *impl
	)

	testDriverQueryRows = [][]driver.Value{}
	t.Cleanup(func() { testDriverQueryRows = nil })
	cfg = testConfig(driverMySQL)
	cfg.Migration = filepath.Join(t.TempDir(), "missing")
	mys = &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err == nil {
		t.Fatal(errMissingDirNoError)
	}
}

func TestImpl_MigrationUpSuccess(t *testing.T) {
	const (
		tplFatalMigrationUp = "MigrationUp() вернул ошибку: %v."
	)
	var (
		migrationDir string
		cfg          *kitModuleDbSqlTypes.Configuration
		mys          *impl
	)

	testDriverQueryRows = [][]driver.Value{{int64(0), true}}
	t.Cleanup(func() { testDriverQueryRows = nil })
	migrationDir = t.TempDir()
	writeTestMigration(t, migrationDir)
	cfg = testConfig(driverMySQL)
	cfg.Migration = migrationDir
	mys = &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err != nil {
		t.Fatalf(tplFatalMigrationUp, err)
	}
}

func TestImpl_MigrationUpApplyError(t *testing.T) {
	const (
		errTextExecError   = "exec error"
		tplFatalApplyError = "MigrationUp() с ошибкой применения вернул %v."
	)
	var (
		migrationDir string
		cfg          *kitModuleDbSqlTypes.Configuration
		mys          *impl
	)

	testDriverQueryRows = [][]driver.Value{{int64(0), true}}
	testDriverExecErr = errors.New(errTextExecError)
	t.Cleanup(func() {
		testDriverQueryRows = nil
		testDriverExecErr = nil
	})
	migrationDir = t.TempDir()
	writeTestMigration(t, migrationDir)
	cfg = testConfig(driverMySQL)
	cfg.Migration = migrationDir
	mys = &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err == nil || !strings.Contains(err.Error(), "файл миграции") {
		t.Fatalf(tplFatalApplyError, err)
	}
}

func TestImpl_MigrationUpEnsureError(t *testing.T) {
	const (
		errTextQueryError = "query error"
		errTextExecError  = "exec error"
		errEnsureNoError  = "MigrationUp() с ошибкой EnsureDBVersion не вернул ошибку."
	)
	var (
		cfg *kitModuleDbSqlTypes.Configuration
		mys *impl
	)

	testDriverQueryErr = errors.New(errTextQueryError)
	testDriverExecErr = errors.New(errTextExecError)
	t.Cleanup(func() {
		testDriverQueryErr = nil
		testDriverExecErr = nil
	})
	cfg = testConfig(driverMySQL)
	cfg.Migration = t.TempDir()
	mys = &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err == nil {
		t.Fatal(errEnsureNoError)
	}
}

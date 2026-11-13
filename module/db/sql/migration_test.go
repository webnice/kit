package sql

import (
	"database/sql/driver"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestImpl_MigrationUpErrors(t *testing.T) {
	var mys impl
	if err := mys.MigrationUp(); err == nil || !strings.Contains(err.Error(), "пустая") {
		t.Fatalf("MigrationUp() с nil cfg вернул %v", err)
	}

	mys = impl{cfg: testConfig(testDriverName)}
	if err := mys.MigrationUp(); err == nil || !strings.Contains(err.Error(), "диалекта") {
		t.Fatalf("MigrationUp() с тестовым драйвером вернул %v", err)
	}

	mys = impl{cfg: testConfig(driverMySQL), connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err == nil {
		t.Fatal("MigrationUp() с fake conn без поддержки goose queries не вернул ошибку")
	}
}

func TestImpl_MigrationUpNoMigrations(t *testing.T) {
	testDriverQueryRows = [][]driver.Value{}
	t.Cleanup(func() { testDriverQueryRows = nil })
	cfg := testConfig(driverMySQL)
	cfg.Migration = t.TempDir()
	mys := &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err != nil {
		t.Fatalf("MigrationUp() без миграций вернул ошибку: %v", err)
	}
}

func TestImpl_MigrationUpCollectError(t *testing.T) {
	testDriverQueryRows = [][]driver.Value{}
	t.Cleanup(func() { testDriverQueryRows = nil })
	cfg := testConfig(driverMySQL)
	cfg.Migration = filepath.Join(t.TempDir(), "missing")
	mys := &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err == nil {
		t.Fatal("MigrationUp() с отсутствующей директорией миграций не вернул ошибку")
	}
}

func TestImpl_MigrationUpSuccess(t *testing.T) {
	testDriverQueryRows = [][]driver.Value{{int64(0), true}}
	t.Cleanup(func() { testDriverQueryRows = nil })
	migrationDir := t.TempDir()
	writeTestMigration(t, migrationDir)
	cfg := testConfig(driverMySQL)
	cfg.Migration = migrationDir
	mys := &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err != nil {
		t.Fatalf("MigrationUp() вернул ошибку: %v", err)
	}
}

func TestImpl_MigrationUpApplyError(t *testing.T) {
	testDriverQueryRows = [][]driver.Value{{int64(0), true}}
	testDriverExecErr = errors.New("exec error")
	t.Cleanup(func() {
		testDriverQueryRows = nil
		testDriverExecErr = nil
	})
	migrationDir := t.TempDir()
	writeTestMigration(t, migrationDir)
	cfg := testConfig(driverMySQL)
	cfg.Migration = migrationDir
	mys := &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err == nil || !strings.Contains(err.Error(), "файл миграции") {
		t.Fatalf("MigrationUp() с ошибкой применения вернул %v", err)
	}
}

func TestImpl_MigrationUpEnsureError(t *testing.T) {
	testDriverQueryErr = errors.New("query error")
	testDriverExecErr = errors.New("exec error")
	t.Cleanup(func() {
		testDriverQueryErr = nil
		testDriverExecErr = nil
	})
	cfg := testConfig(driverMySQL)
	cfg.Migration = t.TempDir()
	mys := &impl{cfg: cfg, connect: newTestSQLDB(t), connectMux: new(sync.RWMutex)}
	if err := mys.MigrationUp(); err == nil {
		t.Fatal("MigrationUp() с ошибкой EnsureDBVersion не вернул ошибку")
	}
}

package sql

import (
	"context"
	"errors"
	"testing"
	"time"

	kmll "github.com/webnice/kit/v4/module/log/level"
	gormLogger "gorm.io/gorm/logger"
)

func TestLoggerGorm(t *testing.T) {
	const (
		errInvalidObject  = "NewLoggerGorm() вернул некорректный объект."
		errLogModeDefault = "LogMode(default) изменил уровень."
		errLogModeSilent  = "LogMode(Silent) некорректен."
		errLogModeError   = "LogMode(Error) некорректен."
		errLogModeWarn    = "LogMode(Warn) некорректен."
		errLogModeInfo    = "LogMode(Info) некорректен."
		tplInfo           = "info %s"
		tplWarn           = "warn %s"
		tplError          = "error %s"
		testMessage       = "msg"
	)
	var (
		parent *impl
		logger *logGorm
	)

	parent = &impl{cfg: testConfig(testDriverName)}
	parent.cfg.Loglevel = kmll.Info
	logger = NewLoggerGorm(parent).(*logGorm)
	if logger == nil || logger.Loglevel != kmll.Info {
		t.Fatal(errInvalidObject)
	}
	logger.LogMode(gormLogger.LogLevel(123))
	if logger.Loglevel != kmll.Info {
		t.Fatal(errLogModeDefault)
	}
	logger.LogMode(gormLogger.Silent)
	if logger.Loglevel != kmll.Off {
		t.Fatal(errLogModeSilent)
	}
	logger.LogMode(gormLogger.Error)
	if logger.Loglevel != kmll.Error {
		t.Fatal(errLogModeError)
	}
	logger.LogMode(gormLogger.Warn)
	if logger.Loglevel != kmll.Warning {
		t.Fatal(errLogModeWarn)
	}
	logger.LogMode(gormLogger.Info)
	if logger.Loglevel != kmll.Info {
		t.Fatal(errLogModeInfo)
	}
	logger.Info(context.Background(), tplInfo, testMessage)
	logger.Warn(context.Background(), tplWarn, testMessage)
	logger.Error(context.Background(), tplError, testMessage)
}

func TestLoggerGorm_Trace(t *testing.T) {
	const (
		errTextTraceError = "trace error"
		tplSelectOne      = `select "1"`
	)
	var (
		parent *impl
		fc     func() (string, int64)
		silent context.Context
	)

	parent = &impl{cfg: testConfig(testDriverName)}
	fc = func() (string, int64) { return tplSelectOne, 1 }
	parent.cfg.Loglevel = kmll.Off
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, nil)
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, errors.New(errTextTraceError))
	parent.cfg.Loglevel = kmll.Error
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, errors.New(errTextTraceError))
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, nil)
	parent.cfg.Loglevel = kmll.Warning
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, errors.New(errTextTraceError))
	parent.cfg.Loglevel = kmll.Info
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, nil)
	parent.cfg.Loglevel = kmll.Level(127)
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, nil)
	silent = context.WithValue(context.Background(), keyContextLogLevel, keyLogSilent)
	NewLoggerGorm(parent).Trace(silent, time.Now(), fc, errors.New(errTextTraceError))
}

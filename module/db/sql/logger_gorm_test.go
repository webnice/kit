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
	parent := &impl{cfg: testConfig(testDriverName)}
	parent.cfg.Loglevel = kmll.Info
	logger := NewLoggerGorm(parent).(*logGorm)
	if logger == nil || logger.Loglevel != kmll.Info {
		t.Fatal("NewLoggerGorm() вернул некорректный объект")
	}

	logger.LogMode(gormLogger.LogLevel(123))
	if logger.Loglevel != kmll.Info {
		t.Fatal("LogMode(default) изменил уровень")
	}
	logger.LogMode(gormLogger.Silent)
	if logger.Loglevel != kmll.Off {
		t.Fatal("LogMode(Silent) некорректен")
	}
	logger.LogMode(gormLogger.Error)
	if logger.Loglevel != kmll.Error {
		t.Fatal("LogMode(Error) некорректен")
	}
	logger.LogMode(gormLogger.Warn)
	if logger.Loglevel != kmll.Warning {
		t.Fatal("LogMode(Warn) некорректен")
	}
	logger.LogMode(gormLogger.Info)
	if logger.Loglevel != kmll.Info {
		t.Fatal("LogMode(Info) некорректен")
	}

	logger.Info(context.Background(), "info %s", "msg")
	logger.Warn(context.Background(), "warn %s", "msg")
	logger.Error(context.Background(), "error %s", "msg")
}

func TestLoggerGorm_Trace(t *testing.T) {
	parent := &impl{cfg: testConfig(testDriverName)}
	fc := func() (string, int64) { return `select "1"`, 1 }

	parent.cfg.Loglevel = kmll.Off
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, nil)
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, errors.New("trace error"))

	parent.cfg.Loglevel = kmll.Error
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, errors.New("trace error"))
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, nil)

	parent.cfg.Loglevel = kmll.Warning
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, errors.New("trace error"))

	parent.cfg.Loglevel = kmll.Info
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, nil)

	parent.cfg.Loglevel = kmll.Level(127)
	NewLoggerGorm(parent).Trace(context.Background(), time.Now(), fc, nil)

	silent := context.WithValue(context.Background(), keyContextLogLevel, keyLogSilent)
	NewLoggerGorm(parent).Trace(silent, time.Now(), fc, errors.New("trace error"))
}

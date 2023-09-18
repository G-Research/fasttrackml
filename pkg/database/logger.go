package database

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	maximumCallerDepth int = 15
	minimumCallerDepth int = 4
)

type loggerAdaptor struct {
	Logger *logrus.Logger
	Config LoggerAdaptorConfig
}

type LoggerAdaptorConfig struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
}

func NewLoggerAdaptor(l *logrus.Logger, cfg LoggerAdaptorConfig) logger.Interface {
	return &loggerAdaptor{l, cfg}
}

// Needed to conform to the gorm logger.Interface interface
func (l *loggerAdaptor) LogMode(level logger.LogLevel) logger.Interface {
	l.Logger.Warn("LogMode does not do anything")
	return l
}

// Needed to conform to the gorm logger.Interface interface
func (l *loggerAdaptor) Info(ctx context.Context, format string, args ...interface{}) {
	l.getLoggerEntry(ctx).Infof(format, args...)
}

// Needed to conform to the gorm logger.Interface interface
func (l *loggerAdaptor) Warn(ctx context.Context, format string, args ...interface{}) {
	l.getLoggerEntry(ctx).Warnf(format, args...)
}

// Needed to conform to the gorm logger.Interface interface
func (l *loggerAdaptor) Error(ctx context.Context, format string, args ...interface{}) {
	l.getLoggerEntry(ctx).Errorf(format, args...)
}

// Needed to conform to the gorm logger.Interface interface
func (l *loggerAdaptor) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	if l.Logger.GetLevel() <= logrus.FatalLevel {
		return
	}

	// This logic is similar to the default logger in gorm.io/gorm/logger.go
	elapsed := time.Since(begin)
	switch {
	case err != nil &&
		l.Logger.IsLevelEnabled(logrus.ErrorLevel) &&
		(!errors.Is(err, gorm.ErrRecordNotFound) || !l.Config.IgnoreRecordNotFoundError):
		l.getLoggerEntryWithSql(ctx, elapsed, fc).WithError(err).Error("SQL error")
	case elapsed > l.Config.SlowThreshold &&
		l.Config.SlowThreshold != 0 &&
		l.Logger.IsLevelEnabled(logrus.WarnLevel):
		l.getLoggerEntryWithSql(ctx, elapsed, fc).Warnf("SLOW SQL >= %v", l.Config.SlowThreshold)
	case l.Logger.IsLevelEnabled(logrus.DebugLevel):
		l.getLoggerEntryWithSql(ctx, elapsed, fc).Debug("SQL trace")
	}
}

// Get a logger entry with context and caller information added
func (l *loggerAdaptor) getLoggerEntry(ctx context.Context) *logrus.Entry {
	e := l.Logger.WithContext(ctx)
	// We want to report the caller of the function that called gorm's logger,
	// not the caller of the loggerAdaptor, so we skip the first few frames and
	// then look for the first frame that is not in the gorm package.
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	for f, again := frames.Next(); again; f, again = frames.Next() {
		if !strings.HasPrefix(f.Function, "gorm.io/gorm") {
			e = e.WithFields(logrus.Fields{
				"app_file": fmt.Sprintf("%s:%d", f.File, f.Line),
				"app_func": fmt.Sprintf("%s()", f.Function),
			})
			break
		}
	}

	return e
}

// Get a logger entry with context, caller information and SQL information added
func (l *loggerAdaptor) getLoggerEntryWithSql(
	ctx context.Context,
	elapsed time.Duration,
	fc func() (sql string, rowsAffected int64),
) *logrus.Entry {
	e := l.getLoggerEntry(ctx)
	if fc != nil {
		sql, rows := fc()
		e = e.WithFields(logrus.Fields{
			"elapsed": fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
			"rows":    rows,
			"sql":     sql,
		})
		if rows == -1 {
			e = e.WithField("rows", "-")
		}
	}

	return e
}

// Needed to conform to the gorm ParamsFilter interface
func (l *loggerAdaptor) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}

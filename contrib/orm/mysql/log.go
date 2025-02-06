package mysql

import (
	"context"
	"errors"
	"fmt"
	"github.com/bpcoder16/Chestnut/v2/core/log"
	cUtils "github.com/bpcoder16/Chestnut/v2/core/utils"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type Logger struct {
	*log.Helper
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func NewLogger(helper *log.Helper, config logger.Config) *Logger {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	return &Logger{
		Helper:       helper,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "MySQL")
		l.Helper.WithContext(ctx).InfoF(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "MySQL")
		l.Helper.WithContext(ctx).WarnF(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "MySQL")
		l.Helper.WithContext(ctx).ErrorF(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	ctx = context.WithValue(ctx, log.DefaultDownstreamKey, "MySQL")
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		l.Helper.WithContext(ctx).ErrorW(
			"caller", utils.FileWithLineNum(),
			"error", err.Error(),
			"costTime", cUtils.ShowDurationString(elapsed),
			"row", func() interface{} {
				if rows == -1 {
					return "-"
				}
				return rows
			}(),
			"sql", sql,
		)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		l.Helper.WithContext(ctx).WarnW(
			"caller", utils.FileWithLineNum(),
			"error", fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold),
			"costTime", cUtils.ShowDurationString(elapsed),
			"row", func() interface{} {
				if rows == -1 {
					return "-"
				}
				return rows
			}(),
			"sql", sql,
		)
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		l.Helper.WithContext(ctx).DebugW(
			"caller", utils.FileWithLineNum(),
			"costTime", cUtils.ShowDurationString(elapsed),
			"row", func() interface{} {
				if rows == -1 {
					return "-"
				}
				return rows
			}(),
			"sql", sql,
		)
	}
}

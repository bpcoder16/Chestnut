package zaplogger

import (
	"io"
	"path"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/file/filerotatelogs"
	"github.com/bpcoder16/Chestnut/v4/core/file/standard"
)

const (
	defaultLogRetentionDays = 30
	logRotationTime         = time.Hour
)

// GetFileRotateLogWriters 使用默认保留时间创建按级别拆分的轮转日志写入器。
func GetFileRotateLogWriters(logDir, appName, logName string) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	return GetFileRotateLogWritersWithRetentionDays(logDir, appName, logName, 0)
}

// GetFileRotateLogWritersWithRetentionDays 使用指定保留天数创建按级别拆分的轮转日志写入器。
// retentionDays 为 0 时保持默认的 30 天保留时间。
func GetFileRotateLogWritersWithRetentionDays(logDir, appName, logName string, retentionDays int) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	maxAge := logRetentionMaxAge(retentionDays)
	debugWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".debug.log"),
		maxAge,
		logRotationTime,
	)
	infoWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".info.log"),
		maxAge,
		logRotationTime,
	)
	warnErrorFatalWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".wf.log"),
		maxAge,
		logRotationTime,
	)
	return
}

// GetFileRotateRequestLogWriter 使用默认保留时间创建请求轮转日志写入器。
func GetFileRotateRequestLogWriter(logDir, appName, logName string) io.Writer {
	return GetFileRotateRequestLogWriterWithRetentionDays(logDir, appName, logName, 0)
}

// GetFileRotateRequestLogWriterWithRetentionDays 使用指定保留天数创建请求轮转日志写入器。
// retentionDays 为 0 时保持默认的 30 天保留时间。
func GetFileRotateRequestLogWriterWithRetentionDays(logDir, appName, logName string, retentionDays int) io.Writer {
	return filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".request.log"),
		logRetentionMaxAge(retentionDays),
		logRotationTime,
	)
}

// GetFileRotateCronLogWriters 使用默认保留时间创建按级别拆分的 Cron 轮转日志写入器。
func GetFileRotateCronLogWriters(logDir, appName, logName string) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	return GetFileRotateCronLogWritersWithRetentionDays(logDir, appName, logName, 0)
}

// GetFileRotateCronLogWritersWithRetentionDays 使用指定保留天数创建按级别拆分的 Cron 轮转日志写入器。
// retentionDays 为 0 时保持默认的 30 天保留时间。
func GetFileRotateCronLogWritersWithRetentionDays(logDir, appName, logName string, retentionDays int) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	maxAge := logRetentionMaxAge(retentionDays)
	debugWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".cron.debug.log"),
		maxAge,
		logRotationTime,
	)
	infoWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".cron.info.log"),
		maxAge,
		logRotationTime,
	)
	warnErrorFatalWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".cron.wf.log"),
		maxAge,
		logRotationTime,
	)
	return
}

func logRetentionMaxAge(retentionDays int) time.Duration {
	if retentionDays <= 0 {
		return defaultLogRetentionDays * 24 * time.Hour
	}
	return time.Duration(retentionDays) * 24 * time.Hour
}

func GetStandardWriters(logDir, appName, logName string) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	debugWriter = standard.NewWriter(
		path.Join(logDir, appName, logName+".debug.log"),
	)
	infoWriter = standard.NewWriter(
		path.Join(logDir, appName, logName+".info.log"),
	)
	warnErrorFatalWriter = standard.NewWriter(
		path.Join(logDir, appName, logName+".wf.log"),
	)
	return
}

func GetStandardRequestWriter(logDir, appName, logName string) io.Writer {
	return standard.NewWriter(
		path.Join(logDir, appName, logName+".request.log"),
	)
}

func GetStandardCronWriters(logDir, appName, logName string) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	debugWriter = standard.NewWriter(
		path.Join(logDir, appName, logName+".cron.debug.log"),
	)
	infoWriter = standard.NewWriter(
		path.Join(logDir, appName, logName+".cron.info.log"),
	)
	warnErrorFatalWriter = standard.NewWriter(
		path.Join(logDir, appName, logName+".cron.wf.log"),
	)
	return
}

package zaplogger

import (
	"io"
	"path"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/file/filerotatelogs"
	"github.com/bpcoder16/Chestnut/v4/core/file/standard"
)

func GetFileRotateLogWriters(logDir, appName, logName string) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	debugWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".debug.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
	infoWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".info.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
	warnErrorFatalWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".wf.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
	return
}

func GetFileRotateRequestLogWriter(logDir, appName, logName string) io.Writer {
	return filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".request.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
}

func GetFileRotateCronLogWriters(logDir, appName, logName string) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	debugWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".cron.debug.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
	infoWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".cron.info.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
	warnErrorFatalWriter = filerotatelogs.NewWriter(
		path.Join(logDir, appName, logName+".cron.wf.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
	return
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

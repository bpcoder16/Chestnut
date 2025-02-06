package zaplogger

import (
	"github.com/bpcoder16/Chestnut/v2/core/file/filerotatelogs"
	"github.com/bpcoder16/Chestnut/v2/core/file/standard"
	"io"
	"path"
	"time"
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

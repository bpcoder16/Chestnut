package zaplogger

import (
	"github.com/bpcoder16/Chestnut/contrib/file/filerotatelogs"
	"io"
	"path"
	"time"
)

func GetWriters(rootPath, appName, logName string) (debugWriter, infoWriter, warnErrorFatalWriter io.Writer) {
	debugWriter = filerotatelogs.NewWriter(
		path.Join(rootPath, "log", appName, logName+".debug.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
	infoWriter = filerotatelogs.NewWriter(
		path.Join(rootPath, "log", appName, logName+".info.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
	warnErrorFatalWriter = filerotatelogs.NewWriter(
		path.Join(rootPath, "log", appName, logName+".wf.log"),
		time.Duration(86400*30)*time.Second,
		time.Duration(3600)*time.Second,
	)
	return
}

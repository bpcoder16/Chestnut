package filerotatelogs

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"io"
	"time"
)

func NewWriter(filePath string, maxAge, rotationTime time.Duration) io.Writer {
	rotateLogsPtr, err := rotatelogs.New(
		filePath+".%Y%m%d%H",
		rotatelogs.WithLinkName(filePath),         // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存份数
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		panic(err)
	}
	return rotateLogsPtr
}

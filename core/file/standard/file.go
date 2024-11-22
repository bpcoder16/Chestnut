package standard

import (
	"os"
	"path/filepath"
)

func NewWriter(filePath string) *os.File {
	dir := filepath.Dir(filePath)
	if errF := os.MkdirAll(dir, 0755); errF != nil {
		panic("Error creating directory:" + errF.Error())
	}

	// 打开文件（写入模式）
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic("open " + filePath + " failed: " + err.Error())
	}

	return file
}

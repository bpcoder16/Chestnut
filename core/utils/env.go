package utils

import "os"

func RootPath() string {
	rootPath, _ := os.Getwd()
	return rootPath
}

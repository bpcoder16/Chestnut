package utils

import (
	"errors"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

// ParseJSONFile
// DEPRECATED: 请使用 ParseFile 替代
func ParseJSONFile(path string, resPtr interface{}) (err error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("json")
	err = v.ReadInConfig()
	if err == nil {
		err = v.Unmarshal(resPtr)
	}
	return
}

func ParseFile(filePath string, resPtr interface{}) (err error) {
	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext)

	v := viper.New()
	v.SetConfigFile(filePath)
	switch ext {
	case "json":
		v.SetConfigType("json")
	case "yaml", "yml":
		v.SetConfigType("yaml")
	case "toml":
		v.SetConfigType("toml")
	default:
		err = errors.New("不支持的配置类型")
		return
	}
	err = v.ReadInConfig()
	if err == nil {
		err = v.Unmarshal(resPtr)
	}
	return
}

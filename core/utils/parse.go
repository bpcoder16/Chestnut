package utils

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func ParseFile(filePath string, resPtr interface{}) (err error) {
	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext)

	v := viper.New()
	v.SetConfigFile(filePath)
	switch ext {
	case ".json":
		v.SetConfigType("json")
	case ".yaml", ".yml":
		v.SetConfigType("yaml")
	case ".toml":
		v.SetConfigType("toml")
	default:
		err = errors.New("不支持的配置类型：" + ext)
		return
	}
	err = v.ReadInConfig()
	if err == nil {
		err = v.Unmarshal(resPtr)
	}
	return
}

func ParseContentYaml(content string, resPtr interface{}) (err error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err = viper.ReadConfig(strings.NewReader(content)); err == nil {
		err = v.Unmarshal(resPtr)
	}
	return
}

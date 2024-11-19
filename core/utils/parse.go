package utils

import "github.com/spf13/viper"

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

package geetest

import (
	"github.com/bpcoder16/Chestnut/v4/core/utils"
)

type Config struct {
	CaptchaID  string `yaml:"captchaId"`
	CaptchaKey string `yaml:"captchaKey"`
}

func LoadConfig(configPath string) *Config {
	var cfg Config
	if err := utils.ParseFile(configPath, &cfg); err != nil {
		panic("load geetest config failed: " + err.Error())
	}
	return &cfg
}

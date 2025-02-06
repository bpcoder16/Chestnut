package websocket

import (
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"time"
)

type Config struct {
	HandshakeTimeoutSec time.Duration `json:"handshakeTimeoutSec"`
	ReadBufferSize      int           `json:"readBufferSize"`
	WriteBufferSize     int           `json:"writeBufferSize"`
	WriteBufferPool     int           `json:"writeBufferPool"`
	AllowedOrigins      []string      `json:"allowedOrigins"`
	EnableCompression   bool          `json:"enableCompression"`
}

func loadConfig(configPath string) *Config {
	var config Config
	err := utils.ParseJSONFile(configPath, &config)
	if err != nil {
		panic("load WebSocket Server conf err:" + err.Error())
	}
	return &config
}

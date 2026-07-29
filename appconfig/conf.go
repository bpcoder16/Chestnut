package appconfig

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
)

type AppConfig struct {
	Env env.Option

	FilterKeys   []string
	Log          Log
	Default      Default
	Bootstrap    Bootstrap
	AsyncService AsyncService
	IPWhitelist  IPWhitelist
	Prometheus   Prometheus
}

type Log struct {
	LogDir                 string
	UseRotateLog           bool
	StdRedirectFileSupport bool
}

type Default struct {
	MySQLSupport               bool
	SQLiteSupport              bool
	ClickhouseSupport          bool
	RedisSupport               bool
	MongoDBSupport             bool
	LRUCacheSupport            bool
	AliyunOSSSupport           bool
	GeeTestSupport             bool
	LocalLockPoolSupport       bool
	SwaggerSupport             bool
	PProfSupport               bool
	CronSupport                bool
	CronDistributedLockSupport bool
	WebSocketSupport           bool
	NATSSupport                bool
}

type Bootstrap struct {
	AutoMigrate bool
}

type AsyncService struct {
	Support         bool
	QueueSize       int
	ConsumerSize    int
	TaskMaxRetryCnt int
}

// IPWhitelist 定义 HTTP 入口的远端 IP 白名单配置。
type IPWhitelist struct {
	Enabled  bool
	AllowIPs []string
}

const (
	minHTTPStatusCode = 100
	maxHTTPStatusCode = 599
)

// Prometheus 定义 Gin HTTP 指标的启停、身份和排除规则。
type Prometheus struct {
	Enabled             bool
	ServiceName         string
	ExcludedPaths       []string
	ExcludedStatusCodes []int
}

func (c *AppConfig) Check() (err error) {
	if len(c.Env.AppName) == 0 {
		err = errors.New("AppName required")
	}
	switch c.Env.RunMode {
	case env.RunModeDebug, env.RunModeTest, env.RunModeRelease:
	default:
		err = errors.New("invalid runMode: " + c.Env.RunMode)
	}
	if c.Default.WebSocketSupport && !c.Default.RedisSupport {
		err = errors.New("WebSocketSupport requires RedisSupport to be enabled")
	}
	if c.Default.CronDistributedLockSupport && !c.Default.CronSupport {
		err = errors.New("CronDistributedLockSupport requires CronSupport to be enabled")
	}
	if c.Default.CronDistributedLockSupport && !c.Default.RedisSupport {
		err = errors.New("CronDistributedLockSupport requires RedisSupport to be enabled")
	}
	if c.Prometheus.Enabled {
		if prometheusErr := c.Prometheus.check(); prometheusErr != nil {
			return prometheusErr
		}
	}
	return err
}

func (c Prometheus) check() error {
	if strings.TrimSpace(c.ServiceName) == "" {
		return errors.New("prometheus.serviceName required")
	}

	for _, excludedPath := range c.ExcludedPaths {
		if excludedPath == "" || strings.TrimSpace(excludedPath) != excludedPath ||
			!strings.HasPrefix(excludedPath, "/") || strings.ContainsAny(excludedPath, "?#") {
			return fmt.Errorf("prometheus.excludedPaths contains invalid URL path %q", excludedPath)
		}
	}
	for _, statusCode := range c.ExcludedStatusCodes {
		if statusCode < minHTTPStatusCode || statusCode > maxHTTPStatusCode {
			return fmt.Errorf("prometheus.excludedStatusCodes contains invalid HTTP status %d", statusCode)
		}
	}
	return nil
}

func ParseConfig(confPath string, configPtr *AppConfig) (err error) {
	if confPath, err = filepath.Abs(confPath); err == nil {
		if err = utils.ParseFile(confPath, configPtr); err == nil {
			err = configPtr.Check()
		}
	}
	if len(configPtr.Log.LogDir) == 0 {
		configPtr.Log.LogDir = path.Join(utils.RootPath(), "log")
	}
	return
}

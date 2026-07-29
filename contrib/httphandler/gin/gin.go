package gin

import (
	"net/http"
	"os"
	"sync"

	"github.com/bpcoder16/Chestnut/v4/appconfig"
	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
	"github.com/bpcoder16/Chestnut/v4/contrib/validator"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var (
	once sync.Once
)

func lazyInit() {
	once.Do(func() {
		switch env.RunMode() {
		case env.RunModeRelease:
			gin.SetMode(gin.ReleaseMode)
		case env.RunModeTest:
			gin.SetMode(gin.TestMode)
		default:
			gin.SetMode(gin.DebugMode)
		}

		binding.Validator = &validator.MultiLangValidator{
			Locale:  "zh",
			TagName: "binding",
		}
	})
}

// HTTPHandler 创建使用 Chestnut 默认中间件的 Gin Engine。
func HTTPHandler(registerFuncs ...func(*gin.RouterGroup)) *gin.Engine {
	return buildHTTPHandler(nil, nil, registerFuncs...)
}

// HTTPHandlerWithMiddlewares 创建 Gin Engine，并在内置路由注册前挂载全局中间件。
func HTTPHandlerWithMiddlewares(middlewares []gin.HandlerFunc, registerFuncs ...func(*gin.RouterGroup)) *gin.Engine {
	return buildHTTPHandler(nil, middlewares, registerFuncs...)
}

// HTTPHandlerWithConfig 根据应用配置创建 Gin Engine，并组合调用方全局中间件。
func HTTPHandlerWithConfig(config *appconfig.AppConfig, middlewares []gin.HandlerFunc, registerFuncs ...func(*gin.RouterGroup)) *gin.Engine {
	var prometheusConfig *appconfig.Prometheus
	if config != nil {
		prometheusConfig = &config.Prometheus
	}
	return buildHTTPHandler(prometheusConfig, middlewares, registerFuncs...)
}

func buildHTTPHandler(prometheusConfig *appconfig.Prometheus, middlewares []gin.HandlerFunc, registerFuncs ...func(*gin.RouterGroup)) *gin.Engine {
	lazyInit()
	h := gin.New()
	h.ContextWithFallback = true

	var metrics *prometheusHTTPMetrics
	if prometheusConfig != nil && prometheusConfig.Enabled {
		metrics = newPrometheusHTTPMetrics(*prometheusConfig)
		// 指标必须位于 Recovery 外层，才能观察未提交响应的 500 或已提交响应的实际状态。
		h.Use(metrics.middleware())
	}
	h.Use(RecoveryWithWriter(os.Stderr))
	for _, middleware := range middlewares {
		h.Use(middleware)
	}
	h.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	if metrics != nil {
		h.GET(prometheusMetricsPath, gin.WrapH(metrics.handler()))
	}
	for _, fn := range registerFuncs {
		fn(&h.RouterGroup)
	}
	return h
}

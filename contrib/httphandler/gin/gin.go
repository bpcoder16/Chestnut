package gin

import (
	"net/http"
	"os"
	"sync"

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

func HTTPHandler(registerFuncs ...func(*gin.RouterGroup)) *gin.Engine {
	lazyInit()
	h := gin.New()
	h.Use(RecoveryWithWriter(os.Stderr))
	h.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	for _, fn := range registerFuncs {
		fn(&h.RouterGroup)
	}
	return h
}

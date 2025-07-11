package gin

import (
	"github.com/bpcoder16/Chestnut/v2/appconfig/env"
	"github.com/bpcoder16/Chestnut/v2/contrib/validator"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"os"
	"sync"
)

var (
	once sync.Once
)

func initGinConfig() {
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

type Router interface {
	RegisterHandler(engine *gin.Engine)
}

func HTTPHandler(routers ...Router) http.Handler {
	initGinConfig()
	h := gin.New()
	h.Use(
		corsPreCheckRequest(),
		recoveryWithWriter(os.Stderr),
	)
	for _, router := range routers {
		router.RegisterHandler(h)
	}
	return h
}

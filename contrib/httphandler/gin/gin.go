package gin

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Router interface {
	RegisterHandler(engine *gin.Engine)
}

var h = gin.New()

func HTTPHandler(routers ...Router) http.Handler {
	for _, router := range routers {
		router.RegisterHandler(h)
	}
	return h
}

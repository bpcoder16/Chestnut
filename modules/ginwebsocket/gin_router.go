package ginwebsocket

import (
	"context"
	ginHandler "github.com/bpcoder16/Chestnut/contrib/httphandler/gin"
	"github.com/bpcoder16/Chestnut/contrib/websocket"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	basePath = "/ws"
)

var _ ginHandler.Router = (*Router)(nil)

type Router struct {
	wsRouter  *ginHandler.DefaultRouter
	wsManager *websocket.WebSocket
	path      string
}

func (r *Router) SetAuthorizationFunc(f func(context.Context) bool) {
	r.wsRouter.Use(func(ctx *gin.Context) {
		if !f(ctx) {
			ctx.String(http.StatusForbidden, "")
			ctx.Abort()
			return
		}
		ctx.Next()
	})
	r.wsManager.SetAuthorizationFunc(f)
}

func (r *Router) SetClientCloseFunc(f func(context.Context, string)) {
	r.wsManager.SetClientCloseFunc(f)
}

func (r *Router) OnTextMessageController(scene string, controller websocket.TextMessageController) {
	r.wsManager.OnTextMessageController(scene, controller)
}

func (r *Router) GetClientManager() *websocket.ClientManager {
	return r.wsManager.GetClientManager()
}

func (r *Router) RegisterHandler(engine *gin.Engine) {
	r.wsRouter.GET(r.path, func(ctx *gin.Context) {
		r.wsManager.Handle(ctx, ctx.Writer, ctx.Request, ctx.Request.Header)
	})
	r.wsRouter.RegisterHandler(engine)
}

func NewRouter(path, configPath string) *Router {
	r := &Router{
		wsRouter:  ginHandler.NewRouterNoLogger(basePath),
		wsManager: websocket.New(configPath),
		path:      path,
	}
	return r
}

package ginwebsocket

import (
	"github.com/bpcoder16/Chestnut/v4/contrib/websocket"
	"github.com/gin-gonic/gin"
)

const (
	basePath = "/ws"
)

type Router struct {
	wsManager *websocket.WebSocket
	path      string
}

func (r *Router) GetClientManager() *websocket.ClientManager {
	return r.wsManager.GetClientManager()
}

func (r *Router) OnTextMessageController(scene string, controller websocket.TextMessageController) error {
	return r.wsManager.OnTextMessageController(scene, controller)
}

func (r *Router) SetAuthorizationFunc(f websocket.AuthorizationFunc) {
	r.wsManager.SetAuthorizationFunc(f)
}

func (r *Router) SetBeforeFunc(f websocket.AuthorizationFunc) {
	r.wsManager.SetBeforeFunc(f)
}

func (r *Router) SetClientCloseFunc(f websocket.ClientCloseFunc) {
	r.wsManager.SetClientCloseFunc(f)
}

// Register 将 WebSocket 路由注册到指定 RouterGroup，供 HTTPHandler 使用。
func (r *Router) Register(rg *gin.RouterGroup) {
	ws := rg.Group(basePath)
	ws.GET(r.path, func(ctx *gin.Context) {
		r.wsManager.Handle(ctx, r.path, ctx.Request, ctx.Writer)
	})
}

func NewRouter(path, configPath string) *Router {
	return &Router{
		wsManager: websocket.New(configPath),
		path:      path,
	}
}

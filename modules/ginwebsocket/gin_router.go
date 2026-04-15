package ginwebsocket

import (
	"context"
	"net/http"

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

func (r *Router) OnTextMessageController(scene string, controller websocket.TextMessageController) {
	r.wsManager.OnTextMessageController(scene, controller)
}

func (r *Router) SetAuthorizationFunc(f func(ctx context.Context, r *http.Request, w http.ResponseWriter) (returnCtx context.Context, isAuthorized bool, userId int64)) {
	r.wsManager.SetAuthorizationFunc(f)
}

func (r *Router) SetBeforeFunc(f func(ctx context.Context, r *http.Request, w http.ResponseWriter) (returnCtx context.Context, isAuthorized bool, userId int64)) {
	r.wsManager.SetBeforeFunc(f)
}

func (r *Router) SetClientCloseFunc(f func(context.Context, string)) {
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

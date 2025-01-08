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

func (r *Router) SetBeforeFunc(f func(ctx context.Context, r *http.Request, w http.ResponseWriter) (returnCtx context.Context, isAuthorized bool, userId int64)) {
	r.wsManager.SetBeforeFunc(f)
}

func (r *Router) SetAuthorizationFunc(f func(ctx context.Context, r *http.Request, w http.ResponseWriter) (returnCtx context.Context, isAuthorized bool, userId int64)) {
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
		r.wsManager.Handle(ctx, r.path, ctx.Request, ctx.Writer)
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

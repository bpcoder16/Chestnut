package websocket

import (
	"context"
	"errors"
	"github.com/bpcoder16/Chestnut/core/gtask"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/bpcoder16/Chestnut/core/utils"
	"github.com/bpcoder16/Chestnut/logit"
	"github.com/gorilla/websocket"
	"net/http"
	"reflect"
	"sync"
	"time"
)

const (
	ConnUUIDCTXKey = "WebSocketConnUUIDCTXKey"
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	maxMessageSize = 1024 * 1024

	readDeadlineDuration = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (readDeadlineDuration * 9) / 10
)

func getUpgrader(config *Config) *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout: config.HandshakeTimeoutSec * time.Second,
		ReadBufferSize:   config.ReadBufferSize,
		WriteBufferSize:  config.WriteBufferSize,
		WriteBufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, config.WriteBufferPool)
			},
		},
		CheckOrigin: func(r *http.Request) bool {
			if len(config.AllowedOrigins) == 0 {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, allowedOrigin := range config.AllowedOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			return false
		},
		EnableCompression: config.EnableCompression,
	}
}

// TODO 临时解决方案，由于 gorilla/websocket 不支持 Sec-WebSocket-Extensions Header
func (ws *WebSocket) filterHeader(h http.Header) http.Header {
	h.Del("Sec-WebSocket-Extensions")
	return h
}

type WebSocket struct {
	config   *Config
	upgrader *websocket.Upgrader

	textMessageControllers map[string]TextMessageController
	authorizationFunc      func(context.Context) bool
	clientCloseFunc        func(ctx context.Context, uuidStr string)
	clientManager          *ClientManager
}

func New(configPath string) *WebSocket {
	config := loadConfig(configPath)
	ws := &WebSocket{
		config:   config,
		upgrader: getUpgrader(config),

		textMessageControllers: make(map[string]TextMessageController),
		authorizationFunc:      nil,
		clientCloseFunc:        nil,
		clientManager:          NewClientManager(),
	}
	return ws
}

func (ws *WebSocket) SetAuthorizationFunc(f func(context.Context) bool) {
	ws.authorizationFunc = f
}

func (ws *WebSocket) SetClientCloseFunc(f func(context.Context, string)) {
	ws.clientCloseFunc = f
}

func (ws *WebSocket) OnTextMessageController(scene string, controller TextMessageController) {
	ws.textMessageControllers[scene] = controller
}

func (ws *WebSocket) getTextMessageController(scene string) (controller TextMessageController, err error) {
	var exist bool
	var controllerTemplate TextMessageController
	controllerTemplate, exist = ws.textMessageControllers[scene]
	if !exist {
		err = errors.New("textMessageController not register")
		return
	}
	controller, _ = reflect.New(reflect.TypeOf(controllerTemplate).Elem()).Interface().(TextMessageController)
	controller.Init(controllerTemplate)
	return
}

func (ws *WebSocket) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request, responseHeader http.Header) {
	// 设置全局日志内容
	ctx = context.WithValue(ctx, log.DefaultMessageKey, "WebSocket")
	var uuidStr string
	var isOK bool
	if uuidStr, isOK = ctx.Value(ConnUUIDCTXKey).(string); !isOK {
		uuidStr = utils.UniqueID()
	}
	ctx = context.WithValue(ctx, log.DefaultLogIdKey, uuidStr)
	ctx = context.WithValue(ctx, log.DefaultWebSocketUUIDKey, uuidStr)
	ctx = context.WithValue(ctx, log.DefaultWebSocketLogIdKey, utils.UniqueID())

	begin := time.Now()
	conn, err := ws.upgrader.Upgrade(w, r, ws.filterHeader(responseHeader))
	elapsed := time.Since(begin)
	if err != nil {
		logit.Context(ctx).InfoW(
			"Connection.Status", "Failed",
			"Connection.CostTime", utils.ShowDurationString(elapsed),
			"Websocket.Upgrade.Err", err,
		)
		logit.Context(ctx).WarnW(
			"Connection.Status", "Failed",
			"Connection.CostTime", utils.ShowDurationString(elapsed),
			"Websocket.Upgrade.Err", err,
		)
		return
	}

	client := NewClient(conn, uuidStr)
	client.ws = ws
	client.ws.clientManager.Store(uuidStr, client)

	client.infoLog(ctx,
		"Connection.Status", "Success",
		"Connection.CostTime", utils.ShowDurationString(elapsed),
		"client.ws.clientManager.Len()", client.ws.clientManager.Len(),
	)

	// 设置连接重要参数
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
	_ = conn.SetReadDeadline(time.Now().Add(readDeadlineDuration))
	conn.SetCloseHandler(func(code int, text string) (err error) {
		client.close(ctx)
		client.debugLog(ctx,
			"function", "SetCloseHandler",
			"code", code,
			"text", text,
			"client.ws.clientManager.Len()", client.ws.clientManager.Len(),
		)
		return
	})
	conn.SetPingHandler(func(appData string) (err error) {
		client.debugLog(ctx,
			"function", "SetPingHandler",
			"appData", appData,
		)
		_ = conn.SetReadDeadline(time.Now().Add(readDeadlineDuration))
		return
	})
	conn.SetPongHandler(func(appData string) (err error) {
		client.debugLog(ctx,
			"function", "SetPongHandler",
			"appData", appData,
		)
		_ = conn.SetReadDeadline(time.Now().Add(readDeadlineDuration))
		return
	})

	g, gCtx := gtask.WithContext(ctx)

	g.Go(func() (err error) {
		client.readPump(gCtx)
		return
	})

	g.Go(func() (err error) {
		client.writePump(gCtx)
		return
	})

	_ = g.Wait()
}

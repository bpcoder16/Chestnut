package websocket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/gtask"
	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/bpcoder16/Chestnut/v4/logit"
	"github.com/gorilla/websocket"
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

func New(configPath string) *WebSocket {
	config := loadConfig(configPath)
	ws := &WebSocket{
		config:   config,
		upgrader: getUpgrader(config),

		clientManager: NewClientManager(),

		textMessageControllers: make(map[string]TextMessageController),
		authorizationFunc:      nil,
		beforeFunc:             nil,
		clientCloseFunc:        nil,
	}
	return ws
}

type WebSocket struct {
	mu sync.RWMutex

	config   *Config
	upgrader *websocket.Upgrader

	clientManager *ClientManager

	textMessageControllers map[string]TextMessageController
	authorizationFunc      AuthorizationFunc
	beforeFunc             AuthorizationFunc
	clientCloseFunc        ClientCloseFunc
}

type AuthorizationFunc func(ctx context.Context, r *http.Request, w http.ResponseWriter) (returnCtx context.Context, isAuthorized bool, userID string)

type ClientCloseFunc func(ctx context.Context, uuidStr string)

func (ws *WebSocket) GetClientManager() *ClientManager {
	return ws.clientManager
}

func (ws *WebSocket) OnTextMessageController(scene string, controller TextMessageController) error {
	if scene == "" {
		return errors.New("websocket scene empty")
	}
	if err := validateTextMessageController(controller); err != nil {
		return err
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.textMessageControllers[scene] = controller
	return nil
}

func (ws *WebSocket) SetAuthorizationFunc(f AuthorizationFunc) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.authorizationFunc = f
}

func (ws *WebSocket) SetBeforeFunc(f AuthorizationFunc) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.beforeFunc = f
}

func (ws *WebSocket) SetClientCloseFunc(f ClientCloseFunc) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.clientCloseFunc = f
}

func (ws *WebSocket) getBeforeFunc() AuthorizationFunc {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.beforeFunc
}

func (ws *WebSocket) getAuthorizationFunc() AuthorizationFunc {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.authorizationFunc
}

func (ws *WebSocket) getClientCloseFunc() ClientCloseFunc {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.clientCloseFunc
}

func (ws *WebSocket) getTextMessageController(scene string) (controller TextMessageController, err error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	var exist bool
	var controllerTemplate TextMessageController
	controllerTemplate, exist = ws.textMessageControllers[scene]
	if !exist {
		err = errors.New("TextMessageController.Not.Register")
		return
	}
	controllerType := reflect.TypeOf(controllerTemplate)
	if controllerType == nil || controllerType.Kind() != reflect.Ptr {
		err = errors.New("TextMessageController.Must.Be.Pointer")
		return
	}
	controller, _ = reflect.New(controllerType.Elem()).Interface().(TextMessageController)
	if controller == nil {
		err = errors.New("TextMessageController.New.Failed")
		return
	}
	controller.Init(controllerTemplate)
	return
}

func (ws *WebSocket) before(ctx context.Context, path string, r *http.Request, w http.ResponseWriter) (ctxNew context.Context, uuidStr string, isAuthorized bool, userID string) {
	ctxNew = context.WithValue(ctx, log.DefaultMessageKey, "WebSocket")
	ctxNew = context.WithValue(ctxNew, log.DefaultWebSocketPathKey, path)
	ctxNew = context.WithValue(ctxNew, log.DefaultWebSocketLogIdKey, utils.UniqueID())
	ctxNew = context.WithValue(ctxNew, log.DefaultLogIdKey, utils.UniqueID())
	if f := ws.getBeforeFunc(); f != nil {
		ctxNew, isAuthorized, userID = f(ctxNew, r, w)
		if !isAuthorized {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	} else {
		isAuthorized = true
	}
	var isOK bool
	if uuidStr, isOK = ctxNew.Value(ConnUUIDCTXKey).(string); !isOK {
		uuidStr = utils.UniqueID()
	}
	ctxNew = context.WithValue(ctxNew, log.DefaultWebSocketUUIDKey, uuidStr)
	return
}

func (ws *WebSocket) Handle(ctx context.Context, path string, r *http.Request, w http.ResponseWriter) {
	ctx, uuidStr, isAuthorized, userID := ws.before(ctx, path, r, w)
	if !isAuthorized {
		return
	}

	begin := time.Now()
	conn, err := ws.upgrader.Upgrade(w, r, w.Header())
	elapsed := time.Since(begin)
	if err != nil {
		logit.Context(ctx).WarnW(
			"Connection.Status", "Failed",
			"Connection.CostTime", utils.ShowDurationString(elapsed),
			"Websocket.Upgrade.Err", err,
		)
		return
	}

	client := NewClient(conn, uuidStr, userID)
	client.ws = ws
	client.ws.clientManager.StoreClient(client)
	defer client.close(ctx, "Handle.Defer")

	client.infoLog(ctx,
		"Connection.Status", "Success",
		"Connection.CostTime", utils.ShowDurationString(elapsed),
	)

	// 设置连接重要参数
	conn.SetReadLimit(maxMessageSize)
	conn.SetCloseHandler(func(code int, text string) (err error) {
		client.close(ctx, "SetCloseHandler")
		client.debugLog(ctx,
			"code", code,
			"text", text,
			"ReceiveMessageType", "Close",
		)
		return
	})
	conn.SetPingHandler(func(appData string) (err error) {
		client.debugLog(ctx,
			"appData", appData,
			"ReceiveMessageType", "Ping",
		)
		_ = conn.SetReadDeadline(time.Now().Add(readDeadlineDuration))
		return
	})
	conn.SetPongHandler(func(appData string) (err error) {
		client.debugLog(ctx,
			"appData", appData,
			"ReceiveMessageType", "Pong",
		)
		_ = conn.SetReadDeadline(time.Now().Add(readDeadlineDuration))
		return
	})

	g, gCtx := gtask.WithContext(ctx)

	g.Go(func() error {
		return client.readPump(gCtx, r, w)
	})
	g.Go(func() error {
		return client.writePump(gCtx, r, w)
	})

	_ = g.Wait()
}

func validateTextMessageController(controller TextMessageController) error {
	if controller == nil {
		return errors.New("websocket text message controller nil")
	}
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() != reflect.Ptr {
		return fmt.Errorf("websocket text message controller must be pointer: %s", controllerType.String())
	}
	if _, ok := reflect.New(controllerType.Elem()).Interface().(TextMessageController); !ok {
		return fmt.Errorf("websocket text message controller must implement TextMessageController: %s", controllerType.String())
	}
	return nil
}

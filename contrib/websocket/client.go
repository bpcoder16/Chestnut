package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bpcoder16/Chestnut/logit"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	ws   *WebSocket
	conn *websocket.Conn

	textMsgCh chan []byte
	isClosed  bool
	uuidStr   string
	State     State // 客户端状态信息

	mu sync.RWMutex
}

type State struct {
	Scene       string
	SID         string
	SceneParams map[string]interface{}
}

func NewClient(conn *websocket.Conn, uuidStr string) *Client {
	return &Client{
		conn: conn,

		textMsgCh: make(chan []byte, 1024),
		isClosed:  false,
		uuidStr:   uuidStr,
		State: State{
			SceneParams: make(map[string]interface{}),
		},
	}
}

func (c *Client) close(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if false == c.isClosed {
		_ = c.conn.Close()
		c.isClosed = true
		close(c.textMsgCh)
		if c.ws.clientCloseFunc != nil {
			c.ws.clientCloseFunc(ctx, c.uuidStr)
		}
	}
	c.ws.clientManager.Delete(c.uuidStr)
}

func (c *Client) log(ctx context.Context, level string, keyValues ...interface{}) {
	newKeyValues := []interface{}{
		"subProtocol", c.conn.Subprotocol(),
		"localAddr", c.conn.LocalAddr().String(),
		"remoteAddr", c.conn.RemoteAddr().String(),
	}
	newKeyValues = append(keyValues, newKeyValues...)

	switch level {
	case "DEBUG":
		logit.Context(ctx).DebugW(newKeyValues...)
	case "INFO":
		logit.Context(ctx).InfoW(newKeyValues...)
	case "WARN":
		logit.Context(ctx).WarnW(newKeyValues...)
	case "ERROR":
		logit.Context(ctx).ErrorW(newKeyValues...)
	}
	return
}

func (c *Client) debugLog(ctx context.Context, keyValues ...interface{}) {
	c.log(ctx, "DEBUG", keyValues...)
}

func (c *Client) infoLog(ctx context.Context, keyValues ...interface{}) {
	c.log(ctx, "INFO", keyValues...)
}

func (c *Client) warnLog(ctx context.Context, keyValues ...interface{}) {
	c.log(ctx, "WARN", keyValues...)
}

func (c *Client) errorLog(ctx context.Context, keyValues ...interface{}) {
	c.log(ctx, "ERROR", keyValues...)
}

// TODO  SetWriteDeadline 设置写超时

func (c *Client) readMessage(ctx context.Context) (messageType int, message []byte, err error) {
	messageType, message, err = c.conn.ReadMessage()
	c.debugLog(ctx,
		"function", "ReadMessage",
		"messageType", messageType,
		"message", string(message),
		"err", err,
	)
	if err != nil {
		c.close(ctx)
	}
	return
}

func (c *Client) WriteTextMessage(ctx context.Context, message []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("c.textMsgCh is closed")
			c.errorLog(ctx,
				"function", "WriteTextMessage",
				"messageType", "Text",
				"message", string(message),
				"err", err,
			)
			c.close(ctx)
		}
	}()
	if !c.isClosed {
		c.textMsgCh <- message
		c.debugLog(ctx,
			"function", "WriteTextMessage",
			"messageType", "Text",
			"message", string(message),
		)
	} else {
		err = errors.New("c.textMsgCh is closed")
	}

	return
}

func (c *Client) receiveTextMessage(ctx context.Context, messageBytes []byte) (err error) {
	if len(messageBytes) == 0 {
		c.warnLog(ctx,
			"function", "receiveTextMessage",
			"messageBytes.Len", 0,
		)
		return
	}
	var receiveMessage ReceiveMessage
	err = json.Unmarshal(messageBytes, &receiveMessage)
	if err != nil {
		c.warnLog(ctx,
			"function", "receiveTextMessage",
			"messageBytes.Err", errors.New("parse text message failed["+err.Error()+"]"),
		)
		return
	}
	c.debugLog(ctx,
		"function", "receiveTextMessage",
		"process", "parse text success",
		"receiveMessage", receiveMessage,
	)
	if len(receiveMessage.Scene) == 0 {
		if len(c.State.Scene) == 0 {
			c.warnLog(ctx,
				"function", "receiveTextMessage",
				"receiveMessage.Scene", receiveMessage.Scene,
				"State.Scene", c.State.Scene,
			)
			err = errors.New("receiveMessage.Scene.Empty")
			return
		}
		// 如果没有传递，说明用户停留在当前场景
		receiveMessage.Scene = c.State.Scene
		receiveMessage.SceneParams = c.State.SceneParams
		receiveMessage.SID = c.State.SID
	}

	var controller TextMessageController
	controller, err = c.ws.getTextMessageController(receiveMessage.Scene)
	if err != nil {
		c.warnLog(ctx,
			"function", "receiveTextMessage",
			"receiveMessage.Scene", receiveMessage.Scene,
			"getTextMessageController.Err", err,
		)
		return
	}
	err = controller.ParsePayload(c, receiveMessage)
	if err != nil {
		c.warnLog(ctx,
			"function", "receiveTextMessage",
			"receiveMessage.Scene", receiveMessage.Scene,
			"controller.ParsePayload.Err", err,
		)
		return
	}

	err = controller.Process()

	return
}

func (c *Client) receiveBinaryMessage(_ context.Context, _ []byte) (err error) {
	return
}

func (c *Client) receiveCloseMessage(_ context.Context, _ []byte) (err error) {
	return
}

func (c *Client) receivePingMessage(_ context.Context, _ []byte) (err error) {
	return
}

func (c *Client) receivePongMessage(_ context.Context, _ []byte) (err error) {
	return
}

func (c *Client) readPump(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.errorLog(ctx,
				"function", "client.readPump",
				"recover", r,
			)
		}
		c.close(ctx)
	}()

	for {
		mt, message, errR := c.readMessage(ctx)
		if errR != nil {
			if websocket.IsUnexpectedCloseError(errR, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.warnLog(ctx,
					"function", "client.readPump",
					"process", "readMessage",
					"err", errR.Error(),
				)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		var messageTypeShow string
		begin := time.Now()
		switch mt {
		case websocket.TextMessage:
			errR = c.receiveTextMessage(ctx, message)
			messageTypeShow = "Text"
		case websocket.BinaryMessage:
			errR = c.receiveBinaryMessage(ctx, message)
			messageTypeShow = "Binary"
		case websocket.CloseMessage:
			errR = c.receiveCloseMessage(ctx, message)
			messageTypeShow = "Close"
		case websocket.PingMessage:
			errR = c.receivePingMessage(ctx, message)
			messageTypeShow = "Ping"
		case websocket.PongMessage:
			errR = c.receivePongMessage(ctx, message)
			messageTypeShow = "Pong"
		}

		elapsed := time.Since(begin)
		c.infoLog(ctx,
			"function", "client.readPump",
			"messageType", messageTypeShow,
			"readMessage.Err", errR,
			"costTime", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
		)

		if errR != nil {
			c.warnLog(ctx,
				"function", "client.readPump",
				"process", "readMessage.receiveMessage",
				"err", errR.Error(),
			)
			break
		}
	}
}

func (c *Client) sendPing(ctx context.Context) error {
	c.debugLog(ctx,
		"function", "sendPing",
		"messageType", "Ping",
	)
	return c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
}

func (c *Client) writePump(ctx context.Context) {
	// 维持心跳
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		if r := recover(); r != nil {
			c.errorLog(ctx,
				"function", "client.writePump",
				"recover", r,
			)
		}
		c.close(ctx)
	}()

	for {
		select {
		case message, ok := <-c.textMsgCh:
			if !ok {
				_ = c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
				return
			}

			c.mu.Lock()
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			n := len(c.textMsgCh)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				_, _ = w.Write(<-c.textMsgCh)
			}
			if errW := w.Close(); errW != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		case <-ticker.C:
			if err := c.sendPing(ctx); err != nil {
				return
			}
			// 鉴权与心跳一个频次校验
			if c.ws.authorizationFunc != nil {
				if isOK := c.ws.authorizationFunc(ctx); !isOK {
					c.close(ctx)
					return
				}
			}
		}
	}
}

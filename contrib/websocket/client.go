package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/bpcoder16/Chestnut/core/log"
	"github.com/bpcoder16/Chestnut/core/utils"
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

func (c *Client) close(ctx context.Context, sourceText string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	isClosedForLog := c.isClosed
	if false == c.isClosed {
		_ = c.sendCloseMessage(ctx)
		_ = c.conn.Close()
		c.isClosed = true
		close(c.textMsgCh)
		if c.ws.clientCloseFunc != nil {
			c.ws.clientCloseFunc(ctx, c.uuidStr)
		}
	}
	c.ws.clientManager.Delete(c.uuidStr)
	c.debugLog(ctx,
		"sourceText", sourceText,
		"function", "Client.close",
		"client.isClosed", isClosedForLog,
		"client.ws.clientManager", "Delete("+c.uuidStr+")",
	)
}

func (c *Client) log(ctx context.Context, level string, keyValues ...interface{}) {
	newKeyValues := []interface{}{
		"subProtocol", c.conn.Subprotocol(),
		"localAddr", c.conn.LocalAddr().String(),
		"remoteAddr", c.conn.RemoteAddr().String(),
		"client.ws.clientManager.Len()", c.ws.clientManager.Len(),
		"client.State", c.State,
	}
	newKeyValues = append(newKeyValues, keyValues...)

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

func (c *Client) getMessageTypeString(messageType int) string {
	return map[int]string{
		websocket.TextMessage:   "Text",
		websocket.BinaryMessage: "Binary",
		websocket.CloseMessage:  "Close",
		websocket.PingMessage:   "Ping",
		websocket.PongMessage:   "Pong",
	}[messageType]
}

func (c *Client) WriteTextMessage(ctx context.Context, message []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("c.textMsgCh.Is.Closed")
			c.close(ctx, "WriteTextMessage.Panic")
		}
		if err != nil {
			c.errorLog(ctx,
				"function", "WriteTextMessage",
				"message", string(message),
				"err", err,
			)
		}
	}()
	c.debugLog(ctx,
		"client.isClosed", c.isClosed,
		"function", "WriteTextMessage",
		"message", string(message),
	)
	if !c.isClosed {
		c.textMsgCh <- message
	} else {
		err = errors.New("c.textMsgCh.Is.Closed")
	}

	return
}

func (c *Client) receiveTextMessage(ctx context.Context, messageBytes []byte) (err error) {
	if len(messageBytes) == 0 {
		err = errors.New("messageBytes.Len(0)")
		return
	}
	var receiveMessage ReceiveMessage
	err = json.Unmarshal(messageBytes, &receiveMessage)
	if err != nil {
		err = errors.New("parse text message failed[" + err.Error() + "]")
		return
	}
	if len(receiveMessage.Scene) == 0 {
		if len(c.State.Scene) == 0 {
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
		return
	}
	err = controller.ParsePayload(ctx, c, receiveMessage)
	if err != nil {
		return
	}

	err = controller.Process(ctx)
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
		c.close(ctx, "ReadPump.Defer")
	}()

	for {
		mt, message, errR := c.conn.ReadMessage()
		rCtx := context.WithValue(ctx, log.DefaultWebSocketLogIdKey, utils.UniqueID())
		if errR != nil {
			if !websocket.IsUnexpectedCloseError(errR, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.warnLog(rCtx,
					"function", "client.readPump",
					"process", "c.readMessage",
					"err", errR.Error(),
				)
			}
			break
		}
		message = bytes.TrimSpace(message)

		begin := time.Now()
		switch mt {
		case websocket.TextMessage:
			errR = c.receiveTextMessage(rCtx, message)
		case websocket.BinaryMessage:
			errR = c.receiveBinaryMessage(rCtx, message)
		case websocket.CloseMessage:
			errR = c.receiveCloseMessage(rCtx, message)
		case websocket.PingMessage:
			errR = c.receivePingMessage(rCtx, message)
		case websocket.PongMessage:
			errR = c.receivePongMessage(rCtx, message)
		}

		elapsed := time.Since(begin)
		c.infoLog(rCtx,
			"function", "client.readPump",
			"process", "readMessage.receiveMessage",
			"readMessage.Err", errR,
			"messageType", c.getMessageTypeString(mt),
			"message", string(message),
			"costTime", utils.ShowDurationString(elapsed),
		)

		if errR != nil {
			c.warnLog(rCtx,
				"function", "client.readPump",
				"process", "readMessage.receiveMessage",
				"readMessage.Err", errR,
			)
			break
		}
	}
}

func (c *Client) sendPingMessage(ctx context.Context) error {
	c.debugLog(ctx,
		"SendMessageType", "Ping",
	)
	return c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
}

func (c *Client) sendCloseMessage(ctx context.Context) error {
	c.debugLog(ctx,
		"SendMessageType", "Close",
	)
	return c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
}

func (c *Client) writePump(ctx context.Context) {
	//defer func() {
	//	if r := recover(); r != nil {
	//		c.errorLog(ctx,
	//			"function", "client.writePump",
	//			"recover", r,
	//		)
	//	}
	//	c.close(ctx, "WritePump.Defer")
	//}()

	// 维持心跳
	//ticker := time.NewTicker(pingPeriod)
	//for {
	//	select {
	//	case message, ok := <-c.textMsgCh:
	//		if !ok {
	//			return
	//		}
	//
	//		c.mu.Lock()
	//		w, err := c.conn.NextWriter(websocket.TextMessage)
	//		if err != nil {
	//			return
	//		}
	//		_, _ = w.Write(message)
	//
	//		n := len(c.textMsgCh)
	//		for i := 0; i < n; i++ {
	//			_, _ = w.Write(newline)
	//			_, _ = w.Write(<-c.textMsgCh)
	//		}
	//		if errW := w.Close(); errW != nil {
	//			c.mu.Unlock()
	//			return
	//		}
	//		c.mu.Unlock()
	//	case <-ticker.C:
	//		if err := c.sendPingMessage(ctx); err != nil {
	//			return
	//		}
	//		// 鉴权与心跳一个频次校验
	//		if c.ws.authorizationFunc != nil {
	//			if isOK := c.ws.authorizationFunc(ctx); !isOK {
	//				return
	//			}
	//		}
	//	}
	//}
}

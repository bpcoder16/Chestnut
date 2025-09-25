package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bpcoder16/Chestnut/v2/core/log"
	"github.com/bpcoder16/Chestnut/v2/core/utils"
	"github.com/bpcoder16/Chestnut/v2/logit"
	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	mu sync.RWMutex
	ws *WebSocket

	conn    *websocket.Conn
	uuidStr string
	userId  int64

	textMsgCh chan []byte
	isClosed  bool
	State     State // 客户端状态信息
}

type State struct {
	SID         string         `json:"sid,omitempty"`
	Scene       string         `json:"scene,omitempty"` // 场景信息
	SceneParams map[string]any `json:"-"`
}

func NewClient(conn *websocket.Conn, uuidStr string, userId int64) *Client {
	return &Client{
		conn:    conn,
		uuidStr: uuidStr,
		userId:  userId,

		textMsgCh: make(chan []byte, 2048),
		isClosed:  false,
		State: State{
			SceneParams: make(map[string]any),
		},
	}
}

func (c *Client) UserID() int64 {
	return c.userId
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
		"client.State.SceneParams", c.State.SceneParams,
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
	if !c.isClosed {
		message = append(message, newline...)
		c.textMsgCh <- message
		c.infoLog(
			context.WithValue(ctx, log.DefaultWebSocketUUIDKey, c.uuidStr),
			"client.isClosed", c.isClosed,
			"function", "WriteTextMessage",
			"sendMessageType", c.getMessageTypeString(websocket.TextMessage),
			"sendMessage", string(message),
		)
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

func (c *Client) readPump(ctx context.Context, _ *http.Request, _ http.ResponseWriter) (err error) {
	defer func() {
		if r := recover(); r != nil {
			c.errorLog(ctx,
				"function", "client.readPump",
				"recover", r,
			)
		}
		c.close(ctx, "ReadPump.Defer")
		if err != nil {
			c.warnLog(ctx,
				"function", "client.readPump",
				"err", "c.conn.ReadMessage().Err:"+err.Error(),
			)
		}
	}()

	// 设置读取超时
	if err = c.conn.SetReadDeadline(time.Now().Add(readDeadlineDuration)); err != nil {
		return fmt.Errorf("set read deadline failed: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			err = errors.New("ctx.Done")
			return
		default:
			mt, message, errR := c.conn.ReadMessage()
			rCtx := context.WithValue(ctx, log.DefaultWebSocketLogIdKey, utils.UniqueID())
			if errR != nil {
				err = errR
				return
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
				"err", errR,
				"ReceiveMessageType", c.getMessageTypeString(mt),
				"ReceiveMessage", string(message),
				"costTime", utils.ShowDurationString(elapsed),
			)

			if errR != nil {
				err = errors.New("ReadMessage.ReceiveMessage.Err:" + errR.Error())
				return
			}
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

func (c *Client) writePump(ctx context.Context, req *http.Request, resp http.ResponseWriter) (err error) {
	defer func() {
		if r := recover(); r != nil {
			c.errorLog(ctx,
				"function", "client.writePump",
				"recover", r,
			)
		}
		c.close(ctx, "WritePump.Defer")
		if err != nil {
			c.warnLog(ctx,
				"function", "client.writePump",
				"err", err,
			)
		}
	}()

	// 维持心跳
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ctx.Done():
			err = errors.New("ctx.Done")
			return
		case message, ok := <-c.textMsgCh:
			if !ok {
				err = errors.New("<-c.textMsgCh.NotOK")
				return
			}
			c.mu.Lock()
			if err = c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.mu.Unlock()
				return
			}
			writer, errW := c.conn.NextWriter(websocket.TextMessage)
			if errW != nil {
				err = errors.New("c.conn.NextWriter.Err:" + errW.Error())
				return
			}
			if _, errW = writer.Write(message); err != nil {
				err = errors.New("writer.Write().Err:" + errW.Error())
				c.mu.Unlock()
				return
			}
			if errC := writer.Close(); errC != nil {
				err = errors.New("writer.Close().Err:" + errC.Error())
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		case <-ticker.C:
			if errS := c.sendPingMessage(ctx); errS != nil {
				err = errors.New("c.sendPingMessage.Err:" + errS.Error())
				return
			}
			// 鉴权与心跳一个频次校验
			if c.ws.authorizationFunc != nil {
				if _, isOK, _ := c.ws.authorizationFunc(ctx, req, resp); !isOK {
					err = errors.New("c.ws.authorizationFunc.NotOK")
					return
				}
			}
		}
	}
}

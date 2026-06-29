package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/bpcoder16/Chestnut/v4/core/utils"
	"github.com/bpcoder16/Chestnut/v4/logit"
	"github.com/gorilla/websocket"
)

var (
	newline          = []byte{'\n'}
	errCtxDone       = errors.New("ctx.Done")
	ErrClientClosed  = errors.New("websocket client closed")
	ErrSendQueueFull = errors.New("websocket send queue full")
)

const sendQueueSize = 2048

type Client struct {
	mu      sync.RWMutex
	writeMu sync.Mutex
	ws      *WebSocket

	conn    *websocket.Conn
	uuidStr string
	userID  string

	textMsgCh chan []byte
	done      chan struct{}
	isClosed  bool

	stateMu sync.RWMutex
	state   State
}

type State struct {
	SID         string         `json:"sid,omitempty"`
	Scene       string         `json:"scene,omitempty"` // 场景信息
	SceneParams map[string]any `json:"-"`
}

// GetState 线程安全地返回当前 State 的深拷贝快照，调用方可以安全修改返回值。
func (c *Client) GetState() State {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return cloneStateDeep(c.state)
}

// GetStateShallow 线程安全地返回当前 State 的浅拷贝快照，仅用于只读场景。
func (c *Client) GetStateShallow() State {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return cloneStateShallow(c.state)
}

// UpdateState 线程安全地更新 State，仅更新非零值字段。
func (c *Client) UpdateState(scene, sid string, sceneParams map[string]any) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	if len(scene) > 0 {
		c.state.Scene = scene
	}
	if len(sid) > 0 {
		c.state.SID = sid
	}
	if len(sceneParams) > 0 {
		c.state.SceneParams = cloneMapDeep(sceneParams)
	}
}

func NewClient(conn *websocket.Conn, uuidStr string, userID string) *Client {
	return &Client{
		conn:    conn,
		uuidStr: uuidStr,
		userID:  userID,

		textMsgCh: make(chan []byte, sendQueueSize),
		done:      make(chan struct{}),
		isClosed:  false,
		state: State{
			SceneParams: make(map[string]any),
		},
	}
}

func (c *Client) UserID() string {
	return c.userID
}

func (c *Client) UUID() string {
	return c.uuidStr
}

func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isClosed
}

// Close 公开关闭方法，供外部（如踢出逻辑）主动关闭连接。
func (c *Client) Close(ctx context.Context) {
	c.close(ctx, "ExternalForceClose")
}

// close 执行连接清理。若本次调用真正触发了清理返回 true，连接已被其他路径关闭则返回 false。
func (c *Client) close(ctx context.Context, sourceText string) bool {
	c.mu.Lock()
	if c.isClosed {
		c.mu.Unlock()
		return false
	}
	c.isClosed = true
	close(c.done)
	c.mu.Unlock()

	if c.conn != nil {
		_ = c.sendCloseMessage(ctx)
		_ = c.conn.Close()
	}
	if c.ws != nil {
		c.ws.clientManager.DeleteClient(c)
		if f := c.ws.getClientCloseFunc(); f != nil {
			f(ctx, c.uuidStr)
		}
	}
	c.infoLog(ctx,
		"sourceText", sourceText,
		"function", "Client.close",
		"Disconnection.Status", "Success",
		"client.ws.clientManager", "Delete("+c.uuidStr+")",
	)
	return true
}

func (c *Client) log(ctx context.Context, level string, keyValues ...interface{}) {
	state := c.GetStateShallow()
	newKeyValues := []interface{}{
		"subProtocol", c.subProtocol(),
		"localAddr", c.localAddr(),
		"remoteAddr", c.remoteAddr(),
		"client.ws.clientManager.Len()", c.clientManagerLen(),
		"client.State", state,
		"client.State.SceneParams", state.SceneParams,
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
	if c.IsClosed() {
		return ErrClientClosed
	}

	msg := append([]byte(nil), message...)
	msg = append(msg, newline...)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return ErrClientClosed
	case c.textMsgCh <- msg:
		c.debugLog(
			context.WithValue(ctx, log.DefaultWebSocketUUIDKey, c.uuidStr),
			"function", "WriteTextMessage",
			"sendMessageType", c.getMessageTypeString(websocket.TextMessage),
			"sendMessageSize", len(msg),
		)
		return nil
	default:
		return ErrSendQueueFull
	}
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
		state := c.GetStateShallow()
		if len(state.Scene) == 0 {
			err = errors.New("receiveMessage.Scene.Empty")
			return
		}
		// 如果没有传递，说明用户停留在当前场景
		receiveMessage.Scene = state.Scene
		receiveMessage.SceneParams = state.SceneParams
		receiveMessage.SID = state.SID
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
		actualClose := c.close(ctx, "ReadPump.Defer")
		if err != nil && actualClose && !errors.Is(err, errCtxDone) {
			keyValues := []interface{}{
				"function", "client.readPump",
				"err", "c.conn.ReadMessage().Err:" + err.Error(),
			}
			if isAbnormalUnexpectedEOFCloseError(err) {
				c.debugLog(ctx, keyValues...)
			} else {
				c.warnLog(ctx, keyValues...)
			}
		}
	}()

	// 设置读取超时
	if err = c.conn.SetReadDeadline(time.Now().Add(readDeadlineDuration)); err != nil {
		return fmt.Errorf("set read deadline failed: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			err = errCtxDone
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
				"ReceiveMessageSize", len(message),
				"costTime", utils.ShowDurationString(elapsed),
			)

			if errR != nil {
				err = errors.New("ReadMessage.ReceiveMessage.Err:" + errR.Error())
				return
			}
		}
	}
}

func isAbnormalUnexpectedEOFCloseError(err error) bool {
	var closeErr *websocket.CloseError
	if !errors.As(err, &closeErr) {
		return false
	}
	// 1006 + unexpected EOF 通常是对端直接断开，不作为服务端告警噪音记录。
	return closeErr.Code == websocket.CloseAbnormalClosure && closeErr.Text == io.ErrUnexpectedEOF.Error()
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
		actualClose := c.close(ctx, "WritePump.Defer")
		if err != nil && actualClose && !errors.Is(err, errCtxDone) {
			c.warnLog(ctx,
				"function", "client.writePump",
				"err", err,
			)
		}
	}()

	// 维持心跳
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			err = errCtxDone
			return
		case <-c.done:
			err = errCtxDone
			return
		case message, ok := <-c.textMsgCh:
			if !ok {
				err = errors.New("<-c.textMsgCh.NotOK")
				return
			}
			if err = c.writeTextFrame(message); err != nil {
				return
			}
		case <-ticker.C:
			if errS := c.sendPingMessage(ctx); errS != nil {
				err = errors.New("c.sendPingMessage.Err:" + errS.Error())
				return
			}
			// 鉴权与心跳一个频次校验
			if f := c.ws.getAuthorizationFunc(); f != nil {
				if _, isOK, _ := f(ctx, req, resp); !isOK {
					err = errors.New("c.ws.authorizationFunc.NotOK")
					return
				}
			}
		}
	}
}

func (c *Client) writeTextFrame(message []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.conn == nil {
		return ErrClientClosed
	}
	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	writer, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return errors.New("c.conn.NextWriter.Err:" + err.Error())
	}
	if _, err = writer.Write(message); err != nil {
		_ = writer.Close()
		return errors.New("writer.Write().Err:" + err.Error())
	}
	if err = writer.Close(); err != nil {
		return errors.New("writer.Close().Err:" + err.Error())
	}
	return nil
}

func (c *Client) subProtocol() string {
	if c.conn == nil {
		return ""
	}
	return c.conn.Subprotocol()
}

func (c *Client) localAddr() string {
	if c.conn == nil || c.conn.LocalAddr() == nil {
		return ""
	}
	return c.conn.LocalAddr().String()
}

func (c *Client) remoteAddr() string {
	if c.conn == nil || c.conn.RemoteAddr() == nil {
		return ""
	}
	return c.conn.RemoteAddr().String()
}

func (c *Client) clientManagerLen() int {
	if c.ws == nil || c.ws.clientManager == nil {
		return 0
	}
	return c.ws.clientManager.Len()
}

func cloneStateDeep(state State) State {
	state.SceneParams = cloneMapDeep(state.SceneParams)
	return state
}

func cloneStateShallow(state State) State {
	state.SceneParams = cloneMapShallow(state.SceneParams)
	return state
}

func cloneMapDeep(src map[string]any) map[string]any {
	if len(src) == 0 {
		return make(map[string]any)
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = deepClone(v)
	}
	return dst
}

func cloneMapShallow(src map[string]any) map[string]any {
	if len(src) == 0 {
		return make(map[string]any)
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func deepClone(src any) any {
	if src == nil {
		return nil
	}
	return deepCloneValue(reflect.ValueOf(src)).Interface()
}

func deepCloneValue(src reflect.Value) reflect.Value {
	if !src.IsValid() {
		return src
	}

	switch src.Kind() {
	case reflect.Interface:
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		cloned := deepCloneValue(src.Elem())
		dst := reflect.New(src.Type()).Elem()
		dst.Set(cloned)
		return dst
	case reflect.Pointer:
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		dst := reflect.New(src.Type().Elem())
		dst.Elem().Set(deepCloneValue(src.Elem()))
		return dst
	case reflect.Map:
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		dst := reflect.MakeMapWithSize(src.Type(), src.Len())
		iter := src.MapRange()
		for iter.Next() {
			dst.SetMapIndex(deepCloneValue(iter.Key()), deepCloneValue(iter.Value()))
		}
		return dst
	case reflect.Slice:
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		dst := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
		for i := 0; i < src.Len(); i++ {
			dst.Index(i).Set(deepCloneValue(src.Index(i)))
		}
		return dst
	case reflect.Array:
		dst := reflect.New(src.Type()).Elem()
		for i := 0; i < src.Len(); i++ {
			dst.Index(i).Set(deepCloneValue(src.Index(i)))
		}
		return dst
	default:
		return src
	}
}

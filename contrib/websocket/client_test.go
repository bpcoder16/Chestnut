package websocket

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type clientReceiveTextController struct {
	BaseTextMessageController
}

func (c *clientReceiveTextController) Process(context.Context) error {
	return nil
}

type clientReceiveParseErrorController struct{}

func (c *clientReceiveParseErrorController) Init(TextMessageController) {}

func (c *clientReceiveParseErrorController) ParsePayload(context.Context, *Client, ReceiveMessage) error {
	return errors.New("parse failed")
}

func (c *clientReceiveParseErrorController) Process(context.Context) error {
	return nil
}

type clientReceiveProcessErrorController struct{}

func (c *clientReceiveProcessErrorController) Init(TextMessageController) {}

func (c *clientReceiveProcessErrorController) ParsePayload(context.Context, *Client, ReceiveMessage) error {
	return nil
}

func (c *clientReceiveProcessErrorController) Process(context.Context) error {
	return errors.New("process failed")
}

func TestNewClientInitialState(t *testing.T) {
	client := NewClient(nil, "uuid-1", "user-1")

	if got := client.UUID(); got != "uuid-1" {
		t.Fatalf("UUID() = %q, want uuid-1", got)
	}
	if got := client.UserID(); got != "user-1" {
		t.Fatalf("UserID() = %q, want user-1", got)
	}
	if client.IsClosed() {
		t.Fatal("IsClosed() = true, want false")
	}
	if client.textMsgCh == nil {
		t.Fatal("textMsgCh = nil, want initialized channel")
	}
	if cap(client.textMsgCh) != sendQueueSize {
		t.Fatalf("textMsgCh cap = %d, want %d", cap(client.textMsgCh), sendQueueSize)
	}
	if client.done == nil {
		t.Fatal("done = nil, want initialized channel")
	}
	if state := client.GetState(); state.SceneParams == nil || len(state.SceneParams) != 0 {
		t.Fatalf("initial SceneParams = %v, want empty non-nil map", state.SceneParams)
	}
}

func TestClientStateSnapshotIsIsolated(t *testing.T) {
	client := NewClient(nil, "uuid-1", "user-1")
	client.UpdateState("scene", "sid", map[string]any{
		"k": "v",
		"nested": map[string]any{
			"status": "open",
		},
		"ids": []any{"1", "2"},
	})

	state := client.GetState()
	state.SceneParams["k"] = "changed"
	state.SceneParams["nested"].(map[string]any)["status"] = "closed"
	state.SceneParams["ids"].([]any)[0] = "changed"

	next := client.GetState()
	if next.SceneParams["k"] != "v" {
		t.Fatalf("state snapshot mutated internal map: got %v, want v", next.SceneParams["k"])
	}
	if next.SceneParams["nested"].(map[string]any)["status"] != "open" {
		t.Fatalf("nested state mutated internal map: got %v, want open", next.SceneParams["nested"])
	}
	if next.SceneParams["ids"].([]any)[0] != "1" {
		t.Fatalf("nested state mutated internal slice: got %v, want 1", next.SceneParams["ids"])
	}
}

func TestClientStateShallowSnapshotOnlyClonesTopLevelMap(t *testing.T) {
	client := NewClient(nil, "uuid-1", "user-1")
	client.UpdateState("scene", "sid", map[string]any{
		"nested": map[string]any{
			"status": "open",
		},
	})

	state := client.GetStateShallow()
	state.SceneParams["top"] = "changed"
	state.SceneParams["nested"].(map[string]any)["status"] = "closed"

	next := client.GetStateShallow()
	if _, ok := next.SceneParams["top"]; ok {
		t.Fatal("shallow state top-level map mutation affected internal map")
	}
	if next.SceneParams["nested"].(map[string]any)["status"] != "closed" {
		t.Fatalf("shallow nested mutation did not share nested value: got %v, want closed", next.SceneParams["nested"])
	}
}

func TestClientUpdateStateOnlyUpdatesNonZeroValues(t *testing.T) {
	client := NewClient(nil, "uuid-1", "user-1")
	client.UpdateState("scene", "sid", map[string]any{"k": "v"})

	client.UpdateState("", "", nil)
	state := client.GetState()
	if state.Scene != "scene" {
		t.Fatalf("Scene = %q, want scene", state.Scene)
	}
	if state.SID != "sid" {
		t.Fatalf("SID = %q, want sid", state.SID)
	}
	if state.SceneParams["k"] != "v" {
		t.Fatalf("SceneParams[k] = %v, want v", state.SceneParams["k"])
	}
}

func TestClientCloseIsIdempotentAndRemovesManagerIndex(t *testing.T) {
	m := NewClientManager()
	ws := &WebSocket{clientManager: m}
	client := NewClient(nil, "uuid-1", "user-1")
	client.ws = ws
	m.StoreClient(client)

	if !client.close(context.Background(), "test") {
		t.Fatal("first close returned false, want true")
	}
	if client.close(context.Background(), "test") {
		t.Fatal("second close returned true, want false")
	}
	if got := m.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestClientCloseCallsCloseFunc(t *testing.T) {
	m := NewClientManager()
	ws := &WebSocket{clientManager: m}
	client := NewClient(nil, "uuid-1", "user-1")
	client.ws = ws
	m.StoreClient(client)

	calls := 0
	ws.SetClientCloseFunc(func(ctx context.Context, uuidStr string) {
		calls++
		if uuidStr != "uuid-1" {
			t.Fatalf("close callback uuid = %q, want uuid-1", uuidStr)
		}
	})

	client.Close(context.Background())
	client.Close(context.Background())

	if calls != 1 {
		t.Fatalf("close callback calls = %d, want 1", calls)
	}
	if got := m.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestWriteTextMessageCopiesMessageAndAppendsNewline(t *testing.T) {
	client := NewClient(nil, "uuid-1", "user-1")
	msg := []byte("msg")

	if err := client.WriteTextMessage(context.Background(), msg); err != nil {
		t.Fatalf("WriteTextMessage returned %v, want nil", err)
	}
	msg[0] = 'x'

	got := <-client.textMsgCh
	if string(got) != "msg\n" {
		t.Fatalf("queued message = %q, want %q", got, "msg\n")
	}
}

func TestWriteTextMessageContextDone(t *testing.T) {
	client := NewClient(nil, "uuid-1", "user-1")
	for i := 0; i < sendQueueSize; i++ {
		if err := client.WriteTextMessage(context.Background(), []byte("msg")); err != nil {
			t.Fatalf("WriteTextMessage fill #%d returned %v, want nil", i, err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.WriteTextMessage(ctx, []byte("msg")); !errors.Is(err, context.Canceled) {
		t.Fatalf("WriteTextMessage canceled err = %v, want context.Canceled", err)
	}
}

func TestWriteTextMessageClosedAndFullQueue(t *testing.T) {
	client := NewClient(nil, "uuid-1", "user-1")
	for i := 0; i < sendQueueSize; i++ {
		if err := client.WriteTextMessage(context.Background(), []byte("msg")); err != nil {
			t.Fatalf("WriteTextMessage fill #%d returned %v, want nil", i, err)
		}
	}
	if err := client.WriteTextMessage(context.Background(), []byte("msg")); !errors.Is(err, ErrSendQueueFull) {
		t.Fatalf("WriteTextMessage full queue err = %v, want ErrSendQueueFull", err)
	}

	client.Close(context.Background())
	if err := client.WriteTextMessage(context.Background(), []byte("msg")); !errors.Is(err, ErrClientClosed) {
		t.Fatalf("WriteTextMessage closed err = %v, want ErrClientClosed", err)
	}
}

func TestClientReceiveTextMessageValidationErrors(t *testing.T) {
	client := NewClient(nil, "uuid-1", "user-1")
	client.ws = &WebSocket{textMessageControllers: make(map[string]TextMessageController)}

	tests := []struct {
		name    string
		message []byte
		want    string
	}{
		{name: "empty", message: nil, want: "messageBytes.Len(0)"},
		{name: "invalid json", message: []byte("{"), want: "parse text message failed"},
		{name: "empty scene", message: []byte(`{"action":"ping"}`), want: "receiveMessage.Scene.Empty"},
		{name: "controller missing", message: []byte(`{"scene":"missing"}`), want: "TextMessageController.Not.Register"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.receiveTextMessage(context.Background(), tt.message)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("receiveTextMessage err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestClientReceiveTextMessageUsesCurrentSceneWhenMessageOmitsScene(t *testing.T) {
	ws := &WebSocket{textMessageControllers: make(map[string]TextMessageController)}
	if err := ws.OnTextMessageController("scene", &clientReceiveTextController{}); err != nil {
		t.Fatalf("OnTextMessageController err = %v, want nil", err)
	}

	client := NewClient(nil, "uuid-1", "user-1")
	client.ws = ws
	client.UpdateState("scene", "sid", map[string]any{"k": "v"})

	if err := client.receiveTextMessage(context.Background(), []byte(`{"action":"ping"}`)); err != nil {
		t.Fatalf("receiveTextMessage err = %v, want nil", err)
	}

	state := client.GetState()
	if state.Scene != "scene" {
		t.Fatalf("Scene = %q, want scene", state.Scene)
	}
	if state.SID != "sid" {
		t.Fatalf("SID = %q, want sid", state.SID)
	}
	if state.SceneParams["k"] != "v" {
		t.Fatalf("SceneParams[k] = %v, want v", state.SceneParams["k"])
	}
}

func TestClientReceiveTextMessageReturnsControllerErrors(t *testing.T) {
	tests := []struct {
		name       string
		controller TextMessageController
		want       string
	}{
		{name: "parse", controller: &clientReceiveParseErrorController{}, want: "parse failed"},
		{name: "process", controller: &clientReceiveProcessErrorController{}, want: "process failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &WebSocket{textMessageControllers: make(map[string]TextMessageController)}
			if err := ws.OnTextMessageController("scene", tt.controller); err != nil {
				t.Fatalf("OnTextMessageController err = %v, want nil", err)
			}
			client := NewClient(nil, "uuid-1", "user-1")
			client.ws = ws

			err := client.receiveTextMessage(context.Background(), []byte(`{"scene":"scene"}`))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("receiveTextMessage err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

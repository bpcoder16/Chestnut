package websocket

import (
	"context"
	"testing"
)

type testTextController struct {
	BaseTextMessageController
}

func (c *testTextController) Process(context.Context) error {
	return nil
}

type valueTextController struct{}

func (valueTextController) Init(TextMessageController) {}

func (valueTextController) ParsePayload(context.Context, *Client, ReceiveMessage) error {
	return nil
}

func (valueTextController) Process(context.Context) error {
	return nil
}

func TestWebSocketControllerRegistrationValidation(t *testing.T) {
	ws := &WebSocket{
		textMessageControllers: make(map[string]TextMessageController),
	}

	if err := ws.OnTextMessageController("", &testTextController{}); err == nil {
		t.Fatal("OnTextMessageController empty scene err = nil, want error")
	}
	if err := ws.OnTextMessageController("scene", nil); err == nil {
		t.Fatal("OnTextMessageController nil controller err = nil, want error")
	}
	if err := ws.OnTextMessageController("scene", valueTextController{}); err == nil {
		t.Fatal("OnTextMessageController non-pointer controller err = nil, want error")
	}
	if err := ws.OnTextMessageController("scene", &testTextController{}); err != nil {
		t.Fatalf("OnTextMessageController valid err = %v, want nil", err)
	}

	controller, err := ws.getTextMessageController("scene")
	if err != nil {
		t.Fatalf("getTextMessageController err = %v, want nil", err)
	}
	if _, ok := controller.(*testTextController); !ok {
		t.Fatalf("controller type = %T, want *testTextController", controller)
	}
}

package websocket

import "context"

type TextMessageController interface {
	Init(base TextMessageController)
	ParsePayload(ctx context.Context, c *Client, message ReceiveMessage) error
	Process(ctx context.Context) error
}

//	{
//		"scene": "test",
//		"sceneParams": {
//			"key1": 1234,
//			"key2": "value"
//		},
//		"action": "test",
//		"actionParams": {
//			"key1": 1234,
//			"key2": "value"
//		}
//	}
type ReceiveMessage struct {
	Scene        string                 `json:"scene"`
	SID          string                 `json:"sid"`
	SceneParams  map[string]interface{} `json:"sceneParams"`
	Action       string                 `json:"action"`
	ActionParams map[string]interface{} `json:"actionParams"`
}

type BaseTextMessageController struct {
	Client       *Client
	Action       string
	ActionParams map[string]interface{}
}

func (b *BaseTextMessageController) Init(_ TextMessageController) {}

func (b *BaseTextMessageController) ParsePayload(_ context.Context, client *Client, message ReceiveMessage) (err error) {
	b.Client = client
	if len(message.Scene) > 0 {
		b.Client.State.Scene = message.Scene
	}
	if len(message.SceneParams) > 0 {
		b.Client.State.SceneParams = message.SceneParams
	}
	if len(message.SID) > 0 {
		b.Client.State.SID = message.SID
	}
	b.Action = message.Action
	b.ActionParams = message.ActionParams
	return
}

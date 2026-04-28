package websocket

import (
	"context"
	"testing"
)

func TestNewRedisPubSubValidation(t *testing.T) {
	if _, err := NewRedisPubSub(); err == nil {
		t.Fatal("NewRedisPubSub() err = nil, want error")
	}
	if _, err := NewRedisPubSub(""); err == nil {
		t.Fatal("NewRedisPubSub(empty channel) err = nil, want error")
	}
	pubSub, err := NewRedisPubSub("channel")
	if err != nil {
		t.Fatalf("NewRedisPubSub(valid) err = %v, want nil", err)
	}
	if channel, err := pubSub.getRandomChannel(); err != nil || channel != "channel" {
		t.Fatalf("getRandomChannel() = %q, %v; want channel, nil", channel, err)
	}
}

func TestRedisPubSubPublishValidation(t *testing.T) {
	pubSub, err := NewRedisPubSub("channel")
	if err != nil {
		t.Fatalf("NewRedisPubSub(valid) err = %v, want nil", err)
	}
	if _, err = pubSub.Publish(context.Background(), nil, "msg"); err == nil {
		t.Fatal("Publish nil redis client err = nil, want error")
	}

	var nilPubSub *RedisPubSub
	if _, err = nilPubSub.Publish(context.Background(), nil, "msg"); err == nil {
		t.Fatal("nil RedisPubSub Publish err = nil, want error")
	}
}

package gonats

import (
	"context"
	"testing"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/nats-io/nats.go"
)

func TestBuildPullConsumerMessageContextPreservesParentContextAndInjectsLogID(t *testing.T) {
	type ctxKey string
	const parentKey ctxKey = "parent"

	parent := context.WithValue(context.Background(), parentKey, "value")
	header := nats.Header{}
	header.Set(LogIdHeader, "log-123")

	msgCtx := buildPullConsumerMessageContext(parent, header)
	if got := msgCtx.Value(parentKey); got != "value" {
		t.Fatalf("parent context value = %v, want value", got)
	}
	if got := log.LogIdFromContext(msgCtx); got != "log-123" {
		t.Fatalf("logId = %q, want log-123", got)
	}
	if got := msgCtx.Value(log.DefaultMessageKey); got != "NATS" {
		t.Fatalf("message type = %v, want NATS", got)
	}
	if got := msgCtx.Value(log.DefaultDownstreamKey); got != "NATS" {
		t.Fatalf("downstream = %v, want NATS", got)
	}
}

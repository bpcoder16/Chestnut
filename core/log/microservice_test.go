package log

import (
	"context"
	"testing"
)

func TestMicroServiceLogIDContextHelpers(t *testing.T) {
	if MicroServiceLogIdHeader != "MicroService-Log-Id" {
		t.Fatalf("MicroServiceLogIdHeader = %q, want %q", MicroServiceLogIdHeader, "MicroService-Log-Id")
	}

	ctx := WithLogId(context.Background(), "log-123")
	logID := LogIdFromContext(ctx)
	if logID != "log-123" {
		t.Fatalf("LogIdFromContext = %q, want %q", logID, "log-123")
	}
}

func TestLogIdFromContextReturnsEmptyWhenMissing(t *testing.T) {
	if logID := LogIdFromContext(context.Background()); logID != "" {
		t.Fatalf("LogIdFromContext = %q, want empty", logID)
	}
}

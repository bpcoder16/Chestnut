package resty

import (
	"context"
	"testing"

	"github.com/bpcoder16/Chestnut/v4/core/log"
)

func TestMicroServiceRequestInjectsLogIDHeader(t *testing.T) {
	SetClient(log.NewHelper(log.DefaultLogger))

	req, err := MicroServiceRequest(log.WithLogId(context.Background(), "log-123"))
	if err != nil {
		t.Fatalf("MicroServiceRequest err = %v, want nil", err)
	}
	if got := req.Header.Get(log.MicroServiceLogIdHeader); got != "log-123" {
		t.Fatalf("header %s = %q, want %q", log.MicroServiceLogIdHeader, got, "log-123")
	}
	if req.Context() == nil {
		t.Fatal("request context is nil")
	}
}

func TestMicroServiceRequestRejectsMissingLogID(t *testing.T) {
	SetClient(log.NewHelper(log.DefaultLogger))

	_, err := MicroServiceRequest(context.Background())
	if err == nil {
		t.Fatal("MicroServiceRequest err = nil, want error")
	}
}

func TestMicroServiceRequestRejectsUninitializedClient(t *testing.T) {
	client = nil

	_, err := MicroServiceRequest(log.WithLogId(context.Background(), "log-123"))
	if err == nil {
		t.Fatal("MicroServiceRequest err = nil, want error")
	}
}

package gin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bpcoder16/Chestnut/v4/core/log"
	"github.com/gin-gonic/gin"
)

func TestRequireMicroServiceLogIDRejectsMissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequireMicroServiceLogID())
	router.GET("/api/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/ping", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestDefaultLoggerDoesNotRewriteRequestContextWhenGinContextHasLogID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type contextKey string
	const markerKey contextKey = "marker"
	markerCtx := context.WithValue(context.Background(), markerKey, "stable")

	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set(log.DefaultLogIdKey, "log-123")
		ctx.Request = ctx.Request.WithContext(markerCtx)
		ctx.Next()
	}, DefaultLogger())
	router.GET("/api/ping", func(ctx *gin.Context) {
		if ctx.Request.Context() != markerCtx {
			t.Fatal("request context was rewritten even though gin context already had logId")
		}
		ctx.String(http.StatusOK, "pong")
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/ping", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestRequireMicroServiceLogIDInjectsGinAndRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequireMicroServiceLogID())
	router.GET("/api/ping", func(ctx *gin.Context) {
		if got := ctx.GetString(log.DefaultLogIdKey); got != "log-123" {
			t.Fatalf("gin context logId = %q, want %q", got, "log-123")
		}
		if got := log.LogIdFromContext(ctx.Request.Context()); got != "log-123" {
			t.Fatalf("request context logId = %q, want %q", got, "log-123")
		}
		ctx.String(http.StatusOK, "pong")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.Header.Set(log.MicroServiceLogIdHeader, "log-123")
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestDefaultLoggerKeepsExistingMicroServiceLogID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequireMicroServiceLogID(), DefaultLogger())
	router.GET("/api/ping", func(ctx *gin.Context) {
		if got := ctx.GetString(log.DefaultLogIdKey); got != "log-123" {
			t.Fatalf("gin context logId = %q, want %q", got, "log-123")
		}
		if got := log.LogIdFromContext(ctx.Request.Context()); got != "log-123" {
			t.Fatalf("request context logId = %q, want %q", got, "log-123")
		}
		ctx.String(http.StatusOK, "pong")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.Header.Set(log.MicroServiceLogIdHeader, "log-123")
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

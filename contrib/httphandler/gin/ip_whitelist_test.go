package gin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewIPWhitelistMiddlewareAllowsRequestsWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware, err := NewIPWhitelistMiddleware(IPWhitelistConfig{
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("NewIPWhitelistMiddleware err = %v, want nil", err)
	}

	router := gin.New()
	router.Use(middleware)
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "198.51.100.10:12345"
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestNewIPWhitelistMiddlewareAllowsConfiguredRemoteIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware, err := NewIPWhitelistMiddleware(IPWhitelistConfig{
		Enabled:  true,
		AllowIPs: []string{"192.0.2.10"},
	})
	if err != nil {
		t.Fatalf("NewIPWhitelistMiddleware err = %v, want nil", err)
	}

	router := gin.New()
	router.Use(middleware)
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestNewIPWhitelistMiddlewareRejectsUnlistedRemoteIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware, err := NewIPWhitelistMiddleware(IPWhitelistConfig{
		Enabled:  true,
		AllowIPs: []string{"192.0.2.10"},
	})
	if err != nil {
		t.Fatalf("NewIPWhitelistMiddleware err = %v, want nil", err)
	}

	router := gin.New()
	router.Use(middleware)
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "198.51.100.10:12345"
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if !strings.Contains(recorder.Body.String(), "权限不足") {
		t.Fatalf("body = %q, want forbidden message", recorder.Body.String())
	}
}

func TestNewIPWhitelistMiddlewareRejectsInvalidAllowIP(t *testing.T) {
	_, err := NewIPWhitelistMiddleware(IPWhitelistConfig{
		Enabled:  true,
		AllowIPs: []string{"not-an-ip"},
	})
	if err == nil {
		t.Fatal("NewIPWhitelistMiddleware err = nil, want error")
	}
}

func TestHTTPHandlerWithMiddlewaresAppliesMiddlewareToHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware, err := NewIPWhitelistMiddleware(IPWhitelistConfig{
		Enabled:  true,
		AllowIPs: []string{"192.0.2.10"},
	})
	if err != nil {
		t.Fatalf("NewIPWhitelistMiddleware err = %v, want nil", err)
	}

	router := HTTPHandlerWithMiddlewares([]gin.HandlerFunc{middleware})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "198.51.100.10:12345"
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("GET /health status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

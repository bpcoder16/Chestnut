package gin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestScalarRegisterUsesEmbeddedAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	ScalarRegister(&r.RouterGroup)

	docsRecorder := httptest.NewRecorder()
	r.ServeHTTP(docsRecorder, httptest.NewRequest(http.MethodGet, "/docs", nil))

	if docsRecorder.Code != http.StatusOK {
		t.Fatalf("GET /docs status = %d, want %d", docsRecorder.Code, http.StatusOK)
	}
	docsBody := docsRecorder.Body.String()
	if strings.Contains(docsBody, "cdn.jsdelivr.net") {
		t.Fatal("GET /docs should not reference CDN assets")
	}
	if !strings.Contains(docsBody, "/docs/assets/api-reference.js") {
		t.Fatal("GET /docs should reference embedded Scalar asset")
	}

	assetRecorder := httptest.NewRecorder()
	r.ServeHTTP(assetRecorder, httptest.NewRequest(http.MethodGet, "/docs/assets/api-reference.js", nil))

	if assetRecorder.Code != http.StatusOK {
		t.Fatalf("GET /docs/assets/api-reference.js status = %d, want %d", assetRecorder.Code, http.StatusOK)
	}
	if !strings.Contains(assetRecorder.Body.String(), "Scalar") {
		t.Fatal("embedded Scalar asset should contain Scalar browser code")
	}
}

func TestScalarSourcesRegisterUsesVirtualSpecSources(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	ScalarSourcesRegister(&r.RouterGroup, []ScalarSource{
		{
			Title:      "管理端 API",
			Slug:       "admin",
			PathPrefix: "/api/v1/admin/",
		},
	})

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/docs", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /docs status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"sources"`) {
		t.Fatal("GET /docs should configure Scalar sources")
	}
	if !strings.Contains(body, `"/docs/specs/admin.json"`) {
		t.Fatal("GET /docs should reference admin virtual spec")
	}
	if strings.Contains(body, `"/swagger/doc.json"`) {
		t.Fatal("GET /docs should not use full swagger doc in sources mode")
	}
}

func TestSplitSwaggerDocByPathPrefix(t *testing.T) {
	raw := []byte(`{
		"swagger": "2.0",
		"info": {
			"title": "full api",
			"version": "1.0"
		},
		"paths": {
			"/api/v1/admin/auth/login": {
				"post": {
					"tags": ["管理端-登录"]
				}
			},
			"/api/v1/user/auth/login": {
				"post": {
					"tags": ["用户端-登录"]
				}
			}
		},
		"tags": [
			{"name": "管理端-登录"},
			{"name": "用户端-登录"}
		],
		"definitions": {
			"Response": {
				"type": "object"
			}
		}
	}`)

	doc, err := splitSwaggerDocByPathPrefix(raw, ScalarSource{
		Title:      "管理端 API",
		PathPrefix: "/api/v1/admin/",
	})
	if err != nil {
		t.Fatalf("splitSwaggerDocByPathPrefix error = %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(doc, &spec); err != nil {
		t.Fatalf("json.Unmarshal split doc error = %v", err)
	}

	info := spec["info"].(map[string]any)
	if info["title"] != "管理端 API" {
		t.Fatalf("info.title = %v, want 管理端 API", info["title"])
	}

	paths := spec["paths"].(map[string]any)
	if _, ok := paths["/api/v1/admin/auth/login"]; !ok {
		t.Fatal("admin path should be kept")
	}
	if _, ok := paths["/api/v1/user/auth/login"]; ok {
		t.Fatal("user path should be filtered out")
	}

	tags := spec["tags"].([]any)
	if len(tags) != 1 {
		t.Fatalf("len(tags) = %d, want 1", len(tags))
	}
	tag := tags[0].(map[string]any)
	if tag["name"] != "管理端-登录" {
		t.Fatalf("tag name = %v, want 管理端-登录", tag["name"])
	}

	if _, ok := spec["definitions"].(map[string]any)["Response"]; !ok {
		t.Fatal("definitions should be kept")
	}
}

package gin

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

const (
	scalarAssetPath = "/docs/assets/api-reference.js"
	scalarSpecPath  = "/docs/specs/"
)

// ScalarSource 定义 Scalar 多文档入口及虚拟 Swagger 文档的路径过滤规则。
type ScalarSource struct {
	Title      string
	Slug       string
	URL        string
	PathPrefix string
}

//go:embed scalar/assets/*
var scalarAssets embed.FS

var scalarAssetFS, _ = fs.Sub(scalarAssets, "scalar/assets")

const scalarHTMLPrefix = `<!doctype html>
<html>
  <head>
    <title>Scalar API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <div id="app"></div>
    <script src="` + scalarAssetPath + `"></script>
    <script>
      Scalar.createApiReference('#app', `

const scalarHTMLSuffix = `)
    </script>
  </body>
</html>`

type scalarConfig struct {
	URL               string               `json:"url,omitempty"`
	Sources           []scalarConfigSource `json:"sources,omitempty"`
	Layout            string               `json:"layout"`
	Theme             string               `json:"theme"`
	PersistAuth       bool                 `json:"persistAuth"`
	Telemetry         bool                 `json:"telemetry"`
	Agent             scalarAgentConfig    `json:"agent"`
	DefaultHTTPClient scalarHTTPClient     `json:"defaultHttpClient"`
}

type scalarConfigSource struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
	URL   string `json:"url"`
}

type scalarAgentConfig struct {
	Disabled bool `json:"disabled"`
}

type scalarHTTPClient struct {
	TargetKey string `json:"targetKey"`
	ClientKey string `json:"clientKey"`
}

// SwaggerRegister 将 Swagger UI 路由注册到指定 RouterGroup。
// 访问路径为 /swagger/index.html。
// 调用方需在自身包中通过 blank import 注册生成的文档，例如：_ "yourmodule/swagger"
func SwaggerRegister(r *gin.RouterGroup) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// ScalarRegister 将 Scalar API Reference 路由注册到指定 RouterGroup。
// 访问路径为 /docs，文档数据复用 /swagger/doc.json。
// 调用方需同时注册 SwaggerRegister 或 SwaggerWithScalarRegister。
func ScalarRegister(r *gin.RouterGroup) {
	r.GET(scalarAssetPath, scalarAssetHandler)
	r.GET("/docs", scalarHandler(defaultScalarConfigJSON()))
}

// ScalarSourcesRegister 将 Scalar 多文档入口和按路径前缀拆分的虚拟 Swagger 文档注册到指定 RouterGroup。
// 访问路径为 /docs，虚拟文档访问路径为 /docs/specs/{slug}.json。
func ScalarSourcesRegister(r *gin.RouterGroup, sources []ScalarSource) {
	r.GET(scalarAssetPath, scalarAssetHandler)
	for _, source := range normalizeScalarSources(sources) {
		r.GET(source.URL, scalarSpecHandler(source))
	}
	r.GET("/docs", scalarHandler(scalarSourcesConfigJSON(sources)))
}

// SwaggerWithScalarRegister 同时注册 Swagger UI 与 Scalar API Reference。
func SwaggerWithScalarRegister(r *gin.RouterGroup) {
	SwaggerRegister(r)
	ScalarRegister(r)
}

// SwaggerWithScalarSourcesRegister 同时注册 Swagger UI、Scalar 多文档入口和虚拟 Swagger 文档。
func SwaggerWithScalarSourcesRegister(sources []ScalarSource) func(*gin.RouterGroup) {
	return func(r *gin.RouterGroup) {
		SwaggerRegister(r)
		ScalarSourcesRegister(r, sources)
	}
}

func scalarHandler(configJSON string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(scalarHTMLPrefix+configJSON+scalarHTMLSuffix))
	}
}

func scalarAssetHandler(ctx *gin.Context) {
	ctx.FileFromFS("api-reference.js", http.FS(scalarAssetFS))
}

func scalarSpecHandler(source ScalarSource) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		doc, err := swag.ReadDoc()
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		data, err := splitSwaggerDocByPathPrefix([]byte(doc), source)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.Data(http.StatusOK, "application/json; charset=utf-8", data)
	}
}

func defaultScalarConfigJSON() string {
	config := newScalarConfig()
	config.URL = "/swagger/doc.json"
	data, _ := json.Marshal(config)
	return string(data)
}

func scalarSourcesConfigJSON(sources []ScalarSource) string {
	normalized := normalizeScalarSources(sources)
	if len(normalized) == 0 {
		return defaultScalarConfigJSON()
	}

	configSources := make([]scalarConfigSource, 0, len(normalized))
	for _, source := range normalized {
		configSources = append(configSources, scalarConfigSource{
			Title: source.Title,
			Slug:  source.Slug,
			URL:   source.URL,
		})
	}

	config := newScalarConfig()
	config.Sources = configSources
	data, _ := json.Marshal(config)
	return string(data)
}

func newScalarConfig() scalarConfig {
	return scalarConfig{
		Layout:      "modern",
		Theme:       "default",
		PersistAuth: true,
		Telemetry:   false,
		Agent: scalarAgentConfig{
			Disabled: true,
		},
		DefaultHTTPClient: scalarHTTPClient{
			TargetKey: "shell",
			ClientKey: "curl",
		},
	}
}

func normalizeScalarSources(sources []ScalarSource) []ScalarSource {
	normalized := make([]ScalarSource, 0, len(sources))
	for _, source := range sources {
		if source.URL == "" {
			source.URL = scalarSpecPath + source.Slug + ".json"
		}
		normalized = append(normalized, source)
	}
	return normalized
}

func splitSwaggerDocByPathPrefix(raw []byte, source ScalarSource) ([]byte, error) {
	var spec map[string]any
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, err
	}

	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		return nil, errors.New("swagger paths missing")
	}

	filteredPaths := make(map[string]any)
	for path, item := range paths {
		if strings.HasPrefix(path, source.PathPrefix) {
			filteredPaths[path] = item
		}
	}
	spec["paths"] = filteredPaths

	if info, ok := spec["info"].(map[string]any); ok && source.Title != "" {
		info["title"] = source.Title
	}
	filterSwaggerTags(spec, filteredPaths)

	return json.Marshal(spec)
}

func filterSwaggerTags(spec map[string]any, paths map[string]any) {
	tags, ok := spec["tags"].([]any)
	if !ok {
		return
	}

	usedTags := collectSwaggerTags(paths)
	filteredTags := make([]any, 0, len(tags))
	for _, tag := range tags {
		tagMap, ok := tag.(map[string]any)
		if !ok {
			continue
		}
		name, ok := tagMap["name"].(string)
		if !ok {
			continue
		}
		if _, ok := usedTags[name]; ok {
			filteredTags = append(filteredTags, tag)
		}
	}
	spec["tags"] = filteredTags
}

func collectSwaggerTags(paths map[string]any) map[string]struct{} {
	usedTags := make(map[string]struct{})
	for _, pathItem := range paths {
		methods, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}
		for _, operation := range methods {
			op, ok := operation.(map[string]any)
			if !ok {
				continue
			}
			tags, ok := op["tags"].([]any)
			if !ok {
				continue
			}
			for _, tag := range tags {
				name, ok := tag.(string)
				if ok {
					usedTags[name] = struct{}{}
				}
			}
		}
	}
	return usedTags
}

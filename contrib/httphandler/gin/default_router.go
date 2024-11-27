package gin

import (
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"path"
)

// abortIndex represents a typical value used in abort functions.
const abortIndex int8 = math.MaxInt8 >> 1

var (
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

type registry struct {
	method  string
	path    string
	handler gin.HandlersChain
}

var _ Router = (*DefaultRouter)(nil)

type DefaultRouter struct {
	RouterGroup
	registries []registry
}

func (d *DefaultRouter) on(method, path string, handler gin.HandlersChain) {
	d.registries = append(d.registries, registry{
		method:  method,
		path:    path,
		handler: handler,
	})
}

func (d *DefaultRouter) RegisterHandler(engine *gin.Engine) {
	for _, registryItem := range d.registries {
		engine.RouterGroup.Handle(registryItem.method, registryItem.path, registryItem.handler...)
	}
}

func NewDefaultRouter(path string) *DefaultRouter {
	r := &DefaultRouter{
		RouterGroup: RouterGroup{
			handlers: nil,
			basePath: path,
			router:   nil,
		},
		registries: make([]registry, 0, 20),
	}
	r.RouterGroup.router = r
	return r
}

type RouterGroup struct {
	handlers gin.HandlersChain
	basePath string
	router   *DefaultRouter
}

func (group *RouterGroup) combineHandlers(handlers gin.HandlersChain) gin.HandlersChain {
	finalSize := len(group.handlers) + len(handlers)
	assert1(finalSize < int(abortIndex), "too many handlers")
	mergedHandlers := make(gin.HandlersChain, finalSize)
	copy(mergedHandlers, group.handlers)
	copy(mergedHandlers[len(group.handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

func (group *RouterGroup) Use(middleware ...gin.HandlerFunc) *RouterGroup {
	group.handlers = append(group.handlers, middleware...)
	return group
}

func (group *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{
		handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		router:   group.router,
	}
}

func (group *RouterGroup) GET(relativePath string, handlers ...gin.HandlerFunc) {
	group.router.on(http.MethodGet, group.calculateAbsolutePath(relativePath), group.combineHandlers(handlers))
}

func (group *RouterGroup) POST(relativePath string, handlers ...gin.HandlerFunc) {
	group.router.on(http.MethodPost, group.calculateAbsolutePath(relativePath), group.combineHandlers(handlers))
}

func (group *RouterGroup) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	group.router.on(http.MethodDelete, group.calculateAbsolutePath(relativePath), group.combineHandlers(handlers))
}

func (group *RouterGroup) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	group.router.on(http.MethodPatch, group.calculateAbsolutePath(relativePath), group.combineHandlers(handlers))
}

func (group *RouterGroup) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	group.router.on(http.MethodPut, group.calculateAbsolutePath(relativePath), group.combineHandlers(handlers))
}

func (group *RouterGroup) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	group.router.on(http.MethodOptions, group.calculateAbsolutePath(relativePath), group.combineHandlers(handlers))
}

func (group *RouterGroup) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	group.router.on(http.MethodHead, group.calculateAbsolutePath(relativePath), group.combineHandlers(handlers))
}

func (group *RouterGroup) Any(relativePath string, handlers ...gin.HandlerFunc) {
	finalPath := group.calculateAbsolutePath(relativePath)
	finalHandlers := group.combineHandlers(handlers)
	for _, method := range anyMethods {
		group.router.on(method, finalPath, finalHandlers)
	}
}

func (group *RouterGroup) Match(methods []string, relativePath string, handlers ...gin.HandlerFunc) {
	finalPath := group.calculateAbsolutePath(relativePath)
	finalHandlers := group.combineHandlers(handlers)
	for _, method := range methods {
		group.router.on(method, finalPath, finalHandlers)
	}
}

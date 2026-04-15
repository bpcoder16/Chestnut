package gin

import (
	ginpprof "github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

// PProfRegister 将标准 pprof 调试路由注册到指定 RouterGroup。
// 路由前缀为 /debug/pprof/，与 Go 官方工具链兼容。
// 仅在需要性能分析时启用，生产环境建议关闭。
func PProfRegister(r *gin.RouterGroup) {
	ginpprof.RouteRegister(r)
}

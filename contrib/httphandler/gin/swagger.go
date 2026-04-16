package gin

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerRegister 将 Swagger UI 路由注册到指定 RouterGroup。
// 访问路径为 /swagger/index.html。
// 调用方需在自身包中通过 blank import 注册生成的文档，例如：_ "yourmodule/swagger"
func SwaggerRegister(r *gin.RouterGroup) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

package gin

import (
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// IPWhitelistConfig 定义 Gin IP 白名单中间件配置。
type IPWhitelistConfig struct {
	Enabled  bool
	AllowIPs []string
}

// NewIPWhitelistMiddleware 创建只信任 TCP 远端地址的 Gin IP 白名单中间件。
func NewIPWhitelistMiddleware(config IPWhitelistConfig) (gin.HandlerFunc, error) {
	allowIPSet := make(map[string]struct{}, len(config.AllowIPs))
	if config.Enabled {
		for _, allowIP := range config.AllowIPs {
			parsedIP := net.ParseIP(allowIP)
			if parsedIP == nil {
				return nil, errors.New("invalid ip whitelist item: " + allowIP)
			}
			allowIPSet[parsedIP.String()] = struct{}{}
		}
	}

	return func(ctx *gin.Context) {
		if !config.Enabled {
			ctx.Next()
			return
		}

		remoteIP := parseRemoteIP(ctx.Request.RemoteAddr)
		if _, ok := allowIPSet[remoteIP]; ok {
			ctx.Next()
			return
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code": http.StatusForbidden,
			"msg":  "权限不足",
		})
	}, nil
}

func parseRemoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	parsedIP := net.ParseIP(host)
	if parsedIP == nil {
		return ""
	}
	return parsedIP.String()
}

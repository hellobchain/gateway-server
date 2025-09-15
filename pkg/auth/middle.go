package auth

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/gateway-server/pkg/config"
)

// Middleware 返回一个可插拔的 gin 中间件
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()
		jwt := cfg.JWT
		if !jwt.Enabled {
			c.Next()
			return
		}
		current := c.Request.URL.Path
		for _, p := range jwt.SkipPaths {
			if matched(p, current) {
				c.Next()
				return
			}
		}
		h := c.GetHeader(cfg.Server.Header)
		if h == "" {
			ResultCode(c, http.StatusUnauthorized, "missing "+cfg.Server.Header)
			return
		}
		claims, err := Validate(h) // 验签 + 状态
		if err != nil {
			ResultCode(c, http.StatusUnauthorized, err.Error())
			return
		}
		// 往请求头写用户数据
		c.Request.Header.Set("X-User-Info", claims["sub"].(string))
		_ = SetClaims(claims["jti"].(string), claims)
		c.Next()
	}
}

// matched 支持通配符 *
func matched(pattern, target string) bool {
	if pattern == target {
		return true
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		return path.Dir(target) == path.Dir(pattern)
	}
	return false
}

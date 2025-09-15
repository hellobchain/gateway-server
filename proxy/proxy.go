package proxy

import (
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// NewReverseProxy 创建针对 target 的反向代理
func NewReverseProxy(target string) *httputil.ReverseProxy {
	u, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(u)
}

// Handler 包装一下，把前缀去掉再转发
func Handler(p *httputil.ReverseProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 例如 /order/create -> 转发到后端 /create
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, c.Param("prefix"))
		p.ServeHTTP(c.Writer, c.Request)
	}
}

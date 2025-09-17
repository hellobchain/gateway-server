package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/gateway-server/pkg/auth"
	"github.com/hellobchain/gateway-server/pkg/lb"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(nil)

// NewReverseProxy 创建针对 target 的反向代理
func NewReverseProxy(target string) *httputil.ReverseProxy {
	u, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(u)
}

// Handler 包装一下，把前缀去掉再转发
func Handler(p *httputil.ReverseProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		p.ServeHTTP(c.Writer, c.Request)
	}
}

func LbHandler(lbBalancer lb.Balancer) gin.HandlerFunc {
	return func(c *gin.Context) {
		addr, ok, isRemovePrex := lbBalancer.Pick()
		if !ok {
			auth.ResultCode(c, http.StatusServiceUnavailable, "no available instance")
			return
		}
		logger.Debugf("pick instance: %s", addr)
		p := NewReverseProxy(addr)
		if isRemovePrex {
			prefix := c.Param("proxyPath")
			if prefix != "" && prefix[0] == '/' {
				prefix = prefix[1:]
			}
			targetPath := "/" + prefix
			c.Request.URL.Path = targetPath
			logger.Debugf("real router path: %s", targetPath)
			c.Request.URL.Path = targetPath
		}
		p.ServeHTTP(c.Writer, c.Request)
	}
}

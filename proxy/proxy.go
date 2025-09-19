package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/gateway-server/pkg/auth"
	"github.com/hellobchain/gateway-server/pkg/breaker"
	"github.com/hellobchain/gateway-server/pkg/lb"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(nil)

// NewReverseProxy 创建针对 target 的反向代理
func NewReverseProxy(target string) *httputil.ReverseProxy {
	u, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(u)
}

// Handler
func Handler(p *httputil.ReverseProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		p.ServeHTTP(c.Writer, c.Request)
	}
}

func LbHandler(lbBalancer lb.Balancer, sreBreaker *breaker.SreBreaker) gin.HandlerFunc {
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
		if sreBreaker.Enabled() {
			p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				w.WriteHeader(http.StatusBadGateway)
			}
			err := sreBreaker.Do(func() error {
				p.ServeHTTP(c.Writer, c.Request)
				if c.Writer.Status() >= 500 {
					return fmt.Errorf("backend 5xx")
				}
				return nil
			})
			if err == breaker.ErrBreakerOpen {
				auth.ResultCode(c, http.StatusServiceUnavailable, "circuit breaker open")
			}
		} else {
			p.ServeHTTP(c.Writer, c.Request)
		}
	}
}

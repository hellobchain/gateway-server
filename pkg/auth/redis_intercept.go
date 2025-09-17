package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/gateway-server/pkg/config"
)

// RedisIntercept 入口
func RedisIntercept() gin.HandlerFunc {
	return func(c *gin.Context) {
		interceptCfg := config.Get().InterceptConfig
		if !interceptCfg.Enabled {
			c.Next()
			return
		}
		// 1. URL 黑白名单
		if _, ok := urlCheck(c.Request.URL.Path, interceptCfg.URL); ok {
			ResultCode(c, http.StatusForbidden, "url forbidden")
			c.Abort()
			return
		}
		// 2. 全局 QPS
		if !globalLimit(interceptCfg) {
			ResultCode(c, http.StatusTooManyRequests, "global limit")
			c.Abort()
			return
		}
		// 3. IP 级流控
		ip := c.ClientIP()
		if interceptCfg.IP.FlowLimit && !ipLimit(ip, interceptCfg.IP.QPS) {
			ResultCode(c, http.StatusTooManyRequests, "ip limit")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ---------- 辅助函数 ----------
func urlCheck(path string, urlCfg config.InterceptUrlConfig) (hit bool, forbidden bool) {
	for _, p := range urlCfg.BlackList {
		if matched(p, path) {
			return true, true
		}
	}
	if len(urlCfg.WhiteList) > 0 {
		for _, p := range urlCfg.WhiteList {
			if matched(p, path) {
				return true, false
			}
		}
		return true, true
	}
	return false, false
}

func globalLimit(interceptConfig config.InterceptConfig) bool {
	key := globalQpsKey()
	n, err := Incr(key)
	if err != nil {
		return false
	}
	if n == 1 {
		Expire(key, time.Second)
	}
	return n <= int64(interceptConfig.Global.QPS)
}

func ipLimit(ip string, qps int) bool {
	key := ipQpsKey(ip)
	n, err := Incr(key)
	if err != nil {
		return false
	}
	if n == 1 {
		Expire(key, time.Second)
	}
	return n <= int64(qps)
}

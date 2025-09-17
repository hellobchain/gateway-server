package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hellobchain/gateway-server/middleware"
	"github.com/hellobchain/gateway-server/pkg/auth"
	"github.com/hellobchain/gateway-server/pkg/config"
	log "github.com/hellobchain/gateway-server/pkg/logger"
	"github.com/hellobchain/gateway-server/proxy"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(log.LogConfig)

// Register 初始化 + 定时同步配置变化
func Register(r *gin.Engine, cfg *config.Cfg) {
	// 全局中间件
	r.Use(middleware.Logger(), gin.Recovery(), middleware.CORS(), auth.Middleware(), auth.RedisIntercept())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	// 首次加载
	loadRoutes(r, cfg)
}

// reloadRoutes 增量更新路由
func loadRoutes(r *gin.Engine, cfg *config.Cfg) {
	// 初始化 JWT 组件
	auth.Init(cfg.JWT)

	newRules := make(map[string]bool)
	for _, rule := range cfg.Routes {
		path := rule.Path
		if _, ok := newRules[path]; ok {
			continue // 已存在
		}
		// 新增路由
		p := proxy.NewReverseProxy(rule.Target)
		newRules[path] = true
		r.Any(path, proxy.Handler(p))
		r.Any(path+"/*proxyPath", proxy.Handler(p))
		logger.Infof("registered route: %s -> %s", path, rule.Target)
	}
}

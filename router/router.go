package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/gateway-server/middleware"
	"github.com/hellobchain/gateway-server/pkg/auth"
	"github.com/hellobchain/gateway-server/pkg/config"
	"github.com/hellobchain/gateway-server/pkg/lb"
	log "github.com/hellobchain/gateway-server/pkg/logger"
	"github.com/hellobchain/gateway-server/proxy"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(log.LogConfig)

func initLB(insts []lb.Instance) lb.Balancer {
	balancer := lb.New(insts)
	// 每 5 秒探活
	checker := lb.NewHealthChecker(balancer, 5*time.Second)
	checker.Start(insts)
	return balancer
}

// Register 初始化 + 定时同步配置变化
func Register(r *gin.Engine, cfg config.Cfg) {
	// 全局中间件
	r.Use(middleware.Logger(), gin.Recovery(), middleware.CORS(), auth.Middleware(), auth.RedisIntercept())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	// 首次加载
	loadRoutes(r, cfg)
}

// reloadRoutes 增量更新路由
func loadRoutes(r *gin.Engine, cfg config.Cfg) {
	// 初始化 JWT 组件
	auth.Init(cfg.JWT)
	newRules := make(map[string]bool)
	for _, rule := range cfg.Routes {
		path := rule.Path
		if _, ok := newRules[path]; ok {
			continue // 已存在
		}
		lbBalancer := initLB(getLbInstances(rule.Targets))
		newRules[path] = true
		r.Any(path, proxy.LbHandler(lbBalancer))
		r.Any(path+"/*proxyPath", proxy.LbHandler(lbBalancer))
		logger.Infof("registered route: %s -> %v", path, rule.Targets)
	}
}

func getLbInstances(rtcs []config.RouterTargetsConfig) []lb.Instance {
	lbInstances := make([]lb.Instance, len(rtcs))
	for i, rtc := range rtcs {
		lbInstances[i].Addr = rtc.Target
		lbInstances[i].Weight = rtc.Weight
		lbInstances[i].Protocol = rtc.Protocol
		lbInstances[i].IsRemovePrex = rtc.IsRemovePrex
	}
	return lbInstances
}

package router

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/gateway-server/middleware"
	"github.com/hellobchain/gateway-server/pkg/auth"
	"github.com/hellobchain/gateway-server/pkg/breaker"
	"github.com/hellobchain/gateway-server/pkg/config"
	"github.com/hellobchain/gateway-server/pkg/lb"
	"github.com/hellobchain/gateway-server/proxy"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(nil)

func initLB(insts []lb.Instance, healthCheckInterval int) lb.Balancer {
	if healthCheckInterval == 0 { // 默认 5 秒
		healthCheckInterval = 5
	}
	balancer := lb.New(insts)
	// 每 5 秒探活
	checker := lb.NewHealthChecker(balancer, time.Duration(healthCheckInterval)*time.Second)
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
	breakerConfig := config.Get().Breaker
	sreBreaker := breaker.New(breaker.Settings{
		Enabled:               breakerConfig.Enabled,
		MaxRequests:           breakerConfig.MaxRequests,
		Interval:              breakerConfig.Interval,
		Timeout:               breakerConfig.Timeout,
		ErrorPercentThreshold: breakerConfig.ErrorPercent,
		MinRequestAmount:      breakerConfig.MinRequestAmount,
	})
	// 初始化 JWT 组件
	auth.Init(cfg.JWT)
	newRules := make(map[string]bool)
	for _, rule := range cfg.Routes {
		path := rule.Path
		if _, ok := newRules[path]; ok {
			continue // 已存在
		}
		lbBalancer := initLB(getLbInstances(rule.Targets), rule.HealthCheckInterval)
		newRules[path] = true
		r.Any(path, proxy.LbHandler(lbBalancer, sreBreaker))
		r.Any(path+"/*proxyPath", proxy.LbHandler(lbBalancer, sreBreaker))
		logger.Infof("registered route: %s -> %s", path, toString(rule.Targets))
	}
}

func toString(rtcs []config.RouterTargetsConfig) string {
	ret, _ := json.Marshal(rtcs)
	return string(ret)
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

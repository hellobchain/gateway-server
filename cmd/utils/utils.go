package utils

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/gateway-server/pkg/auth"
	"github.com/hellobchain/gateway-server/pkg/config"
	log "github.com/hellobchain/gateway-server/pkg/logger"
	"github.com/hellobchain/gateway-server/router"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(log.LogConfig)

func Init() {
	loadConfig()
	cfg := config.Get()
	initTokenStore(cfg)
	startWebServer(cfg, registerGinRouter(cfg))
}

func loadConfig() {
	// 加载配置
	if len(os.Args) > 1 {
		config.Load(os.Args[1])
	} else {
		config.Load("config.yml")
	}
}

// 初始化 TokenStore
func initTokenStore(cfg *config.Cfg) {
	var store auth.TokenStore
	switch cfg.JWT.Store.Type {
	case "redis":
		rStore, err := auth.NewRedisStore(cfg.JWT.Store.Redis)
		if err != nil {
			logger.Fatalf("redis store: %v", err)
		}
		store = rStore
	case "memory":
		mStore, err := auth.NewMemoryStore(cfg.JWT)
		if err != nil {
			logger.Fatalf("memory store: %v", err)
		}
		store = mStore
	default:
		logger.Fatalf("unknown store type: %s", cfg.JWT.Store.Type)
	}
	auth.SetStore(store)
}

// 注册所有路由与中间件
func registerGinRouter(cfg *config.Cfg) *gin.Engine {
	r := gin.New()
	gin.SetMode(cfg.Server.Mode)
	router.Register(r, cfg) // 注册所有路由与中间件
	return r
}

func startWebServer(cfg *config.Cfg, r *gin.Engine) {
	webAddress := config.GetWebServerAddress(cfg)
	logger.Info("Gateway-server listening on " + webAddress)
	if err := r.Run(webAddress); err != nil {
		logger.Fatalf("start failed: %v", err)
	}
}

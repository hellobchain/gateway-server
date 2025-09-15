package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	log "github.com/hellobchain/gateway-server/pkg/logger"
	"github.com/hellobchain/wswlog/wlogging"
	"github.com/spf13/viper"
)

var logger = wlogging.MustGetFileLoggerWithoutName(log.LogConfig)

type Cfg struct {
	Server *ServerConfig   `mapstructure:"server"` // 服务器配置
	Routes *[]RoutesConfig `mapstructure:"routes"` // 路由配置
	JWT    *JWT            `mapstructure:"jwt"`    // JWT 配置
}
type ServerConfig struct {
	Port     int    `mapstructure:"port"`      // 监听端口
	Header   string `mapstructure:"header"`    // token 的 header
	LogLevel string `mapstructure:"log_level"` // 日志级别
	Mode     string `mapstructure:"mode"`      // 运行模式
}
type RoutesConfig struct {
	Path   string `mapstructure:"path"`   // 匹配的路径
	Target string `mapstructure:"target"` // 目标地址
}
type JWT struct {
	Enabled    bool         `mapstructure:"enabled"`
	Algorithm  string       `mapstructure:"algorithm"`   // HS256 / ES256
	Secret     string       `mapstructure:"secret"`      // 对称密钥
	PublicKey  string       `mapstructure:"public_key"`  // 仅 ES256 用
	PrivateKey string       `mapstructure:"private_key"` // 仅 ES256 用
	SkipPaths  []string     `mapstructure:"skip_paths"`
	Store      *StoreConfig `mapstructure:"store"` // 存储配置
}

type StoreConfig struct {
	Type   string        `mapstructure:"type"` // memory | redis
	Memory *MemoryConfig `mapstructure:"memory"`
	Redis  *RedisConfig  `mapstructure:"redis"`
}

type MemoryConfig struct {
	CleanupIntervalSec int `mapstructure:"cleanup_interval_sec"`
}

type RedisConfig struct {
	Addr      string `mapstructure:"addr"`
	Password  string `mapstructure:"password"`
	DB        int    `mapstructure:"db"`
	BufferSec int    `mapstructure:"buffer_sec"`
}

var (
	once sync.Once
	cfg  *Cfg
	mu   sync.RWMutex
)

// Get 读取当前配置（并发安全）
func Get() *Cfg {
	mu.RLock()
	defer mu.RUnlock()
	return cfg
}

const (
	// pre
	cmdPre = "GW"
)

func setEnvVariables() {
	// For environment variables.
	viper.SetEnvPrefix(cmdPre)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
}

// Load 加载并监听热更新
func Load(path string) {
	once.Do(func() {
		setEnvVariables()
		viper.SetConfigFile(path)
		if err := viper.ReadInConfig(); err != nil {
			logger.Fatalf("read config failed: %v", err)
		}
		if err := viper.Unmarshal(&cfg); err != nil {
			logger.Fatalf("unmarshal config failed: %v", err)
		}
		log.SetLogLevel(cfg.Server.LogLevel)
		// 打印
		ret, _ := json.MarshalIndent(cfg, "", "  ")
		logger.Debugf("config: %v", string(ret))
		logger.Info("config loaded")

		// 热加载
		viper.WatchConfig()
		viper.OnConfigChange(func(in fsnotify.Event) {
			mu.Lock()
			if err := viper.Unmarshal(&cfg); err != nil {
				logger.Errorf("reload config error: %v", err)
			} else {
				log.SetLogLevel(cfg.Server.LogLevel)
				ret, _ := json.MarshalIndent(cfg, "", "  ")
				logger.Debugf("config: %v", string(ret))
				logger.Info("config reloaded")
			}
			mu.Unlock()
		})
	})
}

func GetWebServerAddress(cfg *Cfg) string {
	return fmt.Sprintf(":%d", cfg.Server.Port)
}

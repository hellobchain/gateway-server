package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hellobchain/wswlog/wlogging"
	"github.com/spf13/viper"
)

var logger = wlogging.MustGetFileLoggerWithoutName(nil)

type Cfg struct {
	Server          ServerConfig    `mapstructure:"server"`    // 服务器配置
	Routes          []RoutesConfig  `mapstructure:"routes"`    // 路由配置
	JWT             JWT             `mapstructure:"jwt"`       // JWT 配置
	InterceptConfig InterceptConfig `mapstructure:"intercept"` // 拦截配置
	Breaker         Breaker         `mapstructure:"breaker"`   // 熔断配置
}
type ServerConfig struct {
	Port     int    `mapstructure:"port"`      // 监听端口
	LogLevel string `mapstructure:"log_level"` // 日志级别 info debug error
	Mode     string `mapstructure:"mode"`      // 运行模式 debug release test
}
type RoutesConfig struct {
	Path                string                `mapstructure:"path"`                  // 匹配的路径
	Targets             []RouterTargetsConfig `mapstructure:"targets"`               // 目标地址
	IsJwt               bool                  `mapstructure:"is_jwt"`                // 是否需要 JWT
	Header              string                `mapstructure:"header"`                // 请求头
	HealthCheckInterval int                   `mapstructure:"health_check_interval"` // 健康检查间隔
}

type RouterTargetsConfig struct {
	Target       string `mapstructure:"target"`         // 目标地址 192.168.80:80
	Protocol     string `mapstructure:"protocol"`       // 协议 http
	Weight       int    `mapstructure:"weight"`         // 权重
	IsRemovePrex bool   `mapstructure:"is_remove_prex"` // 是否移除前缀
}
type JWT struct {
	Enabled    bool        `mapstructure:"enabled"`     // 是否启用
	Algorithm  string      `mapstructure:"algorithm"`   // HS256 / ES256
	Secret     string      `mapstructure:"secret"`      // 对称密钥
	PublicKey  string      `mapstructure:"public_key"`  // 仅 ES256 用
	PrivateKey string      `mapstructure:"private_key"` // 仅 ES256 用
	SkipPaths  []string    `mapstructure:"skip_paths"`  // 跳过认证的路径
	Store      StoreConfig `mapstructure:"store"`       // 存储配置
}

type StoreConfig struct {
	Type   string       `mapstructure:"type"`   // memory | redis
	Memory MemoryConfig `mapstructure:"memory"` // 内存配置
	Redis  RedisConfig  `mapstructure:"redis"`  // redis 配置
}

type MemoryConfig struct {
	CleanupIntervalSec int `mapstructure:"cleanup_interval_sec"` // 内存清理间隔
}

type RedisConfig struct {
	Addr      string `mapstructure:"addr"`       // redis 地址
	Password  string `mapstructure:"password"`   // redis 密码
	DB        int    `mapstructure:"db"`         // redis db
	BufferSec int    `mapstructure:"buffer_sec"` // redis 缓存时间
}

type InterceptConfig struct {
	Enabled bool                  `mapstructure:"enabled"` // 是否开启拦截
	IP      InterceptIpConfig     `mapstructure:"ip"`      // ip 拦截
	URL     InterceptUrlConfig    `mapstructure:"url"`     // url 拦截
	Global  InterceptGlobalConfig `mapstructure:"global"`  // 全局拦截
}

type InterceptIpConfig struct {
	FlowLimit bool `mapstructure:"flow_limit"` // ip 流量拦截
	QPS       int  `mapstructure:"qps"`        // ip 拦截
}

type InterceptUrlConfig struct {
	WhiteList []string `mapstructure:"white_list"` // url 白名单
	BlackList []string `mapstructure:"black_list"` // url 黑名单
}

type InterceptGlobalConfig struct {
	QPS int `mapstructure:"qps"` // 全局拦截
}

type Breaker struct {
	Enabled          bool          `mapstructure:"enabled"`            // 熔断器是否开启
	MaxRequests      uint32        `mapstructure:"max_requests"`       // 半开时最大探测请求数
	Interval         time.Duration `mapstructure:"interval"`           // 统计窗口
	Timeout          time.Duration `mapstructure:"timeout"`            // 熔断后多久进入半开
	ErrorPercent     float64       `mapstructure:"error_percent"`      // 错误率阈值（0-100）
	MinRequestAmount uint32        `mapstructure:"min_request_amount"` // 最小请求数才触发错误率计算
}

var (
	once sync.Once    // 配置单例
	cfg  Cfg          // 配置
	mu   sync.RWMutex // 读写锁
)

// Get 读取当前配置（并发安全）
func Get() Cfg {
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
		wlogging.SetGlobalLogLevel(cfg.Server.LogLevel)
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
				wlogging.SetGlobalLogLevel(cfg.Server.LogLevel)
				ret, _ := json.MarshalIndent(cfg, "", "  ")
				logger.Debugf("config: %v", string(ret))
				logger.Info("config reloaded")
			}
			mu.Unlock()
		})
	})
}

func GetWebServerAddress(cfg Cfg) string {
	return fmt.Sprintf(":%d", cfg.Server.Port)
}

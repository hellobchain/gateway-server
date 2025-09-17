package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hellobchain/wswlog/wlogging"
)

// Logger 简单请求日志
var logClient = wlogging.MustGetFileLoggerWithoutName(nil)

// 请求日志汇总信息
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := uuid.New().ID()
		// 开始时间
		start := time.Now()
		// path
		path := c.Request.URL.Path
		// ip
		clientIP := c.ClientIP()
		// 方法
		method := c.Request.Method
		// 处理请求
		c.Next()
		// 结束时间
		end := time.Now()
		// 执行时间
		latency := end.Sub(start)
		// 状态
		statusCode := c.Writer.Status()
		logClient.Infof("| %10d | %3d | %13v | %15s | %s  %s |",
			requestId,
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

// CORS 允许跨域
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{"*"},
	})
}

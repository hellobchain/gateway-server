package middleware

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Logger 简单请求日志
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return "[" + param.TimeStamp.Format(time.RFC3339) + "] " +
			param.Method + " " + param.Path + " " + fmt.Sprintf("%d", param.StatusCode) + " " +
			param.Latency.String() + "\n"
	})
}

// CORS 允许跨域
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{"*"},
	})
}

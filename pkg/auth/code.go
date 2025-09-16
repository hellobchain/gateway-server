package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/hellobchain/gateway-server/pkg/logger"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(log.LogConfig)

func ResultCode(ctx *gin.Context, code int, msg string) {
	var httpCode = http.StatusOK
	if code != 200 && code != 0 {
		path := ctx.Request.URL.Path
		logger.Errorf("[ResultCode] path:%s code:%d msg:%s", path, code, msg)
	}
	ctx.JSON(httpCode, gin.H{"code": code, "msg": msg})
}

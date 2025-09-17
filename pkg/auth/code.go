package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(nil)

func ResultCode(ctx *gin.Context, code int, msg string) {
	var httpCode = http.StatusOK
	if code != 200 && code != 0 {
		path := ctx.Request.URL.Path
		logger.Errorf("[ResultCode] path:%s code:%d msg:%s", path, code, msg)
	}
	ctx.JSON(httpCode, gin.H{"code": code, "msg": msg})
}

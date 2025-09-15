package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ResultCode(ctx *gin.Context, code int, msg string) {
	var httpCode = http.StatusOK
	ctx.JSON(httpCode, gin.H{"code": code, "msg": msg})
}

package auth

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
)

func TestJwt(t *testing.T) {
	tokenStr := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjAwODUyMzMsImlhdCI6MTc2MDA4MTYzMywiVXNlcklkIjo0MywiVXNlck5hbWUiOiJrdXZlcmFfYXBwIiwiVXNlclR5cGUiOiIwMSIsIlV1aWQiOiI1NmZjZWQ4ZGZjMmY0MmVmOTFmZDE4NmI4NmYwMWUzMiJ9.Rgz2rqDK4cwjNOaWCww2Gq4JDWLZhmme_TAnJj_9eNY"
	signMethod := jwt.SigningMethodHS256
	secret = []byte("OR56PELLdDcY")
	ret, err := LoadJwtClaims(tokenStr, signMethod)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("success:%v %v", ret, ret.GetUuid())
}

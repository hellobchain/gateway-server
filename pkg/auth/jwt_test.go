package auth

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
)

func TestJwt(t *testing.T) {
	tokenStr := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTgwMDc3MzgsImlhdCI6MTc1NzkyMTMzOCwiVXNlcklkIjoxOCwiVXNlck5hbWUiOiJ6amZfMDMiLCJVc2VyVHlwZSI6IjA0MCIsIlV1aWQiOiJiNDE1Zjc2ZWEwMDg0MTU1YmQyYzYzNDkxZmYzYmVmNCJ9.OUhrsv12d6-LruaX7ETZDu5LvcBtw7xwNaSlhZFV3mY"
	signMethod := jwt.SigningMethodHS256
	ret, err := LoadJwtClaims(tokenStr, signMethod)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("success:%v", ret)
}

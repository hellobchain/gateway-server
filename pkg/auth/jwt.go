package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hellobchain/gateway-server/pkg/config"
)

var (
	signMethod jwt.SigningMethod
	publicKey  *ecdsa.PublicKey // ES256 用
	secret     []byte           // HS256 用
	once       sync.Once
)

// Init 根据配置初始化验签参数
func Init(cfg *config.JWT) {
	once.Do(func() {
		switch cfg.Algorithm {
		case "HS256":
			signMethod = jwt.SigningMethodHS256
			secret = []byte(cfg.Secret)
		case "ES256":
			signMethod = jwt.SigningMethodES256
			pubKeyBytes := []byte(cfg.PublicKey)
			block, _ := pem.Decode(pubKeyBytes)
			if block == nil {
				panic("invalid ES256 public key")
			}
			k, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				panic(err)
			}
			publicKey = k.(*ecdsa.PublicKey)
		default:
			panic("unsupported jwt algorithm: " + cfg.Algorithm)
		}
	})
}

// Validate 验签 + Redis 状态检查，并返回 claims
func Validate(bearer string) (jwt.MapClaims, error) {
	if len(bearer) < 8 || bearer[:7] != "Bearer " {
		return nil, fmt.Errorf("invalid bearer format")
	}
	tokenStr := bearer[7:]

	// 1. 解析 & 验签
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if t.Method != signMethod {
			return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
		}
		switch signMethod {
		case jwt.SigningMethodHS256:
			return secret, nil
		case jwt.SigningMethodES256:
			return publicKey, nil
		default:
			return nil, fmt.Errorf("unknown algorithm")
		}
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	// 2. 取出 jti 做 Redis 校验
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return nil, fmt.Errorf("missing jti")
	}
	valid, err := IsTokenValid(jti)
	if err != nil {
		return nil, fmt.Errorf("redis err: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("token revoked")
	}

	return claims, nil
}

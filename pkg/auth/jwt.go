package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/hellobchain/gateway-server/pkg/config"
)

var (
	signMethod jwt.SigningMethod // / 签名方法
	publicKey  *ecdsa.PublicKey  // ES256 用
	secret     []byte            // HS256 用
	once       sync.Once         // 初始化一次
)

// Init 根据配置初始化验签参数
func Init(cfg config.JWT) {
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
func Validate(bearer string) (JwtMapClaims, error) {
	tokenStr := strings.TrimPrefix(bearer, "Bearer ")
	claims, err := LoadJwtClaims(tokenStr, signMethod)
	if err != nil {
		return nil, err
	}

	jti := claims.GetUuid()
	if jti == "" {
		return nil, fmt.Errorf("missing jti")
	}
	valid, err := IsTokenValid(validTokenKey(jti))
	if err != nil {
		return nil, fmt.Errorf("redis err: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("token revoked")
	}

	return claims, nil
}

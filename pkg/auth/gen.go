package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hellobchain/gateway-server/pkg/config"
)

// GenerateRequest 生成请求参数
type GenerateRequest struct {
	Subject string         // 用户标识
	Extra   map[string]any // 自定义claims
	Expiry  time.Duration  // 有效期，默认 24h
}

// GenerateResponse 返回
type GenerateResponse struct {
	Token     string
	ExpiresAt int64
	JTI       string
}

// GenerateToken 统一生成入口
func GenerateToken(req GenerateRequest) (*GenerateResponse, error) {
	if req.Subject == "" {
		return nil, fmt.Errorf("subject empty")
	}
	if req.Expiry == 0 {
		req.Expiry = 24 * time.Hour
	}
	cfg := config.Get().JWT

	now := time.Now()
	exp := now.Add(req.Expiry).Unix()
	jti := uuid.New().String()

	claims := jwt.MapClaims{
		"sub": req.Subject,
		"iat": now.Unix(),
		"exp": exp,
		"jti": jti,
	}
	maps.Copy(claims, req.Extra)

	var tokenString string
	var err error

	switch cfg.Algorithm {
	case "HS256":
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err = token.SignedString([]byte(cfg.Secret))
	case "ES256":
		privateKey, err := parseECPrivateKey([]byte(cfg.PrivateKey))
		if err != nil {
			return nil, err
		}
		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		tokenString, err = token.SignedString(privateKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported alg: %s", cfg.Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("sign fail: %w", err)
	}

	return &GenerateResponse{
		Token:     tokenString,
		ExpiresAt: exp,
		JTI:       jti,
	}, nil
}

// 解析 EC 私钥（ES256 用）
func parseECPrivateKey(keyPEM []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	k, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return k, nil
}

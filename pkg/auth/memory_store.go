package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hellobchain/gateway-server/pkg/config"
	"github.com/patrickmn/go-cache"
)

type memoryStore struct {
	c      *cache.Cache
	claims *cache.Cache // 单独 namespace 避免冲突
}

func NewMemoryStore(cfg *config.JWT) (TokenStore, error) {
	def := cache.New(cache.NoExpiration, time.Duration(cfg.Store.Memory.CleanupIntervalSec)*time.Second)
	return &memoryStore{
		c:      def,
		claims: def,
	}, nil
}
func (m *memoryStore) AddToken(jti string, exp int64) error {
	ttl := time.Until(time.Unix(exp, 0)) + time.Duration(config.Get().JWT.Store.Redis.BufferSec)*time.Second
	m.c.Set(jti, true, ttl)
	return nil
}
func (m *memoryStore) DelToken(jti string) error {
	m.c.Delete(jti)
	return nil
}
func (m *memoryStore) SetClaims(jti string, claims jwt.MapClaims) error {
	m.claims.Set(jti, claims, cache.DefaultExpiration)
	return nil
}

func (m *memoryStore) GetClaims(jti string) (jwt.MapClaims, error) {
	v, ok := m.claims.Get(jti)
	if !ok {
		return nil, fmt.Errorf("claims not found")
	}
	return v.(jwt.MapClaims), nil
}
func (m *memoryStore) IsTokenValid(jti string) (bool, error) {
	_, ok := m.c.Get(jti)
	return ok, nil
}

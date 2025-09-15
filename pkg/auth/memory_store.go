package auth

import (
	"fmt"
	"time"

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
	ttl := time.Until(time.Unix(exp, 0)) + time.Duration(config.Get().JWT.Store.Memory.CleanupIntervalSec)*time.Second
	m.c.Set(validKey(jti), true, ttl)
	return nil
}
func (m *memoryStore) DelToken(jti string) error {
	m.c.Delete(validKey(jti))
	return nil
}
func (m *memoryStore) IsTokenValid(jti string) (bool, error) {
	_, ok := m.c.Get(validKey(jti))
	return ok, nil
}
func (m *memoryStore) SetClaims(jti string, claims JwtMapClaims) error {
	m.claims.Set(claimsKey(jti), claims, cache.DefaultExpiration)
	return nil
}

func (m *memoryStore) GetClaims(jti string) (JwtMapClaims, error) {
	v, ok := m.claims.Get(claimsKey(jti))
	if !ok {
		return nil, fmt.Errorf("claims not found")
	}
	return v.(JwtMapClaims), nil
}

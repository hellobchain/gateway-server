package auth

import (
	"fmt"
	"time"

	"github.com/hellobchain/gateway-server/pkg/config"
	"github.com/patrickmn/go-cache"
)

type memoryStore struct {
	c      *cache.Cache // 单独 namespace 避免冲突
	claims *cache.Cache // 单独 namespace 避免冲突
}

func NewMemoryStore(cfg config.JWT) (TokenStore, error) {
	def := cache.New(cache.NoExpiration, time.Duration(cfg.Store.Memory.CleanupIntervalSec)*time.Second)
	return &memoryStore{
		c:      def,
		claims: def,
	}, nil
}
func (m *memoryStore) AddToken(key string, exp int64) error {
	ttl := time.Until(time.Unix(exp, 0)) + time.Duration(config.Get().JWT.Store.Memory.CleanupIntervalSec)*time.Second
	m.c.Set(key, true, ttl)
	return nil
}
func (m *memoryStore) DelToken(key string) error {
	m.c.Delete(key)
	return nil
}
func (m *memoryStore) IsTokenValid(key string) (bool, error) {
	_, ok := m.c.Get(key)
	return ok, nil
}
func (m *memoryStore) SetClaims(key string, claims JwtMapClaims) error {
	m.claims.Set(key, claims, cache.DefaultExpiration)
	return nil
}

func (m *memoryStore) GetClaims(key string) (JwtMapClaims, error) {
	v, ok := m.claims.Get(key)
	if !ok {
		return nil, fmt.Errorf("claims not found")
	}
	return v.(JwtMapClaims), nil
}

func (m *memoryStore) Incr(key string) (int64, error) {
	return m.c.IncrementInt64(key, 1)
}

func (m *memoryStore) Expire(key string, expire time.Duration) {
	m.c.Set(key, 1, expire)
}

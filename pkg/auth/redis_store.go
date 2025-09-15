package auth

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hellobchain/gateway-server/pkg/config"
)

type redisStore struct {
	client *redis.Client
	buffer time.Duration
}

func NewRedisStore(cfg *config.RedisConfig) (TokenStore, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &redisStore{
		client: rdb,
		buffer: time.Duration(cfg.BufferSec) * time.Second,
	}, nil
}

func (r *redisStore) AddToken(jti string, exp int64) error {
	ttl := time.Until(time.Unix(exp, 0)) + r.buffer
	return r.client.Set(context.Background(), validKey(jti), "1", ttl).Err()
}
func (r *redisStore) DelToken(jti string) error {
	return r.client.Del(context.Background(), validKey(jti)).Err()
}
func (r *redisStore) IsTokenValid(jti string) (bool, error) {
	n, err := r.client.Exists(context.Background(), validKey(jti)).Result()
	return n == 1, err
}
func (r *redisStore) SetClaims(jti string, claims JwtMapClaims) error {
	return r.client.HSet(context.Background(), claimsKey(jti), "$", claims).Err()
}

func (r *redisStore) GetClaims(jti string) (JwtMapClaims, error) {
	var m JwtMapClaims
	err := r.client.HGet(context.Background(), claimsKey(jti), "$").Scan(&m)
	return m, err
}
func validKey(jti string) string  { return "jwt:valid:" + jti }
func claimsKey(jti string) string { return "jwt:claims:" + jti }

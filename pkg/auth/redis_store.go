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

func NewRedisStore(cfg config.RedisConfig) (TokenStore, error) {
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

func (r *redisStore) AddToken(key string, exp int64) error {
	ttl := time.Until(time.Unix(exp, 0)) + r.buffer
	return r.client.Set(context.Background(), key, "1", ttl).Err()
}
func (r *redisStore) DelToken(key string) error {
	return r.client.Del(context.Background(), key).Err()
}
func (r *redisStore) IsTokenValid(key string) (bool, error) {
	n, err := r.client.Exists(context.Background(), key).Result()
	return n == 1, err
}
func (r *redisStore) SetClaims(key string, claims JwtMapClaims) error {
	return r.client.HSet(context.Background(), key, "$", claims).Err()
}

func (r *redisStore) GetClaims(key string) (JwtMapClaims, error) {
	var m JwtMapClaims
	err := r.client.HGet(context.Background(), key, "$").Scan(&m)
	return m, err
}

func (r *redisStore) Incr(key string) (int64, error) {
	return r.client.Incr(context.Background(), key).Result()
}

func (r *redisStore) Expire(key string, expire time.Duration) {
	r.client.Expire(context.Background(), key, expire)
}

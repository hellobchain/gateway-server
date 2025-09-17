package auth

import "time"

// TokenStore 定义行为
type TokenStore interface {
	AddToken(key string, exp int64) error
	DelToken(key string) error
	IsTokenValid(key string) (bool, error)
	SetClaims(key string, claims JwtMapClaims) error
	GetClaims(key string) (JwtMapClaims, error)
	Incr(key string) (int64, error)
	Expire(key string, expire time.Duration)
}

var store TokenStore

// SetStore 由 main 注入
func SetStore(s TokenStore) {
	store = s
}

// 下方三个函数直接代理到具体实现
func AddToken(key string, exp int64) error {
	return store.AddToken(key, exp)
}
func DelToken(key string) error {
	return store.DelToken(key)
}
func IsTokenValid(key string) (bool, error) {
	return store.IsTokenValid(key)
}

func SetClaims(key string, claims JwtMapClaims) error {
	return store.SetClaims(key, claims)
}
func GetClaims(key string) (JwtMapClaims, error) {
	return store.GetClaims(key)
}

func Incr(key string) (int64, error) { return store.Incr(key) }

func Expire(key string, expire time.Duration) { store.Expire(key, expire) }

func validTokenKey(jti string) string { return LOGIN_TOKEN_KEY + jti }

// func claimsKey(jti string) string     { return JWT_CLAIMS_KEY + jti }

func ipQpsKey(ip string) string { return IP_QPS_KEY + ip }

func globalQpsKey() string { return GLOBAL_QPS_KEY }

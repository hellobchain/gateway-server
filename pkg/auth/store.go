package auth

// TokenStore 定义行为
type TokenStore interface {
	AddToken(jti string, exp int64) error
	DelToken(jti string) error
	IsTokenValid(jti string) (bool, error)
	SetClaims(jti string, claims JwtMapClaims) error
	GetClaims(jti string) (JwtMapClaims, error)
}

var store TokenStore

// SetStore 由 main 注入
func SetStore(s TokenStore) {
	store = s
}

// 下方三个函数直接代理到具体实现
func AddToken(jti string, exp int64) error {
	return store.AddToken(jti, exp)
}
func DelToken(jti string) error {
	return store.DelToken(jti)
}
func IsTokenValid(jti string) (bool, error) {
	return store.IsTokenValid(jti)
}

func SetClaims(jti string, claims JwtMapClaims) error {
	return store.SetClaims(jti, claims)
}
func GetClaims(jti string) (JwtMapClaims, error) {
	return store.GetClaims(jti)
}

func validKey(jti string) string  { return "login_tokens:" + jti }
func claimsKey(jti string) string { return "jwt:claims:" + jti }

// Package jwt 负责签发和解析 JWT token。
// 业务代码不直接碰 jwt 库，只调这里的 Generate / Parse。
package jwt

import (
	"errors"
	"strconv"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken 统一表示「token 不可用」，不区分过期/伪造/格式错，
// 避免把细节暴露给攻击者。
var ErrInvalidToken = errors.New("token 无效或已过期")

// Claims 是 token 里携带的信息（payload 部分）。
// 注意：payload 只是 Base64 编码，不是加密，任何人都能解出来看。
// 所以这里只能放 user_id / username / role 这类非敏感信息，
// 绝对不能放密码、手机号、身份证。
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwtlib.RegisteredClaims
}

// Manager 持有密钥和过期时长，由 main 创建一次，全应用共用。
type Manager struct {
	secret []byte
	expire time.Duration
}

func NewManager(secret string, expireHours int) *Manager {
	if expireHours <= 0 {
		expireHours = 24
	}
	return &Manager{
		secret: []byte(secret),
		expire: time.Duration(expireHours) * time.Hour,
	}
}

// Generate 签发 token，同时返回过期时间（供前端判断何时该重新登录）。
//
// JWT 结构 = Header.Payload.Signature，三段用 . 连接。
// Signature 是用 secret 对「前两段」做的 HMAC 签名：
// 任何人改了 payload 里的 user_id，签名就对不上，服务端验签直接失败。
// 这就是「能防篡改」的原理——不是内容被加密了，而是内容动过就验不过。
func (m *Manager) Generate(userID uint64, username, role string) (string, time.Time, error) {
	now := time.Now()
	expireAt := now.Add(m.expire)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "blog",
			Subject:   strconv.FormatUint(userID, 10),
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(expireAt),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expireAt, nil
}

// Parse 校验签名与有效期，返回 payload。
//
// WithValidMethods 这一行是安全关键：
// 不限定算法的话，攻击者可以把 header 里的 alg 改成 "none" 或换成别的算法，
// 伪造出一个「无需密钥即可通过」的 token（alg 混淆攻击）。
func (m *Manager) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwtlib.ParseWithClaims(
		raw,
		claims,
		func(t *jwtlib.Token) (any, error) {
			// 二次确认：必须是 HMAC 族算法，再用我们的 secret 验签
			if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return m.secret, nil
		},
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		jwtlib.WithIssuer("blog"),
	)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

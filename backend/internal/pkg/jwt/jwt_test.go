package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-for-unit-test-only"

func TestGenerateAndParse(t *testing.T) {
	m := NewManager(testSecret, 24)

	token, expireAt, err := m.Generate(7, "alice", "author")
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}
	if token == "" {
		t.Fatal("token 为空")
	}
	if expireAt.Before(time.Now()) {
		t.Error("过期时间应当在未来")
	}

	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if claims.UserID != 7 {
		t.Errorf("UserID 错误: got %d, want 7", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("Username 错误: got %s", claims.Username)
	}
	if claims.Role != "author" {
		t.Errorf("Role 错误: got %s", claims.Role)
	}
	if claims.Issuer != "blog" {
		t.Errorf("Issuer 错误: got %s", claims.Issuer)
	}
}

// 篡改 payload（比如把自己的 user_id 改大）必须验签失败
func TestParseRejectsTamperedToken(t *testing.T) {
	m := NewManager(testSecret, 24)
	token, _, _ := m.Generate(7, "alice", "author")

	// 手动改掉 payload 里的一个字符
	parts := []byte(token)
	dot1 := -1
	dot2 := -1
	for i, b := range parts {
		if b == '.' {
			if dot1 == -1 {
				dot1 = i
			} else if dot2 == -1 {
				dot2 = i
				break
			}
		}
	}
	if dot1 < 0 || dot2 < 0 {
		t.Fatal("token 格式不对")
	}
	parts[dot1+1] ^= 0x01 // 翻转 payload 第一个字符的一个 bit

	if _, err := m.Parse(string(parts)); err == nil {
		t.Error("篡改后的 token 不应通过校验")
	}
}

// 用别的密钥签发的 token 必须被拒绝
func TestParseRejectsWrongSecret(t *testing.T) {
	attacker := NewManager("attacker-secret", 24)
	token, _, _ := attacker.Generate(1, "admin", "admin")

	victim := NewManager(testSecret, 24)
	if _, err := victim.Parse(token); err == nil {
		t.Error("别人密钥签发的 token 不应通过")
	}
}

// 过期 token 必须被拒绝
func TestParseRejectsExpiredToken(t *testing.T) {
	m := NewManager(testSecret, 24)

	past := time.Now().Add(-2 * time.Hour)
	claims := Claims{
		UserID:   1,
		Username: "admin",
		Role:     "admin",
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "blog",
			ExpiresAt: jwtlib.NewNumericDate(past),
			IssuedAt:  jwtlib.NewNumericDate(past.Add(-time.Hour)),
		},
	}
	raw, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		t.Fatalf("构造过期 token 失败: %v", err)
	}

	if _, err := m.Parse(raw); err == nil {
		t.Error("过期 token 不应通过")
	}
}

// alg 混淆攻击：把算法改成 none 的 token 必须被拒绝
func TestParseRejectsNoneAlgorithm(t *testing.T) {
	m := NewManager(testSecret, 24)

	claims := Claims{
		UserID:   1,
		Username: "admin",
		Role:     "admin",
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "blog",
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	raw, err := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, claims).
		SignedString(jwtlib.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("构造 none token 失败: %v", err)
	}

	if _, err := m.Parse(raw); err == nil {
		t.Error("alg=none 的 token 不应通过（alg 混淆攻击）")
	}
}

// 签名有效但不是本系统签发的（Issuer 不对）也要拒绝
func TestParseRejectsWrongIssuer(t *testing.T) {
	m := NewManager(testSecret, 24)

	claims := Claims{
		UserID:   1,
		Username: "admin",
		Role:     "admin",
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "other-system",
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	raw, _ := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(m.secret)

	if _, err := m.Parse(raw); err == nil {
		t.Error("Issuer 不匹配的 token 不应通过")
	}
}

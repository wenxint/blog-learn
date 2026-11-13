package hash

import (
	"strings"
	"testing"
)

func TestPasswordProducesBcryptHash(t *testing.T) {
	h, err := Password("my-secret-123")
	if err != nil {
		t.Fatalf("哈希失败: %v", err)
	}
	// bcrypt 结果固定 60 个字符，以 $2a$ / $2b$ / $2y$ 开头
	if len(h) != 60 {
		t.Errorf("哈希长度应为 60，实际 %d", len(h))
	}
	if !strings.HasPrefix(h, "$2") {
		t.Errorf("应以 $2 开头，实际: %s", h)
	}
	// 绝不能包含明文
	if strings.Contains(h, "my-secret-123") {
		t.Error("哈希结果里出现了明文密码")
	}
}

func TestCheckAcceptsCorrectPassword(t *testing.T) {
	h, _ := Password("correct-horse")
	if !Check(h, "correct-horse") {
		t.Error("正确密码应当验证通过")
	}
}

func TestCheckRejectsWrongPassword(t *testing.T) {
	h, _ := Password("correct-horse")
	if Check(h, "wrong-horse") {
		t.Error("错误密码不应通过")
	}
}

func TestCheckRejectsEmptyHash(t *testing.T) {
	if Check("", "anything") {
		t.Error("空哈希不应通过（防御未设置密码的账号）")
	}
}

// 同一个密码两次哈希结果必须不同（因为每次随机加盐），
// 但 Check 都要能验证通过。这是 bcrypt 与 MD5 的核心区别之一。
func TestSamePasswordProducesDifferentHashes(t *testing.T) {
	h1, _ := Password("same-password")
	h2, _ := Password("same-password")

	if h1 == h2 {
		t.Error("两次哈希结果相同，说明没有随机盐")
	}
	if !Check(h1, "same-password") || !Check(h2, "same-password") {
		t.Error("两个哈希都应能验证同一个密码")
	}
}

// bcrypt 只取前 72 字节，必须显式拦掉超长密码
func TestTooLongPasswordRejected(t *testing.T) {
	long := strings.Repeat("a", MaxPasswordLen+1)
	if _, err := Password(long); err == nil {
		t.Error("超过 72 字节的密码应当报错")
	}
	// 正好 72 字节应当允许
	if _, err := Password(strings.Repeat("a", MaxPasswordLen)); err != nil {
		t.Errorf("72 字节密码应当允许: %v", err)
	}
}

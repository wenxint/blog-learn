// Package hash 封装密码哈希，业务代码不直接碰 bcrypt。
package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcrypt 的成本因子：越大越慢越难爆破。
// 10 在 2026 年的机器上单次约 50~80ms，是安全与体验的平衡点。
// 每 +1 耗时翻倍。
const cost = 10

// bcrypt 只取输入的前 72 字节，超出部分会被静默忽略。
// 所以必须在入口拦住超长密码，否则「密码前 72 字节相同」的两个不同密码可以互相登录。
const MaxPasswordLen = 72

// Password 把明文密码哈希成 $2a$10$... 形式。
//
// 为什么不能存明文：
//  1. 数据库一旦泄漏（拖库、备份文件外流、开发误连生产），所有用户密码全裸
//  2. 用户常在多个网站用同一个密码，泄漏会连累其他系统
//
// 为什么不用 MD5/SHA256：
//
//	它们算得太快，GPU 每秒能猜几十亿次；bcrypt 天然慢，且内置随机盐。
//
// 为什么同一个密码两次哈希结果不同：
//
//	每次生成都会随机出一个 salt 混进结果里，所以哈希值不相等，但 Check 都能验证通过。
func Password(plain string) (string, error) {
	if len(plain) > MaxPasswordLen {
		return "", fmt.Errorf("密码长度不能超过 %d 字节", MaxPasswordLen)
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Check 验证明文密码是否与哈希匹配。
// 内部是「从哈希里取出 salt，用同样成本重新算一遍再比较」，不是明文比较。
func Check(hashed, plain string) bool {
	if hashed == "" || len(plain) > MaxPasswordLen {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

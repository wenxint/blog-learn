package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

// 注意：go test 的工作目录是当前包目录（internal/config），
// 那里没有 .env 文件，所以 Load() 只会读环境变量 + 默认值，测试结果是可复现的。

func TestLoadWithMinimalEnv(t *testing.T) {
	// t.Setenv 会自动在测试结束时还原，且禁止并行测试，避免互相污染
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("JWT_SECRET", "jwt-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("期望加载成功，实际报错: %v", err)
	}

	if cfg.AppPort != "8090" {
		t.Errorf("AppPort 默认值错误: got %q, want %q", cfg.AppPort, "8090")
	}
	if cfg.AppEnv != "dev" {
		t.Errorf("AppEnv 默认值错误: got %q, want %q", cfg.AppEnv, "dev")
	}
	if cfg.DB.Host != "127.0.0.1" || cfg.DB.Port != "3306" {
		t.Errorf("DB 地址默认值错误: got %s:%s", cfg.DB.Host, cfg.DB.Port)
	}
	if cfg.DB.MaxOpenConns != 50 {
		t.Errorf("MaxOpenConns 默认值错误: got %d, want 50", cfg.DB.MaxOpenConns)
	}
	if cfg.DB.ConnMaxLifetime != 30*time.Minute {
		t.Errorf("ConnMaxLifetime 默认值错误: got %v, want 30m", cfg.DB.ConnMaxLifetime)
	}
}

func TestLoadOverridesFromEnv(t *testing.T) {
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("APP_PORT", "9999")
	t.Setenv("DB_MAX_OPEN_CONNS", "7")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("期望加载成功，实际报错: %v", err)
	}
	if cfg.AppPort != "9999" {
		t.Errorf("APP_PORT 覆盖失败: got %q", cfg.AppPort)
	}
	if cfg.DB.MaxOpenConns != 7 {
		t.Errorf("DB_MAX_OPEN_CONNS 覆盖失败: got %d", cfg.DB.MaxOpenConns)
	}
}

func TestLoadRejectsMissingPassword(t *testing.T) {
	// 清空关键变量：os.Unsetenv 后 t.Setenv 兜底还原
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("JWT_SECRET", "jwt-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("缺少 DB_PASSWORD 时应当报错")
	}
	if !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Errorf("错误信息应指明 DB_PASSWORD，实际: %v", err)
	}
}

func TestLoadRejectsMissingJWTSecret(t *testing.T) {
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("JWT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("缺少 JWT_SECRET 时应当报错")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("错误信息应指明 JWT_SECRET，实际: %v", err)
	}
}

func TestDSN(t *testing.T) {
	db := DB{
		Host:     "127.0.0.1",
		Port:     "3306",
		User:     "blog",
		Password: "pwd123",
		Name:     "blog",
		Charset:  "utf8mb4",
	}
	want := "blog:pwd123@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s"
	if got := db.DSN(); got != want {
		t.Errorf("DSN 拼装错误:\n got: %s\nwant: %s", got, want)
	}
}

func TestSafeDSNHidesPassword(t *testing.T) {
	db := DB{
		Host:     "127.0.0.1",
		Port:     "3306",
		User:     "blog",
		Password: "pwd123",
		Name:     "blog",
		Charset:  "utf8mb4",
	}
	safe := db.SafeDSN()
	if strings.Contains(safe, "pwd123") {
		t.Errorf("SafeDSN 泄漏了密码: %s", safe)
	}
	if !strings.Contains(safe, "***") {
		t.Errorf("SafeDSN 应当用 *** 占位: %s", safe)
	}
}

func TestGetEnvIntFallsBackOnInvalid(t *testing.T) {
	os.Setenv("TEST_BAD_INT", "not-a-number")
	defer os.Unsetenv("TEST_BAD_INT")

	if got := getEnvInt("TEST_BAD_INT", 42); got != 42 {
		t.Errorf("非法整数应回退默认值，得到 %d", got)
	}
}

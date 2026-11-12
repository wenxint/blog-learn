package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config 是应用的全部配置。启动时加载一次，之后只读，不做热更新。
type Config struct {
	AppEnv  string // dev / prod
	AppPort string // HTTP 监听端口
	DB      DB
	JWT     JWT
	Upload  Upload
}

// DB 数据库相关配置
type DB struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	Charset         string
	MaxOpenConns    int           // 连接池最大连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 单个连接最长存活时间
}

// JWT 认证配置（第 9 课使用）
type JWT struct {
	Secret      string
	ExpireHours int
}

// Upload 文件上传配置（第 15 课使用）
type Upload struct {
	Dir   string
	MaxMB int
}

// DSN 拼出 go-sql-driver/mysql 需要的连接串
//
//	user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
func (d DB) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.Charset,
	)
}

// SafeDSN 输出到日志用的 DSN，密码替换成 ***
func (d DB) SafeDSN() string {
	return fmt.Sprintf("%s:***@tcp(%s:%s)/%s?charset=%s",
		d.User, d.Host, d.Port, d.Name, d.Charset)
}

// Load 读取 .env（如果存在）并组装 Config。
// 已存在的操作系统环境变量优先级高于 .env，方便临时覆盖：
//
//	DB_PASSWORD=wrong go run ./cmd/server
func Load() (*Config, error) {
	// .env 不存在不算错误：生产环境通常直接用系统环境变量
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			return nil, fmt.Errorf("加载 .env 失败: %w", err)
		}
	}

	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "dev"),
		AppPort: getEnv("APP_PORT", "8090"),
		DB: DB{
			Host:            getEnv("DB_HOST", "127.0.0.1"),
			Port:            getEnv("DB_PORT", "3306"),
			User:            getEnv("DB_USER", "blog"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "blog"),
			Charset:         getEnv("DB_CHARSET", "utf8mb4"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 30)) * time.Minute,
		},
		JWT: JWT{
			Secret:      getEnv("JWT_SECRET", ""),
			ExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		},
		Upload: Upload{
			Dir:   getEnv("UPLOAD_DIR", "./uploads"),
			MaxMB: getEnvInt("UPLOAD_MAX_MB", 5),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate 校验必填项：缺配置要在启动时炸掉，而不是等第一次查询才报错
func (c *Config) validate() error {
	if c.DB.Password == "" {
		return fmt.Errorf("缺少 DB_PASSWORD，请在 .env 或环境变量中配置")
	}
	if c.DB.Name == "" {
		return fmt.Errorf("缺少 DB_NAME")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("缺少 JWT_SECRET，可执行 openssl rand -hex 32 生成")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

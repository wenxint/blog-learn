// Package config 负责把「运行配置」读进内存。
// 配置来源有两处，优先级从高到低：
//  1. 操作系统环境变量（部署时由系统或容器注入）
//  2. 当前工作目录下的 .env 文件（只在本机开发用）
//
// 不读 .env.example。那个文件是提交到 Git 的模板，告诉别人要配哪些项；
// 真正的密码写在 .env 里，.env 不进 Git。
package config

import (
	"fmt"     // 拼字符串、包装错误
	"os"      // 读文件是否存在、读环境变量
	"strconv" // 把字符串 "50" 转成数字 50
	"strings" // 拆分逗号分隔的配置
	"time"    // 分钟等时间类型

	// 第三方库。godotenv 会把 .env 里的 KEY=VALUE 写进本进程的环境变量。
	"github.com/joho/godotenv"
)

// Config 是应用的全部配置。启动时加载一次，之后只读，不做热更新。
// type ... struct 是在定义一种新的数据类型，花括号里每一行是一个字段。
type Config struct {
	AppEnv  string // dev / prod，来自 APP_ENV
	AppPort string // HTTP 监听端口，来自 APP_PORT，例如 "8090"
	DB      DB     // 数据库相关，嵌套的另一份结构体
	JWT     JWT    // 登录令牌相关（第 9 课才用到）
	Upload  Upload // 图片上传相关（第 15 课才用到）
	Limit   Limit  // 限流相关（第 14 课补充）

	// TrustedProxies 可信代理地址。为空表示「谁都不信」，
	// ClientIP() 直接用 TCP 对端 IP。
	// 部署在 Nginx 后面时填 127.0.0.1，才会采信 X-Forwarded-For。
	TrustedProxies []string
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
	Secret      string // 签发 token 的密钥，泄露后别人就能伪造登录
	ExpireHours int    // token 多少小时后过期
}

// Upload 文件上传配置（第 15 课使用）
type Upload struct {
	Dir   string // 图片存到哪个目录
	MaxMB int    // 单文件最大几 MB
}

// Limit 限流配置。
// 目前只用于评论接口：同一个 IP 在时间窗口内最多发几条。
type Limit struct {
	CommentCount  int           // 窗口内允许的评论条数，默认 3
	CommentWindow time.Duration // 窗口长度，默认 60 秒
}

// DSN 拼出 go-sql-driver/mysql 需要的连接串。
// (d DB) 表示这是 DB 的方法：调用时写成 cfg.DB.DSN()，d 就是那一份 DB。
//
// 拼出来的样子：
//
//	user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
func (d DB) DSN() string {
	// %s 会被后面的参数按顺序替换。密码在这里是明文，只交给数据库驱动，不要打到日志里。
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

// Load 读取配置文件（默认 .env）并组装 Config。
// 返回值 (*Config, error) 是一对：成功时第一份是配置指针、第二份是 nil；
// 失败时第一份是 nil、第二份是错误。调用方必须看第二份。
//
// 配置文件路径默认是当前工作目录的 .env，可用 ENV_FILE 指定别的文件：
//
//	ENV_FILE=.env.prod ./blog-server          # 本地用生产配置试跑
//	ENV_FILE=/etc/blog/blog.env ./blog-server # 配置与代码目录分离
//
// 已存在的操作系统环境变量优先级高于配置文件，方便临时覆盖：
//
//	DB_PASSWORD=wrong go run ./cmd/server
func Load() (*Config, error) {
	// 先看有没有显式指定 ENV_FILE。区分「没设」和「设了」很重要：
	// 没设时默认 .env，文件不存在就静默跳过（生产常直接注入环境变量）；
	// 显式设了却找不到文件，必须报错——否则会误以为加载了生产配置，
	// 实际却用了默认值，这类错误上线后极难发现。
	envFile, explicit := os.LookupEnv("ENV_FILE")
	if !explicit {
		envFile = ".env"
	}

	// os.Stat 问操作系统这个路径在不在。
	// 它返回两个值：文件信息，以及错误。前面的 _ 表示文件信息用不到，丢掉。
	// err == nil 才表示「文件存在并且能访问」。
	if _, err := os.Stat(envFile); err == nil {
		// 文件在，才真正解析。Load 会把 APP_PORT=8090 这类行写进环境变量。
		// 解析失败（文件坏了、没权限）才算启动失败：return nil, 错误。
		// fmt.Errorf 的 %w 会把原始错误包进去，外面还能用 errors.Is 认出来。
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("加载 %s 失败: %w", envFile, err)
		}
	} else if explicit {
		return nil, fmt.Errorf("ENV_FILE 指定的配置文件不存在: %s", envFile)
	}

	// &Config{...} 在堆上造一份 Config，得到它的指针。
	// 每个字段都经 getEnv 读取：环境变量有值就用它，没有就用第二个参数（默认值）。
	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "dev"),
		AppPort: getEnv("APP_PORT", "8090"),
		DB: DB{
			Host:            getEnv("DB_HOST", "127.0.0.1"),
			Port:            getEnv("DB_PORT", "3306"),
			User:            getEnv("DB_USER", "blog"),
			Password:        getEnv("DB_PASSWORD", ""), // 密码没有安全的默认值，留空，交给下面的 validate 拒绝启动
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
		Limit: Limit{
			CommentCount:  getEnvInt("COMMENT_RATE_LIMIT", 3),
			CommentWindow: time.Duration(getEnvInt("COMMENT_RATE_WINDOW_SEC", 60)) * time.Second,
		},
		TrustedProxies: getEnvList("TRUSTED_PROXIES"),
	}

	// 缺必填项就不要返回这份配置。main 看到错误会退出，避免带着空密码跑起来。
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate 校验必填项：缺配置要在启动时炸掉，而不是等第一次查询才报错。
// (c *Config) 表示这是 Config 指针的方法，c 就是刚组装好的那份配置。
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

// getEnv 读一个字符串环境变量。
// LookupEnv 也返回两个值：变量的内容，以及「这个变量到底有没有被设置过」。
// 没设置，或设置成空字符串，都退回 fallback。
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// getEnvInt 读一个整数环境变量。内容不是合法整数时，不报错，退回 fallback。
// 例如 DB_MAX_OPEN_CONNS=abc 时，会安静地用默认值 50，而不是在启动阶段退出。
// getEnvList 把逗号分隔的环境变量切成切片，例如 TRUSTED_PROXIES=127.0.0.1,::1
func getEnvList(key string) []string {
	v := getEnv(key, "")
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v) // Atoi = ASCII to integer
	if err != nil {
		return fallback
	}
	return n
}

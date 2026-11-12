// Package database 负责连上 MySQL，并把连接池参数设好。
// 业务代码拿到的是 *gorm.DB：之后的查询都从这份对象发出去，不要在别处再 Open 一次。
package database

import (
	"fmt"
	"log"
	"time"

	"blog/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect 建立数据库连接，配置连接池，并做一次 Ping 确认真的连通。
// cfg 是 config.Load 交出来的那份配置，这里只用里面的 DB 和 AppEnv。
// 连不上直接返回 error，由 main 决定是否退出——不要在这里 os.Exit，
// 否则这个函数没法被测试，调用方也没机会打自己的日志。
func Connect(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		// dev 环境打印每条 SQL，方便学习时看到 GORM 实际发了什么
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		// 时间统一用本地时区
		NowFunc: func() time.Time { return time.Now().Local() },
	}
	if cfg.AppEnv == "dev" {
		gormCfg.Logger = gormlogger.Default.LogMode(gormlogger.Info)
	}

	// gorm.Open 用 DSN 建一个 GORM 会话。这里失败通常是 DSN 格式不对，
	// 并不代表网络已经连上——真正握手在下面的 Ping。
	db, err := gorm.Open(mysql.Open(cfg.DB.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 拿到底层 *sql.DB，才能设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层 sql.DB 失败: %w", err)
	}

	// 连接池三件套
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)       // 最多同时开多少连接
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)       // 空闲时保留多少连接
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime) // 单个连接最长活多久，到点回收

	// 真正握手一次。gorm.Open 是惰性的，不 Ping 可能到第一次查询才发现连不上
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败(%s): %w", cfg.DB.SafeDSN(), err)
	}

	log.Printf("[DB] 连接成功 %s | 最大连接=%d 空闲=%d 存活=%s",
		cfg.DB.SafeDSN(), cfg.DB.MaxOpenConns, cfg.DB.MaxIdleConns, cfg.DB.ConnMaxLifetime)

	return db, nil
}

// Close 关掉连接池里所有连接。main 在 HTTP 服务排空请求之后才调用它。
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

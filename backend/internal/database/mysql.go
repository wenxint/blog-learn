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
// 连不上直接返回 error，由 main 决定是否退出——不要在这里 os.Exit。
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

// Close 优雅关闭连接池
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

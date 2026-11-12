package main

import (
	"log"

	"blog/internal/config"
	"blog/internal/database"
	"blog/internal/router"
)

func main() {
	// 1. 加载配置（.env + 环境变量）
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 连接数据库（含连接池与 Ping 验证）
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer func() { _ = database.Close(db) }()

	// 3. 组装路由（把 db 注入进去，不用全局变量）
	r := router.New(db)

	addr := ":" + cfg.AppPort
	log.Printf("环境=%s 服务启动: http://127.0.0.1%s", cfg.AppEnv, addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

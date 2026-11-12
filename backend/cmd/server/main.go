package main

import (
	"log"

	"blog/internal/router"
)

func main() {
	r := router.New()

	addr := ":8090"
	log.Printf("服务启动: http://127.0.0.1%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// 3. 组装 HTTP Server
	// 不直接用 r.Run()：它内部固定了一个没有超时的 http.Server，
	// 也没法优雅关闭。这里自己控制。
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router.New(db),
		// 防止慢连接耗尽资源（Slowloris 攻击）
		ReadHeaderTimeout: 5 * time.Second,  // 读完请求头的最长时间
		ReadTimeout:       15 * time.Second, // 读完整个请求（含 body）的最长时间
		WriteTimeout:      30 * time.Second, // 写响应最长时间
		IdleTimeout:       60 * time.Second, // keep-alive 空闲连接保留时间
	}

	// 4. 监听退出信号：Ctrl+C(SIGINT) 或 docker stop / k8s(K8S 发 SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 5. 在后台启动服务（ListenAndServe 会阻塞，不能挡住信号监听）
	go func() {
		log.Printf("环境=%s 服务启动: http://127.0.0.1%s", cfg.AppEnv, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 6. 阻塞等待退出信号
	<-ctx.Done()
	stop()
	log.Println("收到退出信号，开始优雅关闭...")

	// 7. 优雅关闭：不再接收新请求，等待进行中的请求完成（最多 10 秒）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP 关闭未在超时内完成: %v", err)
	} else {
		log.Println("HTTP 服务已关闭（进行中的请求已排空）")
	}

	// 8. 关闭数据库连接池
	if err := database.Close(db); err != nil {
		log.Printf("关闭数据库连接池失败: %v", err)
	} else {
		log.Println("数据库连接池已关闭")
	}

	log.Println("进程退出")
}

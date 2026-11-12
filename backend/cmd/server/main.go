// Command server 是博客后端的进程入口。
// 运行 `go run ./cmd/server` 或 `make run` 时，操作系统从这里的 main 函数开始执行。
// 顺序固定四步：读配置 → 连数据库 → 监听 HTTP → 收到退出信号后把进行中的请求做完再退出。
package main

import (
	// 双引号里不带域名的，是 Go 自带的标准库。
	"context"   // 传递「取消、超时」：谁拿到这个 ctx，谁就能知道该停了
	"errors"    // 判断某个错误是不是某一种已知错误
	"log"       // 往终端打日志
	"net/http"  // HTTP 服务器
	"os"        // 操作系统信号，这里用来接 Ctrl+C
	"os/signal" // 把系统信号转成 Go 能等待的 context
	"syscall"   // SIGTERM：docker stop / k8s 停止容器时发的信号
	"time"      // 秒、分钟这类时间单位

	// 带 "blog/..." 的是本项目自己的包。blog 来自 go.mod 第一行的 module 名。
	// internal/config 就是文件夹 backend/internal/config。
	// 那个文件夹里写着 package config，所以下面才能写 config.Load。
	// config 是包名，不是变量。
	"blog/internal/config"
	"blog/internal/database"
	"blog/internal/router"
)

// main 是整个进程的起点。没有参数，也没有返回值；失败时用 log.Fatalf 直接退出。
func main() {
	// 1. 加载配置（.env + 环境变量）
	// := 是「声明变量并赋值」，左边的 cfg、err 是这里新造出来的变量。
	// Load 一次返回两个结果：第一份是配置，第二份是错误。成功时 err 为 nil（空）。
	cfg, err := config.Load()
	// Go 的约定：出错不抛异常，用返回值告诉你。err != nil 就是失败了。
	// Fatalf 打印这句话后立刻结束进程，下面的代码不会再执行。%v 会被换成 err 的内容。
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 连接数据库（含连接池与 Ping 验证）
	// cfg 传进去，是因为地址、账号、密码都在刚才读到的配置里。
	// 这里再次使用 :=，但 err 已经存在，Go 允许「至少有一个新变量」时复用 :=。
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// 3. 组装 HTTP Server
	// 不直接用 r.Run()：它内部固定了一个没有超时的 http.Server，
	// 也没法优雅关闭。这里自己控制。
	// & 是取地址，得到指针。http.Server 是标准库里的结构体，下面按字段名填值。
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort, // AppPort 来自 .env 的 APP_PORT，例如 "8090"，拼成监听地址 ":8090"
		Handler: router.New(db),    // 真正按网址分发请求的是 router 组装好的 Gin
		// 防止慢连接耗尽资源（Slowloris 攻击）
		ReadHeaderTimeout: 5 * time.Second,  // 读完请求头的最长时间
		ReadTimeout:       15 * time.Second, // 读完整个请求（含 body）的最长时间
		WriteTimeout:      30 * time.Second, // 写响应最长时间
		IdleTimeout:       60 * time.Second, // keep-alive 空闲连接保留时间
	}

	// 4. 监听退出信号：Ctrl+C(SIGINT) 或 docker stop / k8s(K8S 发 SIGTERM)
	// NotifyContext 返回两个东西：ctx 会在收到信号时被取消；stop 用来提前解除监听。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer 把这句话推迟到 main 函数结束时再执行，用来释放信号监听，避免泄漏。
	defer stop()

	// 5. 在后台启动服务（ListenAndServe 会阻塞，不能挡住信号监听）
	// go 关键字：另开一条协程跑这个匿名函数，main 自己继续往下走。
	go func() {
		log.Printf("环境=%s 服务启动: http://127.0.0.1%s", cfg.AppEnv, srv.Addr)
		// ListenAndServe 正常时一直堵住，直到服务被关掉。
		// 优雅关闭时它返回的是 http.ErrServerClosed，这是预期结果，不能当成启动失败。
		// errors.Is 用来比较「是不是这一种错误」（错误可能被包了一层，不能直接用 !=）。
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 6. 阻塞等待退出信号
	// <-ctx.Done() 会卡在这里，直到第 4 步那个 ctx 被取消（也就是你按了 Ctrl+C）。
	<-ctx.Done()
	stop()
	log.Println("收到退出信号，开始优雅关闭...")

	// 7. 优雅关闭：不再接收新请求，等待进行中的请求完成（最多 10 秒）
	// WithTimeout 造一个 10 秒后自动取消的 ctx。cancel 必须调用，所以用 defer。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown 和「直接掐断」不同：已接进来的请求给它们时间写完响应。
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP 关闭未在超时内完成: %v", err)
	} else {
		log.Println("HTTP 服务已关闭（进行中的请求已排空）")
	}

	// 8. 关闭数据库连接池
	// 放在 HTTP 关掉之后：进行中的请求可能还要查库，先断库会让那些请求失败。
	if err := database.Close(db); err != nil {
		log.Printf("关闭数据库连接池失败: %v", err)
	} else {
		log.Println("数据库连接池已关闭")
	}

	log.Println("进程退出")
}

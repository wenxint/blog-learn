# 博客后端常用命令（在 backend/ 目录下执行）

APP_NAME := blog-server
BIN_DIR  := bin

# 数据库连接（会交互式提示输入密码，不把密码写进命令行历史）
MYSQL_USER ?= blog
MYSQL_DB   ?= blog
MYSQL      := mysql -u$(MYSQL_USER) -p $(MYSQL_DB)

.PHONY: help run build vet tidy fmt clean migrate seed db

help: ## 显示所有命令
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

run: ## 启动开发服务（默认 :8090，Ctrl+C 停止）
	go run ./cmd/server

build: ## 编译成单个可执行文件到 bin/
	@mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags "-s -w" -o $(BIN_DIR)/$(APP_NAME) ./cmd/server
	@echo "编译完成: $(BIN_DIR)/$(APP_NAME)"

vet: ## 静态检查
	go vet ./...

tidy: ## 整理依赖
	go mod tidy

fmt: ## 格式化代码
	gofmt -l -w .

clean: ## 清理编译产物
	rm -rf $(BIN_DIR)

migrate: ## 建表（执行 migrations/001_init.sql）
	$(MYSQL) < migrations/001_init.sql
	@echo "建表完成"

seed: ## 灌种子数据（执行 migrations/002_seed.sql）
	$(MYSQL) < migrations/002_seed.sql
	@echo "种子数据完成"

db: ## 进入 MySQL 交互终端
	$(MYSQL)

SHELL := /bin/sh

# 应用名称以及本地编译产物目录。
APP_NAME := cash
BIN_DIR := bin

# Docker Compose 文件。
COMPOSE := docker compose -f deployments/docker-compose.yml

.PHONY: help run build test test-race coverage fmt vet lint tidy docker-up docker-down

help: ## 显示全部可用命令及用途
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## 在本机启动 API 服务（存在 .env 时自动加载）
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; go run ./cmd/cash-core

build: ## 将 API 编译到 bin/cash
	mkdir -p $(BIN_DIR)
	go build -trimpath -o $(BIN_DIR)/$(APP_NAME) ./cmd/cash-core

test: ## 运行全部单元测试
	go test ./...

test-race: ## 使用竞态检测器运行全部测试
	go test -race ./...

coverage: ## 生成 coverage.html 测试覆盖率报告
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fmt: ## 使用 gofmt 格式化所有 Go 代码
	gofmt -w .

vet: ## 使用 go vet 执行静态检查
	go vet ./...

lint: ## 使用 golangci-lint 检查代码（需要先安装）
	golangci-lint run

tidy: ## 整理 go.mod 和 go.sum 依赖
	go mod tidy

docker-down: ## 停止 Docker 服务但保留 SQLite 数据卷
	$(COMPOSE) down

docker-up: ## 构建并在容器中启动 API
	$(COMPOSE) up -d --build api

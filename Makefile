SHELL := /bin/sh

# 应用名称以及本地编译产物目录。
APP_NAME := cash
BIN_DIR := bin

# Docker Compose 文件以及创建迁移文件时使用的宿主机用户。
COMPOSE := docker compose -f deployments/docker-compose.yml
LOCAL_USER := $(shell id -u):$(shell id -g)

# migrate 在 Compose 网络内通过服务名 postgres 连接数据库。
# 私有部署可通过环境变量覆盖这些值，或直接覆盖 MIGRATION_DATABASE_URL。
DB_NAME ?= cash
DB_USER ?= cash
DB_PASSWORD ?= cash
DB_SSLMODE ?= disable
MIGRATION_DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: help run build test test-race coverage fmt vet lint tidy db-up db-down docker-up migrate-up migrate-down migrate-create

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

db-up: ## 使用 Docker 启动 PostgreSQL
	$(COMPOSE) up -d postgres

db-down: ## 停止 Docker 服务但保留数据库 volume
	$(COMPOSE) down

docker-up: db-up migrate-up ## 构建并在容器中启动 API
	$(COMPOSE) --profile app up -d --build api

migrate-up: ## 使用 Docker 执行所有尚未应用的数据库迁移
	$(COMPOSE) run --rm migrate -path /migrations -database '$(MIGRATION_DATABASE_URL)' up

migrate-down: ## 使用 Docker 回退最近一个数据库迁移版本
	$(COMPOSE) run --rm migrate -path /migrations -database '$(MIGRATION_DATABASE_URL)' down 1

migrate-create: ## 使用 Docker 创建新迁移，例如：make migrate-create name=add_budgets
	@test -n "$(name)" || (echo "必须通过 name 指定迁移名称"; exit 1)
	$(COMPOSE) run --rm --no-deps --user '$(LOCAL_USER)' migrate create -ext sql -dir /migrations -seq $(name)

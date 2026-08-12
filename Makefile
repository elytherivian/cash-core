SHELL := /bin/sh

# 应用名称以及本地编译产物目录。
APP_NAME := cash
BIN_DIR := bin

# Docker Compose 文件。生产命令从 Docker Hub 拉取镜像；开发命令额外叠加源码构建配置。
COMPOSE := docker compose --env-file .env -f deployments/docker-compose.yml
DEV_COMPOSE := docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.dev.yml

# Docker Hub 镜像。发布时通过命令行或环境变量设置真实仓库名。
IMAGE_REPOSITORY ?= your-dockerhub-user/cash-core
IMAGE_TAG ?= dev
PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: help run build test test-race coverage fmt vet lint tidy docker-build docker-push docker-up docker-build-up docker-down deploy

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

docker-build: ## 将当前源码构建为本地 dev 镜像
	docker build --tag $(IMAGE_REPOSITORY):$(IMAGE_TAG) .

docker-push: ## 多架构构建当前源码并推送到 Docker Hub（默认 dev 标签）
	docker buildx build --platform $(PLATFORMS) --tag $(IMAGE_REPOSITORY):$(IMAGE_TAG) --push .

docker-up: ## 从 Docker Hub 拉取镜像并启动生产 API
	$(COMPOSE) pull api
	$(COMPOSE) up -d api

docker-build-up: ## 使用当前源码构建并启动本地 API
	APP_IMAGE=$(IMAGE_REPOSITORY):$(IMAGE_TAG) $(DEV_COMPOSE) up -d --build api

docker-down: ## 停止 API；不会删除宿主机 SQLite 数据目录
	$(COMPOSE) down

deploy: docker-up ## VPS 从 Docker Hub 拉取镜像并更新 API（不在 VPS 构建）

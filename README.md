# cash-core

cash-core 是一个使用 Gin、GORM 和 PostgreSQL 构建的日常记账流水后端。项目按照业务模块纵向组织，每个模块独立拥有 HTTP、业务、数据访问和模型代码，不允许业务模块之间直接依赖。

## 目录结构

```text
cash-core/
├── cmd/cash-core/main.go          # 程序入口、依赖初始化、优雅退出
├── internal/
│   ├── app/                       # 业务模块，模块之间不互相 import
│   │   ├── user/
│   │   ├── account/
│   │   ├── category/
│   │   └── transaction/
│   │       ├── handler.go         # Gin 路由与 HTTP 参数处理
│   │       ├── service.go         # 业务规则与流程编排
│   │       ├── repository.go      # GORM 数据访问
│   │       └── model.go           # 数据模型与请求结构
│   ├── common/                    # 统一响应、错误、生命周期、分页
│   ├── pkg/                       # 跨模块共享的基础组件
│   │   ├── database/              # GORM/PostgreSQL 初始化
│   │   ├── logger/                # slog 日志初始化
│   │   ├── middleware/            # 请求 ID、日志、恢复、CORS 等
│   │   └── utils/                 # JSON、UUID、分页工具
│   └── router/                    # 统一路由和模块依赖组装
├── migrations/                    # golang-migrate SQL 文件
├── deployments/                   # Docker Compose 部署文件
├── scripts/                       # 迁移等开发脚本
└── test/integration/              # 跨模块 HTTP 集成测试
```

## 模块依赖规则

- `internal/app/user`、`account`、`category`、`transaction` 之间禁止互相导入。
- 各业务模块只能依赖标准库、第三方库、`internal/common` 和 `internal/pkg`。
- `internal/router` 是统一组装入口，可以导入所有业务模块。
- 跨模块数据一致性依靠 service 编排、数据库事务或数据库约束完成。
- 数据库 schema 只通过 `migrations` 变更，不使用 GORM `AutoMigrate`。

当前 migration 的复合外键已经保证流水、账户、分类属于同一个用户，并保证流水类型和分类类型一致。
`internal/pkg/database.WithinTransaction` 可用于需要同时写入多个仓储的用例；`internal/pkg/middleware.Authentication` 提供了与 JWT/Session 实现无关的认证扩展接口，接入具体 TokenVerifier 后再挂到受保护路由组。

## 配置优先级

配置加载顺序为：代码默认值 → 环境变量，后加载的值覆盖前面的值。

```bash
cp .env.example .env
```

`make run` 会自动加载根目录的 `.env`；直接执行 `go run` 时需要自行导出环境变量。生产环境应使用环境变量或密钥管理系统覆盖数据库密码。

## 本地启动

需要安装 Go 1.25 和 Docker。数据库迁移通过 Compose 中的 `migrate/migrate` 容器执行，不要求宿主机安装 `golang-migrate`。

```bash
make db-up
make migrate-up
make run
```

第一次执行 `make migrate-up` 时 Docker 会自动拉取 migration 工具镜像。迁移容器通过 Compose 网络中的 `postgres` 服务名连接数据库，所以这里不能使用 `localhost`。

默认 PostgreSQL 配置：

- 数据库：`cash`
- 用户：`cash`
- 密码：`cash`
- 地址：`localhost:5432`

这些默认值只能用于本地开发。

## API

系统端点：

```text
GET /health/live
GET /health/ready
```

业务端点：

```text
POST   /api/v1/users
GET    /api/v1/users/:user_id
DELETE /api/v1/users/:user_id

POST   /api/v1/users/:user_id/accounts
GET    /api/v1/users/:user_id/accounts
GET    /api/v1/users/:user_id/accounts/:id
DELETE /api/v1/users/:user_id/accounts/:id

POST   /api/v1/users/:user_id/categories
GET    /api/v1/users/:user_id/categories?type=expense
GET    /api/v1/users/:user_id/categories/:id
DELETE /api/v1/users/:user_id/categories/:id

POST   /api/v1/users/:user_id/transactions
GET    /api/v1/users/:user_id/transactions?from=2026-08-01T00:00:00Z&to=2026-09-01T00:00:00Z
GET    /api/v1/users/:user_id/transactions/:id
DELETE /api/v1/users/:user_id/transactions/:id
```

所有响应使用统一结构：

```json
{
  "version": "dev",
  "code": 0,
  "message": "ok",
  "data": {}
}
```

HTTP 状态码与响应体中的 `code` 相互独立：HTTP 状态码遵循 RESTful 语义，`code` 是应用业务码。成功响应使用 `0`；例如重复注册用户时 HTTP 返回 `409 Conflict`，响应体中的业务码返回 `40001`。没有数据的响应会省略 `data` 字段。业务码统一定义在 `internal/common/code.go`。

## Makefile

执行 `make help` 可以查看全部中文说明。常用命令：

```bash
make run              # 启动 API
make build            # 编译 bin/cash
make test             # 单元测试
make test-race        # 竞态检测
make coverage         # 覆盖率报告
make vet              # 静态检查
make db-up             # 启动 PostgreSQL
make migrate-up       # 执行数据库迁移
make migrate-down     # 回退一个迁移版本
```

## 开发新模块

新增例如 `budget` 模块时，在 `internal/app/budget` 中建立 `model.go`、`repository.go`、`service.go`、`handler.go`，然后只在 `internal/router/router.go` 中完成构造和路由注册。不要从已有的 account 或 transaction 模块直接调用代码；确有跨模块流程时，应在单独的编排 service 中通过接口完成。

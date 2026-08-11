# cash-core

cash-core 是一个使用 Gin、GORM 和 SQLite 构建的日常记账流水后端。项目按照业务模块纵向组织，每个模块独立拥有 HTTP、业务、数据访问和模型代码，不允许业务模块之间直接依赖。

完整的接口说明、认证方式和调用示例见 [API 文档](docs/API.md)。

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
│   ├── common/                    # 统一响应、错误和生命周期
│   ├── pkg/                       # 跨模块共享的基础组件
│   │   ├── database/              # GORM/SQLite 初始化
│   │   ├── logger/                # slog 日志初始化
│   │   ├── middleware/            # 请求 ID、日志、恢复、CORS 等
│   │   └── utils/                 # JSON、UUID 等 HTTP 工具
│   └── router/                    # 统一路由和模块依赖组装
├── deployments/                   # Docker Compose 部署文件
└── test/integration/              # 跨模块 HTTP 集成测试
```

## 模块依赖规则

- `internal/app/user`、`account`、`category`、`transaction` 之间禁止互相导入。
- 各业务模块只能依赖标准库、第三方库、`internal/common` 和 `internal/pkg`。
- `internal/router` 是统一组装入口，可以导入所有业务模块。
- 跨模块数据一致性依靠 service 编排、数据库事务或数据库约束完成。
- SQLite schema 会在应用启动时自动初始化；初始化是幂等的，不需要单独执行迁移命令。

SQLite 的复合外键保证流水、账户和分类属于同一个用户。
`internal/pkg/middleware.Authentication` 提供了与 JWT/Session 实现无关的认证扩展接口，接入具体 TokenVerifier 后再挂到受保护路由组。

## 配置优先级

配置加载顺序为：代码默认值 → 环境变量，后加载的值覆盖前面的值。

```bash
cp .env.example .env
```

`make run` 会自动加载根目录的 `.env`；直接执行 `go run` 时需要自行导出环境变量。`DB_PATH` 指定 SQLite 文件路径，父目录不存在时应用会自动创建。

`API_TIMEZONE` 控制 API 响应中时间字段的展示时区（默认 `Asia/Shanghai`，使用 IANA 时区名）。数据库仍统一以 UTC 存储，避免因部署地区改变而影响数据语义。

## 本地启动

需要安装 Go 1.25。SQLite 文件和 schema 会在应用首次启动时自动创建。

```bash
make run
```

也可以使用容器完整启动本项目：

```bash
make docker-up
```

该命令会构建并启动 API；SQLite 数据文件保存在 Docker 命名卷中。API 默认监听 `http://localhost:8080`，不再运行独立的数据库服务。

默认 SQLite 配置：

- 本机运行：`data/cash.db`
- Docker 运行：Docker 命名卷 `sqlite_data` 中的 `/data/cash.db`

SQLite 适合单机、低并发的 VPS 部署；同一数据库文件不应放在网络文件系统上，也不应同时由多个 API 容器写入。

## 私有部署

Compose 配置可通过环境变量覆盖应用配置，例如 `APP_ENV`、`JWT_SECRET` 与 `API_HOST_PORT`。SQLite 连接数固定为 1，避免单文件数据库的并发写入锁竞争。生产环境至少应设置：

```bash
export APP_ENV=production
export JWT_SECRET="$(openssl rand -base64 48)"
make docker-up
```

SQLite 不使用数据库账号、密码或端口。不要提交 `.env`、私钥、`data/` 或 `deployments/docker-compose.override.yml`；这些文件已被 Git 忽略。私有部署不依赖 GitHub Actions，仓库中的 CI 工作流已移除。

### VPS + Cloudflare + Caddy

仓库已经提供 Caddy 反向代理配置。Caddy 是唯一对公网暴露的容器：它占用 VPS 的 `80/tcp`、`443/tcp` 和 `443/udp`，API 仅绑定宿主机回环地址；SQLite 数据保存在 Docker 命名卷中。

假设 Cloudflare DNS 中的橙云 A 记录为 `dash-rn.elytherivian.top -> <VPS IPv4>`，在 VPS 上执行：

```bash
git clone <你的仓库地址> cash-core
cd cash-core
cp deployments/.env.production.example .env
chmod 600 .env
# 编辑 .env：填写 ACME_EMAIL，并替换 JWT_SECRET。
openssl rand -base64 48

# 构建并启动 API 与 Caddy；SQLite 会自动初始化。
set -a && . ./.env && set +a
docker compose --env-file .env -f deployments/docker-compose.yml --profile app --profile proxy up -d --build
```

Cloudflare 的 SSL/TLS 模式应改为 **Full (strict)**。Caddy 会通过 HTTP-01 自动取得受信任的源站证书；Cloudflare 的橙云代理不会阻止该验证。VPS 防火墙或云厂商安全组必须放行 `80/tcp` 与 `443/tcp`；不要公开 `8080`。首次完成后，用以下命令验证：

```bash
curl -i https://dash-rn.elytherivian.top/health/live
curl -i https://dash-rn.elytherivian.top/health/ready
docker compose --env-file .env -f deployments/docker-compose.yml logs -f caddy api
```

本地 App 的 API 基地址设置为 `https://dash-rn.elytherivian.top`，接口路径保持原样，例如 `https://dash-rn.elytherivian.top/api/v1/auth/login`。桌面端和移动端直接使用 HTTPS，不需要加入端口号，也不需要配置 CORS；若后续增加 Web 前端，再将其完整 HTTPS 域名加入 `HTTP_ALLOWED_ORIGINS`。

## API

系统端点：

```text
GET /health/live
GET /health/ready
```

业务端点：

```text
POST   /api/v1/users/register
DELETE /api/v1/users/delete
POST   /api/v1/users/restore

POST   /api/v1/auth/login
POST   /api/v1/auth/refresh

POST   /api/v1/accounts/create
GET    /api/v1/accounts/list

POST   /api/v1/categories/create
GET    /api/v1/categories/list

POST   /api/v1/transactions/create
GET    /api/v1/transactions/getByAccountID?account_id=<account_id>
GET    /api/v1/transactions/getByCategoryID?category_id=<category_id>
GET    /api/v1/transactions/getByAccountIDAndCategoryID?account_id=<account_id>&category_id=<category_id>
```

注册、删除、恢复、登录和刷新 token 不要求 JWT。账户、分类和流水接口必须携带：

```http
Authorization: Bearer <access_token>
```

用户身份只从验证通过的 JWT 中解析，客户端不再通过 URL 或请求体提交 `user_id`。

分类表示具体的收支项目，例如 `日用品`、`工资`。创建分类时传入：

```json
{
  "category_name": "日用品"
}
```

这里之所以不使用 `category_type` 代替 `category_name`，是因为这个类别应该是人为定义，
而不是硬编码到代码中，而满足不同用户的需求

流水的 `type` 才表示该笔流水是 `income` 还是 `expense`；分类本身不再区分收入或支出。

创建流水时必须同时指定账户和分类，例如“从微信 wechat1 支出 20 元购买零食”：

```json
{
  "type": "expense",
  "amount": "20",
  "account_id": "<wechat1 的账户 ID>",
  "category_id": "<零食的分类 ID>"
}
```

`occurred_at` 可选；不传时服务端使用当前 UTC 时间。传入时可用于补记过去实际发生的流水。

三个流水查询接口会返回全部匹配的有效流水，并按 `occurred_at` 正序排列，供客户端计算总收入、总支出和账户余额。

创建账户时，`account_type` 表示账户渠道或银行类型，目前仅支持 `WeChat`、`AliPay`、`BOC`；`account_name` 用于区分同一类型下的多个账户：

```json
{
  "account_type": "WeChat",
  "account_name": "wechat1",
  "initial_balance": "0"
}
```

同一用户可以创建 `wechat1`、`wechat2` 等多个 `WeChat` 账户；有效账户的 `(account_type, account_name)` 组合必须唯一。

登录请求：

```json
{
  "username": "cash-user",
  "password": "password123"
}
```

登录和刷新成功后返回 `user_id`、15 分钟有效的 `access_token`，以及默认 30 天有效的 `refresh_token`。刷新请求：

```json
{
  "refresh_token": "<refresh_token>"
}
```

JWT 参数由 `JWT_ISSUER`、`JWT_SECRET`、`JWT_ACCESS_TOKEN_TTL` 和 `JWT_REFRESH_TOKEN_TTL` 控制。生产环境必须使用 HTTPS，并将 `JWT_SECRET` 替换为至少 32 字节的随机值，例如 `openssl rand -base64 48` 的输出。

桌面端和移动端不应保存用户密码。access token 仅保存在进程内存；refresh token 在 macOS 保存到 Keychain，在 Android 使用 Android Keystore 中的 AES 密钥加密后保存密文。应用启动时读取 refresh token 调用刷新接口，刷新失败则清除本地 token 并重新显示登录页。

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
make docker-up        # 构建并在容器中启动 API
make docker-down      # 停止 Docker 服务并保留 SQLite 数据卷
```

## 开发新模块

新增例如 `budget` 模块时，在 `internal/app/budget` 中建立 `model.go`、`repository.go`、`service.go`、`handler.go`，然后只在 `internal/router/router.go` 中完成构造和路由注册。不要从已有的 account 或 transaction 模块直接调用代码；确有跨模块流程时，应在单独的编排 service 中通过接口完成。

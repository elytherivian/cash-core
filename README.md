# cash-core

cash-core 是一个使用 Gin、GORM 和 SQLite 构建的日常记账流水后端。项目按照业务模块纵向组织，每个模块独立拥有 HTTP、业务、数据访问和模型代码，不允许业务模块之间直接依赖。

完整的接口说明、认证方式和调用示例见 [API 文档](docs/API.md)，VPS 首次部署、升级和回滚命令见 [部署手册](docs/DEPLOYMENT.md)。

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
├── .github/workflows/ci.yml       # 测试、构建并推送 Docker Hub 镜像
├── deployments/                   # 生产 Compose、本地开发 override 与 Caddy 站点示例
├── docs/DEPLOYMENT.md             # VPS 首次部署、更新、备份和回滚
├── scripts/deploy.sh              # VPS 拉取镜像并更新服务
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

### 本地 `.env` 与生产 `.env`

项目中有两类 `.env`，文件名相同但用途不同，不能直接混用：

| 对比项 | 本地开发 `.env` | VPS 生产 `.env` |
| --- | --- | --- |
| 示例来源 | 根目录 `.env.example` | `deployments/.env.production.example` |
| 建议位置 | 项目根目录 `.env` | `/root/cash-core/.env` |
| 使用者 | `make run` 和本地 Go 进程 | Docker Compose 和 API 容器 |
| 环境 | `APP_ENV=local` | `APP_ENV=production` |
| SQLite | `DB_PATH=data/cash.db` | `SQLITE_DATA_DIR=/root/cash-core/data`；容器内固定 `DB_PATH=/data/cash.db` |
| HTTP | 可设置 `HTTP_HOST`、`HTTP_PORT` 和超时参数 | Compose 固定监听容器内 `0.0.0.0:8080`，不 publish 到宿主机 |
| 镜像与网络 | 不需要 `APP_IMAGE` 或 `CADDY_PROXY_NETWORK` | 必须设置完整 `APP_IMAGE`，并连接 `CADDY_PROXY_NETWORK` |
| 密钥 | 只能使用本地开发密钥 | `JWT_SECRET` 必须使用独立随机强密钥 |

本地 `.env` 由应用直接读取，重点是 Go 进程、HTTP 超时和本地数据库路径。生产 `.env` 首先由 Docker Compose 用来解析镜像、宿主机挂载目录和外部网络，然后 Compose 把应用所需变量传入容器。生产 Compose 会强制将 `DB_PATH`、`HTTP_HOST` 和 `HTTP_PORT` 设置为容器内的固定值，因此生产 `.env` 不需要重复设置这三个变量。

两份 `.env` 都已被 Git 忽略，不应提交。不要把本地 `.env` 直接复制到 VPS，也不要把生产 `JWT_SECRET` 写回 `.env.example` 或 README。

## 本地启动

需要安装 Go 1.25。SQLite 文件和 schema 会在应用首次启动时自动创建。

```bash
make run
```

也可以使用容器启动本地源码。开发命令叠加 `deployments/docker-compose.dev.yml`，构建当前工作区代码：

```bash
make docker-build-up
```

本地 SQLite 默认保存在仓库 `data/cash.db`。Compose 只 `expose` 容器的 `8080`，不会 publish 到宿主机；如需从宿主机直接调试，建议使用 `make run`。容器模式需要提前创建 `caddy_proxy` 网络，或连接已有的本地反向代理：

```bash
docker network inspect caddy_proxy >/dev/null 2>&1 || docker network create caddy_proxy
```

本机已经执行过 `docker login` 时，可以直接把当前源码构建为多架构 `dev` 镜像并推送到 Docker Hub：

```bash
make docker-push
```

也可以显式传入自己的 Docker Hub 仓库名和目标平台。标签只允许 `dev` 或 `latest`，日常开发应使用默认的 `dev`：

```bash
make docker-push \
  IMAGE_REPOSITORY=your-dockerhub-user/cash-core \
  IMAGE_TAG=dev \
  PLATFORMS=linux/amd64,linux/arm64
```

默认 SQLite 配置：

- 本机运行：`data/cash.db`
- 本地开发容器：宿主机仓库 `data/` bind mount 到容器 `/data`
- VPS 生产容器：宿主机 `/root/cash-core/data` bind mount 到容器 `/data`

SQLite 适合单机、低并发的 VPS 部署；同一数据库文件不应放在网络文件系统上，也不应同时由多个 API 容器写入。

## 开发、发布与部署流程

整个流程将源码开发、镜像构建和生产运行分开：

```text
本地 codex/dev 开发
  -> 本地测试
  -> push codex/dev
  -> GitHub CI 测试并发布 :dev
  -> 合并到 main
  -> GitHub CI 测试并发布 :latest
  -> 创建 X.Y.Z Git tag（只标记源码版本，不构建新镜像标签）
  -> 备份 SQLite
  -> VPS pull :latest + up
```

### 1. 日常开发

所有日常修改在 `codex/dev` 分支进行：

```bash
git switch codex/dev
git pull --ff-only
cp .env.example .env  # 仅首次需要，之后保留自己的本地配置
make run
```

提交前执行：

```bash
make fmt
make vet
make test
git status --short
git add <本次修改的文件>
git commit -m "<变更说明>"
git push origin codex/dev
```

如果需要验证本地 Docker 构建，先确保存在 `caddy_proxy` 网络，再运行 `make docker-build-up`。本地源码构建和生产镜像拉取是两个独立命令，避免误把工作区代码当成生产版本。

### 2. 开发镜像

推送 `codex/dev` 后，GitHub Actions 自动执行格式检查、`go vet`、测试和双架构镜像构建。全部通过后发布：

```text
your-dockerhub-user/cash-core:dev
```

`:dev` 会随每次开发推送移动，只用于开发或测试环境，不应部署到生产 VPS。

### 3. 稳定分支与正式版本

开发版本验证完成后，通过 Pull Request 或经过审查的合并把 `codex/dev` 合入 `main`。推送 `main` 后，CI 只发布 `your-dockerhub-user/cash-core:latest`。

正式发布使用 Git tag，而不是为每个版本创建永久分支：

```bash
git switch main
git pull --ff-only
git tag -a 1.0.0 -m "Release 1.0.0"
git push origin main
git push origin 1.0.0
```

Git Tag 只记录源码发布版本，不触发镜像构建，也不会生成 `:1.0.0` 镜像。仓库始终只有两个 CI 管理的 Docker 标签：`codex/dev` 对应 `:dev`，`main` 对应 `:latest`。Docker 标签不能包含 `/`，所以开发镜像使用 `dev`，而不是 `codex/dev`。

### 4. VPS 更新

VPS 不 clone 完整源码、不安装 Go，也不执行 `docker build`。它只保留 `docker-compose.yml`、生产 `.env`、全局 Caddy site 文件和 SQLite 数据目录。生产 `.env` 固定使用 `APP_IMAGE=your-dockerhub-user/cash-core:latest`；更新时先备份数据库，再执行 `docker compose pull api` 和 `docker compose up -d api`。完整步骤见 [部署手册](docs/DEPLOYMENT.md)。

### GitHub Actions CI

流水线位于 `.github/workflows/ci.yml`，执行以下流程：

1. Pull Request：运行格式检查、`go vet`、测试和容器构建验证，但不推送镜像。
2. 推送到 `codex/dev`：测试通过后只推送 `dev`，供开发环境使用。
3. 推送到 `main`：测试通过后只推送 `latest`，供生产 VPS 使用。
4. 推送 Git Tag 不触发流水线，也不生成版本号或 SHA 镜像。

先在 GitHub 仓库的 **Settings → Secrets and variables → Actions** 中配置：

- Secret `DOCKERHUB_USERNAME`：Docker Hub 用户名。
- Secret `DOCKERHUB_TOKEN`：Docker Hub Access Token，不要使用或提交账号密码。
- 可选 Variable `DOCKERHUB_IMAGE`：完整镜像仓库名，例如 `your-dockerhub-user/cash-core`；未设置时由 `DOCKERHUB_USERNAME` Secret 自动生成 `<用户名>/cash-core`。流水线不在仓库文件中硬编码真实账号。

源码发布版本使用 Git Tag 标记：

```bash
git tag -a 1.0.0 -m "Release 1.0.0"
git push origin 1.0.0
```

CI 同时构建 `linux/amd64` 和 `linux/arm64`，因此常见 x86_64、ARM64 VPS 可以使用同一个标签。

## VPS 生产部署

生产架构由两套独立 Compose 组成：`/root/caddy` 管理全局 Caddy，cash-core Compose 只管理一个 API 容器。二者通过外部 Docker 网络 `caddy_proxy` 通信。只有全局 Caddy 对宿主机发布 `80/tcp` 和 `443/tcp`；cash-core API 不 publish `8080`，只在 Docker 网络内 `expose 8080`。

Caddy 统一通过网络别名 `cash-core:8080` 访问 API。生产环境不配置 `replicas` 或 `scale`，因为多个进程同时写一个 SQLite 文件不适合当前架构。

### 首次部署

1. 启动全局 Caddy。其 Compose 应创建或加入外部网络 `caddy_proxy`，并且只发布 `80/tcp`、`443/tcp`：

```bash
cd /root/caddy
docker compose up -d
docker network inspect caddy_proxy
```

如果 `/root/caddy` 尚未创建该网络，可先执行一次：

```bash
docker network create caddy_proxy
```

2. 准备 cash-core 部署目录。VPS 无需完整 Git 仓库：

```bash
mkdir -p /root/cash-core
cd /root/cash-core
```

只需把仓库中的 `deployments/docker-compose.yml` 保存为 `/root/cash-core/docker-compose.yml`，并准备后续的 `.env`；应用源码、Go 工具链、Dockerfile 和构建上下文都不需要放到 VPS。

3. 创建 SQLite 数据目录，并确保镜像内的非 root 用户可以读写：

```bash
install -d -o 65532 -g 65532 -m 750 /root/cash-core/data
```

容器内 `/data` bind mount 到宿主机 `/root/cash-core/data`，数据库路径固定为 `/data/cash.db`，对应宿主机文件 `/root/cash-core/data/cash.db`。

4. 创建权限为 `600` 的 `/root/cash-core/.env`，至少包含：

```dotenv
APP_ENV=production
APP_IMAGE=your-dockerhub-user/cash-core:latest
APP_VERSION=1.0.0
JWT_SECRET=<openssl rand -base64 48 生成的值>
SQLITE_DATA_DIR=/root/cash-core/data
CADDY_PROXY_NETWORK=caddy_proxy
```

完整模板见 `deployments/.env.production.example`。当前 CI 只发布 `dev` 和 `latest`，生产必须使用 `latest`，不能使用 `dev`。`APP_VERSION` 仅用于 API 展示，可以填写对应源码 Git Tag。如果 Docker Hub 仓库是私有的，VPS 首次部署前需使用只读 Access Token 执行 `docker login`。

5. 拉取镜像并启动 API。生产 Compose 没有 `build:`，不会在 VPS 编译：

```bash
cd /root/cash-core
docker compose --env-file .env pull api
docker compose --env-file .env up -d api
```

6. 在全局 Caddy 中配置站点。仓库示例位于 `deployments/caddy/cash-core.caddy.example`；VPS 当前站点文件可使用：

```text
/root/caddy/sites/api.example.com.caddy
```

内容为：

```caddyfile
api.example.com {
    reverse_proxy cash-core:8080
}
```

7. Reload 全局 Caddy：

```bash
docker compose -f /root/caddy/compose.yaml exec caddy \
  caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile
```

8. 验证公网健康检查，同时确认 API 没有宿主机端口映射：

```bash
curl -i https://api.example.com/health/live
curl -i https://api.example.com/health/ready
docker compose --env-file /root/cash-core/.env \
  -f /root/cash-core/docker-compose.yml ps
```

### 版本更新与回滚

更新前应先记录当前镜像 digest 并备份整个 SQLite 数据目录，再拉取 `latest`：

```bash
cd /root/cash-core
docker compose --env-file .env pull api
docker compose --env-file .env stop api
cp -a /root/cash-core/data \
  /root/cash-core/data.backup.$(date +%Y%m%d%H%M%S)
docker compose --env-file .env up -d api
```

容器删除、镜像更新和重新 `up` 都不会删除 `/root/cash-core/data/cash.db`，因为它位于宿主机 bind mount，而不是 Docker named volume。禁止执行：

```bash
docker compose down -v
rm -rf /root/cash-core/data
```

因为 `latest` 是移动标签，镜像回滚必须使用更新前记录的 Docker Hub digest，例如临时把 `APP_IMAGE` 改成 `your-dockerhub-user/cash-core@sha256:<旧摘要>`。数据回滚必须先停止 API，再恢复完整备份目录。详细的记录 digest、备份、升级、验证和回滚命令见 [部署手册](docs/DEPLOYMENT.md)。应用启动时会自动幂等初始化当前 schema；所有 schema 变更必须保持向后兼容和幂等。如果未来加入破坏性数据库迁移，发布文档必须要求先备份，并提供对应的数据及镜像回滚步骤。

SQLite 不使用数据库账号、密码或端口。不要提交 `.env`、Docker Hub Token、私钥或 `data/`；这些路径已被 Git 忽略。

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
PATCH  /api/v1/transactions/update
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
make docker-build     # 构建本地镜像，不推送
make docker-push      # 构建多架构镜像并推送 Docker Hub
make docker-up        # 从 Docker Hub 拉取镜像并启动生产 API
make docker-build-up  # 用当前源码构建并启动本地 API
make docker-down      # 停止 API，不删除宿主机 SQLite 数据目录
make deploy           # VPS 拉取镜像并更新 API
```

## 开发新模块

新增例如 `budget` 模块时，在 `internal/app/budget` 中建立 `model.go`、`repository.go`、`service.go`、`handler.go`，然后只在 `internal/router/router.go` 中完成构造和路由注册。不要从已有的 account 或 transaction 模块直接调用代码；确有跨模块流程时，应在单独的编排 service 中通过接口完成。

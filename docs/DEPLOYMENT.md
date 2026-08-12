# cash-core VPS 部署与升级手册

本文描述生产环境的首次部署、日常升级、SQLite 备份和回滚。示例中的 Docker Hub 账号、域名和密钥都是占位符，执行前必须替换。

## 1. 架构与发布约定

- VPS 的全局 Caddy 位于 `/root/caddy`，它是唯一发布公网端口的容器，只开放 `80/tcp` 和 `443/tcp`。
- Caddy 和 API 加入外部 Docker 网络 `caddy_proxy`。
- API 只 `expose 8080`，不把 `8080` publish 到宿主机。
- Caddy 统一反代 `cash-core:8080`。
- SQLite 宿主机目录是 `/root/cash-core/data`，bind mount 到容器 `/data`；数据库文件是 `/root/cash-core/data/cash.db`。
- API 只运行一个副本。不要设置 `replicas` 或 `scale`，避免多个进程同时写 SQLite。
- VPS 不构建镜像，也不需要完整源码，只保留 `docker-compose.yml`、`.env`、Caddy site 文件和数据目录。

CI 只维护两个 Docker Hub 标签：

| Git 分支 | Docker 镜像 | 用途 |
| --- | --- | --- |
| `codex/dev` | `your-dockerhub-user/cash-core:dev` | 开发与测试 |
| `main` | `your-dockerhub-user/cash-core:latest` | 生产 |

`1.0.0` 这样的 Git Tag 只标记源码版本，不触发 CI，也不会生成同名 Docker 镜像。生产环境固定拉取 `latest`。由于 `latest` 会移动，升级前必须记录旧镜像 digest，才能可靠回滚应用镜像。

## 2. VPS 首次部署

### 2.1 启动全局 Caddy

确认 `/root/caddy/compose.yaml` 只对公网发布 `80/tcp` 和 `443/tcp`，并让 Caddy 加入 `caddy_proxy` 网络：

```bash
docker network inspect caddy_proxy >/dev/null 2>&1 || docker network create caddy_proxy
cd /root/caddy
docker compose up -d
```

### 2.2 准备 API 目录

```bash
mkdir -p /root/cash-core
cd /root/cash-core
```

把仓库中的 `deployments/docker-compose.yml` 复制到 `/root/cash-core/docker-compose.yml`。不要在 VPS 放置 Dockerfile 或执行 `docker build`。

### 2.3 创建 SQLite 数据目录

镜像使用 UID/GID `65532:65532` 的非 root 用户运行：

```bash
install -d -o 65532 -g 65532 -m 750 /root/cash-core/data
```

不要改成 Docker named volume。宿主机 bind mount 便于备份和迁移，并保证删除容器、拉取镜像和重新 `up` 时数据库不会消失。

### 2.4 创建生产 `.env`

先生成 JWT 密钥：

```bash
openssl rand -base64 48
```

创建 `/root/cash-core/.env`：

```dotenv
APP_ENV=production
APP_IMAGE=your-dockerhub-user/cash-core:latest
APP_VERSION=1.0.0
JWT_SECRET=replace-with-the-generated-random-secret
SQLITE_DATA_DIR=/root/cash-core/data
CADDY_PROXY_NETWORK=caddy_proxy
```

然后限制权限：

```bash
chmod 600 /root/cash-core/.env
```

完整配置可从 `deployments/.env.production.example` 复制。生产 `.env` 不应提交到 Git。若 Docker Hub 仓库是私有的，使用只读 Access Token 在 VPS 执行一次 `docker login`。

### 2.5 拉取并启动 API

```bash
cd /root/cash-core
docker compose --env-file .env pull api
docker compose --env-file .env up -d api
docker compose --env-file .env ps
```

生产 Compose 不包含 `build:`，所以以上命令只会从 Docker Hub 拉取镜像。

### 2.6 配置并 reload Caddy

创建类似 `/root/caddy/sites/api.example.com.caddy` 的站点文件：

```caddyfile
api.example.com {
    reverse_proxy cash-core:8080
}
```

Reload Caddy：

```bash
docker compose -f /root/caddy/compose.yaml exec caddy \
  caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile
```

### 2.7 验证首次部署

```bash
curl -i https://api.example.com/health/live
curl -i https://api.example.com/health/ready
docker compose --env-file /root/cash-core/.env \
  -f /root/cash-core/docker-compose.yml ps
docker port cash-core 2>/dev/null || true
```

健康检查应成功；Compose 的 `PORTS` 不应出现宿主机 `8080` 映射。容器名可能由 Compose 自动添加项目前缀，因此诊断时优先使用 `docker compose ps`。

## 3. 生产版本更新

推送或合并到 `main` 后，先等待 GitHub CI 成功发布新的 `:latest`，再操作 VPS。

### 3.1 记录当前镜像 digest

在更新之前执行：

```bash
cd /root/cash-core
docker compose --env-file .env images
docker image inspect your-dockerhub-user/cash-core:latest \
  --format '{{index .RepoDigests 0}}'
```

把输出的 `your-dockerhub-user/cash-core@sha256:<摘要>` 保存到发布记录。它是本次失败时的应用镜像回滚目标。

### 3.2 提前拉取新镜像

拉取可以在停机前完成，以缩短维护时间：

```bash
cd /root/cash-core
docker compose --env-file .env pull api
```

### 3.3 停止写入并备份 SQLite

SQLite 可能使用 WAL 文件。为了得到一致备份，应短暂停止 API，并复制整个数据目录，而不是只在应用运行时复制 `cash.db`：

```bash
cd /root/cash-core
docker compose --env-file .env stop api
cp -a /root/cash-core/data \
  /root/cash-core/data.backup.$(date +%Y%m%d%H%M%S)
```

确认备份目录存在后再继续。

### 3.4 使用新镜像重建容器

```bash
cd /root/cash-core
docker compose --env-file .env up -d api
docker compose --env-file .env ps
docker compose --env-file .env logs --tail=100 api
```

如需让 API 展示新的源码版本，可以在执行 `up` 前更新 `.env` 中的 `APP_VERSION`；`APP_IMAGE` 保持 `your-dockerhub-user/cash-core:latest`。

### 3.5 验证升级

```bash
curl -i https://api.example.com/health/live
curl -i https://api.example.com/health/ready
```

同时验证登录、查询等关键业务路径。确认新版本稳定后再按备份保留策略清理旧备份。

## 4. 回滚

### 4.1 只回滚应用镜像

把 `.env` 中的 `APP_IMAGE` 临时改为升级前记录的 digest：

```dotenv
APP_IMAGE=your-dockerhub-user/cash-core@sha256:replace-with-previous-digest
```

然后重建 API：

```bash
cd /root/cash-core
docker compose --env-file .env up -d api
docker compose --env-file .env logs --tail=100 api
```

验证完成后保留 digest，直到下一次明确发布；不要立即改回可能已经指向新镜像的 `latest`。

### 4.2 回滚 SQLite 数据

只有确认新版本改变了数据且必须恢复时，才回滚数据库。先停止 API，并把当前数据目录移作故障现场备份，再恢复升级前的完整目录：

```bash
cd /root/cash-core
docker compose --env-file .env stop api
mv /root/cash-core/data /root/cash-core/data.failed
cp -a /root/cash-core/data.backup.YYYYMMDDHHMMSS /root/cash-core/data
chown -R 65532:65532 /root/cash-core/data
chmod 750 /root/cash-core/data
docker compose --env-file .env up -d api
```

把 `YYYYMMDDHHMMSS` 替换为实际备份时间。数据库回滚会丢失备份时间之后写入的数据，执行前必须确认业务影响。

## 5. 禁止操作与迁移约束

不要执行：

```bash
docker compose down -v
rm -rf /root/cash-core/data
```

虽然当前 SQLite 使用 bind mount，不是 named volume，但仍统一禁止 `down -v`，避免未来配置变化时误删数据。容器删除、镜像更新和重新 `up` 不会删除 `/root/cash-core/data`。

应用启动时可以自动初始化或迁移 schema，但迁移必须向后兼容且幂等。如果未来加入破坏性数据库迁移，发布说明必须明确：

1. 升级前的完整备份命令。
2. 兼容的旧镜像 digest。
3. 数据回滚步骤和可能丢失的数据范围。

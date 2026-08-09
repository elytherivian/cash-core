# cash-core API 文档

## 概览

- 基础地址：`http://localhost:8080`
- API 前缀：`/api/v1`
- 请求与响应均使用 JSON；`GET` 请求的筛选条件放在 URL 查询参数中，不放在请求体。
- UUID 参数必须是标准 UUID 字符串。
- 所有响应时间都按 `API_TIMEZONE` 配置输出；数据库内部以 UTC 保存。

除健康检查、用户管理和认证接口外，所有业务接口都需要访问令牌：

```http
Authorization: Bearer <access_token>
```

## 统一响应

成功响应：

```json
{
  "version": "dev",
  "code": 0,
  "message": "ok",
  "data": {}
}
```

错误响应没有 `data` 字段：

```json
{
  "version": "dev",
  "code": 40000,
  "message": "invalid input: ..."
}
```

常见业务码：

| HTTP 状态 | `code` | 含义 |
| --- | --- | --- |
| 400 | 40000 | 请求参数无效 |
| 401 | 40100 | 未认证、令牌无效或过期 |
| 404 | 40400 | 资源不存在 |
| 405 | 40500 | HTTP 方法不允许 |
| 409 | 40900 | 资源冲突 |
| 409 | 40001 | 用户名已存在 |
| 500 | 50000 | 服务内部错误 |
| 503 | 50300 | 服务暂不可用 |

## 健康检查

### 存活检查

```http
GET /health/live
```

无需认证。服务进程可用时返回 `200`。

### 就绪检查

```http
GET /health/ready
```

无需认证。数据库可连接时返回 `200`；否则返回 `503`。

## 用户

### 注册

```http
POST /api/v1/users/register
Content-Type: application/json
```

```json
{
  "username": "cash-user",
  "password": "password123"
}
```

`username` 长度为 1–50 个字符；注册密码长度为 8–72 字节。成功返回 `201`：

```json
{
  "code": 0,
  "message": "user registered",
  "data": {
    "id": "<user_id>",
    "username": "cash-user",
    "created_at": "2026-08-09T20:00:00+08:00",
    "updated_at": "2026-08-09T20:00:00+08:00",
    "is_active": true
  }
}
```

### 删除用户

```http
DELETE /api/v1/users/delete
Content-Type: application/json
```

```json
{
  "username": "cash-user",
  "password": "password123"
}
```

这是软删除操作。成功返回 `200`。

### 恢复用户

```http
POST /api/v1/users/restore
Content-Type: application/json
```

```json
{
  "username": "cash-user",
  "password": "password123"
}
```

仅能恢复已删除用户。成功返回 `200`。

## 认证

### 登录

```http
POST /api/v1/auth/login
Content-Type: application/json
```

```json
{
  "username": "cash-user",
  "password": "password123"
}
```

成功返回 `200`，`access_token` 用于业务请求，`refresh_token` 仅用于刷新：

```json
{
  "code": 0,
  "message": "login successful",
  "data": {
    "user_id": "<user_id>",
    "access_token": "<access_token>",
    "refresh_token": "<refresh_token>",
    "token_type": "Bearer",
    "expires_in": 900,
    "refresh_expires_in": 2592000
  }
}
```

### 刷新令牌

```http
POST /api/v1/auth/refresh
Content-Type: application/json
```

```json
{
  "refresh_token": "<refresh_token>"
}
```

成功后会返回一组新的 `access_token` 和 `refresh_token`。不要把 `refresh_token` 放入 Bearer 头调用业务接口。

## 账户

账户表示资金归属，例如 `WeChat/wechat1`。每个用户有效账户的 `(account_type, account_name)` 必须唯一。

### 创建账户

```http
POST /api/v1/accounts/create
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "account_type": "WeChat",
  "account_name": "wechat1",
  "initial_balance": "0"
}
```

`account_type` 仅支持 `WeChat`、`AliPay`、`BOC`；`account_name` 长度为 1–100 个字符。成功返回 `201` 及创建的账户。

### 获取账户列表

```http
GET /api/v1/accounts/list
Authorization: Bearer <access_token>
```

成功返回 `200` 和当前用户的全部有效账户数组。

## 分类

分类是用户定义的具体项目，例如 `零食`、`日用品`、`工资`；分类本身不区分收入或支出。

### 创建分类

```http
POST /api/v1/categories/create
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "category_name": "零食"
}
```

`category_name` 长度为 1–80 个字符。成功返回 `201` 及创建的分类。

### 获取分类列表

```http
GET /api/v1/categories/list
Authorization: Bearer <access_token>
```

成功返回 `200` 和当前用户的全部有效分类数组。

## 流水

流水记录实际的收入或支出，并关联一个账户与一个分类。`type` 表示流水方向：`income` 为收入，`expense` 为支出。

### 创建流水

```http
POST /api/v1/transactions/create
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "type": "expense",
  "amount": "20",
  "account_id": "<wechat1 的 account_id>",
  "category_id": "<零食的 category_id>",
  "occurred_at": "2026-08-09T12:00:00Z"
}
```

`amount` 必须大于 0；`account_id` 和 `category_id` 必须属于当前用户。`occurred_at` 可选，未传时服务端使用当前 UTC 时间。成功返回 `201` 及创建的流水。

### 按账户查询流水

```http
GET /api/v1/transactions/getByAccountID?account_id=<account_id>
Authorization: Bearer <access_token>
```

返回该账户下当前用户的全部有效流水，按 `occurred_at`、`id` 正序排列。

### 按分类查询流水

```http
GET /api/v1/transactions/getByCategoryID?category_id=<category_id>
Authorization: Bearer <access_token>
```

返回该分类下当前用户的全部有效流水，按 `occurred_at`、`id` 正序排列。

### 按账户和分类查询流水

```http
GET /api/v1/transactions/getByAccountIDAndCategoryID?account_id=<account_id>&category_id=<category_id>
Authorization: Bearer <access_token>
```

两个查询参数都必填。返回同时符合账户和分类条件的全部有效流水，按 `occurred_at`、`id` 正序排列。

流水查询不分页，供客户端按完整数据计算：

```text
账户余额 = initial_balance + income 金额之和 - expense 金额之和
```

### 流水响应字段

创建与查询接口中的流水对象包含：

```json
{
  "id": "<transaction_id>",
  "user_id": "<user_id>",
  "type": "expense",
  "amount": "20",
  "account_id": "<account_id>",
  "category_id": "<category_id>",
  "occurred_at": "2026-08-09T20:00:00+08:00",
  "created_at": "2026-08-09T20:00:00+08:00",
  "updated_at": "2026-08-09T20:00:00+08:00",
  "is_active": true
}
```

## 调用顺序示例

1. 调用注册接口创建用户。
2. 调用登录接口，保存返回的 `access_token` 与 `refresh_token`。
3. 使用 `access_token` 创建账户和分类。
4. 使用账户、分类返回的 `id` 创建流水。
5. 通过按账户查询流水接口获取完整流水，在客户端计算账户余额。
6. `access_token` 过期后，使用 `refresh_token` 调用刷新接口，再替换客户端保存的 token。

# Go 多服务微服务架构

基于 CloudWeGo 生态（Hertz + Kitex）的微服务项目，使用 etcd 服务注册与发现，nginx + Docker 部署。

## 架构概览

```
                    ┌──────────┐
     HTTP 请求 ───▶ │  Nginx   │
                    └────┬─────┘
                         │
                    ┌────▼─────┐
                    │ Hertz    │
                    │ API 网关  │  byte.dance.api (:8080)
                    └──┬───┬───┘
              ┌────────┘   └────────┐
         Kitex RPC            Kitex RPC
              │                     │
     ┌────────▼────────┐  ┌────────▼────────┐
     │  User Service   │  │ Article Service  │
     │ byte.dance.user │  │byte.dance.article│
     │    (:8881)      │  │    (:8882)       │
     └────────┬────────┘  └────────┬─────────┘
              │                     │
              └──────┐   ┌──────────┘
                  ┌──▼───▼──┐
                  │  etcd   │
                  │ (:2379) │
                  └─────────┘
```

## 项目结构

```
.
├── idl/                           # Thrift IDL 定义
│   ├── api.thrift                 # Hertz HTTP 接口 IDL (api.* 注解)
│   ├── user.thrift                # 用户 RPC 服务 IDL
│   └── article.thrift             # 文章 RPC 服务 IDL
├── kitex_gen/                     # Kitex 生成代码 (module: byte.dance/kitex_gen)
│   ├── user/
│   └── article/
├── pkg/                           # 公共包 (module: byte.dance/pkg)
│   ├── consts/
│   ├── errno/
│   └── middleware/
├── api/                           # Hertz HTTP 网关 (module: byte.dance/api)
│   ├── biz/
│   │   ├── model/api/api.go       # hz 生成的请求/响应结构体 (带 binding + vd tags)
│   │   ├── handler/api/           # hz 生成的 handler 桩 → 填入 RPC 调用
│   │   ├── handler/ping.go        # hz 生成的健康检查
│   │   ├── router/api/api.go      # hz 自动生成路由 (基于 api.post/get/put/delete)
│   │   ├── router/api/middleware.go  # 每路由中间件钩子 (JWT 鉴权)
│   │   ├── router/register.go     # hz 生成的路由注册入口
│   │   └── rpc/                   # RPC 客户端初始化 (etcd 服务发现)
│   ├── router.go                  # hz 生成: 自定义路由
│   ├── router_gen.go              # hz 生成: 路由聚合 (DO NOT EDIT)
│   ├── main.go
│   └── Dockerfile
├── service/
│   ├── user/                      # Kitex 用户 RPC 服务 (module: byte.dance/user)
│   └── article/                   # Kitex 文章 RPC 服务 (module: byte.dance/article)
├── deploy/nginx/nginx.conf
├── go.work
├── docker-compose.yml
└── Makefile
```

## 技术栈

| 组件 | 技术 |
|------|------|
| HTTP 框架 | [Hertz](https://github.com/cloudwego/hertz) |
| RPC 框架 | [Kitex](https://github.com/cloudwego/kitex) |
| IDL | Apache Thrift |
| HTTP 代码生成 | `hz` (Hertz IDL 注解生成工具) |
| RPC 代码生成 | `kitex` + `thriftgo` |
| 服务注册/发现 | etcd |
| 参数校验 | go-tagexpr (`vd` tag) |
| 认证 | JWT (HS256) |
| 部署 | Docker + Nginx |

## Hertz IDL 注解说明 (`idl/api.thrift`)

`hz` 工具通过 Thrift IDL 中的 `api.*` 注解自动生成 Hertz HTTP 路由和请求/响应结构体。

### 参数绑定注解 (struct field)

| 注解 | 生成的 tag | 作用 | 示例 |
|------|-----------|------|------|
| `api.body` | `json:"x" form:"x"` | 从 JSON/Form 请求体绑定 | `api.body="username"` |
| `api.path` | `path:"x"` | 从 URL 路径参数绑定 | `api.path="id"` |
| `api.query` | `query:"x"` | 从 URL 查询参数绑定 | `api.query="page"` |
| `api.header` | `header:"x"` | 从请求头绑定 | `api.header="Token"` |
| `api.cookie` | `cookie:"x"` | 从 Cookie 绑定 | `api.cookie="session"` |
| `api.form` | `form:"x"` | 从 Form 表单绑定 | `api.form="file"` |

### 参数校验注解 (struct field)

| 注解 | 生成的 tag | 作用 | 示例 |
|------|-----------|------|------|
| `api.vd` | `vd:"expr"` | go-tagexpr 校验表达式 | `api.vd="len($)>1 && len($)<33"` |

校验表达式中 `$` 代表当前字段值，支持 `len($)`、`$>0`、`regexp(...)` 等。可用 `msg:'...'` 自定义错误消息。

### 路由注解 (service method)

| 注解 | 作用 | 示例 |
|------|------|------|
| `api.post` | 生成 POST 路由 | `(api.post="/api/v1/user/register")` |
| `api.get` | 生成 GET 路由 | `(api.get="/api/v1/user/:id")` |
| `api.put` | 生成 PUT 路由 | `(api.put="/api/v1/user/:id")` |
| `api.delete` | 生成 DELETE 路由 | `(api.delete="/api/v1/article/:id")` |

### 示例 IDL 片段

```thrift
struct RegisterReq {
    1: string username (api.body="username", api.vd="len($)>1 && len($)<33; msg:'username length must be 2-32'")
    2: string email    (api.body="email",    api.vd="len($)>4 && len($)<65; msg:'email length must be 5-64'")
    3: string password (api.body="password", api.vd="len($)>5 && len($)<129; msg:'password length must be 6-128'")
}

struct GetUserReq {
    1: i64 id (api.path="id", api.vd="$>0; msg:'invalid user id'")
}

struct ListArticleReq {
    1: i64 author_id (api.query="author_id")
    2: i32 page      (api.query="page",      api.vd="$>0; msg:'page must be > 0'")
    3: i32 page_size (api.query="page_size",  api.vd="$>0 && $<=100; msg:'page_size must be 1-100'")
}

service ApiService {
    RegisterResp   Register(1: RegisterReq req) (api.post="/api/v1/user/register")
    GetUserResp    GetUser(1: GetUserReq req)   (api.get="/api/v1/user/:id")
    ListArticleResp ListArticle(1: ListArticleReq req) (api.get="/api/v1/articles")
}
```

`hz` 读取这些注解后自动生成：
1. **`biz/model/`** — Go 结构体，字段带有 `json`/`form`/`path`/`query`/`vd` tag
2. **`biz/router/`** — 路由注册代码，按 `api.post`/`api.get` 等注解映射到 handler
3. **`biz/handler/`** — handler 桩函数，自动调用 `c.BindAndValidate(&req)` 完成绑定+校验

开发者只需在 handler 桩中填入业务逻辑（RPC 调用）。

## API 接口

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/user/register` | 用户注册 |
| POST | `/api/v1/user/login` | 用户登录 |
| GET  | `/api/v1/user/:id` | 获取用户信息 |
| GET  | `/api/v1/articles` | 文章列表 |
| GET  | `/api/v1/article/:id` | 获取文章详情 |

### 需要认证（Bearer Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| PUT    | `/api/v1/user/:id` | 更新用户信息 |
| POST   | `/api/v1/article` | 创建文章 |
| PUT    | `/api/v1/article/:id` | 更新文章 |
| DELETE | `/api/v1/article/:id` | 删除文章 |

## 快速开始

### Docker Compose（推荐）

```bash
docker-compose up --build -d
```

服务启动后：
- Nginx 代理: http://localhost:80
- API 网关直连: http://localhost:8080
- etcd: localhost:2379

### 本地开发

1. 启动 etcd：

```bash
docker run -d --name etcd \
  -p 2379:2379 \
  -e ALLOW_NONE_AUTHENTICATION=yes \
  bitnami/etcd:3.5
```

2. 分别启动三个服务：

```bash
make run-user      # 终端 1
make run-article   # 终端 2
make run-api       # 终端 3
```

### 代码生成

```bash
# 生成全部（Kitex RPC + Hertz HTTP）
make gen

# 仅生成 Kitex RPC 代码
make gen-kitex

# 仅生成 Hertz HTTP 代码（从 idl/api.thrift 的 api.* 注解）
make gen-hz
```

### 编译

```bash
make all
```

输出二进制文件在 `output/` 目录。

## 使用示例

```bash
# 注册
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@byte.dance","password":"123456"}'

# 登录
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"123456"}'

# 获取用户
curl http://localhost:8080/api/v1/user/1

# 创建文章（需要 token）
curl -X POST http://localhost:8080/api/v1/article \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"Hello World","content":"This is my first article."}'

# 获取文章列表（参数校验: page>0, page_size 1-100）
curl "http://localhost:8080/api/v1/articles?page=1&page_size=10"

# 获取文章详情
curl http://localhost:8080/api/v1/article/1
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `ETCD_ENDPOINT` | `127.0.0.1:2379` | etcd 地址 |

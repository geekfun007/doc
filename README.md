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
├── idl/                    # Thrift IDL 定义
│   ├── user.thrift
│   └── article.thrift
├── kitex_gen/              # Kitex 生成代码 (module: byte.dance/kitex_gen)
│   ├── user/
│   └── article/
├── pkg/                    # 公共包 (module: byte.dance/pkg)
│   ├── consts/             # 常量与服务配置
│   ├── errno/              # 统一错误码
│   └── middleware/         # JWT 中间件
├── api/                    # Hertz HTTP 网关 (module: byte.dance/api)
│   ├── biz/
│   │   ├── handler/        # HTTP 处理器
│   │   ├── router/         # 路由注册
│   │   └── rpc/            # RPC 客户端初始化
│   ├── main.go
│   └── Dockerfile
├── service/
│   ├── user/               # Kitex 用户 RPC 服务 (module: byte.dance/user)
│   │   ├── dal/            # 数据访问层
│   │   ├── handler.go
│   │   ├── main.go
│   │   └── Dockerfile
│   └── article/            # Kitex 文章 RPC 服务 (module: byte.dance/article)
│       ├── dal/
│       ├── handler.go
│       ├── main.go
│       └── Dockerfile
├── deploy/
│   └── nginx/nginx.conf
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
| 服务注册/发现 | etcd |
| 认证 | JWT (HS256) |
| 部署 | Docker + Nginx |

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
# 使用 Docker 启动单节点 etcd
docker run -d --name etcd \
  -p 2379:2379 \
  -e ALLOW_NONE_AUTHENTICATION=yes \
  bitnami/etcd:3.5
```

2. 分别启动三个服务：

```bash
# 终端 1 - 用户服务
make run-user

# 终端 2 - 文章服务
make run-article

# 终端 3 - API 网关
make run-api
```

### 重新生成 Kitex 代码

```bash
make gen
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

# 创建文章（需要 token）
curl -X POST http://localhost:8080/api/v1/article \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"Hello World","content":"This is my first article."}'

# 获取文章列表
curl http://localhost:8080/api/v1/articles?page=1&page_size=10
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `ETCD_ENDPOINT` | `127.0.0.1:2379` | etcd 地址 |

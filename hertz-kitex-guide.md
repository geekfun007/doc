# Go Hertz & Kitex 详解与实战

本文档详细介绍 CloudWeGo 开源的高性能 HTTP 框架 Hertz 和 RPC 框架 Kitex 的使用方法和最佳实践。

## 目录

1. [框架概述](#框架概述)
2. [Hertz 快速开始](#hertz-快速开始)
3. [Hertz 路由与中间件](#hertz-路由与中间件)
4. [Hertz 请求处理](#hertz-请求处理)
5. [Hertz 高级特性](#hertz-高级特性)
6. [Kitex 快速开始](#kitex-快速开始)
7. [Kitex IDL 与代码生成](#kitex-idl-与代码生成)
8. [Kitex 服务端开发](#kitex-服务端开发)
9. [Kitex 客户端开发](#kitex-客户端开发)
10. [Kitex 高级特性](#kitex-高级特性)
11. [Hertz + Kitex 微服务实战](#hertz--kitex-微服务实战)
12. [性能优化与最佳实践](#性能优化与最佳实践)

---

## 框架概述

### CloudWeGo 生态

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        CloudWeGo 生态                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │   Hertz     │  │   Kitex     │  │   Netpoll   │  │   Volo      │   │
│  │ HTTP 框架   │  │  RPC 框架   │  │  网络库     │  │ Rust 框架   │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │  Thriftgo   │  │   Sonic     │  │  Frugal     │  │  Shmipc     │   │
│  │ Thrift 编译 │  │ JSON 库     │  │ Thrift 编码 │  │  共享内存   │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Hertz vs Kitex

| 特性 | Hertz | Kitex |
|------|-------|-------|
| 类型 | HTTP 框架 | RPC 框架 |
| 协议 | HTTP/1.1, HTTP/2, HTTP/3 | Thrift, gRPC, HTTP |
| 使用场景 | API 网关、Web 服务 | 微服务内部通信 |
| 序列化 | JSON, Protobuf, Form | Thrift, Protobuf |
| 代码生成 | hz 工具 | kitex 工具 |
| 网络层 | Netpoll/go net | Netpoll |

---

## Hertz 快速开始

### 安装

```bash
# 安装 Hertz
go get github.com/cloudwego/hertz

# 安装 hz 代码生成工具
go install github.com/cloudwego/hertz/cmd/hz@latest

# 验证安装
hz --version
```

### 项目初始化

```bash
# 使用 hz 创建新项目
hz new -mod github.com/example/myproject

# 项目结构
myproject/
├── biz/
│   ├── handler/         # 请求处理器
│   ├── model/           # 数据模型
│   └── router/          # 路由注册
├── main.go
├── router.go
├── router_gen.go
├── go.mod
└── go.sum
```

### Hello World

```go
package main

import (
    "context"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
    h := server.Default()

    h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
        c.JSON(consts.StatusOK, map[string]interface{}{
            "message": "pong",
        })
    })

    h.Spin()
}
```

### 服务器配置

```go
package main

import (
    "time"

    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/cloudwego/hertz/pkg/common/config"
    "github.com/cloudwego/hertz/pkg/network/standard"
)

func main() {
    h := server.New(
        // 基本配置
        server.WithHostPorts(":8080"),
        server.WithMaxRequestBodySize(20<<20),          // 20MB
        server.WithReadTimeout(10*time.Second),
        server.WithWriteTimeout(10*time.Second),
        server.WithIdleTimeout(120*time.Second),
        
        // 优雅关闭
        server.WithExitWaitTime(5*time.Second),
        
        // 网络配置
        server.WithTransport(standard.NewTransporter),   // 使用标准库
        // server.WithTransport(netpoll.NewTransporter), // 使用 netpoll（默认）
        
        // HTTP/2
        server.WithH2C(true),
        
        // TLS
        // server.WithTLS(&tls.Config{...}),
        
        // 流式处理
        server.WithStreamBody(true),
        
        // 链路追踪
        server.WithTraceLevel(stats.LevelDetailed),
    )
    
    // 注册路由
    registerRoutes(h)
    
    h.Spin()
}
```

---

## Hertz 路由与中间件

### 路由注册

```go
package main

import (
    "context"
    
    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
    h := server.Default()
    
    // ============ 基本路由 ============
    h.GET("/get", handler)
    h.POST("/post", handler)
    h.PUT("/put", handler)
    h.DELETE("/delete", handler)
    h.PATCH("/patch", handler)
    h.HEAD("/head", handler)
    h.OPTIONS("/options", handler)
    
    // 任意方法
    h.Any("/any", handler)
    
    // ============ 路由参数 ============
    
    // 命名参数
    h.GET("/user/:id", func(ctx context.Context, c *app.RequestContext) {
        id := c.Param("id")
        c.String(consts.StatusOK, "User ID: %s", id)
    })
    
    // 通配符参数
    h.GET("/files/*filepath", func(ctx context.Context, c *app.RequestContext) {
        filepath := c.Param("filepath")
        c.String(consts.StatusOK, "File: %s", filepath)
    })
    
    // ============ 路由组 ============
    
    v1 := h.Group("/api/v1")
    {
        v1.GET("/users", listUsers)
        v1.GET("/users/:id", getUser)
        v1.POST("/users", createUser)
        v1.PUT("/users/:id", updateUser)
        v1.DELETE("/users/:id", deleteUser)
    }
    
    v2 := h.Group("/api/v2")
    v2.Use(authMiddleware()) // 组级中间件
    {
        v2.GET("/users", listUsersV2)
    }
    
    // 嵌套路由组
    admin := h.Group("/admin")
    admin.Use(adminAuthMiddleware())
    {
        users := admin.Group("/users")
        {
            users.GET("/", listAdminUsers)
            users.POST("/", createAdminUser)
        }
    }
    
    h.Spin()
}

func handler(ctx context.Context, c *app.RequestContext) {
    c.String(consts.StatusOK, "OK")
}
```

### 中间件

```go
package middleware

import (
    "context"
    "log"
    "time"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ============ 日志中间件 ============
func Logger() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        start := time.Now()
        path := string(c.Request.URI().Path())
        method := string(c.Request.Method())
        
        // 处理请求
        c.Next(ctx)
        
        // 记录日志
        latency := time.Since(start)
        statusCode := c.Response.StatusCode()
        
        log.Printf("[%s] %s %d %v", method, path, statusCode, latency)
    }
}

// ============ 恢复中间件 ============
func Recovery() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Panic recovered: %v", r)
                c.JSON(consts.StatusInternalServerError, map[string]interface{}{
                    "error": "Internal Server Error",
                })
                c.Abort()
            }
        }()
        c.Next(ctx)
    }
}

// ============ 认证中间件 ============
func Auth() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        token := c.GetHeader("Authorization")
        if len(token) == 0 {
            c.JSON(consts.StatusUnauthorized, map[string]interface{}{
                "error": "Missing authorization token",
            })
            c.Abort()
            return
        }
        
        // 验证 token
        userID, err := validateToken(string(token))
        if err != nil {
            c.JSON(consts.StatusUnauthorized, map[string]interface{}{
                "error": "Invalid token",
            })
            c.Abort()
            return
        }
        
        // 设置用户信息到上下文
        c.Set("user_id", userID)
        c.Next(ctx)
    }
}

// ============ CORS 中间件 ============
func CORS() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        origin := string(c.GetHeader("Origin"))
        
        // 设置 CORS 头
        c.Header("Access-Control-Allow-Origin", origin)
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Max-Age", "86400")
        
        // 处理预检请求
        if string(c.Method()) == "OPTIONS" {
            c.AbortWithStatus(consts.StatusNoContent)
            return
        }
        
        c.Next(ctx)
    }
}

// ============ 限流中间件 ============
func RateLimit(rate int, burst int) app.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(rate), burst)
    
    return func(ctx context.Context, c *app.RequestContext) {
        if !limiter.Allow() {
            c.JSON(consts.StatusTooManyRequests, map[string]interface{}{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next(ctx)
    }
}

// ============ 请求 ID 中间件 ============
func RequestID() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        requestID := c.GetHeader("X-Request-ID")
        if len(requestID) == 0 {
            requestID = []byte(uuid.New().String())
        }
        c.Set("request_id", string(requestID))
        c.Header("X-Request-ID", string(requestID))
        c.Next(ctx)
    }
}

// ============ 使用中间件 ============
func main() {
    h := server.Default()
    
    // 全局中间件
    h.Use(Recovery())
    h.Use(Logger())
    h.Use(RequestID())
    h.Use(CORS())
    
    // 路由组中间件
    api := h.Group("/api")
    api.Use(Auth())
    {
        api.GET("/profile", getProfile)
    }
    
    h.Spin()
}
```

---

## Hertz 请求处理

### 请求参数获取

```go
package handler

import (
    "context"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ============ 路径参数 ============
func GetUser(ctx context.Context, c *app.RequestContext) {
    id := c.Param("id")
    c.JSON(consts.StatusOK, map[string]interface{}{
        "id": id,
    })
}

// ============ 查询参数 ============
func ListUsers(ctx context.Context, c *app.RequestContext) {
    // 获取单个参数
    page := c.DefaultQuery("page", "1")
    pageSize := c.DefaultQuery("page_size", "10")
    
    // 获取参数（不存在返回空字符串）
    keyword := c.Query("keyword")
    
    // 获取所有参数
    queryParams := c.QueryArgs()
    
    // 获取数组参数 ?ids=1&ids=2&ids=3
    ids := c.QueryArgs().PeekMulti("ids")
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "page":      page,
        "page_size": pageSize,
        "keyword":   keyword,
    })
}

// ============ 表单参数 ============
func CreateUser(ctx context.Context, c *app.RequestContext) {
    // 获取表单字段
    name := c.PostForm("name")
    email := c.DefaultPostForm("email", "")
    
    // 获取多值
    tags := c.PostArgs().PeekMulti("tags")
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "name":  name,
        "email": email,
        "tags":  tags,
    })
}

// ============ JSON Body ============
type CreateUserRequest struct {
    Name     string   `json:"name" vd:"len($)>0"`
    Email    string   `json:"email" vd:"email($)"`
    Age      int      `json:"age" vd:"$>=0 && $<=150"`
    Tags     []string `json:"tags"`
}

func CreateUserJSON(ctx context.Context, c *app.RequestContext) {
    var req CreateUserRequest
    
    // 绑定 JSON
    if err := c.BindAndValidate(&req); err != nil {
        c.JSON(consts.StatusBadRequest, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    c.JSON(consts.StatusCreated, map[string]interface{}{
        "message": "User created",
        "data":    req,
    })
}

// ============ 请求头 ============
func GetHeaders(ctx context.Context, c *app.RequestContext) {
    // 获取单个头
    contentType := c.GetHeader("Content-Type")
    authorization := c.GetHeader("Authorization")
    
    // 获取所有头
    c.Request.Header.VisitAll(func(key, value []byte) {
        // 处理每个头
    })
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "content_type":  string(contentType),
        "authorization": string(authorization),
    })
}

// ============ Cookie ============
func GetCookies(ctx context.Context, c *app.RequestContext) {
    // 获取 Cookie
    sessionID := string(c.Cookie("session_id"))
    
    // 设置 Cookie
    c.SetCookie("user_token", "xxx", 3600, "/", "localhost", false, true)
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "session_id": sessionID,
    })
}

// ============ 文件上传 ============
func UploadFile(ctx context.Context, c *app.RequestContext) {
    // 单文件
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(consts.StatusBadRequest, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    // 保存文件
    dst := "./uploads/" + file.Filename
    if err := c.SaveUploadedFile(file, dst); err != nil {
        c.JSON(consts.StatusInternalServerError, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "filename": file.Filename,
        "size":     file.Size,
    })
}

func UploadMultipleFiles(ctx context.Context, c *app.RequestContext) {
    // 多文件
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(consts.StatusBadRequest, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    files := form.File["files"]
    for _, file := range files {
        dst := "./uploads/" + file.Filename
        c.SaveUploadedFile(file, dst)
    }
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "count": len(files),
    })
}
```

### 响应处理

```go
package handler

import (
    "context"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ============ JSON 响应 ============
func JSONResponse(ctx context.Context, c *app.RequestContext) {
    c.JSON(consts.StatusOK, map[string]interface{}{
        "code":    0,
        "message": "success",
        "data": map[string]interface{}{
            "id":   1,
            "name": "Alice",
        },
    })
}

// 结构体响应
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func StructResponse(ctx context.Context, c *app.RequestContext) {
    c.JSON(consts.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data:    map[string]string{"name": "Alice"},
    })
}

// ============ 其他响应格式 ============

// 纯文本
func TextResponse(ctx context.Context, c *app.RequestContext) {
    c.String(consts.StatusOK, "Hello, %s!", "World")
}

// HTML
func HTMLResponse(ctx context.Context, c *app.RequestContext) {
    c.HTML(consts.StatusOK, "index.html", map[string]interface{}{
        "title": "Home",
    })
}

// XML
func XMLResponse(ctx context.Context, c *app.RequestContext) {
    c.XML(consts.StatusOK, map[string]interface{}{
        "message": "success",
    })
}

// 文件下载
func FileDownload(ctx context.Context, c *app.RequestContext) {
    c.File("./files/document.pdf")
}

// 文件流
func FileStream(ctx context.Context, c *app.RequestContext) {
    c.FileAttachment("./files/document.pdf", "download.pdf")
}

// 重定向
func Redirect(ctx context.Context, c *app.RequestContext) {
    c.Redirect(consts.StatusFound, []byte("/new-location"))
}

// ============ 响应头 ============
func CustomHeaders(ctx context.Context, c *app.RequestContext) {
    c.Header("X-Custom-Header", "value")
    c.Header("X-Request-ID", "12345")
    c.JSON(consts.StatusOK, map[string]interface{}{
        "message": "success",
    })
}

// ============ 流式响应 ============
func StreamResponse(ctx context.Context, c *app.RequestContext) {
    c.SetStatusCode(consts.StatusOK)
    c.Response.Header.Set("Content-Type", "text/event-stream")
    c.Response.Header.Set("Cache-Control", "no-cache")
    
    c.SetBodyStreamWriter(func(w *bufio.Writer) {
        for i := 0; i < 10; i++ {
            fmt.Fprintf(w, "data: Message %d\n\n", i)
            w.Flush()
            time.Sleep(time.Second)
        }
    })
}

// ============ 统一响应封装 ============
type APIResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func Success(c *app.RequestContext, data interface{}) {
    c.JSON(consts.StatusOK, APIResponse{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

func Error(c *app.RequestContext, code int, message string) {
    c.JSON(code, APIResponse{
        Code:    code,
        Message: "error",
        Error:   message,
    })
}

func BadRequest(c *app.RequestContext, message string) {
    Error(c, consts.StatusBadRequest, message)
}

func Unauthorized(c *app.RequestContext, message string) {
    Error(c, consts.StatusUnauthorized, message)
}

func InternalError(c *app.RequestContext, message string) {
    Error(c, consts.StatusInternalServerError, message)
}
```

### 参数验证

```go
package handler

import (
    "context"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// 使用 vd 标签进行验证
type CreateUserRequest struct {
    Name     string `json:"name" vd:"len($)>0; msg:'name is required'"`
    Email    string `json:"email" vd:"email($); msg:'invalid email format'"`
    Age      int    `json:"age" vd:"$>=0 && $<=150; msg:'age must be between 0 and 150'"`
    Password string `json:"password" vd:"len($)>=6 && len($)<=20; msg:'password must be 6-20 characters'"`
    Phone    string `json:"phone" vd:"regexp('^1[3-9]\\d{9}$'); msg:'invalid phone number'"`
    Status   string `json:"status" vd:"in($, 'active', 'inactive'); msg:'status must be active or inactive'"`
}

func CreateUser(ctx context.Context, c *app.RequestContext) {
    var req CreateUserRequest
    
    if err := c.BindAndValidate(&req); err != nil {
        c.JSON(consts.StatusBadRequest, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    // 处理请求...
    c.JSON(consts.StatusCreated, map[string]interface{}{
        "message": "User created",
    })
}

// ============ 验证标签说明 ============
/*
vd 验证标签语法：

比较运算符：
  $>0, $>=0, $<100, $<=100, $==0, $!=0

长度验证：
  len($)>0          - 字符串/数组长度大于 0
  len($)>=6         - 长度至少 6
  len($)<=20        - 长度最多 20

正则验证：
  regexp('pattern') - 正则匹配

枚举验证：
  in($, 'a', 'b')   - 值在枚举中

邮箱验证：
  email($)          - 邮箱格式

组合验证：
  $>0 && $<100      - AND 组合
  $==0 || $>10      - OR 组合

自定义消息：
  ; msg:'错误消息'  - 自定义错误消息

嵌套验证：
  dive              - 验证数组/切片中的每个元素
*/

type OrderRequest struct {
    UserID uint   `json:"user_id" vd:"$>0"`
    Items  []Item `json:"items" vd:"len($)>0; dive"`  // dive 验证每个元素
}

type Item struct {
    ProductID uint `json:"product_id" vd:"$>0"`
    Quantity  int  `json:"quantity" vd:"$>0 && $<=100"`
}
```

---

## Hertz 高级特性

### 使用 IDL 生成代码

```bash
# 定义 Thrift IDL
# idl/api.thrift

namespace go api

struct User {
    1: i64 id
    2: string name
    3: string email
}

struct CreateUserRequest {
    1: string name (api.body="name", api.vd="len($)>0")
    2: string email (api.body="email", api.vd="email($)")
}

struct CreateUserResponse {
    1: i64 code
    2: string message
    3: User user
}

service UserService {
    CreateUserResponse CreateUser(1: CreateUserRequest req) (
        api.post="/api/users"
    )
}

# 生成代码
hz new -idl idl/api.thrift -mod github.com/example/myproject

# 更新代码
hz update -idl idl/api.thrift
```

### WebSocket

```go
package main

import (
    "context"
    "log"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/hertz-contrib/websocket"
)

var upgrader = websocket.HertzUpgrader{
    CheckOrigin: func(c *app.RequestContext) bool {
        return true
    },
}

func main() {
    h := server.Default()

    h.GET("/ws", func(ctx context.Context, c *app.RequestContext) {
        err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
            defer conn.Close()
            
            for {
                // 读取消息
                messageType, message, err := conn.ReadMessage()
                if err != nil {
                    log.Println("read error:", err)
                    break
                }
                
                log.Printf("received: %s", message)
                
                // 发送消息
                err = conn.WriteMessage(messageType, message)
                if err != nil {
                    log.Println("write error:", err)
                    break
                }
            }
        })
        
        if err != nil {
            log.Println("upgrade error:", err)
        }
    })

    h.Spin()
}
```

### HTTP/2 与 TLS

```go
package main

import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/cloudwego/hertz/pkg/network/standard"
)

func main() {
    // HTTP/2 with TLS
    h := server.Default(
        server.WithHostPorts(":443"),
        server.WithTLS(&tls.Config{
            Certificates: []tls.Certificate{cert},
        }),
        server.WithALPN(true),
        server.WithTransport(standard.NewTransporter),
    )
    
    // HTTP/2 without TLS (H2C)
    h := server.Default(
        server.WithHostPorts(":8080"),
        server.WithH2C(true),
    )
    
    h.Spin()
}
```

### 客户端

```go
package main

import (
    "context"
    "time"

    "github.com/cloudwego/hertz/pkg/app/client"
    "github.com/cloudwego/hertz/pkg/protocol"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
    // 创建客户端
    c, err := client.NewClient()
    if err != nil {
        panic(err)
    }
    
    // GET 请求
    status, body, err := c.Get(context.Background(), nil, "http://localhost:8080/ping")
    
    // POST 请求
    req := &protocol.Request{}
    resp := &protocol.Response{}
    
    req.SetMethod(consts.MethodPost)
    req.SetRequestURI("http://localhost:8080/api/users")
    req.Header.SetContentTypeBytes([]byte("application/json"))
    req.SetBody([]byte(`{"name": "Alice", "email": "alice@example.com"}`))
    
    err = c.Do(context.Background(), req, resp)
    
    // 带超时的请求
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    status, body, err = c.Get(ctx, nil, "http://localhost:8080/ping")
    
    // 客户端配置
    c, err = client.NewClient(
        client.WithDialTimeout(5*time.Second),
        client.WithMaxConnsPerHost(100),
        client.WithMaxIdleConnDuration(30*time.Second),
        client.WithKeepAlive(true),
    )
}
```

---

## Kitex 快速开始

### 安装

```bash
# 安装 Kitex
go install github.com/cloudwego/kitex/tool/cmd/kitex@latest

# 安装 Thriftgo
go install github.com/cloudwego/thriftgo@latest

# 验证安装
kitex --version
thriftgo --version
```

### IDL 定义

```thrift
// idl/user.thrift
namespace go user

struct User {
    1: i64 id
    2: string name
    3: string email
    4: i32 age
}

struct CreateUserRequest {
    1: string name
    2: string email
    3: i32 age
}

struct CreateUserResponse {
    1: bool success
    2: string message
    3: User user
}

struct GetUserRequest {
    1: i64 id
}

struct GetUserResponse {
    1: bool success
    2: string message
    3: User user
}

struct ListUsersRequest {
    1: i32 page
    2: i32 page_size
}

struct ListUsersResponse {
    1: bool success
    2: string message
    3: list<User> users
    4: i64 total
}

service UserService {
    CreateUserResponse CreateUser(1: CreateUserRequest req)
    GetUserResponse GetUser(1: GetUserRequest req)
    ListUsersResponse ListUsers(1: ListUsersRequest req)
}
```

### 生成代码

```bash
# 生成服务端代码
kitex -module github.com/example/user-service -service user idl/user.thrift

# 生成客户端代码
kitex -module github.com/example/user-client idl/user.thrift

# 项目结构
user-service/
├── build.sh
├── handler.go           # 业务逻辑
├── main.go
├── kitex_gen/           # 生成的代码
│   └── user/
│       ├── user.go
│       ├── userservice/
│       │   ├── client.go
│       │   ├── invoker.go
│       │   └── server.go
│       └── k-user.go
├── kitex_info.yaml
└── script/
```

---

## Kitex 服务端开发

### 实现 Handler

```go
// handler.go
package main

import (
    "context"
    "sync"
    "sync/atomic"

    user "github.com/example/user-service/kitex_gen/user"
)

// UserServiceImpl 实现 UserService 接口
type UserServiceImpl struct {
    users   sync.Map
    counter int64
}

// CreateUser 创建用户
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
    // 生成 ID
    id := atomic.AddInt64(&s.counter, 1)
    
    // 创建用户
    newUser := &user.User{
        Id:    id,
        Name:  req.Name,
        Email: req.Email,
        Age:   req.Age,
    }
    
    // 存储
    s.users.Store(id, newUser)
    
    return &user.CreateUserResponse{
        Success: true,
        Message: "User created successfully",
        User:    newUser,
    }, nil
}

// GetUser 获取用户
func (s *UserServiceImpl) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
    value, ok := s.users.Load(req.Id)
    if !ok {
        return &user.GetUserResponse{
            Success: false,
            Message: "User not found",
        }, nil
    }
    
    return &user.GetUserResponse{
        Success: true,
        Message: "Success",
        User:    value.(*user.User),
    }, nil
}

// ListUsers 获取用户列表
func (s *UserServiceImpl) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
    var users []*user.User
    var total int64
    
    s.users.Range(func(key, value interface{}) bool {
        users = append(users, value.(*user.User))
        total++
        return true
    })
    
    // 分页
    start := (req.Page - 1) * req.PageSize
    end := start + req.PageSize
    if start > int32(len(users)) {
        start = int32(len(users))
    }
    if end > int32(len(users)) {
        end = int32(len(users))
    }
    
    return &user.ListUsersResponse{
        Success: true,
        Message: "Success",
        Users:   users[start:end],
        Total:   total,
    }, nil
}
```

### 服务端启动

```go
// main.go
package main

import (
    "log"
    "net"

    "github.com/cloudwego/kitex/pkg/rpcinfo"
    "github.com/cloudwego/kitex/server"
    user "github.com/example/user-service/kitex_gen/user/userservice"
)

func main() {
    addr, _ := net.ResolveTCPAddr("tcp", ":8888")
    
    svr := user.NewServer(
        &UserServiceImpl{},
        server.WithServiceAddr(addr),
        server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
            ServiceName: "user-service",
        }),
        // 更多配置...
    )
    
    if err := svr.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### 服务端配置

```go
package main

import (
    "time"

    "github.com/cloudwego/kitex/pkg/limit"
    "github.com/cloudwego/kitex/pkg/rpcinfo"
    "github.com/cloudwego/kitex/server"
    user "github.com/example/user-service/kitex_gen/user/userservice"
)

func main() {
    svr := user.NewServer(
        &UserServiceImpl{},
        
        // 基本配置
        server.WithServiceAddr(addr),
        server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
            ServiceName: "user-service",
        }),
        
        // 超时配置
        server.WithReadWriteTimeout(5*time.Second),
        server.WithExitWaitTime(5*time.Second),
        
        // 连接限制
        server.WithMaxConnIdleTime(60*time.Second),
        server.WithLimit(&limit.Option{
            MaxConnections: 10000,
            MaxQPS:         5000,
        }),
        
        // 中间件
        server.WithMiddleware(loggingMiddleware),
        server.WithMiddleware(recoveryMiddleware),
        
        // 服务注册
        // server.WithRegistry(registry),
        
        // 链路追踪
        // server.WithTracer(tracer),
    )
    
    svr.Run()
}

// 日志中间件
func loggingMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
    return func(ctx context.Context, req, resp interface{}) error {
        ri := rpcinfo.GetRPCInfo(ctx)
        log.Printf("Method: %s, From: %s", ri.To().Method(), ri.From().ServiceName())
        
        start := time.Now()
        err := next(ctx, req, resp)
        log.Printf("Cost: %v, Error: %v", time.Since(start), err)
        
        return err
    }
}

// 恢复中间件
func recoveryMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
    return func(ctx context.Context, req, resp interface{}) (err error) {
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("panic: %v", r)
            }
        }()
        return next(ctx, req, resp)
    }
}
```

---

## Kitex 客户端开发

### 基本调用

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/cloudwego/kitex/client"
    user "github.com/example/user-service/kitex_gen/user"
    userservice "github.com/example/user-service/kitex_gen/user/userservice"
)

func main() {
    // 创建客户端
    c, err := userservice.NewClient(
        "user-service",
        client.WithHostPorts("127.0.0.1:8888"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建用户
    createResp, err := c.CreateUser(context.Background(), &user.CreateUserRequest{
        Name:  "Alice",
        Email: "alice@example.com",
        Age:   25,
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Created user: %+v", createResp.User)
    
    // 获取用户
    getResp, err := c.GetUser(context.Background(), &user.GetUserRequest{
        Id: createResp.User.Id,
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Got user: %+v", getResp.User)
    
    // 获取用户列表
    listResp, err := c.ListUsers(context.Background(), &user.ListUsersRequest{
        Page:     1,
        PageSize: 10,
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Total users: %d", listResp.Total)
}
```

### 客户端配置

```go
package main

import (
    "time"

    "github.com/cloudwego/kitex/client"
    "github.com/cloudwego/kitex/pkg/connpool"
    "github.com/cloudwego/kitex/pkg/retry"
    "github.com/cloudwego/kitex/pkg/rpcinfo"
    userservice "github.com/example/user-service/kitex_gen/user/userservice"
)

func NewUserClient() (userservice.Client, error) {
    return userservice.NewClient(
        "user-service",
        
        // 地址配置
        client.WithHostPorts("127.0.0.1:8888"),
        // 或使用服务发现
        // client.WithResolver(resolver),
        
        // 超时配置
        client.WithRPCTimeout(3*time.Second),
        client.WithConnectTimeout(500*time.Millisecond),
        
        // 连接池配置
        client.WithLongConnection(connpool.IdleConfig{
            MaxIdlePerAddress: 10,
            MaxIdleGlobal:     100,
            MaxIdleTimeout:    60*time.Second,
        }),
        
        // 重试配置
        client.WithFailureRetry(retry.NewFailurePolicy()),
        client.WithRetryMethodPolicies(map[string]retry.Policy{
            "GetUser": retry.BuildFailurePolicy(retry.NewFailurePolicy()),
        }),
        
        // 熔断配置
        // client.WithCircuitBreaker(cb),
        
        // 负载均衡
        // client.WithLoadBalancer(lb),
        
        // 中间件
        client.WithMiddleware(clientLoggingMiddleware),
        
        // 客户端信息
        client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
            ServiceName: "api-gateway",
        }),
    )
}

// 客户端日志中间件
func clientLoggingMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
    return func(ctx context.Context, req, resp interface{}) error {
        ri := rpcinfo.GetRPCInfo(ctx)
        log.Printf("Calling: %s.%s", ri.To().ServiceName(), ri.To().Method())
        
        start := time.Now()
        err := next(ctx, req, resp)
        log.Printf("Response time: %v", time.Since(start))
        
        return err
    }
}
```

### 泛化调用

```go
package main

import (
    "context"

    "github.com/cloudwego/kitex/client/genericclient"
    "github.com/cloudwego/kitex/pkg/generic"
)

func main() {
    // 加载 IDL
    p, err := generic.NewThriftFileProvider("./idl/user.thrift")
    if err != nil {
        panic(err)
    }
    
    // 创建泛化客户端（JSON）
    g, err := generic.JSONThriftGeneric(p)
    if err != nil {
        panic(err)
    }
    
    cli, err := genericclient.NewClient(
        "user-service",
        g,
        client.WithHostPorts("127.0.0.1:8888"),
    )
    if err != nil {
        panic(err)
    }
    
    // 泛化调用
    resp, err := cli.GenericCall(context.Background(), "CreateUser", `{
        "name": "Alice",
        "email": "alice@example.com",
        "age": 25
    }`)
    if err != nil {
        panic(err)
    }
    
    // resp 是 JSON 字符串
    log.Printf("Response: %s", resp)
}
```

---

## Kitex 高级特性

### 服务注册与发现

```go
// 使用 etcd 作为注册中心
import (
    "github.com/cloudwego/kitex/pkg/registry"
    "github.com/kitex-contrib/registry-etcd"
)

// 服务端注册
func main() {
    r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"})
    if err != nil {
        panic(err)
    }
    
    svr := userservice.NewServer(
        &UserServiceImpl{},
        server.WithRegistry(r),
        server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
            ServiceName: "user-service",
        }),
    )
    svr.Run()
}

// 客户端发现
func NewUserClient() (userservice.Client, error) {
    r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
    if err != nil {
        return nil, err
    }
    
    return userservice.NewClient(
        "user-service",
        client.WithResolver(r),
        client.WithLoadBalancer(loadbalance.NewWeightedRoundRobinBalancer()),
    )
}

// 使用 Nacos
import nacos "github.com/kitex-contrib/registry-nacos/registry"

r, err := nacos.NewDefaultNacosRegistry()
svr := userservice.NewServer(&UserServiceImpl{}, server.WithRegistry(r))
```

### 链路追踪

```go
// 使用 OpenTelemetry
import (
    "github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {
    // 初始化 tracer
    provider := tracing.NewServerProvider(
        tracing.WithServiceName("user-service"),
        tracing.WithExportEndpoint("localhost:4317"),
    )
    defer provider.Shutdown(context.Background())
    
    svr := userservice.NewServer(
        &UserServiceImpl{},
        server.WithSuite(tracing.NewServerSuite()),
    )
    svr.Run()
}

// 客户端
func NewUserClient() (userservice.Client, error) {
    provider := tracing.NewClientProvider(
        tracing.WithServiceName("api-gateway"),
        tracing.WithExportEndpoint("localhost:4317"),
    )
    
    return userservice.NewClient(
        "user-service",
        client.WithSuite(tracing.NewClientSuite()),
    )
}
```

### 熔断与限流

```go
// 熔断
import (
    "github.com/cloudwego/kitex/pkg/circuitbreak"
)

// 客户端熔断
cbSuite := circuitbreak.NewCBSuite(circuitbreak.RPCInfo2Key)
cli, _ := userservice.NewClient(
    "user-service",
    client.WithCircuitBreaker(cbSuite),
)

// 服务端限流
import (
    "github.com/cloudwego/kitex/pkg/limit"
)

svr := userservice.NewServer(
    &UserServiceImpl{},
    server.WithLimit(&limit.Option{
        MaxConnections: 10000,
        MaxQPS:         5000,
    }),
    server.WithMuxTransport(), // 使用多路复用
)
```

---

## Hertz + Kitex 微服务实战

### 项目架构

```
microservices/
├── api-gateway/              # Hertz API 网关
│   ├── main.go
│   ├── handler/
│   ├── middleware/
│   └── router/
├── user-service/             # Kitex 用户服务
│   ├── main.go
│   ├── handler.go
│   └── kitex_gen/
├── order-service/            # Kitex 订单服务
│   ├── main.go
│   ├── handler.go
│   └── kitex_gen/
├── idl/                      # IDL 定义
│   ├── user.thrift
│   └── order.thrift
└── pkg/                      # 公共包
    ├── config/
    ├── logger/
    └── response/
```

### API 网关 (Hertz)

```go
// api-gateway/main.go
package main

import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/example/api-gateway/handler"
    "github.com/example/api-gateway/middleware"
)

func main() {
    h := server.Default(server.WithHostPorts(":8080"))
    
    // 全局中间件
    h.Use(middleware.Recovery())
    h.Use(middleware.Logger())
    h.Use(middleware.CORS())
    h.Use(middleware.RequestID())
    
    // API 路由
    api := h.Group("/api")
    {
        // 用户服务
        users := api.Group("/users")
        {
            users.POST("/", handler.CreateUser)
            users.GET("/:id", handler.GetUser)
            users.GET("/", handler.ListUsers)
        }
        
        // 订单服务
        orders := api.Group("/orders")
        orders.Use(middleware.Auth())
        {
            orders.POST("/", handler.CreateOrder)
            orders.GET("/:id", handler.GetOrder)
            orders.GET("/", handler.ListOrders)
        }
    }
    
    h.Spin()
}
```

```go
// api-gateway/handler/user.go
package handler

import (
    "context"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
    "github.com/example/api-gateway/rpc"
    user "github.com/example/user-service/kitex_gen/user"
)

type CreateUserReq struct {
    Name  string `json:"name" vd:"len($)>0"`
    Email string `json:"email" vd:"email($)"`
    Age   int32  `json:"age" vd:"$>=0 && $<=150"`
}

func CreateUser(ctx context.Context, c *app.RequestContext) {
    var req CreateUserReq
    if err := c.BindAndValidate(&req); err != nil {
        c.JSON(consts.StatusBadRequest, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    // 调用 Kitex 服务
    resp, err := rpc.UserClient.CreateUser(ctx, &user.CreateUserRequest{
        Name:  req.Name,
        Email: req.Email,
        Age:   req.Age,
    })
    if err != nil {
        c.JSON(consts.StatusInternalServerError, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    if !resp.Success {
        c.JSON(consts.StatusBadRequest, map[string]interface{}{
            "error": resp.Message,
        })
        return
    }
    
    c.JSON(consts.StatusCreated, map[string]interface{}{
        "code":    0,
        "message": "success",
        "data":    resp.User,
    })
}

func GetUser(ctx context.Context, c *app.RequestContext) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(consts.StatusBadRequest, map[string]interface{}{
            "error": "invalid user id",
        })
        return
    }
    
    resp, err := rpc.UserClient.GetUser(ctx, &user.GetUserRequest{
        Id: id,
    })
    if err != nil {
        c.JSON(consts.StatusInternalServerError, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    if !resp.Success {
        c.JSON(consts.StatusNotFound, map[string]interface{}{
            "error": resp.Message,
        })
        return
    }
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "code":    0,
        "message": "success",
        "data":    resp.User,
    })
}

func ListUsers(ctx context.Context, c *app.RequestContext) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
    
    resp, err := rpc.UserClient.ListUsers(ctx, &user.ListUsersRequest{
        Page:     int32(page),
        PageSize: int32(pageSize),
    })
    if err != nil {
        c.JSON(consts.StatusInternalServerError, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "code":    0,
        "message": "success",
        "data": map[string]interface{}{
            "users": resp.Users,
            "total": resp.Total,
            "page":  page,
        },
    })
}
```

```go
// api-gateway/rpc/client.go
package rpc

import (
    "log"
    "sync"
    "time"

    "github.com/cloudwego/kitex/client"
    "github.com/cloudwego/kitex/pkg/connpool"
    "github.com/cloudwego/kitex/pkg/retry"
    etcd "github.com/kitex-contrib/registry-etcd"
    userservice "github.com/example/user-service/kitex_gen/user/userservice"
    orderservice "github.com/example/order-service/kitex_gen/order/orderservice"
)

var (
    UserClient  userservice.Client
    OrderClient orderservice.Client
    once        sync.Once
)

func Init() {
    once.Do(func() {
        initUserClient()
        initOrderClient()
    })
}

func initUserClient() {
    r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
    if err != nil {
        log.Fatal(err)
    }
    
    UserClient, err = userservice.NewClient(
        "user-service",
        client.WithResolver(r),
        client.WithRPCTimeout(3*time.Second),
        client.WithConnectTimeout(500*time.Millisecond),
        client.WithFailureRetry(retry.NewFailurePolicy()),
        client.WithLongConnection(connpool.IdleConfig{
            MaxIdlePerAddress: 10,
            MaxIdleGlobal:     100,
            MaxIdleTimeout:    60*time.Second,
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
}

func initOrderClient() {
    r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
    if err != nil {
        log.Fatal(err)
    }
    
    OrderClient, err = orderservice.NewClient(
        "order-service",
        client.WithResolver(r),
        client.WithRPCTimeout(3*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### 用户服务 (Kitex)

```go
// user-service/main.go
package main

import (
    "log"
    "net"

    "github.com/cloudwego/kitex/pkg/rpcinfo"
    "github.com/cloudwego/kitex/server"
    etcd "github.com/kitex-contrib/registry-etcd"
    user "github.com/example/user-service/kitex_gen/user/userservice"
)

func main() {
    // 注册中心
    r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"})
    if err != nil {
        log.Fatal(err)
    }
    
    addr, _ := net.ResolveTCPAddr("tcp", ":8888")
    
    svr := user.NewServer(
        &UserServiceImpl{},
        server.WithServiceAddr(addr),
        server.WithRegistry(r),
        server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
            ServiceName: "user-service",
        }),
    )
    
    if err := svr.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Docker Compose 部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  etcd:
    image: bitnami/etcd:latest
    ports:
      - "2379:2379"
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379

  api-gateway:
    build: ./api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - etcd
      - user-service
      - order-service
    environment:
      - ETCD_ENDPOINTS=etcd:2379

  user-service:
    build: ./user-service
    ports:
      - "8888:8888"
    depends_on:
      - etcd
    environment:
      - ETCD_ENDPOINTS=etcd:2379

  order-service:
    build: ./order-service
    ports:
      - "8889:8889"
    depends_on:
      - etcd
      - user-service
    environment:
      - ETCD_ENDPOINTS=etcd:2379
```

---

## 性能优化与最佳实践

### Hertz 性能优化

```go
// 1. 使用 Netpoll（默认）
h := server.Default() // 默认使用 netpoll

// 2. 开启预读
server.WithReadBufferSize(4096)

// 3. 复用 JSON encoder
import "github.com/cloudwego/hertz/pkg/common/json"
// Hertz 默认使用 sonic

// 4. 流式处理大请求
server.WithStreamBody(true)

// 5. 合理设置超时
server.WithReadTimeout(10*time.Second)
server.WithWriteTimeout(10*time.Second)

// 6. 使用对象池
var responsePool = sync.Pool{
    New: func() interface{} {
        return &Response{}
    },
}

func handler(ctx context.Context, c *app.RequestContext) {
    resp := responsePool.Get().(*Response)
    defer responsePool.Put(resp)
    // 使用 resp
}
```

### Kitex 性能优化

```go
// 1. 使用多路复用
server.WithMuxTransport()
client.WithMuxConnection(2)

// 2. 连接池配置
client.WithLongConnection(connpool.IdleConfig{
    MaxIdlePerAddress: 10,
    MaxIdleGlobal:     100,
    MaxIdleTimeout:    60*time.Second,
})

// 3. 合理设置超时
client.WithRPCTimeout(3*time.Second)
client.WithConnectTimeout(500*time.Millisecond)

// 4. 使用 Frugal（高性能 Thrift 编解码）
// 需要安装 frugal 插件
kitex -thrift frugal ...

// 5. 服务端限流
server.WithLimit(&limit.Option{
    MaxConnections: 10000,
    MaxQPS:         5000,
})

// 6. 熔断配置
client.WithCircuitBreaker(circuitbreak.NewCBSuite(circuitbreak.RPCInfo2Key))
```

### 最佳实践

```go
// 1. 统一错误处理
type BizError struct {
    Code    int32
    Message string
}

func (e *BizError) Error() string {
    return e.Message
}

// 2. 上下文传递
func (s *UserServiceImpl) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
    // 从上下文获取追踪信息
    span := trace.SpanFromContext(ctx)
    span.SetAttributes(attribute.Int64("user.id", req.Id))
    
    // 业务逻辑
    return &user.GetUserResponse{}, nil
}

// 3. 优雅关闭
func main() {
    // 监听信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-quit
        // 优雅关闭
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        svr.Shutdown(ctx)
    }()
    
    svr.Run()
}

// 4. 健康检查
h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
    c.JSON(consts.StatusOK, map[string]string{"status": "ok"})
})

// 5. 配置管理
type Config struct {
    Server struct {
        Port string `yaml:"port"`
    } `yaml:"server"`
    Etcd struct {
        Endpoints []string `yaml:"endpoints"`
    } `yaml:"etcd"`
}
```

---

## 总结

### Hertz 速查

```go
// 创建服务器
h := server.Default(server.WithHostPorts(":8080"))

// 路由
h.GET("/path", handler)
h.POST("/path", handler)
h.Group("/api").Use(middleware)

// 请求处理
c.Param("id")           // 路径参数
c.Query("key")          // 查询参数
c.PostForm("field")     // 表单字段
c.BindAndValidate(&req) // 绑定并验证

// 响应
c.JSON(200, data)
c.String(200, "text")
c.Redirect(302, []byte("/new"))
```

### Kitex 速查

```bash
# 生成代码
kitex -module xxx -service svc idl/xxx.thrift

# 服务端
svr := xxxservice.NewServer(&Handler{}, server.WithServiceAddr(addr))
svr.Run()

# 客户端
cli, _ := xxxservice.NewClient("svc", client.WithHostPorts("addr"))
resp, err := cli.Method(ctx, req)
```

### 技术选型建议

| 场景 | 推荐方案 |
|------|---------|
| API 网关 | Hertz |
| 微服务内部通信 | Kitex |
| 对外 HTTP API | Hertz |
| 高性能 RPC | Kitex + Thrift |
| 跨语言通信 | Kitex + gRPC |

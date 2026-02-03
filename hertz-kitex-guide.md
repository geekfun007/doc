# Go Hertz & Kitex 详解与实战

本文档详细介绍 CloudWeGo 开源的高性能 HTTP 框架 Hertz 和 RPC 框架 Kitex 的使用方法和最佳实践。

## 目录

1. [框架概述](#框架概述)
2. [Hertz 快速开始](#hertz-快速开始)
3. [Hertz 路由与中间件](#hertz-路由与中间件)
4. [Hertz 请求处理](#hertz-请求处理)
5. [参数校验详解](#参数校验详解)
6. [Hertz 高级特性](#hertz-高级特性)
7. [Kitex 快速开始](#kitex-快速开始)
8. [Kitex IDL 与代码生成](#kitex-idl-与代码生成)
9. [Kitex 服务端开发](#kitex-服务端开发)
10. [Kitex 客户端开发](#kitex-客户端开发)
11. [Kitex 高级特性](#kitex-高级特性)
12. [GORM MySQL 数据库集成](#gorm-mysql-数据库集成)
13. [Hertz + Kitex 微服务实战](#hertz--kitex-微服务实战)
14. [性能优化与最佳实践](#性能优化与最佳实践)

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
```

---

## 参数校验详解

本节详细介绍 Hertz 和 Kitex 中的参数校验机制，包括内置验证器、自定义验证规则、错误处理等。

### Hertz 参数校验

#### vd 验证器基础

Hertz 使用 `go-tagexpr/v2` 作为默认验证器，通过 `vd` 标签定义验证规则。

```go
import (
    "github.com/cloudwego/hertz/pkg/app/server/binding"
)

// ============ vd 标签完整语法 ============

type UserRequest struct {
    // 必填验证
    Name string `json:"name" vd:"len($)>0; msg:'姓名不能为空'"`
    
    // 数值范围
    Age    int     `json:"age" vd:"$>=0 && $<=150; msg:'年龄必须在 0-150 之间'"`
    Score  float64 `json:"score" vd:"$>=0.0 && $<=100.0; msg:'分数范围 0-100'"`
    Amount int64   `json:"amount" vd:"$>0; msg:'金额必须大于 0'"`
    
    // 字符串长度
    Username string `json:"username" vd:"len($)>=3 && len($)<=20; msg:'用户名长度 3-20'"`
    Bio      string `json:"bio" vd:"len($)<=500; msg:'简介最多 500 字符'"`
    
    // 正则匹配
    Phone    string `json:"phone" vd:"regexp('^1[3-9]\\d{9}$'); msg:'手机号格式错误'"`
    IDCard   string `json:"id_card" vd:"regexp('^\\d{17}[\\dXx]$'); msg:'身份证格式错误'"`
    PostCode string `json:"post_code" vd:"regexp('^\\d{6}$'); msg:'邮编格式错误'"`
    
    // 内置函数
    Email    string `json:"email" vd:"email($); msg:'邮箱格式错误'"`
    
    // 枚举值
    Gender string `json:"gender" vd:"in($, 'male', 'female', 'other'); msg:'性别值无效'"`
    Status int    `json:"status" vd:"in($, 0, 1, 2); msg:'状态值必须是 0/1/2'"`
    
    // 可选字段（允许零值）
    Nickname string `json:"nickname" vd:"len($)==0 || (len($)>=2 && len($)<=30); msg:'昵称长度 2-30'"`
    
    // 条件验证
    Password        string `json:"password" vd:"len($)>=8; msg:'密码至少 8 位'"`
    ConfirmPassword string `json:"confirm_password" vd:"$==Password; msg:'两次密码不一致'"`
}
```

#### vd 验证器运算符

```go
// ============ 比较运算符 ============
type CompareExample struct {
    A int `vd:"$>0"`           // 大于
    B int `vd:"$>=0"`          // 大于等于
    C int `vd:"$<100"`         // 小于
    D int `vd:"$<=100"`        // 小于等于
    E int `vd:"$==10"`         // 等于
    F int `vd:"$!=0"`          // 不等于
}

// ============ 逻辑运算符 ============
type LogicExample struct {
    // AND：两个条件都满足
    Age int `vd:"$>=18 && $<=60; msg:'年龄必须在 18-60 之间'"`
    
    // OR：满足任一条件
    Type string `vd:"$=='vip' || $=='normal'; msg:'类型无效'"`
    
    // NOT：取反
    Status int `vd:"!($==0); msg:'状态不能为 0'"`
    
    // 复合条件
    Level int `vd:"($>=1 && $<=10) || $==99; msg:'等级 1-10 或 99'"`
}

// ============ 内置函数 ============
type FunctionExample struct {
    // len() - 长度
    Name  string   `vd:"len($)>0"`
    Items []string `vd:"len($)>0 && len($)<=10"`
    
    // email() - 邮箱格式
    Email string `vd:"email($)"`
    
    // regexp() - 正则匹配
    Phone string `vd:"regexp('^1\\d{10}$')"`
    
    // in() - 枚举
    Type string `vd:"in($, 'a', 'b', 'c')"`
    
    // sprintf() - 格式化消息
    Code string `vd:"len($)==6; msg:sprintf('验证码长度必须为 6，当前长度 %d', len($))"`
}
```

#### 嵌套结构与数组验证

```go
// ============ 嵌套结构验证 ============
type Order struct {
    OrderNo string `json:"order_no" vd:"len($)>0; msg:'订单号不能为空'"`
    
    // 嵌套结构自动验证
    User    UserInfo `json:"user"`
    Address Address  `json:"address"`
    
    // 数组/切片验证
    Items []OrderItem `json:"items" vd:"len($)>0 && len($)<=100; msg:'商品数量 1-100'"`
}

type UserInfo struct {
    ID   int64  `json:"id" vd:"$>0; msg:'用户ID无效'"`
    Name string `json:"name" vd:"len($)>0; msg:'用户名不能为空'"`
}

type Address struct {
    Province string `json:"province" vd:"len($)>0; msg:'省份不能为空'"`
    City     string `json:"city" vd:"len($)>0; msg:'城市不能为空'"`
    Detail   string `json:"detail" vd:"len($)>=5 && len($)<=200; msg:'详细地址 5-200 字符'"`
}

type OrderItem struct {
    ProductID int64   `json:"product_id" vd:"$>0; msg:'商品ID无效'"`
    Quantity  int     `json:"quantity" vd:"$>0 && $<=999; msg:'数量 1-999'"`
    Price     float64 `json:"price" vd:"$>0; msg:'价格必须大于 0'"`
}

// ============ Map 验证 ============
type ConfigRequest struct {
    // Map 键值验证
    Settings map[string]string `json:"settings" vd:"len($)>0 && len($)<=50; msg:'配置项 1-50 个'"`
    
    // 复杂 Map
    Metadata map[string]interface{} `json:"metadata"`
}

// ============ 指针字段验证 ============
type UpdateRequest struct {
    // 指针字段：nil 时跳过验证，非 nil 时验证
    Name  *string `json:"name" vd:"@:len($)>=2; msg:'姓名至少 2 字符'"`
    Age   *int    `json:"age" vd:"@:$>=0 && $<=150; msg:'年龄 0-150'"`
    Email *string `json:"email" vd:"@:email($); msg:'邮箱格式错误'"`
}
// 注意：@: 前缀表示仅当字段非 nil 时才验证
```

#### 自定义验证函数

```go
package validator

import (
    "reflect"
    "regexp"
    "unicode"
    
    "github.com/bytedance/go-tagexpr/v2/validator"
)

// ============ 注册自定义验证函数 ============
func init() {
    // 注册自定义函数
    validator.RegFunc("mobile", validateMobile)
    validator.RegFunc("idcard", validateIDCard)
    validator.RegFunc("password_strength", validatePasswordStrength)
    validator.RegFunc("chinese", validateChinese)
    validator.RegFunc("url", validateURL)
    validator.RegFunc("ip", validateIP)
}

// 手机号验证
func validateMobile(args ...interface{}) error {
    if len(args) != 1 {
        return fmt.Errorf("mobile() requires 1 argument")
    }
    
    phone, ok := args[0].(string)
    if !ok {
        return fmt.Errorf("mobile() argument must be string")
    }
    
    if phone == "" {
        return nil // 允许空值，必填用 len($)>0
    }
    
    matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
    if !matched {
        return fmt.Errorf("invalid mobile number")
    }
    return nil
}

// 身份证验证（含校验位）
func validateIDCard(args ...interface{}) error {
    if len(args) != 1 {
        return fmt.Errorf("idcard() requires 1 argument")
    }
    
    idcard, ok := args[0].(string)
    if !ok {
        return fmt.Errorf("idcard() argument must be string")
    }
    
    if idcard == "" {
        return nil
    }
    
    // 18 位身份证校验
    if len(idcard) != 18 {
        return fmt.Errorf("ID card must be 18 digits")
    }
    
    // 校验位计算
    weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
    checkCodes := "10X98765432"
    
    sum := 0
    for i := 0; i < 17; i++ {
        n := int(idcard[i] - '0')
        if n < 0 || n > 9 {
            return fmt.Errorf("invalid ID card format")
        }
        sum += n * weights[i]
    }
    
    checkCode := checkCodes[sum%11]
    if idcard[17] != byte(checkCode) && !(checkCode == 'X' && (idcard[17] == 'x' || idcard[17] == 'X')) {
        return fmt.Errorf("invalid ID card check digit")
    }
    
    return nil
}

// 密码强度验证
func validatePasswordStrength(args ...interface{}) error {
    if len(args) != 1 {
        return fmt.Errorf("password_strength() requires 1 argument")
    }
    
    password, ok := args[0].(string)
    if !ok {
        return fmt.Errorf("password_strength() argument must be string")
    }
    
    if len(password) < 8 {
        return fmt.Errorf("password must be at least 8 characters")
    }
    
    var hasUpper, hasLower, hasNumber, hasSpecial bool
    for _, c := range password {
        switch {
        case unicode.IsUpper(c):
            hasUpper = true
        case unicode.IsLower(c):
            hasLower = true
        case unicode.IsNumber(c):
            hasNumber = true
        case unicode.IsPunct(c) || unicode.IsSymbol(c):
            hasSpecial = true
        }
    }
    
    strength := 0
    if hasUpper {
        strength++
    }
    if hasLower {
        strength++
    }
    if hasNumber {
        strength++
    }
    if hasSpecial {
        strength++
    }
    
    if strength < 3 {
        return fmt.Errorf("password must contain at least 3 of: uppercase, lowercase, number, special character")
    }
    
    return nil
}

// 中文字符验证
func validateChinese(args ...interface{}) error {
    if len(args) != 1 {
        return fmt.Errorf("chinese() requires 1 argument")
    }
    
    str, ok := args[0].(string)
    if !ok {
        return fmt.Errorf("chinese() argument must be string")
    }
    
    for _, r := range str {
        if !unicode.Is(unicode.Han, r) {
            return fmt.Errorf("must be Chinese characters only")
        }
    }
    
    return nil
}

// ============ 使用自定义验证函数 ============
type RegisterRequest struct {
    Phone    string `json:"phone" vd:"mobile($); msg:'手机号格式错误'"`
    IDCard   string `json:"id_card" vd:"idcard($); msg:'身份证号无效'"`
    Password string `json:"password" vd:"password_strength($); msg:'密码强度不够'"`
    RealName string `json:"real_name" vd:"chinese($) && len($)>=2 && len($)<=10; msg:'请输入 2-10 个汉字'"`
}
```

#### 参数绑定与验证方法

```go
package handler

import (
    "context"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ============ 绑定方法对比 ============

func BindingExamples(ctx context.Context, c *app.RequestContext) {
    var req UserRequest
    
    // 方法 1: Bind - 只绑定，不验证
    err := c.Bind(&req)
    
    // 方法 2: Validate - 只验证，不绑定
    err = c.Validate(&req)
    
    // 方法 3: BindAndValidate - 绑定 + 验证（推荐）
    err = c.BindAndValidate(&req)
    
    // 方法 4: BindQuery - 只绑定查询参数
    err = c.BindQuery(&req)
    
    // 方法 5: BindForm - 只绑定表单参数
    err = c.BindForm(&req)
    
    // 方法 6: BindJSON - 只绑定 JSON Body
    err = c.BindJSON(&req)
    
    // 方法 7: BindPath - 只绑定路径参数
    err = c.BindPath(&req)
    
    // 方法 8: BindHeader - 只绑定请求头
    err = c.BindHeader(&req)
}

// ============ 混合参数绑定 ============

type QueryParams struct {
    Page     int    `query:"page" vd:"$>=1; msg:'页码从 1 开始'"`
    PageSize int    `query:"page_size" vd:"$>=1 && $<=100; msg:'每页 1-100 条'"`
    Keyword  string `query:"keyword"`
}

type PathParams struct {
    ID int64 `path:"id" vd:"$>0; msg:'ID 无效'"`
}

type HeaderParams struct {
    Authorization string `header:"Authorization" vd:"len($)>0; msg:'缺少认证信息'"`
    UserAgent     string `header:"User-Agent"`
}

type BodyParams struct {
    Name  string `json:"name" vd:"len($)>0; msg:'名称不能为空'"`
    Email string `json:"email" vd:"email($); msg:'邮箱格式错误'"`
}

// 组合请求结构
type ComplexRequest struct {
    QueryParams
    PathParams
    HeaderParams
    BodyParams
}

func HandleComplexRequest(ctx context.Context, c *app.RequestContext) {
    var req ComplexRequest
    
    // 一次性绑定所有参数
    if err := c.BindAndValidate(&req); err != nil {
        c.JSON(consts.StatusBadRequest, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    // req.Page, req.ID, req.Authorization, req.Name 都已填充
}
```

#### 验证错误处理

```go
package handler

import (
    "context"
    "errors"
    "strings"

    "github.com/bytedance/go-tagexpr/v2/validator"
    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ============ 错误响应结构 ============
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type ErrorResponse struct {
    Code    int               `json:"code"`
    Message string            `json:"message"`
    Errors  []ValidationError `json:"errors,omitempty"`
}

// ============ 解析验证错误 ============
func ParseValidationError(err error) []ValidationError {
    if err == nil {
        return nil
    }
    
    var validationErrors []ValidationError
    
    // go-tagexpr 验证错误格式: "field_name: error message"
    errStr := err.Error()
    parts := strings.Split(errStr, "; ")
    
    for _, part := range parts {
        idx := strings.Index(part, ": ")
        if idx > 0 {
            validationErrors = append(validationErrors, ValidationError{
                Field:   part[:idx],
                Message: part[idx+2:],
            })
        } else {
            validationErrors = append(validationErrors, ValidationError{
                Message: part,
            })
        }
    }
    
    return validationErrors
}

// ============ 统一验证处理 ============
func ValidateRequest(c *app.RequestContext, req interface{}) bool {
    if err := c.BindAndValidate(req); err != nil {
        errors := ParseValidationError(err)
        c.JSON(consts.StatusBadRequest, ErrorResponse{
            Code:    400,
            Message: "参数验证失败",
            Errors:  errors,
        })
        return false
    }
    return true
}

// 使用示例
func CreateUser(ctx context.Context, c *app.RequestContext) {
    var req CreateUserRequest
    
    if !ValidateRequest(c, &req) {
        return
    }
    
    // 业务逻辑...
}

// ============ 自定义验证错误处理器 ============
func SetupValidation() {
    // 自定义绑定错误处理
    binding.SetLooseZeroMode(true) // 允许零值
    
    // 配置验证器
    vd := binding.NewValidator()
    vd.SetErrorFactory(func(failPath, msg string) error {
        return fmt.Errorf("字段 '%s' 验证失败: %s", failPath, msg)
    })
}
```

### Kitex 参数校验

#### IDL 级别验证

```thrift
// idl/user.thrift
namespace go user

// ============ 使用 api 注解定义验证规则 ============
struct CreateUserRequest {
    1: string name (
        api.body = "name",
        api.vd = "len($)>0 && len($)<=50; msg:'姓名 1-50 字符'"
    )
    
    2: string email (
        api.body = "email",
        api.vd = "email($); msg:'邮箱格式错误'"
    )
    
    3: i32 age (
        api.body = "age",
        api.vd = "$>=0 && $<=150; msg:'年龄 0-150'"
    )
    
    4: string phone (
        api.body = "phone",
        api.vd = "regexp('^1[3-9]\\d{9}$'); msg:'手机号格式错误'"
    )
    
    5: string status (
        api.body = "status",
        api.vd = "in($, 'active', 'inactive', 'pending'); msg:'状态值无效'"
    )
    
    6: optional string avatar (
        api.body = "avatar",
        api.vd = "@:len($)<=500; msg:'头像 URL 最长 500 字符'"
    )
}

// 嵌套结构验证
struct Address {
    1: string province (api.body = "province", api.vd = "len($)>0")
    2: string city (api.body = "city", api.vd = "len($)>0")
    3: string detail (api.body = "detail", api.vd = "len($)>=5 && len($)<=200")
}

struct CreateOrderRequest {
    1: i64 user_id (api.body = "user_id", api.vd = "$>0")
    2: Address address (api.body = "address")
    3: list<OrderItem> items (api.body = "items", api.vd = "len($)>0 && len($)<=100")
    4: double total_amount (api.body = "total_amount", api.vd = "$>0")
}

struct OrderItem {
    1: i64 product_id (api.body = "product_id", api.vd = "$>0")
    2: i32 quantity (api.body = "quantity", api.vd = "$>0 && $<=999")
    3: double price (api.body = "price", api.vd = "$>0")
}
```

#### Handler 级别验证

```go
// handler.go
package main

import (
    "context"
    "fmt"
    "regexp"

    user "github.com/example/user-service/kitex_gen/user"
)

// ============ 业务验证器 ============
type UserValidator struct{}

func (v *UserValidator) ValidateCreateUser(req *user.CreateUserRequest) error {
    // 必填验证
    if req.Name == "" {
        return fmt.Errorf("name is required")
    }
    
    // 长度验证
    if len(req.Name) > 50 {
        return fmt.Errorf("name must be less than 50 characters")
    }
    
    // 邮箱格式
    if !isValidEmail(req.Email) {
        return fmt.Errorf("invalid email format")
    }
    
    // 手机号格式
    if req.Phone != "" && !isValidPhone(req.Phone) {
        return fmt.Errorf("invalid phone number")
    }
    
    // 年龄范围
    if req.Age < 0 || req.Age > 150 {
        return fmt.Errorf("age must be between 0 and 150")
    }
    
    // 枚举验证
    validStatus := map[string]bool{"active": true, "inactive": true, "pending": true}
    if !validStatus[req.Status] {
        return fmt.Errorf("invalid status")
    }
    
    return nil
}

func isValidEmail(email string) bool {
    pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    matched, _ := regexp.MatchString(pattern, email)
    return matched
}

func isValidPhone(phone string) bool {
    pattern := `^1[3-9]\d{9}$`
    matched, _ := regexp.MatchString(pattern, phone)
    return matched
}

// ============ 在 Handler 中使用 ============
type UserServiceImpl struct {
    validator *UserValidator
}

func NewUserServiceImpl() *UserServiceImpl {
    return &UserServiceImpl{
        validator: &UserValidator{},
    }
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
    // 参数验证
    if err := s.validator.ValidateCreateUser(req); err != nil {
        return &user.CreateUserResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }
    
    // 业务逻辑...
    return &user.CreateUserResponse{
        Success: true,
        Message: "User created successfully",
    }, nil
}
```

#### 使用 validator 库

```go
package validator

import (
    "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
    validate = validator.New()
    
    // 注册自定义验证器
    validate.RegisterValidation("mobile", validateMobile)
    validate.RegisterValidation("idcard", validateIDCard)
}

// 自定义手机号验证
func validateMobile(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    if phone == "" {
        return true // 空值由 required 处理
    }
    matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
    return matched
}

// 自定义身份证验证
func validateIDCard(fl validator.FieldLevel) bool {
    idcard := fl.Field().String()
    if idcard == "" {
        return true
    }
    // 校验逻辑...
    return len(idcard) == 18
}

// ============ 验证请求结构 ============
type CreateUserReq struct {
    Name     string `validate:"required,min=1,max=50"`
    Email    string `validate:"required,email"`
    Age      int    `validate:"gte=0,lte=150"`
    Phone    string `validate:"omitempty,mobile"`
    Password string `validate:"required,min=8,max=32"`
    Status   string `validate:"required,oneof=active inactive pending"`
}

func ValidateCreateUserReq(req *CreateUserReq) error {
    return validate.Struct(req)
}

// ============ 翻译验证错误 ============
import (
    "github.com/go-playground/locales/zh"
    ut "github.com/go-playground/universal-translator"
    zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
    uni      *ut.UniversalTranslator
    trans    ut.Translator
)

func InitTranslator() {
    zhLocale := zh.New()
    uni = ut.New(zhLocale, zhLocale)
    trans, _ = uni.GetTranslator("zh")
    
    zh_translations.RegisterDefaultTranslations(validate, trans)
    
    // 自定义翻译
    validate.RegisterTranslation("mobile", trans, func(ut ut.Translator) error {
        return ut.Add("mobile", "{0}必须是有效的手机号", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("mobile", fe.Field())
        return t
    })
}

func TranslateError(err error) string {
    if err == nil {
        return ""
    }
    
    errs, ok := err.(validator.ValidationErrors)
    if !ok {
        return err.Error()
    }
    
    var messages []string
    for _, e := range errs {
        messages = append(messages, e.Translate(trans))
    }
    return strings.Join(messages, "; ")
}
```

### 通用校验模式

#### 校验中间件

```go
// Hertz 验证中间件
package middleware

import (
    "context"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// 通用验证中间件
func ValidateMiddleware[T any]() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        var req T
        if err := c.BindAndValidate(&req); err != nil {
            c.JSON(consts.StatusBadRequest, map[string]interface{}{
                "code":    400,
                "message": "参数验证失败",
                "error":   err.Error(),
            })
            c.Abort()
            return
        }
        
        // 存入上下文
        c.Set("validated_request", &req)
        c.Next(ctx)
    }
}

// 使用示例
func main() {
    h := server.Default()
    
    h.POST("/users", ValidateMiddleware[CreateUserRequest](), createUserHandler)
}

func createUserHandler(ctx context.Context, c *app.RequestContext) {
    req := c.MustGet("validated_request").(*CreateUserRequest)
    // 使用已验证的 req
}
```

#### 业务规则验证

```go
// 业务规则验证器
package validator

import (
    "context"
    "fmt"
)

// 验证规则接口
type Rule interface {
    Validate(ctx context.Context) error
}

// 规则链
type RuleChain struct {
    rules []Rule
}

func NewRuleChain(rules ...Rule) *RuleChain {
    return &RuleChain{rules: rules}
}

func (c *RuleChain) Validate(ctx context.Context) error {
    for _, rule := range c.rules {
        if err := rule.Validate(ctx); err != nil {
            return err
        }
    }
    return nil
}

// ============ 具体业务规则 ============

// 用户名唯一性验证
type UniqueUsernameRule struct {
    Username string
    UserRepo UserRepository
}

func (r *UniqueUsernameRule) Validate(ctx context.Context) error {
    exists, err := r.UserRepo.ExistsByUsername(ctx, r.Username)
    if err != nil {
        return fmt.Errorf("failed to check username: %w", err)
    }
    if exists {
        return fmt.Errorf("username '%s' already exists", r.Username)
    }
    return nil
}

// 邮箱唯一性验证
type UniqueEmailRule struct {
    Email    string
    UserRepo UserRepository
}

func (r *UniqueEmailRule) Validate(ctx context.Context) error {
    exists, err := r.UserRepo.ExistsByEmail(ctx, r.Email)
    if err != nil {
        return fmt.Errorf("failed to check email: %w", err)
    }
    if exists {
        return fmt.Errorf("email '%s' already registered", r.Email)
    }
    return nil
}

// 库存检查规则
type StockAvailableRule struct {
    ProductID int64
    Quantity  int
    StockRepo StockRepository
}

func (r *StockAvailableRule) Validate(ctx context.Context) error {
    stock, err := r.StockRepo.GetStock(ctx, r.ProductID)
    if err != nil {
        return fmt.Errorf("failed to check stock: %w", err)
    }
    if stock < r.Quantity {
        return fmt.Errorf("insufficient stock: available %d, requested %d", stock, r.Quantity)
    }
    return nil
}

// ============ 在 Handler 中使用 ============
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    // 1. 基础参数验证（由框架完成）
    
    // 2. 业务规则验证
    rules := NewRuleChain(
        &UniqueUsernameRule{Username: req.Username, UserRepo: s.userRepo},
        &UniqueEmailRule{Email: req.Email, UserRepo: s.userRepo},
    )
    
    if err := rules.Validate(ctx); err != nil {
        return &CreateUserResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }
    
    // 3. 执行业务逻辑
    user, err := s.userRepo.Create(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return &CreateUserResponse{
        Success: true,
        Message: "User created",
        User:    user,
    }, nil
}
```

### 校验最佳实践

```go
// ============ 1. 分层验证 ============
/*
┌─────────────────────────────────────────┐
│           HTTP/RPC 层                    │
│  - 格式验证（JSON/Thrift 解析）         │
│  - 类型验证（字段类型匹配）             │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│           参数验证层                     │
│  - 必填验证                             │
│  - 格式验证（邮箱、手机号）             │
│  - 范围验证（长度、数值范围）           │
│  - 枚举验证                             │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│           业务验证层                     │
│  - 唯一性验证                           │
│  - 关联性验证                           │
│  - 权限验证                             │
│  - 状态机验证                           │
└─────────────────────────────────────────┘
*/

// ============ 2. 验证错误码定义 ============
const (
    ErrCodeValidation      = 40001 // 参数验证失败
    ErrCodeMissingField    = 40002 // 缺少必填字段
    ErrCodeInvalidFormat   = 40003 // 格式错误
    ErrCodeOutOfRange      = 40004 // 超出范围
    ErrCodeDuplicate       = 40005 // 重复值
    ErrCodeNotFound        = 40006 // 关联资源不存在
    ErrCodePermissionDenied = 40007 // 权限不足
)

// ============ 3. 验证消息国际化 ============
type ValidationMessages struct {
    Required     string
    MinLength    string
    MaxLength    string
    Email        string
    Phone        string
    Range        string
}

var messages = map[string]ValidationMessages{
    "zh": {
        Required:  "不能为空",
        MinLength: "长度至少 %d 个字符",
        MaxLength: "长度最多 %d 个字符",
        Email:     "邮箱格式错误",
        Phone:     "手机号格式错误",
        Range:     "必须在 %d 到 %d 之间",
    },
    "en": {
        Required:  "is required",
        MinLength: "must be at least %d characters",
        MaxLength: "must be at most %d characters",
        Email:     "must be a valid email",
        Phone:     "must be a valid phone number",
        Range:     "must be between %d and %d",
    },
}

// ============ 4. 安全验证 ============
type SecureInput struct {
    // XSS 防护：限制特殊字符
    Content string `json:"content" vd:"regexp('^[^<>]*$'); msg:'内容包含非法字符'"`
    
    // SQL 注入防护：使用参数化查询（在数据库层处理）
    
    // 路径穿越防护
    Filename string `json:"filename" vd:"regexp('^[a-zA-Z0-9_.-]+$'); msg:'文件名只能包含字母数字和._-'"`
}

// ============ 5. 性能优化 ============
/*
1. 使用编译时验证而非运行时反射
2. 缓存验证器实例
3. 批量验证时使用并行处理
4. 简单验证优先，复杂验证（如数据库查询）后置
*/

// 缓存验证器
var validatorInstance = validator.New()

func GetValidator() *validator.Validate {
    return validatorInstance
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

## GORM MySQL 数据库集成

本节详细介绍如何在 Hertz/Kitex 微服务中集成 GORM 和 MySQL，包括连接配置、模型定义、CRUD 操作、事务处理等。

### 安装与配置

```bash
# 安装 GORM 和 MySQL 驱动
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
```

### 数据库连接

```go
package db

import (
    "fmt"
    "log"
    "time"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "gorm.io/gorm/schema"
)

// Config 数据库配置
type Config struct {
    Host         string
    Port         int
    User         string
    Password     string
    DBName       string
    MaxIdleConns int
    MaxOpenConns int
    MaxLifetime  time.Duration
    LogLevel     logger.LogLevel
}

var DB *gorm.DB

// Init 初始化数据库连接
func Init(cfg *Config) error {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        cfg.User,
        cfg.Password,
        cfg.Host,
        cfg.Port,
        cfg.DBName,
    )
    
    // GORM 配置
    gormConfig := &gorm.Config{
        // 命名策略
        NamingStrategy: schema.NamingStrategy{
            TablePrefix:   "t_",     // 表前缀
            SingularTable: true,     // 使用单数表名
        },
        // 日志配置
        Logger: logger.Default.LogMode(cfg.LogLevel),
        // 禁用默认事务（提高性能）
        SkipDefaultTransaction: true,
        // 预编译语句缓存
        PrepareStmt: true,
    }
    
    var err error
    DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
    if err != nil {
        return fmt.Errorf("failed to connect database: %w", err)
    }
    
    // 获取底层 sql.DB
    sqlDB, err := DB.DB()
    if err != nil {
        return fmt.Errorf("failed to get sql.DB: %w", err)
    }
    
    // 连接池配置
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)   // 最大空闲连接数
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)   // 最大打开连接数
    sqlDB.SetConnMaxLifetime(cfg.MaxLifetime) // 连接最大生命周期
    
    log.Println("Database connected successfully")
    return nil
}

// Close 关闭数据库连接
func Close() error {
    sqlDB, err := DB.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}

// ============ 多数据库连接 ============
type DBManager struct {
    master  *gorm.DB
    slaves  []*gorm.DB
    current int
}

func NewDBManager(masterDSN string, slaveDSNs []string) (*DBManager, error) {
    master, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    var slaves []*gorm.DB
    for _, dsn := range slaveDSNs {
        slave, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
        if err != nil {
            return nil, err
        }
        slaves = append(slaves, slave)
    }
    
    return &DBManager{
        master: master,
        slaves: slaves,
    }, nil
}

// Master 获取主库（写操作）
func (m *DBManager) Master() *gorm.DB {
    return m.master
}

// Slave 获取从库（读操作，轮询）
func (m *DBManager) Slave() *gorm.DB {
    if len(m.slaves) == 0 {
        return m.master
    }
    m.current = (m.current + 1) % len(m.slaves)
    return m.slaves[m.current]
}
```

### 模型定义

```go
package model

import (
    "database/sql"
    "time"

    "gorm.io/gorm"
)

// ============ 基础模型 ============
type BaseModel struct {
    ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
    CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// ============ 用户模型 ============
type User struct {
    BaseModel
    Username  string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
    Email     string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
    Password  string         `gorm:"type:varchar(255);not null" json:"-"`
    Nickname  string         `gorm:"type:varchar(50)" json:"nickname"`
    Avatar    string         `gorm:"type:varchar(255)" json:"avatar"`
    Phone     sql.NullString `gorm:"type:varchar(20);uniqueIndex" json:"phone"`
    Status    int8           `gorm:"type:tinyint;default:1;index" json:"status"` // 1:正常 0:禁用
    LastLogin *time.Time     `json:"last_login"`
    
    // 关联
    Profile   *UserProfile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
    Orders    []Order      `gorm:"foreignKey:UserID" json:"orders,omitempty"`
    Roles     []Role       `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

// TableName 自定义表名
func (User) TableName() string {
    return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 密码加密等操作
    return nil
}

// ============ 用户详情模型 ============
type UserProfile struct {
    ID       uint64 `gorm:"primaryKey" json:"id"`
    UserID   uint64 `gorm:"uniqueIndex;not null" json:"user_id"`
    RealName string `gorm:"type:varchar(50)" json:"real_name"`
    IDCard   string `gorm:"type:varchar(20)" json:"id_card"`
    Birthday *time.Time `json:"birthday"`
    Gender   int8   `gorm:"type:tinyint;default:0" json:"gender"` // 0:未知 1:男 2:女
    Address  string `gorm:"type:varchar(255)" json:"address"`
    Bio      string `gorm:"type:text" json:"bio"`
}

// ============ 订单模型 ============
type Order struct {
    BaseModel
    OrderNo     string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"order_no"`
    UserID      uint64    `gorm:"index;not null" json:"user_id"`
    TotalAmount float64   `gorm:"type:decimal(10,2);not null" json:"total_amount"`
    Status      int8      `gorm:"type:tinyint;default:0;index" json:"status"`
    PayTime     *time.Time `json:"pay_time"`
    
    // 关联
    User  *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
    ID        uint64  `gorm:"primaryKey" json:"id"`
    OrderID   uint64  `gorm:"index;not null" json:"order_id"`
    ProductID uint64  `gorm:"index;not null" json:"product_id"`
    Quantity  int     `gorm:"not null" json:"quantity"`
    Price     float64 `gorm:"type:decimal(10,2);not null" json:"price"`
    
    Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// ============ 商品模型 ============
type Product struct {
    BaseModel
    Name        string  `gorm:"type:varchar(100);not null" json:"name"`
    Description string  `gorm:"type:text" json:"description"`
    Price       float64 `gorm:"type:decimal(10,2);not null" json:"price"`
    Stock       int     `gorm:"default:0" json:"stock"`
    CategoryID  uint64  `gorm:"index" json:"category_id"`
    Status      int8    `gorm:"type:tinyint;default:1" json:"status"`
    
    Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// ============ 分类模型（树形结构）============
type Category struct {
    ID       uint64     `gorm:"primaryKey" json:"id"`
    Name     string     `gorm:"type:varchar(50);not null" json:"name"`
    ParentID *uint64    `gorm:"index" json:"parent_id"`
    Sort     int        `gorm:"default:0" json:"sort"`
    
    Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// ============ 角色模型（多对多）============
type Role struct {
    ID          uint64 `gorm:"primaryKey" json:"id"`
    Name        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
    Description string `gorm:"type:varchar(255)" json:"description"`
    
    Users []User `gorm:"many2many:user_roles" json:"users,omitempty"`
}

// ============ 自动迁移 ============
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &User{},
        &UserProfile{},
        &Order{},
        &OrderItem{},
        &Product{},
        &Category{},
        &Role{},
    )
}
```

### CRUD 操作

```go
package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

// ============ 用户仓储 ============
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

// BatchCreate 批量创建
func (r *UserRepository) BatchCreate(ctx context.Context, users []*User) error {
    return r.db.WithContext(ctx).CreateInBatches(users, 100).Error
}

// GetByID 根据 ID 查询
func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

// GetByUsername 根据用户名查询
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

// GetWithProfile 查询用户及详情
func (r *UserRepository) GetWithProfile(ctx context.Context, id uint64) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).
        Preload("Profile").
        First(&user, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

// GetWithRoles 查询用户及角色
func (r *UserRepository) GetWithRoles(ctx context.Context, id uint64) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).
        Preload("Roles").
        First(&user, id).Error
    return &user, err
}

// List 分页查询
func (r *UserRepository) List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]*User, int64, error) {
    var users []*User
    var total int64
    
    query := r.db.WithContext(ctx).Model(&User{})
    
    // 动态条件
    if status, ok := conditions["status"]; ok {
        query = query.Where("status = ?", status)
    }
    if keyword, ok := conditions["keyword"]; ok {
        query = query.Where("username LIKE ? OR nickname LIKE ?", "%"+keyword.(string)+"%", "%"+keyword.(string)+"%")
    }
    
    // 统计总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // 分页查询
    offset := (page - 1) * pageSize
    err := query.
        Order("id DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&users).Error
    
    return users, total, err
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, id uint64, updates map[string]interface{}) error {
    return r.db.WithContext(ctx).
        Model(&User{}).
        Where("id = ?", id).
        Updates(updates).Error
}

// UpdateSelective 选择性更新（只更新非零值）
func (r *UserRepository) UpdateSelective(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).
        Model(user).
        Updates(user).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(ctx context.Context, id uint64) error {
    return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

// HardDelete 硬删除
func (r *UserRepository) HardDelete(ctx context.Context, id uint64) error {
    return r.db.WithContext(ctx).Unscoped().Delete(&User{}, id).Error
}

// Exists 检查是否存在
func (r *UserRepository) Exists(ctx context.Context, field string, value interface{}) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&User{}).
        Where(field+" = ?", value).
        Count(&count).Error
    return count > 0, err
}
```

### 复杂查询

```go
package repository

import (
    "context"
    "time"

    "gorm.io/gorm"
)

// ============ 关联查询 ============

// GetOrderWithDetails 获取订单详情（包含用户、商品信息）
func (r *OrderRepository) GetOrderWithDetails(ctx context.Context, id uint64) (*Order, error) {
    var order Order
    err := r.db.WithContext(ctx).
        Preload("User").
        Preload("Items").
        Preload("Items.Product").
        First(&order, id).Error
    return &order, err
}

// GetUserOrders 获取用户的订单列表
func (r *OrderRepository) GetUserOrders(ctx context.Context, userID uint64, page, pageSize int) ([]*Order, error) {
    var orders []*Order
    err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Preload("Items", func(db *gorm.DB) *gorm.DB {
            return db.Limit(5) // 每个订单只加载前 5 个商品
        }).
        Order("created_at DESC").
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find(&orders).Error
    return orders, err
}

// ============ 聚合查询 ============

type OrderStats struct {
    TotalOrders  int64   `json:"total_orders"`
    TotalAmount  float64 `json:"total_amount"`
    AvgAmount    float64 `json:"avg_amount"`
    PaidOrders   int64   `json:"paid_orders"`
    UnpaidOrders int64   `json:"unpaid_orders"`
}

// GetUserOrderStats 获取用户订单统计
func (r *OrderRepository) GetUserOrderStats(ctx context.Context, userID uint64) (*OrderStats, error) {
    var stats OrderStats
    err := r.db.WithContext(ctx).
        Model(&Order{}).
        Where("user_id = ?", userID).
        Select(`
            COUNT(*) as total_orders,
            COALESCE(SUM(total_amount), 0) as total_amount,
            COALESCE(AVG(total_amount), 0) as avg_amount,
            SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as paid_orders,
            SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as unpaid_orders
        `).
        Scan(&stats).Error
    return &stats, err
}

// GetDailyOrderStats 按日统计订单
func (r *OrderRepository) GetDailyOrderStats(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
    var results []map[string]interface{}
    err := r.db.WithContext(ctx).
        Model(&Order{}).
        Where("created_at BETWEEN ? AND ?", startDate, endDate).
        Select(`
            DATE(created_at) as date,
            COUNT(*) as order_count,
            SUM(total_amount) as total_amount
        `).
        Group("DATE(created_at)").
        Order("date").
        Find(&results).Error
    return results, err
}

// ============ 子查询 ============

// GetActiveUsers 获取有订单的活跃用户
func (r *UserRepository) GetActiveUsers(ctx context.Context) ([]*User, error) {
    var users []*User
    
    // 子查询：获取有订单的用户 ID
    subQuery := r.db.Model(&Order{}).Select("DISTINCT user_id")
    
    err := r.db.WithContext(ctx).
        Where("id IN (?)", subQuery).
        Find(&users).Error
    
    return users, err
}

// GetTopBuyers 获取消费最高的用户
func (r *UserRepository) GetTopBuyers(ctx context.Context, limit int) ([]map[string]interface{}, error) {
    var results []map[string]interface{}
    err := r.db.WithContext(ctx).
        Table("users u").
        Select("u.id, u.username, u.nickname, COALESCE(SUM(o.total_amount), 0) as total_spent").
        Joins("LEFT JOIN orders o ON u.id = o.user_id AND o.deleted_at IS NULL").
        Group("u.id").
        Order("total_spent DESC").
        Limit(limit).
        Find(&results).Error
    return results, err
}

// ============ 原生 SQL ============

// ExecuteRawSQL 执行原生 SQL
func (r *UserRepository) ExecuteRawSQL(ctx context.Context) error {
    // 查询
    var results []map[string]interface{}
    r.db.WithContext(ctx).Raw(`
        SELECT u.*, 
               (SELECT COUNT(*) FROM orders WHERE user_id = u.id) as order_count
        FROM users u
        WHERE u.status = ?
    `, 1).Scan(&results)
    
    // 执行
    r.db.WithContext(ctx).Exec("UPDATE users SET status = ? WHERE last_login < ?", 0, time.Now().AddDate(0, -6, 0))
    
    return nil
}

// ============ 锁 ============

// GetForUpdate 悲观锁查询
func (r *ProductRepository) GetForUpdate(ctx context.Context, tx *gorm.DB, id uint64) (*Product, error) {
    var product Product
    err := tx.WithContext(ctx).
        Clauses(clause.Locking{Strength: "UPDATE"}).
        First(&product, id).Error
    return &product, err
}

// DecrementStock 扣减库存（带锁）
func (r *ProductRepository) DecrementStock(ctx context.Context, id uint64, quantity int) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var product Product
        // 加锁查询
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&product, id).Error; err != nil {
            return err
        }
        
        // 检查库存
        if product.Stock < quantity {
            return errors.New("insufficient stock")
        }
        
        // 扣减库存
        return tx.Model(&product).
            Update("stock", gorm.Expr("stock - ?", quantity)).Error
    })
}

// OptimisticUpdate 乐观锁更新
func (r *ProductRepository) OptimisticUpdate(ctx context.Context, product *Product) error {
    result := r.db.WithContext(ctx).
        Model(product).
        Where("id = ? AND updated_at = ?", product.ID, product.UpdatedAt).
        Updates(product)
    
    if result.RowsAffected == 0 {
        return errors.New("concurrent update conflict")
    }
    return result.Error
}
```

### 事务处理

```go
package service

import (
    "context"
    "fmt"

    "gorm.io/gorm"
)

// ============ 基本事务 ============

func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    var order *Order
    
    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 1. 创建订单
        order = &Order{
            OrderNo:     generateOrderNo(),
            UserID:      req.UserID,
            TotalAmount: req.TotalAmount,
            Status:      0,
        }
        if err := tx.Create(order).Error; err != nil {
            return err
        }
        
        // 2. 创建订单项
        for _, item := range req.Items {
            orderItem := &OrderItem{
                OrderID:   order.ID,
                ProductID: item.ProductID,
                Quantity:  item.Quantity,
                Price:     item.Price,
            }
            if err := tx.Create(orderItem).Error; err != nil {
                return err
            }
            
            // 3. 扣减库存
            result := tx.Model(&Product{}).
                Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
                Update("stock", gorm.Expr("stock - ?", item.Quantity))
            
            if result.RowsAffected == 0 {
                return fmt.Errorf("product %d stock insufficient", item.ProductID)
            }
        }
        
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    return order, nil
}

// ============ 嵌套事务（保存点）============

func (s *OrderService) CreateOrderWithSavepoint(ctx context.Context, req *CreateOrderRequest) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 外层事务：创建订单
        order := &Order{...}
        if err := tx.Create(order).Error; err != nil {
            return err
        }
        
        // 嵌套事务：处理订单项
        err := tx.Transaction(func(tx2 *gorm.DB) error {
            for _, item := range req.Items {
                // 这里的错误不会回滚外层事务
                if err := tx2.Create(&OrderItem{...}).Error; err != nil {
                    return err // 只回滚到这个保存点
                }
            }
            return nil
        })
        
        if err != nil {
            // 可以选择继续或回滚
            return err
        }
        
        return nil
    })
}

// ============ 手动事务控制 ============

func (s *OrderService) ManualTransaction(ctx context.Context) error {
    tx := s.db.WithContext(ctx).Begin()
    if tx.Error != nil {
        return tx.Error
    }
    
    // 使用 defer 确保事务结束
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    // 执行操作
    if err := tx.Create(&Order{...}).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    if err := tx.Create(&OrderItem{...}).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit().Error
}

// ============ 事务传播 ============

type TxKey struct{}

// GetTxFromContext 从上下文获取事务
func GetTxFromContext(ctx context.Context) *gorm.DB {
    tx, ok := ctx.Value(TxKey{}).(*gorm.DB)
    if ok {
        return tx
    }
    return nil
}

// WithTx 在上下文中设置事务
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
    return context.WithValue(ctx, TxKey{}, tx)
}

// 使用事务上下文
func (s *OrderService) CreateOrderWithTxContext(ctx context.Context, req *CreateOrderRequest) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 将事务放入上下文
        txCtx := WithTx(ctx, tx)
        
        // 子服务使用相同事务
        if err := s.orderRepo.Create(txCtx, &Order{...}); err != nil {
            return err
        }
        
        if err := s.stockService.Decrement(txCtx, req.Items); err != nil {
            return err
        }
        
        return nil
    })
}

// 仓储层自动使用事务
func (r *OrderRepository) Create(ctx context.Context, order *Order) error {
    db := GetTxFromContext(ctx)
    if db == nil {
        db = r.db
    }
    return db.WithContext(ctx).Create(order).Error
}
```

### 性能优化

```go
package db

import (
    "gorm.io/gorm"
    "gorm.io/hints"
)

// ============ 索引优化 ============

// 强制使用索引
func (r *UserRepository) ListWithIndex(ctx context.Context) ([]*User, error) {
    var users []*User
    err := r.db.WithContext(ctx).
        Clauses(hints.UseIndex("idx_status")).
        Where("status = ?", 1).
        Find(&users).Error
    return users, err
}

// ============ 批量操作 ============

// 批量插入
func (r *UserRepository) BatchInsert(ctx context.Context, users []*User) error {
    return r.db.WithContext(ctx).CreateInBatches(users, 100).Error
}

// 批量更新
func (r *UserRepository) BatchUpdate(ctx context.Context, ids []uint64, updates map[string]interface{}) error {
    return r.db.WithContext(ctx).
        Model(&User{}).
        Where("id IN ?", ids).
        Updates(updates).Error
}

// 批量删除
func (r *UserRepository) BatchDelete(ctx context.Context, ids []uint64) error {
    return r.db.WithContext(ctx).Delete(&User{}, ids).Error
}

// ============ 分批处理大数据 ============

func (r *UserRepository) ProcessInBatches(ctx context.Context, batchSize int, fn func([]*User) error) error {
    return r.db.WithContext(ctx).
        FindInBatches(&[]*User{}, batchSize, func(tx *gorm.DB, batch int) error {
            users := tx.Statement.Dest.(*[]*User)
            return fn(*users)
        }).Error
}

// 使用游标分批处理
func (r *UserRepository) ProcessWithCursor(ctx context.Context, fn func(*User) error) error {
    rows, err := r.db.WithContext(ctx).Model(&User{}).Rows()
    if err != nil {
        return err
    }
    defer rows.Close()
    
    for rows.Next() {
        var user User
        if err := r.db.ScanRows(rows, &user); err != nil {
            return err
        }
        if err := fn(&user); err != nil {
            return err
        }
    }
    return nil
}

// ============ 查询优化 ============

// 只查询需要的字段
func (r *UserRepository) ListBasicInfo(ctx context.Context) ([]map[string]interface{}, error) {
    var results []map[string]interface{}
    err := r.db.WithContext(ctx).
        Model(&User{}).
        Select("id", "username", "nickname", "avatar").
        Find(&results).Error
    return results, err
}

// 使用 Pluck 获取单列
func (r *UserRepository) GetAllUsernames(ctx context.Context) ([]string, error) {
    var usernames []string
    err := r.db.WithContext(ctx).
        Model(&User{}).
        Pluck("username", &usernames).Error
    return usernames, err
}

// 使用 Count 代替 Find
func (r *UserRepository) CountByStatus(ctx context.Context, status int) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&User{}).
        Where("status = ?", status).
        Count(&count).Error
    return count, err
}

// ============ 预加载优化 ============

// 条件预加载
func (r *OrderRepository) GetWithConditionalPreload(ctx context.Context, id uint64) (*Order, error) {
    var order Order
    err := r.db.WithContext(ctx).
        Preload("Items", "quantity > ?", 0).
        Preload("Items.Product", func(db *gorm.DB) *gorm.DB {
            return db.Select("id", "name", "price")
        }).
        First(&order, id).Error
    return &order, err
}

// 使用 Joins 代替 Preload（单层关联）
func (r *OrderRepository) GetWithJoins(ctx context.Context, id uint64) (*Order, error) {
    var order Order
    err := r.db.WithContext(ctx).
        Joins("User").
        First(&order, id).Error
    return &order, err
}
```

### 在 Hertz/Kitex 中集成

```go
// ============ Hertz Handler 集成 ============
package handler

import (
    "context"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
)

type UserHandler struct {
    userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUser(ctx context.Context, c *app.RequestContext) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
    
    user, err := h.userService.GetByID(ctx, id)
    if err != nil {
        c.JSON(consts.StatusInternalServerError, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    if user == nil {
        c.JSON(consts.StatusNotFound, map[string]interface{}{
            "error": "user not found",
        })
        return
    }
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "code": 0,
        "data": user,
    })
}

func (h *UserHandler) ListUsers(ctx context.Context, c *app.RequestContext) {
    var req ListUsersRequest
    if err := c.BindAndValidate(&req); err != nil {
        c.JSON(consts.StatusBadRequest, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    users, total, err := h.userService.List(ctx, req.Page, req.PageSize, req.ToConditions())
    if err != nil {
        c.JSON(consts.StatusInternalServerError, map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    
    c.JSON(consts.StatusOK, map[string]interface{}{
        "code": 0,
        "data": map[string]interface{}{
            "users": users,
            "total": total,
            "page":  req.Page,
        },
    })
}

// ============ Kitex Handler 集成 ============
package main

import (
    "context"
    
    user "github.com/example/user-service/kitex_gen/user"
)

type UserServiceImpl struct {
    userRepo *repository.UserRepository
}

func (s *UserServiceImpl) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
    u, err := s.userRepo.GetByID(ctx, uint64(req.Id))
    if err != nil {
        return nil, err
    }
    
    if u == nil {
        return &user.GetUserResponse{
            Success: false,
            Message: "user not found",
        }, nil
    }
    
    return &user.GetUserResponse{
        Success: true,
        User: &user.User{
            Id:       int64(u.ID),
            Username: u.Username,
            Email:    u.Email,
            Nickname: u.Nickname,
        },
    }, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
    newUser := &model.User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashPassword(req.Password),
        Nickname: req.Nickname,
    }
    
    if err := s.userRepo.Create(ctx, newUser); err != nil {
        return nil, err
    }
    
    return &user.CreateUserResponse{
        Success: true,
        Message: "user created",
        User: &user.User{
            Id:       int64(newUser.ID),
            Username: newUser.Username,
        },
    }, nil
}

// ============ 初始化数据库 ============
package main

import (
    "log"
    "time"

    "github.com/cloudwego/hertz/pkg/app/server"
    "gorm.io/gorm/logger"
)

func main() {
    // 初始化数据库
    dbConfig := &db.Config{
        Host:         "localhost",
        Port:         3306,
        User:         "root",
        Password:     "password",
        DBName:       "myapp",
        MaxIdleConns: 10,
        MaxOpenConns: 100,
        MaxLifetime:  time.Hour,
        LogLevel:     logger.Info,
    }
    
    if err := db.Init(dbConfig); err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // 自动迁移
    if err := model.AutoMigrate(db.DB); err != nil {
        log.Fatal(err)
    }
    
    // 初始化仓储和服务
    userRepo := repository.NewUserRepository(db.DB)
    userService := service.NewUserService(userRepo)
    userHandler := handler.NewUserHandler(userService)
    
    // 启动 Hertz
    h := server.Default(server.WithHostPorts(":8080"))
    
    api := h.Group("/api")
    {
        users := api.Group("/users")
        users.GET("/:id", userHandler.GetUser)
        users.GET("/", userHandler.ListUsers)
        users.POST("/", userHandler.CreateUser)
    }
    
    h.Spin()
}
```

### GORM 最佳实践

```go
// ============ 1. 使用上下文传递数据库连接 ============
type contextKey string

const dbKey contextKey = "db"

func DBMiddleware(db *gorm.DB) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        c.Set(string(dbKey), db)
        c.Next(ctx)
    }
}

func GetDB(c *app.RequestContext) *gorm.DB {
    return c.MustGet(string(dbKey)).(*gorm.DB)
}

// ============ 2. 统一错误处理 ============
func HandleDBError(err error) (int, string) {
    if err == nil {
        return 0, ""
    }
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return 404, "record not found"
    }
    
    // MySQL 错误
    var mysqlErr *mysql.MySQLError
    if errors.As(err, &mysqlErr) {
        switch mysqlErr.Number {
        case 1062:
            return 409, "duplicate entry"
        case 1452:
            return 400, "foreign key constraint fails"
        }
    }
    
    return 500, "database error"
}

// ============ 3. 软删除与恢复 ============
func (r *UserRepository) Restore(ctx context.Context, id uint64) error {
    return r.db.WithContext(ctx).
        Unscoped().
        Model(&User{}).
        Where("id = ?", id).
        Update("deleted_at", nil).Error
}

// 查询包含已删除记录
func (r *UserRepository) GetAllIncludingDeleted(ctx context.Context) ([]*User, error) {
    var users []*User
    err := r.db.WithContext(ctx).Unscoped().Find(&users).Error
    return users, err
}

// ============ 4. 审计日志 ============
type AuditLog struct {
    ID        uint64    `gorm:"primaryKey"`
    TableName string    `gorm:"type:varchar(50)"`
    RecordID  uint64
    Action    string    `gorm:"type:varchar(20)"` // create, update, delete
    OldValue  string    `gorm:"type:text"`
    NewValue  string    `gorm:"type:text"`
    UserID    uint64
    CreatedAt time.Time
}

func RegisterAuditCallbacks(db *gorm.DB) {
    db.Callback().Create().After("gorm:create").Register("audit:create", auditCreate)
    db.Callback().Update().After("gorm:update").Register("audit:update", auditUpdate)
    db.Callback().Delete().After("gorm:delete").Register("audit:delete", auditDelete)
}

// ============ 5. 通用仓储 ============
type Repository[T any] struct {
    db *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
    return &Repository[T]{db: db}
}

func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
    return r.db.WithContext(ctx).Create(entity).Error
}

func (r *Repository[T]) GetByID(ctx context.Context, id uint64) (*T, error) {
    var entity T
    err := r.db.WithContext(ctx).First(&entity, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &entity, err
}

func (r *Repository[T]) Update(ctx context.Context, id uint64, updates map[string]interface{}) error {
    var entity T
    return r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository[T]) Delete(ctx context.Context, id uint64) error {
    var entity T
    return r.db.WithContext(ctx).Delete(&entity, id).Error
}

func (r *Repository[T]) List(ctx context.Context, page, pageSize int) ([]*T, int64, error) {
    var entities []*T
    var total int64
    var entity T
    
    r.db.WithContext(ctx).Model(&entity).Count(&total)
    err := r.db.WithContext(ctx).
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find(&entities).Error
    
    return entities, total, err
}

// 使用
userRepo := NewRepository[User](db)
user, _ := userRepo.GetByID(ctx, 1)
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

# Web CORS 跨域配置详解

本文档详细介绍如何在 TypeScript 前端和 Go Hertz 后端配置 CORS（跨域资源共享）。

## 目录

1. [CORS 基础概念](#cors-基础概念)
2. [TypeScript 前端配置](#typescript-前端配置)
3. [Go Hertz 后端配置](#go-hertz-后端配置)
4. [Thrift IDL 与 TLB 配置](#thrift-idl-与-tlb-配置)
5. [常见问题与解决方案](#常见问题与解决方案)

---

## CORS 基础概念

### 什么是 CORS?

CORS (Cross-Origin Resource Sharing) 是一种基于 HTTP 头的机制，允许服务器指示浏览器允许从哪些源加载资源。

### 同源策略

浏览器的同源策略要求：
- **协议**相同 (http/https)
- **域名**相同
- **端口**相同

```
https://example.com:443/path
  ↓        ↓        ↓
协议     域名      端口
```

### CORS 请求类型

| 类型 | 条件 | 预检请求 |
|------|------|---------|
| 简单请求 | GET/HEAD/POST + 简单头 | 否 |
| 预检请求 | 其他方法或自定义头 | 是 (OPTIONS) |

### CORS 相关 HTTP 头

**请求头：**
```
Origin: https://example.com
Access-Control-Request-Method: POST
Access-Control-Request-Headers: Content-Type, Authorization
```

**响应头：**
```
Access-Control-Allow-Origin: https://example.com
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 86400
Access-Control-Expose-Headers: X-Custom-Header
```

---

## TypeScript 前端配置

### 1. Fetch API 配置

```typescript
// utils/request.ts

interface RequestConfig extends RequestInit {
  baseURL?: string;
  timeout?: number;
}

class HttpClient {
  private baseURL: string;
  private timeout: number;

  constructor(config: { baseURL: string; timeout?: number }) {
    this.baseURL = config.baseURL;
    this.timeout = config.timeout || 30000;
  }

  async request<T>(endpoint: string, options: RequestConfig = {}): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal,
        // CORS 关键配置
        mode: 'cors',                    // 启用 CORS
        credentials: 'include',          // 发送 cookies
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      clearTimeout(timeoutId);
      throw error;
    }
  }

  get<T>(endpoint: string, options?: RequestConfig): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'GET' });
  }

  post<T>(endpoint: string, data?: unknown, options?: RequestConfig): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  put<T>(endpoint: string, data?: unknown, options?: RequestConfig): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  delete<T>(endpoint: string, options?: RequestConfig): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' });
  }
}

// 创建实例
export const apiClient = new HttpClient({
  baseURL: 'https://api.example.com',
  timeout: 10000,
});
```

### 2. Axios 配置

```typescript
// utils/axios.ts
import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

// 创建 axios 实例
const axiosInstance: AxiosInstance = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'https://api.example.com',
  timeout: 10000,
  // CORS 关键配置
  withCredentials: true,  // 允许跨域携带 cookies
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
axiosInstance.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    // 添加认证 token
    const token = localStorage.getItem('token');
    if (token && config.headers) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    return response.data;
  },
  (error) => {
    if (error.response) {
      // CORS 错误通常表现为 network error
      if (!error.response.status) {
        console.error('CORS Error: 请检查服务器 CORS 配置');
      }
      
      switch (error.response.status) {
        case 401:
          // 未授权，跳转登录
          window.location.href = '/login';
          break;
        case 403:
          console.error('禁止访问');
          break;
        case 404:
          console.error('资源不存在');
          break;
        case 500:
          console.error('服务器错误');
          break;
      }
    }
    return Promise.reject(error);
  }
);

export default axiosInstance;
```

### 3. 开发环境代理配置 (Vite)

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    // 代理配置 - 开发环境绕过 CORS
    proxy: {
      '/api': {
        target: 'http://localhost:8080',  // Hertz 后端地址
        changeOrigin: true,                // 修改 origin 头
        rewrite: (path) => path.replace(/^\/api/, ''),
        // 配置 WebSocket 代理
        ws: true,
        // 自定义代理配置
        configure: (proxy, options) => {
          proxy.on('proxyReq', (proxyReq, req, res) => {
            // 添加自定义请求头
            proxyReq.setHeader('X-Forwarded-For', req.socket.remoteAddress || '');
          });
        },
      },
      // 多个代理目标
      '/auth': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
});
```

### 4. 开发环境代理配置 (Webpack/CRA)

```typescript
// src/setupProxy.js (Create React App)
const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function(app) {
  app.use(
    '/api',
    createProxyMiddleware({
      target: 'http://localhost:8080',
      changeOrigin: true,
      pathRewrite: {
        '^/api': '',
      },
      onProxyReq: (proxyReq, req, res) => {
        // 可以修改请求头
        proxyReq.setHeader('X-Custom-Header', 'value');
      },
      onProxyRes: (proxyRes, req, res) => {
        // 可以修改响应头
        proxyRes.headers['X-Proxy-Header'] = 'value';
      },
    })
  );
};
```

### 5. TypeScript 类型定义

```typescript
// types/api.ts

// CORS 配置类型
interface CORSConfig {
  mode: RequestMode;           // 'cors' | 'no-cors' | 'same-origin'
  credentials: RequestCredentials;  // 'include' | 'same-origin' | 'omit'
  headers?: HeadersInit;
}

// API 响应类型
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

// 分页响应
interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

// 错误响应
interface ApiError {
  code: number;
  message: string;
  details?: Record<string, string[]>;
}

// 请求配置
interface ApiRequestConfig {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  url: string;
  data?: unknown;
  params?: Record<string, string | number>;
  headers?: Record<string, string>;
  timeout?: number;
  withCredentials?: boolean;
}
```

---

## Go Hertz 后端配置

### 1. 安装依赖

```bash
# 安装 Hertz
go get github.com/cloudwego/hertz

# 安装 CORS 中间件
go get github.com/hertz-contrib/cors

# 安装其他常用中间件
go get github.com/hertz-contrib/logger/accesslog
go get github.com/hertz-contrib/recovery
```

### 2. 基础 CORS 配置

```go
// main.go
package main

import (
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

func main() {
	h := server.Default(
		server.WithHostPorts(":8080"),
	)

	// CORS 中间件配置
	h.Use(cors.New(cors.Config{
		// 允许的源
		AllowOrigins: []string{
			"http://localhost:3000",
			"https://example.com",
		},
		
		// 允许的 HTTP 方法
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		
		// 允许的请求头
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
		},
		
		// 暴露给客户端的响应头
		ExposeHeaders: []string{
			"Content-Length",
			"X-Request-ID",
			"X-Response-Time",
		},
		
		// 是否允许携带凭证 (cookies)
		AllowCredentials: true,
		
		// 预检请求缓存时间
		MaxAge: 12 * time.Hour,
	}))

	// 注册路由
	registerRoutes(h)

	h.Spin()
}
```

### 3. 完整项目结构

```
project/
├── main.go
├── go.mod
├── go.sum
├── config/
│   └── config.go
├── middleware/
│   ├── cors.go
│   ├── auth.go
│   └── logger.go
├── handler/
│   └── user.go
├── router/
│   └── router.go
└── pkg/
    └── response/
        └── response.go
```

### 4. CORS 中间件封装

```go
// middleware/cors.go
package middleware

import (
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/cors"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig 默认 CORS 配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-CSRF-Token",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Request-ID",
		},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

// NewCORSMiddleware 创建 CORS 中间件
func NewCORSMiddleware(cfg CORSConfig) app.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
		// 自定义 AllowOriginFunc 实现动态源验证
		AllowOriginFunc: func(origin string) bool {
			// 可以在这里实现更复杂的逻辑
			// 例如：从数据库或配置中心读取允许的源
			return true
		},
	})
}

// DevelopmentCORS 开发环境 CORS（允许所有）
func DevelopmentCORS() app.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	})
}

// ProductionCORS 生产环境 CORS（严格限制）
func ProductionCORS(allowedOrigins []string) app.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	})
}
```

### 5. 手动实现 CORS（不使用中间件）

```go
// middleware/cors_manual.go
package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ManualCORS 手动实现 CORS
func ManualCORS(allowedOrigins []string) app.HandlerFunc {
	originsMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originsMap[origin] = true
	}

	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))

		// 检查源是否允许
		if originsMap[origin] || originsMap["*"] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "43200") // 12 小时

		// 处理预检请求
		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}

		c.Next(ctx)
	}
}

// CORSWithWildcard 支持子域名通配符
func CORSWithWildcard(patterns []string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))

		allowed := false
		for _, pattern := range patterns {
			if matchOrigin(origin, pattern) {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "43200")
		}

		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}

		c.Next(ctx)
	}
}

// matchOrigin 匹配源（支持通配符）
func matchOrigin(origin, pattern string) bool {
	if pattern == "*" {
		return true
	}
	
	// 支持子域名通配符，如 *.example.com
	if strings.HasPrefix(pattern, "*.") {
		domain := pattern[1:] // .example.com
		return strings.HasSuffix(origin, domain)
	}
	
	return origin == pattern
}
```

### 6. 路由配置

```go
// router/router.go
package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"myproject/handler"
	"myproject/middleware"
)

func RegisterRoutes(h *server.Hertz) {
	// 全局 CORS 中间件
	h.Use(middleware.NewCORSMiddleware(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "https://app.example.com"},
		AllowCredentials: true,
	}))

	// API v1 路由组
	v1 := h.Group("/api/v1")
	{
		// 用户相关
		users := v1.Group("/users")
		{
			users.GET("", handler.ListUsers)
			users.GET("/:id", handler.GetUser)
			users.POST("", handler.CreateUser)
			users.PUT("/:id", handler.UpdateUser)
			users.DELETE("/:id", handler.DeleteUser)
		}

		// 需要认证的路由
		auth := v1.Group("/auth")
		auth.Use(middleware.JWTAuth())
		{
			auth.GET("/profile", handler.GetProfile)
			auth.PUT("/profile", handler.UpdateProfile)
		}
	}

	// 健康检查（不需要 CORS）
	h.GET("/health", handler.HealthCheck)
}
```

### 7. Handler 示例

```go
// handler/user.go
package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// User 用户模型
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// ListUsers 获取用户列表
func ListUsers(ctx context.Context, c *app.RequestContext) {
	users := []User{
		{ID: 1, Username: "alice", Email: "alice@example.com"},
		{ID: 2, Username: "bob", Email: "bob@example.com"},
	}

	c.JSON(consts.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    users,
	})
}

// GetUser 获取单个用户
func GetUser(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	
	user := User{
		ID:       1,
		Username: "alice",
		Email:    "alice@example.com",
	}

	c.JSON(consts.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    user,
	})
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" vd:"len($)>0"`
	Email    string `json:"email" vd:"email($)"`
	Password string `json:"password" vd:"len($)>=6"`
}

// CreateUser 创建用户
func CreateUser(ctx context.Context, c *app.RequestContext) {
	var req CreateUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 创建用户逻辑...
	
	c.JSON(consts.StatusCreated, Response{
		Code:    0,
		Message: "用户创建成功",
		Data: User{
			ID:       1,
			Username: req.Username,
			Email:    req.Email,
		},
	})
}

// UpdateUser 更新用户
func UpdateUser(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	
	var req CreateUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, Response{
		Code:    0,
		Message: "用户更新成功",
	})
}

// DeleteUser 删除用户
func DeleteUser(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	
	c.JSON(consts.StatusOK, Response{
		Code:    0,
		Message: "用户删除成功",
	})
}

// HealthCheck 健康检查
func HealthCheck(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, map[string]string{
		"status": "ok",
	})
}
```

### 8. 配置文件管理

```go
// config/config.go
package config

import (
	"os"
	"strings"
	"time"
)

// Config 应用配置
type Config struct {
	Server ServerConfig
	CORS   CORSConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string
	Port string
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		CORS: CORSConfig{
			AllowOrigins:     getEnvSlice("CORS_ALLOW_ORIGINS", []string{"http://localhost:3000"}),
			AllowMethods:     getEnvSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowHeaders:     getEnvSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Authorization"}),
			ExposeHeaders:    getEnvSlice("CORS_EXPOSE_HEADERS", []string{"Content-Length"}),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           time.Duration(getEnvInt("CORS_MAX_AGE", 43200)) * time.Second,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	// 简化实现
	return defaultValue
}
```

---

## Thrift IDL 与 TLB 配置

### 1. Thrift IDL 定义

```thrift
// idl/user.thrift
namespace go user

// 用户结构
struct User {
    1: required i64 id
    2: required string username
    3: required string email
    4: optional string avatar
    5: required i64 created_at
}

// 创建用户请求
struct CreateUserRequest {
    1: required string username
    2: required string email
    3: required string password
}

// 创建用户响应
struct CreateUserResponse {
    1: required i32 code
    2: required string message
    3: optional User user
}

// 获取用户请求
struct GetUserRequest {
    1: required i64 user_id
}

// 获取用户响应
struct GetUserResponse {
    1: required i32 code
    2: required string message
    3: optional User user
}

// 用户列表请求
struct ListUsersRequest {
    1: required i32 page
    2: required i32 page_size
    3: optional string keyword
}

// 用户列表响应
struct ListUsersResponse {
    1: required i32 code
    2: required string message
    3: required list<User> users
    4: required i64 total
}

// 用户服务定义
service UserService {
    CreateUserResponse CreateUser(1: CreateUserRequest req)
    GetUserResponse GetUser(1: GetUserRequest req)
    ListUsersResponse ListUsers(1: ListUsersRequest req)
}
```

### 2. HTTP 注解配置

```thrift
// idl/api.thrift
namespace go api

include "user.thrift"

// HTTP 服务定义（用于 Hertz 生成）
service UserAPI {
    // 创建用户
    user.CreateUserResponse CreateUser(1: user.CreateUserRequest req) (
        api.post="/api/v1/users"
        api.serializer="json"
    )
    
    // 获取用户
    user.GetUserResponse GetUser(1: user.GetUserRequest req) (
        api.get="/api/v1/users/:user_id"
        api.serializer="json"
    )
    
    // 用户列表
    user.ListUsersResponse ListUsers(1: user.ListUsersRequest req) (
        api.get="/api/v1/users"
        api.serializer="json"
    )
}
```

### 3. 使用 hz 生成代码

```bash
# 安装 hz 工具
go install github.com/cloudwego/hertz/cmd/hz@latest

# 从 Thrift IDL 生成代码
hz new -module myproject -idl idl/api.thrift

# 更新代码
hz update -idl idl/api.thrift
```

### 4. 生成的项目结构

```
myproject/
├── biz/
│   ├── handler/
│   │   └── api/
│   │       └── user_api.go    # 生成的 handler
│   ├── model/
│   │   └── api/
│   │       └── api.go         # 生成的模型
│   └── router/
│       └── api/
│           └── user_api.go    # 生成的路由
├── idl/
│   ├── api.thrift
│   └── user.thrift
├── main.go
├── router.go
└── router_gen.go
```

---

## 常见问题与解决方案

### 1. 预检请求失败

**问题：** OPTIONS 请求返回 405 或被拦截

**解决方案：**
```go
// 确保 OPTIONS 方法被正确处理
h.OPTIONS("/*path", func(ctx context.Context, c *app.RequestContext) {
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
    c.Status(consts.StatusNoContent)
})
```

### 2. Credentials 与通配符冲突

**问题：** `Access-Control-Allow-Credentials: true` 时不能使用 `*`

**解决方案：**
```go
// 错误示例
AllowOrigins:     []string{"*"},
AllowCredentials: true,  // 这会导致错误

// 正确示例
AllowOrigins:     []string{"http://localhost:3000"},
AllowCredentials: true,

// 或者使用 AllowOriginFunc 动态返回 Origin
AllowOriginFunc: func(origin string) bool {
    // 验证逻辑
    return isAllowedOrigin(origin)
},
```

### 3. 自定义请求头被拦截

**问题：** 自定义头如 `X-Request-ID` 未被允许

**解决方案：**
```go
AllowHeaders: []string{
    "Origin",
    "Content-Type",
    "Authorization",
    "X-Request-ID",     // 添加自定义头
    "X-Custom-Header",
},
```

### 4. 响应头无法在前端访问

**问题：** 自定义响应头在 JavaScript 中无法读取

**解决方案：**
```go
// 后端配置
ExposeHeaders: []string{
    "X-Request-ID",
    "X-Total-Count",
    "X-Page-Count",
},

// 前端读取
const response = await fetch(url);
const requestId = response.headers.get('X-Request-ID');
```

### 5. Cookie 跨域问题

**问题：** Cookie 无法跨域发送或接收

**解决方案：**

后端：
```go
AllowCredentials: true,
AllowOrigins: []string{"http://localhost:3000"},  // 具体域名，不能用 *
```

前端：
```typescript
// Fetch API
fetch(url, {
  credentials: 'include',
});

// Axios
axios.defaults.withCredentials = true;
```

Cookie 设置：
```go
c.SetCookie("token", value, 3600, "/", "example.com", true, true)
// SameSite 属性设置
c.SetCookie("token", value, 3600, "/", "example.com", true, true)
// 对于跨站点，需要设置 SameSite=None; Secure
```

### 6. 调试 CORS 问题

```typescript
// 前端调试
fetch(url)
  .then(response => {
    console.log('Response headers:');
    response.headers.forEach((value, key) => {
      console.log(`${key}: ${value}`);
    });
  })
  .catch(error => {
    console.error('CORS Error:', error);
  });
```

```go
// 后端调试中间件
func DebugCORS() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        fmt.Printf("Request Origin: %s\n", c.GetHeader("Origin"))
        fmt.Printf("Request Method: %s\n", c.Method())
        fmt.Printf("Request Headers: %v\n", c.GetHeader("Access-Control-Request-Headers"))
        
        c.Next(ctx)
        
        fmt.Printf("Response Headers:\n")
        c.Response.Header.VisitAll(func(key, value []byte) {
            fmt.Printf("  %s: %s\n", key, value)
        })
    }
}
```

---

## 环境变量配置示例

```bash
# .env.development
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:5173
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=43200

# .env.production
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
CORS_ALLOW_ORIGINS=https://app.example.com,https://admin.example.com
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400
```

---

## 总结

| 配置项 | 开发环境 | 生产环境 |
|--------|---------|---------|
| AllowOrigins | 使用代理或 `localhost` | 具体域名列表 |
| AllowCredentials | true | true |
| AllowMethods | 全部 | 按需配置 |
| AllowHeaders | 宽松 | 严格限制 |
| MaxAge | 短 (便于测试) | 长 (减少预检) |

**最佳实践：**

1. 开发环境使用代理避免 CORS 问题
2. 生产环境严格限制允许的源
3. 使用 `AllowOriginFunc` 实现动态源验证
4. 合理设置 `MaxAge` 减少预检请求
5. 不要在生产环境使用 `AllowAllOrigins`
6. 使用环境变量管理 CORS 配置

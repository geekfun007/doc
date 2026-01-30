# Web CORS 跨域配置详解

本文档详细介绍如何在 TypeScript 前端和 Go Hertz 后端配置 CORS（跨域资源共享）。

## 目录

1. [CORS 基础概念](#cors-基础概念)
2. [CORS HTTP Header 详解](#cors-http-header-详解)
3. [Nginx CORS 配置](#nginx-cors-配置)
4. [TypeScript 前端配置](#typescript-前端配置)
5. [Go Hertz 后端配置](#go-hertz-后端配置)
6. [Thrift IDL 与 TLB 配置](#thrift-idl-与-tlb-配置)
7. [常见问题与解决方案](#常见问题与解决方案)

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

## CORS HTTP Header 详解

### 请求头 (Request Headers)

#### 1. Origin

```http
Origin: https://example.com
```

| 属性 | 说明 |
|------|------|
| **作用** | 标识请求的来源（协议 + 域名 + 端口） |
| **发送时机** | 跨域请求时自动添加 |
| **可修改** | 否（浏览器自动设置，无法通过 JavaScript 修改） |

**示例场景：**
```
页面地址: https://app.example.com
请求地址: https://api.example.com/users

浏览器自动添加:
Origin: https://app.example.com
```

**注意事项：**
- 同源请求通常不发送 `Origin` 头
- `Origin` 与 `Referer` 不同：`Origin` 不包含路径信息
- 某些隐私模式或安全设置可能阻止发送 `Origin`

---

#### 2. Access-Control-Request-Method

```http
Access-Control-Request-Method: POST
```

| 属性 | 说明 |
|------|------|
| **作用** | 预检请求中告知服务器实际请求将使用的 HTTP 方法 |
| **发送时机** | 仅在预检请求 (OPTIONS) 中发送 |
| **触发条件** | 非简单请求方法（PUT、DELETE、PATCH 等） |

**预检请求示例：**
```http
OPTIONS /api/users HTTP/1.1
Host: api.example.com
Origin: https://app.example.com
Access-Control-Request-Method: DELETE
Access-Control-Request-Headers: Authorization, Content-Type
```

---

#### 3. Access-Control-Request-Headers

```http
Access-Control-Request-Headers: Authorization, Content-Type, X-Custom-Header
```

| 属性 | 说明 |
|------|------|
| **作用** | 预检请求中告知服务器实际请求将携带的自定义请求头 |
| **发送时机** | 仅在预检请求 (OPTIONS) 中发送 |
| **触发条件** | 请求包含非简单头（如 Authorization、自定义头等） |

**简单头（不触发预检）：**
- `Accept`
- `Accept-Language`
- `Content-Language`
- `Content-Type`（仅限 `text/plain`、`multipart/form-data`、`application/x-www-form-urlencoded`）

**非简单头（触发预检）：**
- `Authorization`
- `Content-Type: application/json`
- 任何自定义头（如 `X-Request-ID`、`X-API-Key`）

---

### 响应头 (Response Headers)

#### 1. Access-Control-Allow-Origin

```http
Access-Control-Allow-Origin: https://example.com
```

| 属性 | 说明 |
|------|------|
| **作用** | 指定允许访问资源的源 |
| **必需** | 是（CORS 响应必须包含） |
| **可选值** | 具体源、`*`（通配符）、`null` |

**配置方式：**

```nginx
# 方式1: 允许单个源
add_header Access-Control-Allow-Origin "https://example.com";

# 方式2: 允许所有源（不支持 credentials）
add_header Access-Control-Allow-Origin "*";

# 方式3: 动态返回请求的 Origin（推荐）
set $cors_origin "";
if ($http_origin ~* "^https://(app|admin)\.example\.com$") {
    set $cors_origin $http_origin;
}
add_header Access-Control-Allow-Origin $cors_origin;
```

**重要限制：**

| 场景 | Access-Control-Allow-Origin | Access-Control-Allow-Credentials | 是否有效 |
|------|----------------------------|----------------------------------|---------|
| 无凭证 | `*` | 不设置或 `false` | ✅ 有效 |
| 有凭证 | `*` | `true` | ❌ 无效 |
| 有凭证 | 具体源 | `true` | ✅ 有效 |

---

#### 2. Access-Control-Allow-Methods

```http
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS
```

| 属性 | 说明 |
|------|------|
| **作用** | 指定允许的 HTTP 方法 |
| **使用场景** | 预检请求的响应 |
| **注意** | 简单方法（GET、HEAD、POST）始终允许 |

**完整配置示例：**
```http
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD
```

**各方法说明：**

| 方法 | 用途 | 是否简单方法 |
|------|------|-------------|
| GET | 获取资源 | ✅ 是 |
| HEAD | 获取响应头 | ✅ 是 |
| POST | 创建资源 | ✅ 是（限简单 Content-Type） |
| PUT | 完整更新资源 | ❌ 否 |
| PATCH | 部分更新资源 | ❌ 否 |
| DELETE | 删除资源 | ❌ 否 |
| OPTIONS | 预检请求 | - |

---

#### 3. Access-Control-Allow-Headers

```http
Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With, X-Request-ID
```

| 属性 | 说明 |
|------|------|
| **作用** | 指定允许的请求头 |
| **使用场景** | 预检请求的响应 |
| **通配符** | 可使用 `*`（不支持 credentials 场景） |

**常用请求头分类：**

```
# 认证相关
Authorization          # Bearer token, Basic auth
X-API-Key             # API 密钥
X-CSRF-Token          # CSRF 防护令牌

# 内容相关
Content-Type          # 请求体类型
Accept                # 期望响应类型
Accept-Language       # 期望语言
Accept-Encoding       # 期望编码

# 追踪相关
X-Request-ID          # 请求追踪 ID
X-Correlation-ID      # 关联 ID
X-Trace-ID            # 链路追踪 ID

# 缓存相关
If-Modified-Since     # 条件请求
If-None-Match         # ETag 匹配
Cache-Control         # 缓存控制

# 自定义业务
X-Tenant-ID           # 多租户 ID
X-Client-Version      # 客户端版本
X-Device-ID           # 设备 ID
```

---

#### 4. Access-Control-Allow-Credentials

```http
Access-Control-Allow-Credentials: true
```

| 属性 | 说明 |
|------|------|
| **作用** | 指示是否允许发送 Cookie 和 HTTP 认证信息 |
| **可选值** | `true`（允许）或不设置（不允许） |
| **前端配合** | 需要设置 `credentials: 'include'` |

**工作流程：**

```
前端请求:
fetch(url, { credentials: 'include' })

↓

浏览器发送:
Cookie: session_id=abc123
Origin: https://app.example.com

↓

服务器响应:
Access-Control-Allow-Origin: https://app.example.com  (不能是 *)
Access-Control-Allow-Credentials: true
Set-Cookie: session_id=xyz789; SameSite=None; Secure

↓

浏览器:
✅ 接受响应并保存 Cookie
```

**Cookie SameSite 属性：**

| SameSite | 跨站发送 | 跨站接收 | 安全要求 |
|----------|---------|---------|---------|
| `Strict` | ❌ | ❌ | 无 |
| `Lax` | 部分（顶级导航） | ✅ | 无 |
| `None` | ✅ | ✅ | 必须 `Secure` |

---

#### 5. Access-Control-Expose-Headers

```http
Access-Control-Expose-Headers: X-Request-ID, X-Total-Count, X-Page-Count, Content-Disposition
```

| 属性 | 说明 |
|------|------|
| **作用** | 指定哪些响应头可以被前端 JavaScript 访问 |
| **默认可访问** | `Cache-Control`、`Content-Language`、`Content-Type`、`Expires`、`Last-Modified`、`Pragma` |
| **通配符** | 可使用 `*`（不支持 credentials 场景） |

**使用场景：**

```javascript
// 后端响应头
// X-Total-Count: 100
// X-Request-ID: req-123456

// 前端读取（需要 Expose）
const response = await fetch(url);
const totalCount = response.headers.get('X-Total-Count');   // 需要暴露
const requestId = response.headers.get('X-Request-ID');     // 需要暴露
const contentType = response.headers.get('Content-Type');   // 默认可访问
```

**常用暴露头：**
```
# 分页信息
X-Total-Count         # 总记录数
X-Page-Count          # 总页数
X-Current-Page        # 当前页
Link                  # 分页链接

# 请求追踪
X-Request-ID          # 请求 ID
X-Response-Time       # 响应时间

# 下载相关
Content-Disposition   # 文件下载名
Content-Length        # 内容长度

# 限流信息
X-RateLimit-Limit     # 限流上限
X-RateLimit-Remaining # 剩余次数
X-RateLimit-Reset     # 重置时间
```

---

#### 6. Access-Control-Max-Age

```http
Access-Control-Max-Age: 86400
```

| 属性 | 说明 |
|------|------|
| **作用** | 预检请求结果的缓存时间（秒） |
| **默认值** | 浏览器默认通常为 5 秒 |
| **最大值** | 浏览器有上限（Chrome: 7200秒，Firefox: 86400秒） |

**缓存策略建议：**

| 环境 | Max-Age | 说明 |
|------|---------|------|
| 开发环境 | 0 或 60 | 便于调试，快速看到配置变化 |
| 测试环境 | 3600 | 1 小时 |
| 生产环境 | 86400 | 24 小时，减少预检请求 |

**性能影响：**
```
没有缓存时:
请求1: OPTIONS → 200 → POST → 200  (2个请求)
请求2: OPTIONS → 200 → POST → 200  (2个请求)
请求3: OPTIONS → 200 → POST → 200  (2个请求)
总计: 6 个请求

有缓存时 (Max-Age: 86400):
请求1: OPTIONS → 200 → POST → 200  (2个请求)
请求2: POST → 200                  (1个请求，使用缓存)
请求3: POST → 200                  (1个请求，使用缓存)
总计: 4 个请求
```

---

### 完整 CORS 请求流程图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        浏览器 (Browser)                              │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │ JavaScript 发起跨域请求        │
                    │ fetch('https://api.example.com')│
                    └───────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │ 判断是否为简单请求？            │
                    └───────────────────────────────┘
                           │              │
                    是     │              │  否
                           ▼              ▼
              ┌─────────────────┐   ┌─────────────────────────┐
              │ 直接发送请求     │   │ 发送预检请求 (OPTIONS)   │
              │                 │   │ Origin: ...              │
              │ Origin: ...     │   │ Access-Control-Request-  │
              │ 其他请求头...    │   │   Method: DELETE         │
              └─────────────────┘   │ Access-Control-Request-  │
                      │             │   Headers: Authorization  │
                      │             └─────────────────────────┘
                      │                        │
                      │                        ▼
                      │             ┌─────────────────────────┐
                      │             │ 服务器响应预检请求       │
                      │             │ Access-Control-Allow-   │
                      │             │   Origin: ...            │
                      │             │ Access-Control-Allow-   │
                      │             │   Methods: ...           │
                      │             │ Access-Control-Allow-   │
                      │             │   Headers: ...           │
                      │             │ Access-Control-Max-Age   │
                      │             └─────────────────────────┘
                      │                        │
                      │                        ▼
                      │             ┌─────────────────────────┐
                      │             │ 预检通过？               │
                      │             └─────────────────────────┘
                      │                  │           │
                      │           是     │           │ 否
                      │                  ▼           ▼
                      │    ┌──────────────────┐  ┌────────────┐
                      │    │ 发送实际请求      │  │ CORS 错误  │
                      │    └──────────────────┘  └────────────┘
                      │                  │
                      ▼                  ▼
              ┌───────────────────────────────────┐
              │ 服务器处理请求并返回响应            │
              │ Access-Control-Allow-Origin: ...  │
              │ Access-Control-Expose-Headers: ...│
              └───────────────────────────────────┘
                                    │
                                    ▼
              ┌───────────────────────────────────┐
              │ 浏览器检查 CORS 响应头             │
              └───────────────────────────────────┘
                           │              │
                    通过   │              │  失败
                           ▼              ▼
              ┌─────────────────┐   ┌─────────────────┐
              │ JavaScript 获得 │   │ 抛出 CORS 错误   │
              │ 响应数据        │   │ TypeError:       │
              │                 │   │ Failed to fetch  │
              └─────────────────┘   └─────────────────┘
```

---

### HTTP Header 速查表

| Header | 方向 | 必需 | 说明 |
|--------|------|------|------|
| `Origin` | 请求 | 自动 | 请求来源 |
| `Access-Control-Request-Method` | 请求 | 预检 | 实际请求方法 |
| `Access-Control-Request-Headers` | 请求 | 预检 | 实际请求头 |
| `Access-Control-Allow-Origin` | 响应 | ✅ | 允许的源 |
| `Access-Control-Allow-Methods` | 响应 | 预检 | 允许的方法 |
| `Access-Control-Allow-Headers` | 响应 | 预检 | 允许的请求头 |
| `Access-Control-Allow-Credentials` | 响应 | 可选 | 允许凭证 |
| `Access-Control-Expose-Headers` | 响应 | 可选 | 暴露的响应头 |
| `Access-Control-Max-Age` | 响应 | 可选 | 预检缓存时间 |

---

## Nginx CORS 配置

### 1. 基础配置

```nginx
# /etc/nginx/conf.d/cors.conf

server {
    listen 80;
    server_name api.example.com;

    # CORS 基础配置
    location /api/ {
        # 允许的源
        add_header 'Access-Control-Allow-Origin' 'https://app.example.com' always;
        
        # 允许的方法
        add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, PATCH, OPTIONS' always;
        
        # 允许的请求头
        add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID' always;
        
        # 暴露的响应头
        add_header 'Access-Control-Expose-Headers' 'Content-Length, X-Request-ID, X-Total-Count' always;
        
        # 允许携带凭证
        add_header 'Access-Control-Allow-Credentials' 'true' always;
        
        # 预检请求缓存时间
        add_header 'Access-Control-Max-Age' '86400' always;

        # 处理预检请求
        if ($request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' 'https://app.example.com';
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, PATCH, OPTIONS';
            add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID';
            add_header 'Access-Control-Allow-Credentials' 'true';
            add_header 'Access-Control-Max-Age' '86400';
            add_header 'Content-Type' 'text/plain charset=UTF-8';
            add_header 'Content-Length' '0';
            return 204;
        }

        # 代理到后端
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. 多源动态配置

```nginx
# /etc/nginx/conf.d/cors-multi-origin.conf

# 定义允许的源列表
map $http_origin $cors_origin {
    default "";
    "https://app.example.com" $http_origin;
    "https://admin.example.com" $http_origin;
    "https://m.example.com" $http_origin;
    "http://localhost:3000" $http_origin;
    "http://localhost:5173" $http_origin;
}

# 定义是否允许凭证
map $http_origin $cors_credentials {
    default "";
    "https://app.example.com" "true";
    "https://admin.example.com" "true";
    "https://m.example.com" "true";
    "http://localhost:3000" "true";
    "http://localhost:5173" "true";
}

server {
    listen 80;
    server_name api.example.com;

    location /api/ {
        # 动态设置允许的源
        if ($cors_origin != "") {
            add_header 'Access-Control-Allow-Origin' $cors_origin always;
            add_header 'Access-Control-Allow-Credentials' $cors_credentials always;
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, PATCH, OPTIONS' always;
            add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With' always;
            add_header 'Access-Control-Expose-Headers' 'Content-Length, X-Request-ID' always;
            add_header 'Access-Control-Max-Age' '86400' always;
        }

        # 预检请求处理
        if ($request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' $cors_origin;
            add_header 'Access-Control-Allow-Credentials' $cors_credentials;
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, PATCH, OPTIONS';
            add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With';
            add_header 'Access-Control-Max-Age' '86400';
            add_header 'Content-Length' '0';
            add_header 'Content-Type' 'text/plain';
            return 204;
        }

        proxy_pass http://backend:8080;
    }
}
```

### 3. 正则匹配源

```nginx
# /etc/nginx/conf.d/cors-regex.conf

# 使用正则匹配子域名
map $http_origin $cors_origin_regex {
    default "";
    # 匹配 *.example.com
    "~^https://([a-z0-9-]+\.)?example\.com$" $http_origin;
    # 匹配开发环境
    "~^http://localhost:\d+$" $http_origin;
    # 匹配 192.168.x.x 内网
    "~^http://192\.168\.\d+\.\d+(:\d+)?$" $http_origin;
}

server {
    listen 80;
    server_name api.example.com;

    location /api/ {
        set $cors_method "";
        
        # 检查是否为允许的源
        if ($cors_origin_regex != "") {
            set $cors_method "allowed";
        }
        
        # 如果是 OPTIONS 请求
        if ($request_method = 'OPTIONS') {
            set $cors_method "${cors_method}_preflight";
        }
        
        # 处理允许的预检请求
        if ($cors_method = "allowed_preflight") {
            add_header 'Access-Control-Allow-Origin' $cors_origin_regex;
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, PATCH, OPTIONS';
            add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID, X-CSRF-Token';
            add_header 'Access-Control-Allow-Credentials' 'true';
            add_header 'Access-Control-Max-Age' '86400';
            add_header 'Content-Type' 'text/plain';
            add_header 'Content-Length' '0';
            return 204;
        }
        
        # 处理允许的实际请求
        if ($cors_method = "allowed") {
            add_header 'Access-Control-Allow-Origin' $cors_origin_regex always;
            add_header 'Access-Control-Allow-Credentials' 'true' always;
            add_header 'Access-Control-Expose-Headers' 'Content-Length, X-Request-ID, X-Total-Count' always;
        }

        proxy_pass http://backend:8080;
    }
}
```

### 4. 可复用的 CORS 配置片段

```nginx
# /etc/nginx/snippets/cors.conf

# CORS 配置片段 - 在其他配置中 include 使用

# 通用 CORS 头
add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, PATCH, OPTIONS' always;
add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID, X-CSRF-Token, Cache-Control, If-Modified-Since' always;
add_header 'Access-Control-Expose-Headers' 'Content-Length, X-Request-ID, X-Total-Count, X-Page-Count, Content-Disposition' always;
add_header 'Access-Control-Max-Age' '86400' always;
```

```nginx
# /etc/nginx/snippets/cors-preflight.conf

# 预检请求处理片段
if ($request_method = 'OPTIONS') {
    add_header 'Access-Control-Allow-Origin' $cors_origin;
    add_header 'Access-Control-Allow-Credentials' 'true';
    add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, PATCH, OPTIONS';
    add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID, X-CSRF-Token';
    add_header 'Access-Control-Max-Age' '86400';
    add_header 'Content-Type' 'text/plain';
    add_header 'Content-Length' '0';
    return 204;
}
```

```nginx
# /etc/nginx/conf.d/api.conf

# 使用配置片段
map $http_origin $cors_origin {
    default "";
    "https://app.example.com" $http_origin;
    "https://admin.example.com" $http_origin;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate /etc/ssl/certs/api.example.com.crt;
    ssl_certificate_key /etc/ssl/private/api.example.com.key;

    location /api/ {
        # 设置动态 Origin
        add_header 'Access-Control-Allow-Origin' $cors_origin always;
        add_header 'Access-Control-Allow-Credentials' 'true' always;
        
        # 包含通用 CORS 配置
        include snippets/cors.conf;
        
        # 包含预检处理
        include snippets/cors-preflight.conf;

        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 5. 完整生产配置示例

```nginx
# /etc/nginx/conf.d/production-api.conf

# 上游服务器
upstream backend_servers {
    least_conn;
    server backend1:8080 weight=5;
    server backend2:8080 weight=5;
    keepalive 32;
}

# 允许的源映射
map $http_origin $cors_allowed_origin {
    default "";
    "https://www.example.com" $http_origin;
    "https://app.example.com" $http_origin;
    "https://admin.example.com" $http_origin;
    "https://m.example.com" $http_origin;
}

# 请求限制
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/s;
limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

server {
    listen 443 ssl http2;
    server_name api.example.com;

    # SSL 配置
    ssl_certificate /etc/ssl/certs/api.example.com.crt;
    ssl_certificate_key /etc/ssl/private/api.example.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # 日志
    access_log /var/log/nginx/api.access.log combined;
    error_log /var/log/nginx/api.error.log warn;

    # API 路由
    location /api/ {
        # 请求限制
        limit_req zone=api_limit burst=50 nodelay;
        limit_conn conn_limit 20;

        # CORS 配置
        set $cors_method "";
        
        if ($cors_allowed_origin != "") {
            set $cors_method "origin_allowed";
        }
        
        if ($request_method = 'OPTIONS') {
            set $cors_method "${cors_method}_preflight";
        }

        # 预检请求
        if ($cors_method = "origin_allowed_preflight") {
            add_header 'Access-Control-Allow-Origin' $cors_allowed_origin;
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, PATCH, OPTIONS';
            add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID, X-CSRF-Token';
            add_header 'Access-Control-Allow-Credentials' 'true';
            add_header 'Access-Control-Max-Age' '86400';
            add_header 'Content-Type' 'text/plain';
            add_header 'Content-Length' '0';
            return 204;
        }

        # 实际请求
        if ($cors_method = "origin_allowed") {
            add_header 'Access-Control-Allow-Origin' $cors_allowed_origin always;
            add_header 'Access-Control-Allow-Credentials' 'true' always;
            add_header 'Access-Control-Expose-Headers' 'Content-Length, X-Request-ID, X-Total-Count, X-RateLimit-Limit, X-RateLimit-Remaining' always;
        }

        # 代理配置
        proxy_pass http://backend_servers;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Request-ID $request_id;
        
        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # 缓冲配置
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # 健康检查（不需要 CORS）
    location /health {
        access_log off;
        return 200 "OK\n";
        add_header Content-Type text/plain;
    }

    # 禁止访问隐藏文件
    location ~ /\. {
        deny all;
        return 404;
    }
}

# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name api.example.com;
    return 301 https://$server_name$request_uri;
}
```

### 6. WebSocket CORS 配置

```nginx
# /etc/nginx/conf.d/websocket.conf

map $http_origin $ws_cors_origin {
    default "";
    "https://app.example.com" $http_origin;
    "wss://app.example.com" $http_origin;
}

map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

server {
    listen 443 ssl http2;
    server_name ws.example.com;

    ssl_certificate /etc/ssl/certs/ws.example.com.crt;
    ssl_certificate_key /etc/ssl/private/ws.example.com.key;

    location /ws/ {
        # CORS 头
        add_header 'Access-Control-Allow-Origin' $ws_cors_origin always;
        add_header 'Access-Control-Allow-Credentials' 'true' always;

        # WebSocket 代理
        proxy_pass http://websocket_backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # WebSocket 超时
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
    }
}
```

### 7. Nginx + Hertz 联合部署架构

```
                    ┌─────────────────────────────────────┐
                    │            Internet                  │
                    └─────────────────────────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────────┐
                    │         Load Balancer               │
                    │      (AWS ALB / Nginx LB)           │
                    └─────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
                    ▼                 ▼                 ▼
        ┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐
        │   Nginx Node 1    │ │   Nginx Node 2    │ │   Nginx Node 3    │
        │   (CORS + SSL)    │ │   (CORS + SSL)    │ │   (CORS + SSL)    │
        │   Port: 443       │ │   Port: 443       │ │   Port: 443       │
        └───────────────────┘ └───────────────────┘ └───────────────────┘
                    │                 │                 │
                    └─────────────────┼─────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
                    ▼                 ▼                 ▼
        ┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐
        │  Hertz Backend 1  │ │  Hertz Backend 2  │ │  Hertz Backend 3  │
        │   (Go App)        │ │   (Go App)        │ │   (Go App)        │
        │   Port: 8080      │ │   Port: 8080      │ │   Port: 8080      │
        └───────────────────┘ └───────────────────┘ └───────────────────┘
```

**配置职责分工：**

| 职责 | Nginx | Hertz |
|------|-------|-------|
| SSL 终止 | ✅ | ❌ |
| CORS 处理 | ✅ (推荐) | 可选（作为备份） |
| 负载均衡 | ✅ | ❌ |
| 静态文件 | ✅ | ❌ |
| 限流 | ✅ | 可选（应用级） |
| 压缩 | ✅ | 可选 |
| 业务逻辑 | ❌ | ✅ |

### 8. Docker Compose 部署示例

```yaml
# docker-compose.yml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - ./nginx/snippets:/etc/nginx/snippets:ro
      - ./nginx/ssl:/etc/ssl:ro
      - ./nginx/logs:/var/log/nginx
    depends_on:
      - backend
    networks:
      - app-network
    restart: unless-stopped

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    expose:
      - "8080"
    environment:
      - SERVER_HOST=0.0.0.0
      - SERVER_PORT=8080
      # 如果 Nginx 处理 CORS，后端可以不配置
      - CORS_ENABLED=false
    networks:
      - app-network
    restart: unless-stopped
    deploy:
      replicas: 3

networks:
  app-network:
    driver: bridge
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

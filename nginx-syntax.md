# Nginx 配置语法详解

本文档详细介绍 Nginx 配置文件的语法，包括指令、变量、条件判断、重写规则等核心概念。

## 目录

1. [基础语法结构](#基础语法结构)
2. [核心指令详解](#核心指令详解)
3. [变量系统](#变量系统)
4. [条件判断 (if)](#条件判断-if)
5. [set 指令](#set-指令)
6. [map 指令](#map-指令)
7. [rewrite 指令](#rewrite-指令)
8. [location 匹配规则](#location-匹配规则)
9. [正则表达式](#正则表达式)
10. [常用配置模式](#常用配置模式)
11. [图片处理与水印](#图片处理与水印)

---

## 基础语法结构

### 配置文件结构

```nginx
# 全局块 - 影响整个 Nginx 服务器
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

# events 块 - 影响网络连接
events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

# http 块 - HTTP 服务器配置
http {
    # http 全局配置
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # server 块 - 虚拟主机配置
    server {
        listen 80;
        server_name example.com;

        # location 块 - URI 匹配配置
        location / {
            root /var/www/html;
            index index.html;
        }
    }
}

# stream 块 - TCP/UDP 代理（可选）
stream {
    server {
        listen 3306;
        proxy_pass mysql_backend;
    }
}
```

### 配置层级关系

```
┌─────────────────────────────────────────────────────────────┐
│                        main (全局)                           │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                      events                            │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                        http                            │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │                    server                        │  │  │
│  │  │  ┌───────────────────────────────────────────┐  │  │  │
│  │  │  │              location                      │  │  │  │
│  │  │  │  ┌─────────────────────────────────────┐  │  │  │  │
│  │  │  │  │         if / limit_except           │  │  │  │  │
│  │  │  │  └─────────────────────────────────────┘  │  │  │  │
│  │  │  └───────────────────────────────────────────┘  │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 基本语法规则

```nginx
# 1. 指令以分号结尾
worker_processes 4;

# 2. 块指令使用大括号
server {
    listen 80;
}

# 3. 注释使用 #
# 这是注释
server_name example.com;  # 行尾注释

# 4. 字符串可以不加引号（除非包含空格或特殊字符）
root /var/www/html;
root "/var/www/my site";  # 包含空格需要引号

# 5. 布尔值: on/off
gzip on;
sendfile off;

# 6. 大小单位
client_max_body_size 10m;   # m = 兆字节
client_body_buffer_size 8k; # k = 千字节
# 可用单位: k/K (千字节), m/M (兆字节), g/G (吉字节)

# 7. 时间单位
keepalive_timeout 65s;      # s = 秒
proxy_read_timeout 60;      # 默认秒
send_timeout 30m;           # m = 分钟
# 可用单位: ms (毫秒), s (秒), m (分钟), h (小时), d (天)

# 8. include 引入其他配置
include /etc/nginx/conf.d/*.conf;
include /etc/nginx/sites-enabled/*;
```

---

## 核心指令详解

### 全局指令

```nginx
# ============ 进程管理 ============

# 运行用户
user nginx nginx;              # 用户 组

# 工作进程数（通常设置为 CPU 核心数）
worker_processes auto;         # auto 自动检测
worker_processes 4;            # 手动指定

# CPU 亲和性绑定
worker_cpu_affinity auto;
worker_cpu_affinity 0001 0010 0100 1000;  # 4 核绑定

# 进程优先级 (-20 到 20，越小优先级越高)
worker_priority -10;

# 单个进程最大打开文件数
worker_rlimit_nofile 65535;

# PID 文件位置
pid /var/run/nginx.pid;

# ============ 错误日志 ============

# 日志级别: debug, info, notice, warn, error, crit, alert, emerg
error_log /var/log/nginx/error.log warn;
error_log /var/log/nginx/debug.log debug;  # 调试日志
error_log syslog:server=192.168.1.1 info;  # 发送到 syslog

# ============ 事件模型 ============

events {
    # 单个进程最大连接数
    worker_connections 10240;
    
    # 事件模型（Linux 推荐 epoll）
    use epoll;           # Linux
    # use kqueue;        # FreeBSD/macOS
    
    # 是否同时接受多个连接
    multi_accept on;
    
    # 是否使用互斥锁（负载均衡用）
    accept_mutex on;
    accept_mutex_delay 500ms;
}
```

### HTTP 块指令

```nginx
http {
    # ============ MIME 类型 ============
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # ============ 日志格式 ============
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';
    
    log_format json escape=json '{'
        '"time": "$time_iso8601",'
        '"remote_addr": "$remote_addr",'
        '"request": "$request",'
        '"status": "$status",'
        '"body_bytes_sent": "$body_bytes_sent",'
        '"request_time": "$request_time",'
        '"upstream_response_time": "$upstream_response_time"'
    '}';

    access_log /var/log/nginx/access.log main;
    access_log /var/log/nginx/access.json.log json;
    access_log off;  # 关闭访问日志

    # ============ 性能优化 ============
    
    # 零拷贝
    sendfile on;
    
    # 减少网络包数量
    tcp_nopush on;
    tcp_nodelay on;
    
    # 长连接
    keepalive_timeout 65;
    keepalive_requests 100;
    
    # 客户端超时
    client_body_timeout 60s;
    client_header_timeout 60s;
    send_timeout 60s;
    
    # 缓冲区大小
    client_body_buffer_size 16k;
    client_header_buffer_size 1k;
    client_max_body_size 100m;
    large_client_header_buffers 4 8k;

    # ============ Gzip 压缩 ============
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_min_length 1000;
    gzip_types text/plain text/css text/xml application/json 
               application/javascript application/xml+rss 
               application/atom+xml image/svg+xml;

    # ============ 代理配置 ============
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;
    proxy_buffer_size 4k;
    proxy_buffers 4 32k;
    proxy_busy_buffers_size 64k;
    proxy_temp_file_write_size 64k;

    # 代理头设置
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # ============ 包含其他配置 ============
    include /etc/nginx/conf.d/*.conf;
}
```

### Server 块指令

```nginx
server {
    # ============ 监听配置 ============
    
    # 基本监听
    listen 80;
    listen 443 ssl;
    
    # 完整语法
    listen 80 default_server;              # 默认服务器
    listen 443 ssl http2;                  # SSL + HTTP/2
    listen [::]:80;                        # IPv6
    listen 80 reuseport;                   # 端口复用
    listen unix:/var/run/nginx.sock;       # Unix Socket
    
    # SSL 监听选项
    listen 443 ssl http2 backlog=1024 so_keepalive=on;

    # ============ 服务器名 ============
    
    # 精确匹配
    server_name example.com www.example.com;
    
    # 通配符
    server_name *.example.com;             # 前缀通配符
    server_name example.*;                 # 后缀通配符
    
    # 正则表达式（以 ~ 开头）
    server_name ~^www\d+\.example\.com$;
    server_name ~^(?<subdomain>.+)\.example\.com$;  # 命名捕获
    
    # 匹配所有
    server_name _;
    
    # 匹配优先级（从高到低）:
    # 1. 精确匹配
    # 2. 以 * 开头的最长通配符
    # 3. 以 * 结尾的最长通配符
    # 4. 按配置文件顺序的第一个正则匹配

    # ============ SSL 配置 ============
    ssl_certificate /etc/ssl/certs/example.com.crt;
    ssl_certificate_key /etc/ssl/private/example.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;
    
    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    ssl_trusted_certificate /etc/ssl/certs/ca-bundle.crt;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;

    # ============ 根目录与索引 ============
    root /var/www/html;
    index index.html index.htm index.php;

    # ============ 错误页面 ============
    error_page 404 /404.html;
    error_page 500 502 503 504 /50x.html;
    error_page 404 = @notfound;  # 重定向到命名 location

    # ============ 访问控制 ============
    allow 192.168.1.0/24;
    deny all;
    
    # 认证
    auth_basic "Restricted Area";
    auth_basic_user_file /etc/nginx/.htpasswd;
}
```

---

## 变量系统

### 内置变量分类

#### 请求相关变量

```nginx
# ============ 请求行变量 ============
$request            # 完整请求行: "GET /path?query HTTP/1.1"
$request_method     # 请求方法: GET, POST, PUT, DELETE 等
$request_uri        # 原始 URI（包含参数）: /path?query
$uri                # 当前 URI（不含参数，可被 rewrite 修改）
$document_uri       # 同 $uri
$args               # 查询字符串: key1=value1&key2=value2
$query_string       # 同 $args
$is_args            # 如果有参数则为 "?"，否则为空

# 示例: GET /api/users?page=1&limit=10 HTTP/1.1
# $request_method = "GET"
# $request_uri    = "/api/users?page=1&limit=10"
# $uri            = "/api/users"
# $args           = "page=1&limit=10"
# $is_args        = "?"

# ============ 请求参数变量 ============
$arg_name           # 获取 GET 参数: $arg_page = "1"
$arg_PARAMETER      # 参数名不区分大小写

# ============ 请求头变量 ============
$http_host          # Host 头
$http_user_agent    # User-Agent 头
$http_referer       # Referer 头
$http_cookie        # Cookie 头
$http_accept        # Accept 头
$http_authorization # Authorization 头
$http_x_forwarded_for      # X-Forwarded-For 头
$http_x_real_ip            # X-Real-IP 头
$http_x_requested_with     # X-Requested-With 头（AJAX 检测）
$http_HEADER        # 任意请求头（小写，- 替换为 _）

# ============ 请求体变量 ============
$request_body       # 请求体内容
$request_body_file  # 请求体临时文件路径
$content_type       # Content-Type 头
$content_length     # Content-Length 头
```

#### 连接与客户端变量

```nginx
# ============ 客户端信息 ============
$remote_addr        # 客户端 IP 地址
$remote_port        # 客户端端口
$remote_user        # Basic 认证的用户名

$binary_remote_addr # 客户端 IP（二进制格式，用于 limit_req_zone）

# ============ 服务器信息 ============
$server_addr        # 服务器 IP
$server_port        # 服务器端口
$server_name        # 匹配的 server_name
$server_protocol    # 协议版本: HTTP/1.0, HTTP/1.1, HTTP/2.0

$host               # 请求的主机名（按优先级）:
                    # 1. 请求行中的主机名
                    # 2. Host 头
                    # 3. 匹配的 server_name

$hostname           # 服务器的 hostname

# ============ 连接信息 ============
$connection         # 连接序号
$connection_requests # 当前连接的请求数
$scheme             # 协议: http 或 https
$https              # 如果是 HTTPS 则为 "on"，否则为空

# ============ SSL 相关 ============
$ssl_protocol       # SSL 协议版本: TLSv1.2, TLSv1.3
$ssl_cipher         # SSL 加密套件
$ssl_client_cert    # 客户端证书（PEM 格式）
$ssl_client_s_dn    # 客户端证书的 Subject DN
$ssl_session_id     # SSL 会话 ID
$ssl_session_reused # 会话是否重用: "r" 或 "."
```

#### 响应相关变量

```nginx
# ============ 响应信息 ============
$status             # 响应状态码: 200, 404, 500 等
$body_bytes_sent    # 发送的响应体字节数
$bytes_sent         # 发送的总字节数（含响应头）

# ============ 响应头变量 ============
$sent_http_content_type     # Content-Type 响应头
$sent_http_content_length   # Content-Length 响应头
$sent_http_location         # Location 响应头
$sent_http_HEADER           # 任意响应头

# ============ 时间变量 ============
$time_iso8601       # ISO 8601 格式时间: 2024-01-15T10:30:00+08:00
$time_local         # 本地格式时间: 15/Jan/2024:10:30:00 +0800
$msec               # 当前时间（Unix 时间戳，带毫秒）
$request_time       # 请求处理时间（秒，精确到毫秒）
$request_length     # 请求长度（包括请求行、头、体）
```

#### 代理与上游变量

```nginx
# ============ 上游信息 ============
$upstream_addr              # 上游服务器地址
$upstream_status            # 上游响应状态码
$upstream_response_time     # 上游响应时间
$upstream_response_length   # 上游响应长度
$upstream_connect_time      # 与上游建立连接的时间
$upstream_header_time       # 接收上游响应头的时间
$upstream_cache_status      # 缓存状态: HIT, MISS, BYPASS, EXPIRED 等

# ============ 代理信息 ============
$proxy_host                 # proxy_pass 的主机名
$proxy_port                 # proxy_pass 的端口
$proxy_add_x_forwarded_for  # 客户端 IP + 现有 X-Forwarded-For
```

#### 路径与文件变量

```nginx
# ============ 文件路径 ============
$document_root      # root 或 alias 指定的根目录
$realpath_root      # document_root 的绝对路径
$request_filename   # 请求文件的完整路径

# 示例: root /var/www; 请求 /images/logo.png
# $document_root    = "/var/www"
# $request_filename = "/var/www/images/logo.png"

# ============ 限制 ============
$limit_rate         # 响应速率限制（字节/秒）
```

#### Nginx 内部变量

```nginx
# ============ 内部状态 ============
$nginx_version      # Nginx 版本号
$pid                # 当前 worker 进程 ID
$request_id         # 16 位随机请求 ID（1.11.0+）

# ============ 正则捕获 ============
$1, $2, $3 ...      # 正则表达式捕获组
```

### 变量使用示例

```nginx
# 1. 日志格式中使用变量
log_format detailed '$remote_addr - [$time_local] '
                    '"$request" $status $body_bytes_sent '
                    '"$http_referer" "$http_user_agent" '
                    'rt=$request_time ut=$upstream_response_time';

# 2. 在响应头中使用变量
add_header X-Request-ID $request_id;
add_header X-Server $hostname;

# 3. 条件判断中使用变量
if ($request_method = POST) {
    # ...
}

# 4. 在 proxy_pass 中使用变量
location /api/ {
    proxy_pass http://backend$request_uri;
}

# 5. 组合使用变量
set $cache_key "$scheme$host$request_uri";
```

---

## 条件判断 (if)

### if 语法

```nginx
if (condition) {
    # 指令
}
```

### 条件表达式

#### 1. 变量判断

```nginx
# 变量存在且不为空字符串
if ($variable) {
    # $variable 不为空时执行
}

# 变量为空或不存在
if ($variable = "") {
    # $variable 为空时执行
}
```

#### 2. 字符串比较

```nginx
# 相等
if ($request_method = POST) {
    return 405;
}

# 不相等
if ($request_method != GET) {
    return 405;
}

# 注意: 比较运算符两边必须有空格
# 正确: if ($var = "value")
# 错误: if ($var="value")
```

#### 3. 正则匹配

```nginx
# 区分大小写匹配
if ($http_user_agent ~ MSIE) {
    # 包含 MSIE
}

# 不区分大小写匹配
if ($http_user_agent ~* chrome) {
    # 包含 chrome/Chrome/CHROME 等
}

# 不匹配
if ($http_user_agent !~ MSIE) {
    # 不包含 MSIE
}

# 不区分大小写不匹配
if ($http_user_agent !~* bot) {
    # 不包含 bot
}

# 使用捕获组
if ($request_uri ~* ^/api/v(\d+)/(.*)$) {
    set $api_version $1;
    set $api_path $2;
}
```

#### 4. 文件/目录判断

```nginx
# 文件存在
if (-f $request_filename) {
    # 文件存在
}

# 文件不存在
if (!-f $request_filename) {
    # 文件不存在
}

# 目录存在
if (-d $request_filename) {
    # 是目录
}

# 文件或目录存在
if (-e $request_filename) {
    # 存在
}

# 文件可执行
if (-x $request_filename) {
    # 可执行
}
```

### if 使用场景

#### 1. 请求方法限制

```nginx
location /api/ {
    # 只允许 GET 和 POST
    if ($request_method !~ ^(GET|POST)$) {
        return 405;
    }
    
    proxy_pass http://backend;
}
```

#### 2. User-Agent 过滤

```nginx
# 阻止恶意爬虫
if ($http_user_agent ~* (wget|curl|scrapy|spider)) {
    return 403;
}

# 移动设备重定向
if ($http_user_agent ~* (mobile|android|iphone)) {
    return 302 https://m.example.com$request_uri;
}
```

#### 3. Referer 防盗链

```nginx
location ~* \.(jpg|jpeg|png|gif|webp)$ {
    valid_referers none blocked server_names *.example.com;
    
    if ($invalid_referer) {
        return 403;
    }
}
```

#### 4. CORS 处理

```nginx
# 动态设置 CORS Origin
set $cors_origin "";
if ($http_origin ~* "^https://(app|admin)\.example\.com$") {
    set $cors_origin $http_origin;
}

if ($cors_origin != "") {
    add_header Access-Control-Allow-Origin $cors_origin;
}
```

#### 5. HTTPS 重定向

```nginx
# 方法1: 使用 if
if ($scheme = http) {
    return 301 https://$host$request_uri;
}

# 方法2: 推荐方式（独立 server 块）
server {
    listen 80;
    server_name example.com;
    return 301 https://$host$request_uri;
}
```

### if 的陷阱与注意事项

#### "If is Evil" 问题

Nginx 的 `if` 指令在 location 上下文中可能导致意外行为：

```nginx
# ⚠️ 危险示例
location / {
    set $true 1;
    
    if ($true) {
        # 这里创建了一个新的内部 location
        # 某些指令不会继承
    }
    
    # 这些指令可能不会在 if 块内生效
    proxy_pass http://backend;
}
```

#### if 中安全的指令

```nginx
# ✅ 安全使用
if ($condition) {
    return ...;      # 安全
    rewrite ...;     # 安全
    set $var ...;    # 安全
}

# ⚠️ 可能有问题
if ($condition) {
    proxy_pass ...;   # 可能有问题
    fastcgi_pass ...; # 可能有问题
    root ...;         # 可能有问题
    # 其他内容处理指令
}
```

#### 替代方案

```nginx
# ❌ 不推荐
location / {
    if ($request_uri ~* ^/api/) {
        proxy_pass http://api_backend;
    }
    proxy_pass http://web_backend;
}

# ✅ 推荐: 使用多个 location
location /api/ {
    proxy_pass http://api_backend;
}

location / {
    proxy_pass http://web_backend;
}
```

```nginx
# ❌ 不推荐
location / {
    if ($http_user_agent ~* mobile) {
        root /var/www/mobile;
    }
    root /var/www/desktop;
}

# ✅ 推荐: 使用 map
map $http_user_agent $root_path {
    default "/var/www/desktop";
    ~*mobile "/var/www/mobile";
}

server {
    root $root_path;
}
```

---

## set 指令

### 基本语法

```nginx
set $variable value;
```

### Nginx 变量类型详解

#### 核心概念：Nginx 变量只有字符串类型

**重要：Nginx 中所有变量都是字符串类型，没有整数、布尔、数组等类型。**

```nginx
# 所有这些都是字符串
set $string "hello";
set $number "42";           # 这是字符串 "42"，不是数字
set $boolean "true";        # 这是字符串 "true"，不是布尔值
set $empty "";              # 空字符串
```

#### 变量的"假值"判断

在 Nginx 的 `if` 语句中，以下值被视为"假"：

```nginx
# 假值 (false)
set $var "";              # 空字符串
set $var "0";             # 字符串 "0"
# 未定义的变量

# 真值 (true) - 其他所有非空字符串
set $var "1";
set $var "true";
set $var "false";         # 注意：字符串 "false" 是真值！
set $var "hello";
set $var " ";             # 空格也是真值
```

**示例：**

```nginx
location / {
    set $flag "";
    
    # 空字符串被视为假
    if ($flag) {
        # 不会执行，因为 $flag 是空字符串
        return 200 "flag is set";
    }
    
    set $flag "0";
    if ($flag) {
        # 不会执行，因为 "0" 被视为假
        return 200 "flag is 0";
    }
    
    set $flag "1";
    if ($flag) {
        # 会执行，因为 "1" 是真值
        return 200 "flag is 1";
    }
}
```

#### 变量类型转换

```nginx
# ============ 数字运算 ============
# Nginx 本身不支持算术运算，需要借助其他模块

# 方法1: 使用 ngx_http_lua_module
set_by_lua_block $sum {
    return tonumber(ngx.var.a) + tonumber(ngx.var.b)
}

# 方法2: 使用 ngx_http_set_misc_module
set $a 10;
set $b 20;
set_eval $sum "$a + $b";

# 方法3: 使用 map 模拟
map $arg_count $next_count {
    default "1";
    "1"     "2";
    "2"     "3";
    "3"     "4";
    # ... 有限的映射
}

# ============ 字符串操作 ============
# 字符串拼接 - 直接拼接
set $full_url "$scheme://$host$request_uri";
set $greeting "Hello, $arg_name!";

# 字符串包含变量时的大括号语法
set $file "${arg_name}_file.txt";     # 避免歧义
set $path "/data/${env}_config.json";
```

#### 变量作用域

```nginx
# ============ 变量作用域规则 ============

# 1. 变量在整个请求生命周期内有效
# 2. 变量在不同 location 之间共享（同一请求）
# 3. 变量不在不同请求之间共享

server {
    # server 级别设置变量
    set $server_var "server_value";
    
    location /first {
        set $loc_var "first_value";
        # 可以访问 $server_var
        return 200 "server: $server_var, loc: $loc_var";
    }
    
    location /second {
        # 可以访问 $server_var
        # 如果是内部跳转，可以访问 $loc_var
        return 200 "server: $server_var";
    }
    
    location /redirect {
        set $my_var "original";
        # 内部重定向后，变量保持
        rewrite ^ /target last;
    }
    
    location /target {
        # 如果是从 /redirect 重定向来的，$my_var 仍然可用
        return 200 "my_var: $my_var";
    }
}
```

#### 变量初始化时机

```nginx
# ============ 变量在请求处理时才求值 ============

map $uri $backend {
    default "backend_a";
    /api    "backend_b";
}

server {
    # $backend 在这里并没有被计算
    # 而是在实际使用时才求值
    
    location / {
        # 此时 $backend 才被求值
        proxy_pass http://$backend;
    }
}
```

#### 特殊变量行为

```nginx
# ============ 内置变量是只读的 ============

location / {
    # ❌ 错误：不能修改内置变量
    # set $uri "/new/path";
    # set $host "new.example.com";
    
    # ✅ 正确：使用自定义变量
    set $my_uri $uri;
    set $my_host $host;
    
    # 然后修改自定义变量
    if ($my_uri ~ ^/old/) {
        set $my_uri "/new/";
    }
}

# ============ 某些变量可写 ============

location / {
    # $args 可以被修改
    set $args "modified=true&$args";
    
    # $limit_rate 可以被修改
    set $limit_rate 100k;
}
```

#### 变量类型速查表

| 类型 | Nginx 表示 | 示例 | 说明 |
|------|-----------|------|------|
| 字符串 | `"text"` 或 `text` | `set $var "hello";` | 唯一的实际类型 |
| 数字 | `"123"` | `set $var "42";` | 实际是字符串 |
| 布尔-真 | 非空非零字符串 | `set $var "1";` | "1", "true", "yes" 等 |
| 布尔-假 | 空或 "0" | `set $var "";` | "", "0" |
| 空/未定义 | 空字符串 | `set $var "";` | 等同于假值 |
| 列表/数组 | 不支持 | - | 使用 map 或多个变量 |

#### 模拟布尔逻辑

```nginx
# ============ 模拟布尔值 ============

location / {
    # 初始化为 "假"
    set $is_allowed "";
    set $is_admin "";
    
    # 设置为 "真"
    if ($remote_addr ~ ^192\.168\.) {
        set $is_allowed "1";
    }
    
    if ($http_x_admin_token = "secret") {
        set $is_admin "1";
    }
    
    # AND 逻辑
    set $check "$is_allowed$is_admin";
    if ($check = "11") {
        # 两个条件都为真
        return 200 "Access granted";
    }
    
    # OR 逻辑
    if ($is_allowed) {
        return 200 "Allowed by IP";
    }
    if ($is_admin) {
        return 200 "Allowed by token";
    }
    
    return 403 "Forbidden";
}
```

#### 模拟数组/列表

```nginx
# ============ 使用 map 模拟数组 ============

# 方法1: 使用 map 做查找
map $arg_color $color_hex {
    default "#000000";
    "red"    "#FF0000";
    "green"  "#00FF00";
    "blue"   "#0000FF";
    "white"  "#FFFFFF";
}

# 方法2: 使用分隔符字符串
set $allowed_methods "GET,POST,PUT,DELETE";

# 检查是否在列表中（使用正则）
if ($allowed_methods !~ $request_method) {
    return 405;
}

# 方法3: 使用多个变量
set $item_0 "first";
set $item_1 "second";
set $item_2 "third";
```

#### 变量调试技巧

```nginx
# ============ 调试变量值 ============

# 方法1: 返回变量值
location /debug {
    default_type text/plain;
    return 200 "
uri: $uri
args: $args
host: $host
my_var: $my_var
";
}

# 方法2: 添加响应头
location / {
    add_header X-Debug-Var $my_var;
    add_header X-Debug-URI $uri;
    proxy_pass http://backend;
}

# 方法3: 写入日志
log_format debug_log '$remote_addr - $request - my_var=$my_var';
access_log /var/log/nginx/debug.log debug_log;

# 方法4: 使用 echo 模块（需要安装）
location /echo {
    echo "Variable value: $my_var";
    echo "Request URI: $uri";
}
```

#### 常见陷阱

```nginx
# ============ 陷阱1: 字符串 "0" vs 空字符串 ============

set $count "0";
if ($count) {
    # ❌ 不会执行！"0" 被视为假
}

set $count "00";
if ($count) {
    # ✅ 会执行！"00" 不是 "0"，被视为真
}

# ============ 陷阱2: 变量未定义 ============

# 未定义的变量被视为空字符串
if ($undefined_var) {
    # 不会执行
}

# 使用未定义变量不会报错
return 200 "Value: $nonexistent";  # 输出: "Value: "

# ============ 陷阱3: 引号处理 ============

# 引号内的变量会被解析
set $greeting "Hello, $name!";  # 变量会被替换

# 单引号不是特殊语法（Nginx 使用双引号）
set $text 'single quotes';  # ❌ 语法错误

# ============ 陷阱4: 变量名大小写 ============

# 变量名区分大小写
set $MyVar "value1";
set $myvar "value2";
# $MyVar 和 $myvar 是不同的变量

# 但 HTTP 头变量转换为小写
# X-Custom-Header 变成 $http_x_custom_header

# ============ 陷阱5: 正则捕获组覆盖 ============

location ~ ^/user/(\d+) {
    set $user_id $1;  # 保存捕获组
    
    if ($uri ~ ^/user/(\d+)/profile) {
        # ⚠️ 这里的 $1 会覆盖之前的 $1
        # 使用之前保存的 $user_id
    }
}
```

### 使用场景

#### 1. 定义常量

```nginx
server {
    set $backend_host "api.internal.example.com";
    set $backend_port "8080";
    
    location /api/ {
        proxy_pass http://$backend_host:$backend_port;
    }
}
```

#### 2. 条件赋值

```nginx
location / {
    # 默认值
    set $cors_origin "";
    set $cors_credentials "";
    
    # 条件赋值
    if ($http_origin ~* "^https://.*\.example\.com$") {
        set $cors_origin $http_origin;
        set $cors_credentials "true";
    }
    
    add_header Access-Control-Allow-Origin $cors_origin;
    add_header Access-Control-Allow-Credentials $cors_credentials;
}
```

#### 3. 变量组合

```nginx
location / {
    # 组合请求信息
    set $cache_key "$scheme://$host$request_uri";
    
    # 组合代理目标
    set $upstream_url "http://backend:8080$request_uri";
    
    proxy_cache_key $cache_key;
    proxy_pass $upstream_url;
}
```

#### 4. 复杂条件判断（状态机模式）

```nginx
location /protected/ {
    # 初始化状态
    set $allow "";
    
    # 检查 IP
    if ($remote_addr ~ "^192\.168\.") {
        set $allow "ip_ok";
    }
    
    # 检查 Token
    if ($http_authorization ~ "^Bearer\s+valid_token") {
        set $allow "${allow}_token_ok";
    }
    
    # 检查最终状态
    if ($allow != "ip_ok_token_ok") {
        return 403;
    }
    
    proxy_pass http://backend;
}
```

#### 5. 多条件组合

```nginx
# 需求: 同时满足多个条件才执行操作
location / {
    set $flag "";
    
    # 条件1: POST 请求
    if ($request_method = POST) {
        set $flag "A";
    }
    
    # 条件2: 特定 Content-Type
    if ($content_type ~* "application/json") {
        set $flag "${flag}B";
    }
    
    # 条件3: 特定路径
    if ($uri ~* "^/api/") {
        set $flag "${flag}C";
    }
    
    # 检查是否满足所有条件
    if ($flag = "ABC") {
        # 满足所有条件
        rewrite ^ /api-handler last;
    }
}
```

#### 6. 请求参数处理

```nginx
location /search {
    # 提取查询参数
    set $search_query $arg_q;
    set $page $arg_page;
    set $limit $arg_limit;
    
    # 设置默认值
    if ($page = "") {
        set $page "1";
    }
    
    if ($limit = "") {
        set $limit "10";
    }
    
    # 构建后端请求
    proxy_pass http://search_backend?query=$search_query&page=$page&limit=$limit;
}
```

#### 7. 动态代理目标

```nginx
# 根据路径选择后端
location ~ ^/service/(?<service_name>[^/]+)/(?<service_path>.*)$ {
    set $backend "";
    
    if ($service_name = "users") {
        set $backend "user_service:8080";
    }
    
    if ($service_name = "orders") {
        set $backend "order_service:8080";
    }
    
    if ($service_name = "products") {
        set $backend "product_service:8080";
    }
    
    if ($backend = "") {
        return 404;
    }
    
    proxy_pass http://$backend/$service_path$is_args$args;
}
```

---

## map 指令

### 基本语法

```nginx
map $source_variable $target_variable {
    default default_value;
    value1  result1;
    value2  result2;
    ~regex  result3;
}
```

### map 使用示例

#### 1. 简单值映射

```nginx
# 根据文件扩展名设置缓存时间
map $uri $cache_control {
    default         "no-cache";
    ~*\.html$       "no-cache";
    ~*\.css$        "max-age=31536000";
    ~*\.js$         "max-age=31536000";
    ~*\.(jpg|png)$  "max-age=86400";
}

server {
    location / {
        add_header Cache-Control $cache_control;
    }
}
```

#### 2. CORS 源白名单

```nginx
map $http_origin $cors_origin {
    default "";
    "https://app.example.com"    $http_origin;
    "https://admin.example.com"  $http_origin;
    "https://m.example.com"      $http_origin;
    "http://localhost:3000"      $http_origin;
    ~^https://.*\.example\.com$  $http_origin;
}

map $http_origin $cors_credentials {
    default "";
    "https://app.example.com"    "true";
    "https://admin.example.com"  "true";
    ~^https://.*\.example\.com$  "true";
}
```

#### 3. 设备类型检测

```nginx
map $http_user_agent $device_type {
    default "desktop";
    ~*mobile "mobile";
    ~*tablet "tablet";
    ~*ipad   "tablet";
    ~*android.*mobile "mobile";
    ~*android "tablet";
}

map $device_type $root_path {
    default "/var/www/desktop";
    mobile  "/var/www/mobile";
    tablet  "/var/www/tablet";
}

server {
    root $root_path;
}
```

#### 4. 后端服务路由

```nginx
map $uri $backend {
    default         "default_backend";
    ~^/api/users    "user_service";
    ~^/api/orders   "order_service";
    ~^/api/products "product_service";
    ~^/api/auth     "auth_service";
}

upstream default_backend {
    server backend:8080;
}

upstream user_service {
    server user-svc:8080;
}

upstream order_service {
    server order-svc:8080;
}

# ... 其他 upstream

server {
    location /api/ {
        proxy_pass http://$backend;
    }
}
```

#### 5. 限流配置

```nginx
# 根据客户端类型设置不同限流
map $http_x_api_key $limit_key {
    default         $binary_remote_addr;  # 普通用户按 IP 限流
    "premium_key_1" "";                   # VIP 不限流
    "premium_key_2" "";
    ~^internal_     "";                   # 内部服务不限流
}

limit_req_zone $limit_key zone=api_limit:10m rate=10r/s;

server {
    location /api/ {
        limit_req zone=api_limit burst=20 nodelay;
        proxy_pass http://backend;
    }
}
```

#### 6. A/B 测试

```nginx
# 基于 Cookie 的 A/B 测试
map $cookie_ab_test $ab_backend {
    default "backend_a";
    "a"     "backend_a";
    "b"     "backend_b";
}

# 如果没有 Cookie，随机分配
split_clients "${remote_addr}AAA" $ab_group {
    50%     "a";
    *       "b";
}

map $cookie_ab_test $ab_final_backend {
    default $ab_backend;
    ""      $ab_group;
}

server {
    location / {
        # 设置 Cookie
        add_header Set-Cookie "ab_test=$ab_final_backend; Path=/; Max-Age=86400";
        proxy_pass http://$ab_final_backend;
    }
}
```

#### 7. 多级 map

```nginx
# 第一级: 提取主域名
map $host $main_domain {
    default                     "unknown";
    ~^(?<sub>.+\.)?example\.com$ "example.com";
    ~^(?<sub>.+\.)?another\.com$ "another.com";
}

# 第二级: 根据域名选择配置
map $main_domain $site_config {
    default       "default_config";
    "example.com" "example_config";
    "another.com" "another_config";
}
```

### map 高级选项

```nginx
map $var $result {
    # 大小写敏感（默认）
    default "not_found";
    Value   "found_case_sensitive";
}

map $var $result {
    # 主机名匹配（类似 server_name）
    hostnames;
    default          "not_found";
    .example.com     "found";
    *.example.com    "found";
}

map $var $result {
    # 包含正则时的挥发性（需要每次计算）
    volatile;
    default "not_found";
    ~regex  "found";
}
```

---

## rewrite 指令

### 基本语法

```nginx
rewrite regex replacement [flag];
```

### Flag 类型

| Flag | 说明 |
|------|------|
| `last` | 停止当前 rewrite，重新搜索 location |
| `break` | 停止所有 rewrite，继续当前 location |
| `redirect` | 302 临时重定向 |
| `permanent` | 301 永久重定向 |

### rewrite 示例

#### 1. 基本重写

```nginx
# 移除 .html 扩展名
rewrite ^/(.*)\.html$ /$1 last;

# 添加尾部斜杠
rewrite ^([^.]*[^/])$ $1/ permanent;

# 去除尾部斜杠
rewrite ^/(.*)/$ /$1 permanent;
```

#### 2. URL 标准化

```nginx
server {
    # 强制 www
    if ($host !~* ^www\.) {
        rewrite ^(.*)$ https://www.$host$1 permanent;
    }
    
    # 去除 www
    if ($host ~* ^www\.(.+)$) {
        set $host_without_www $1;
        rewrite ^(.*)$ https://$host_without_www$1 permanent;
    }
}
```

#### 3. 旧 URL 重定向

```nginx
location / {
    # 旧博客 URL 重定向
    rewrite ^/blog/(\d{4})/(\d{2})/(.*)$ /posts/$1-$2-$3 permanent;
    
    # 旧产品 URL
    rewrite ^/products/(\d+)\.html$ /product/$1 permanent;
    
    # 批量重定向
    rewrite ^/old-path/(.*)$ /new-path/$1 permanent;
}
```

#### 4. 内部重写

```nginx
# 前端路由支持（SPA）
location / {
    try_files $uri $uri/ /index.html;
}

# 等效于
location / {
    if (!-e $request_filename) {
        rewrite ^(.*)$ /index.html last;
    }
}
```

#### 5. API 版本路由

```nginx
location /api/ {
    # /api/users -> /api/v1/users
    rewrite ^/api/(?!v\d+/)(.*)$ /api/v1/$1 last;
}

location /api/v1/ {
    proxy_pass http://api_v1_backend;
}

location /api/v2/ {
    proxy_pass http://api_v2_backend;
}
```

#### 6. 语言路由

```nginx
# 检测 Accept-Language
map $http_accept_language $lang {
    default en;
    ~^zh    zh;
    ~^ja    ja;
    ~^ko    ko;
}

server {
    # 根路径重定向到语言目录
    location = / {
        rewrite ^ /$lang/ redirect;
    }
    
    location ~ ^/(en|zh|ja|ko)/ {
        try_files $uri $uri/ /$1/index.html;
    }
}
```

### return vs rewrite

```nginx
# return - 立即返回响应
return 301 https://example.com$request_uri;  # 简单重定向
return 200 "OK";                             # 返回内容
return 403;                                  # 返回状态码

# rewrite - 用于复杂 URL 改写
rewrite ^/old/(.*)$ /new/$1 permanent;

# 性能: return > rewrite
# 功能: rewrite 更强大（支持正则捕获）

# 推荐用法:
# - 简单重定向用 return
# - 需要正则捕获用 rewrite
```

---

## location 匹配规则

### 语法

```nginx
location [ = | ~ | ~* | ^~ ] uri { ... }
location @name { ... }
```

### 修饰符

| 修饰符 | 说明 | 优先级 |
|--------|------|--------|
| `=` | 精确匹配 | 最高 |
| `^~` | 前缀匹配（不检查正则） | 高 |
| `~` | 正则匹配（区分大小写） | 中 |
| `~*` | 正则匹配（不区分大小写） | 中 |
| 无 | 前缀匹配 | 低 |
| `@` | 命名 location（内部使用） | - |

### 匹配优先级

```nginx
# 优先级从高到低:
# 1. = 精确匹配
# 2. ^~ 前缀匹配（最长匹配，找到后停止搜索正则）
# 3. ~ 或 ~* 正则匹配（按配置顺序，第一个匹配即停止）
# 4. 无修饰符前缀匹配（最长匹配）

server {
    # 精确匹配 - 只匹配 /
    location = / {
        # 处理首页
    }
    
    # 精确匹配 - 只匹配 /favicon.ico
    location = /favicon.ico {
        log_not_found off;
        access_log off;
    }
    
    # 前缀匹配 + 停止正则搜索
    location ^~ /static/ {
        # 匹配 /static/* 且不继续检查正则
    }
    
    # 正则匹配（区分大小写）
    location ~ \.php$ {
        # 匹配 .php 文件
    }
    
    # 正则匹配（不区分大小写）
    location ~* \.(jpg|jpeg|png|gif)$ {
        # 匹配图片文件
    }
    
    # 普通前缀匹配
    location /api/ {
        # 匹配 /api/*
    }
    
    # 默认匹配
    location / {
        # 匹配所有其他请求
    }
}
```

### 匹配示例

```nginx
# 请求: /static/js/app.js

location = /static/js/app.js { }  # ✅ 精确匹配（如果存在，最优先）
location ^~ /static/ { }          # ✅ 前缀匹配（不继续检查正则）
location ~ \.js$ { }              # 不会检查（因为 ^~ 已匹配）
location /static/ { }             # 不会使用（^~ 优先级更高）

# 请求: /images/logo.PNG

location = /images/logo.PNG { }   # ✅ 精确匹配
location ^~ /images/ { }          # ✅ 如果无精确匹配
location ~* \.(png|jpg)$ { }      # ✅ 如果无 ^~ 匹配
location /images/ { }             # 优先级最低
```

### 命名 location

```nginx
server {
    location / {
        try_files $uri $uri/ @backend;
    }
    
    # 命名 location - 只能内部跳转使用
    location @backend {
        proxy_pass http://backend;
    }
    
    # 错误页面处理
    error_page 404 @notfound;
    location @notfound {
        return 404 "Page not found";
    }
    
    error_page 500 502 503 504 @error;
    location @error {
        root /var/www/error;
        try_files /50x.html =500;
    }
}
```

### 嵌套 location

```nginx
location /api/ {
    # /api/* 的通用配置
    proxy_set_header Host $host;
    
    # 嵌套 location
    location /api/v1/ {
        proxy_pass http://api_v1;
    }
    
    location /api/v2/ {
        proxy_pass http://api_v2;
    }
    
    # 正则也可以嵌套
    location ~ ^/api/v\d+/admin {
        auth_basic "Admin Area";
        auth_basic_user_file /etc/nginx/.htpasswd;
        proxy_pass http://admin_api;
    }
}
```

---

## 正则表达式

### PCRE 正则语法

```nginx
# ============ 字符类 ============
.       # 任意单个字符
\d      # 数字 [0-9]
\D      # 非数字
\w      # 单词字符 [a-zA-Z0-9_]
\W      # 非单词字符
\s      # 空白字符
\S      # 非空白字符

# ============ 量词 ============
*       # 0 或多个
+       # 1 或多个
?       # 0 或 1 个
{n}     # 恰好 n 个
{n,}    # n 个或更多
{n,m}   # n 到 m 个

# ============ 锚点 ============
^       # 字符串开始
$       # 字符串结束
\b      # 单词边界

# ============ 分组 ============
(...)   # 捕获组
(?:...) # 非捕获组
(?<name>...) # 命名捕获组

# ============ 字符集 ============
[abc]   # a、b 或 c
[^abc]  # 非 a、b、c
[a-z]   # a 到 z
[A-Za-z0-9]  # 字母和数字

# ============ 转义 ============
\.      # 匹配点号
\/      # 匹配斜杠
\-      # 匹配连字符
```

### Nginx 中的正则使用

```nginx
# ============ location 正则 ============

# 匹配 PHP 文件
location ~ \.php$ {
    fastcgi_pass php:9000;
}

# 匹配图片（不区分大小写）
location ~* \.(jpg|jpeg|png|gif|ico|svg|webp)$ {
    expires 30d;
}

# 匹配 API 版本
location ~ ^/api/v(\d+)/ {
    # $1 = 版本号
    set $api_version $1;
}

# ============ rewrite 正则 ============

# 捕获组使用
rewrite ^/user/(\d+)/profile$ /profile?id=$1 last;
# /user/123/profile -> /profile?id=123

# 多个捕获组
rewrite ^/(\d{4})/(\d{2})/(\d{2})/(.*)$ /archive?year=$1&month=$2&day=$3&slug=$4 last;
# /2024/01/15/hello-world -> /archive?year=2024&month=01&day=15&slug=hello-world

# 命名捕获组
location ~ ^/(?<lang>en|zh|ja)/(?<path>.*)$ {
    # $lang = 语言代码
    # $path = 路径
    try_files /$lang/$path /$lang/index.html =404;
}

# ============ if 正则 ============

# User-Agent 检测
if ($http_user_agent ~* (bot|spider|crawler)) {
    return 403;
}

# 提取子域名
if ($host ~* ^(.+)\.example\.com$) {
    set $subdomain $1;
}

# ============ map 正则 ============

map $uri $cache_time {
    default         "1h";
    ~*\.html$       "10m";
    ~*\.(css|js)$   "1y";
    ~*\.(jpg|png)$  "30d";
}
```

### 正则性能优化

```nginx
# ❌ 低效正则
location ~ ^/(.*)/(.*)/(.*)/(.*)$ { }

# ✅ 更具体的正则
location ~ ^/api/v\d+/users/\d+$ { }

# ❌ 过度使用正则
location ~ ^/static/.*$ { }

# ✅ 使用前缀匹配
location /static/ { }
location ^~ /static/ { }  # 更好，不检查正则

# 正则排序原则:
# 1. 把最常访问的放前面
# 2. 把最具体的放前面
# 3. 优先使用前缀匹配
```

---

## 常用配置模式

### 1. 完整的 Server 配置模板

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name example.com www.example.com;
    
    # 强制 HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name example.com www.example.com;
    
    # SSL 配置
    ssl_certificate /etc/ssl/certs/example.com.crt;
    ssl_certificate_key /etc/ssl/private/example.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:10m;
    
    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000" always;
    
    # 日志
    access_log /var/log/nginx/example.com.access.log;
    error_log /var/log/nginx/example.com.error.log;
    
    # 根目录
    root /var/www/example.com;
    index index.html index.php;
    
    # 静态文件
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff2?)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
    
    # PHP 处理
    location ~ \.php$ {
        fastcgi_pass unix:/var/run/php/php8.1-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    
    # 禁止访问隐藏文件
    location ~ /\. {
        deny all;
    }
    
    # 默认处理
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }
}
```

### 2. 反向代理模板

```nginx
upstream backend {
    least_conn;
    server backend1:8080 weight=5;
    server backend2:8080 weight=3;
    server backend3:8080 backup;
    
    keepalive 32;
}

server {
    listen 80;
    server_name api.example.com;
    
    # 代理配置
    location / {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        
        # 转发真实信息
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Request-ID $request_id;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # 缓冲设置
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;
    }
    
    # WebSocket 支持
    location /ws/ {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400;
    }
    
    # 健康检查
    location /health {
        access_log off;
        return 200 "OK";
    }
}
```

### 3. SPA 应用配置

```nginx
server {
    listen 80;
    server_name app.example.com;
    root /var/www/spa;
    
    # Gzip 压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript;
    
    # 安全头
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";
    
    # 静态资源缓存
    location ~* \.(?:css|js)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    location ~* \.(?:jpg|jpeg|png|gif|ico|svg|webp)$ {
        expires 30d;
        add_header Cache-Control "public";
    }
    
    # API 代理
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    # SPA 路由回退
    location / {
        try_files $uri $uri/ /index.html;
        
        # index.html 不缓存
        location = /index.html {
            add_header Cache-Control "no-cache, no-store, must-revalidate";
        }
    }
}
```

### 4. 负载均衡配置

```nginx
# 轮询（默认）
upstream backend_round_robin {
    server backend1:8080;
    server backend2:8080;
    server backend3:8080;
}

# 加权轮询
upstream backend_weighted {
    server backend1:8080 weight=5;
    server backend2:8080 weight=3;
    server backend3:8080 weight=2;
}

# 最少连接
upstream backend_least_conn {
    least_conn;
    server backend1:8080;
    server backend2:8080;
}

# IP 哈希（会话保持）
upstream backend_ip_hash {
    ip_hash;
    server backend1:8080;
    server backend2:8080;
}

# 一致性哈希
upstream backend_hash {
    hash $request_uri consistent;
    server backend1:8080;
    server backend2:8080;
}

# 服务器状态参数
upstream backend_full {
    server backend1:8080 weight=5 max_fails=3 fail_timeout=30s;
    server backend2:8080 weight=3 max_fails=3 fail_timeout=30s;
    server backend3:8080 backup;  # 备份服务器
    server backend4:8080 down;    # 下线服务器
    
    keepalive 32;                 # 长连接数
}
```

### 5. 限流配置

```nginx
# 定义限流区域
limit_req_zone $binary_remote_addr zone=ip_limit:10m rate=10r/s;
limit_req_zone $server_name zone=server_limit:10m rate=1000r/s;
limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

server {
    listen 80;
    
    # 全局连接限制
    limit_conn conn_limit 20;
    
    location /api/ {
        # 请求速率限制
        limit_req zone=ip_limit burst=20 nodelay;
        
        # 超出限制的处理
        limit_req_status 429;
        
        proxy_pass http://backend;
    }
    
    location /download/ {
        # 限制连接数
        limit_conn conn_limit 1;
        
        # 限制下载速度
        limit_rate 500k;
        limit_rate_after 10m;  # 前 10MB 不限速
    }
}
```

### 6. 缓存配置

```nginx
# 定义缓存区域
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m 
                 max_size=10g inactive=60m use_temp_path=off;

server {
    listen 80;
    
    location /api/ {
        proxy_pass http://backend;
        
        # 启用缓存
        proxy_cache my_cache;
        proxy_cache_key "$scheme$request_method$host$request_uri";
        
        # 缓存时间
        proxy_cache_valid 200 302 10m;
        proxy_cache_valid 404 1m;
        proxy_cache_valid any 5m;
        
        # 缓存条件
        proxy_cache_min_uses 3;
        proxy_cache_use_stale error timeout updating http_500 http_502;
        
        # 缓存绕过
        proxy_cache_bypass $http_cache_control;
        proxy_no_cache $http_pragma;
        
        # 添加缓存状态头
        add_header X-Cache-Status $upstream_cache_status;
    }
}
```

---

## 图片处理与水印

Nginx 可以通过内置模块和第三方模块实现图片处理、水印添加、缩略图生成等功能。

### ngx_http_image_filter_module

Nginx 内置的图片处理模块，支持裁剪、缩放、旋转等基本操作。

```bash
# 检查是否已编译该模块
nginx -V 2>&1 | grep -o 'http_image_filter_module'

# 如果没有，需要重新编译 Nginx
./configure --with-http_image_filter_module
make && make install

# 或使用包管理器安装完整版
# Ubuntu/Debian
apt install nginx-extras

# CentOS/RHEL
yum install nginx-mod-http-image-filter
```

#### 基本图片处理

```nginx
http {
    server {
        listen 80;
        server_name images.example.com;
        
        # ============ 图片缩放 ============
        location /resize/ {
            # 将图片缩放到指定尺寸
            # image_filter resize width height;
            image_filter resize 300 200;
            
            # 图片质量 (1-100)
            image_filter_jpeg_quality 85;
            image_filter_webp_quality 80;
            
            # 处理缓冲区大小
            image_filter_buffer 10M;
            
            # 代理到原始图片
            alias /var/www/images/;
        }
        
        # ============ 等比例缩放 ============
        location /scale/ {
            # 按比例缩放，保持宽高比
            # 只指定宽度，高度自动计算
            image_filter resize 400 -;
            
            # 只指定高度，宽度自动计算
            # image_filter resize - 300;
            
            alias /var/www/images/;
        }
        
        # ============ 裁剪图片 ============
        location /crop/ {
            # 裁剪到指定尺寸（从中心裁剪）
            image_filter crop 200 200;
            image_filter_jpeg_quality 90;
            
            alias /var/www/images/;
        }
        
        # ============ 旋转图片 ============
        location /rotate/ {
            # 旋转角度：90, 180, 270
            image_filter rotate 90;
            
            alias /var/www/images/;
        }
        
        # ============ 获取图片尺寸 ============
        location /size/ {
            # 返回 JSON 格式的图片信息
            image_filter size;
            
            alias /var/www/images/;
        }
        # 返回示例: {"img":{"width":1920,"height":1080,"type":"jpeg"}}
    }
}
```

#### 动态参数处理

```nginx
server {
    listen 80;
    server_name images.example.com;
    
    # ============ URL 参数动态处理 ============
    # 格式: /image/filename.jpg?w=300&h=200&q=85
    location ~ ^/image/(.+)$ {
        set $file $1;
        set $width $arg_w;
        set $height $arg_h;
        set $quality $arg_q;
        
        # 设置默认值
        if ($width = "") {
            set $width "-";
        }
        if ($height = "") {
            set $height "-";
        }
        if ($quality = "") {
            set $quality "85";
        }
        
        # 内部重定向到处理位置
        rewrite ^ /internal/process last;
    }
    
    location /internal/process {
        internal;
        
        image_filter resize $width $height;
        image_filter_jpeg_quality $quality;
        image_filter_buffer 20M;
        
        alias /var/www/images/$file;
    }
    
    # ============ 路径参数处理 ============
    # 格式: /thumb/300x200/filename.jpg
    location ~ ^/thumb/(\d+)x(\d+)/(.+)$ {
        set $w $1;
        set $h $2;
        set $img $3;
        
        image_filter resize $w $h;
        image_filter_jpeg_quality 80;
        
        alias /var/www/images/$img;
    }
    
    # 格式: /thumb/300w/filename.jpg (只指定宽度)
    location ~ ^/thumb/(\d+)w/(.+)$ {
        set $w $1;
        set $img $2;
        
        image_filter resize $w -;
        alias /var/www/images/$img;
    }
}
```

### 使用 Lua 添加水印 (OpenResty)

OpenResty 集成了 LuaJIT，可以实现更复杂的图片处理功能。

```bash
# 安装 OpenResty
# Ubuntu
apt install openresty

# CentOS
yum install openresty

# 安装 lua-resty-imagick (基于 ImageMagick)
opm get toruneko/lua-resty-imagick
```

#### 文字水印

```nginx
http {
    lua_package_path "/usr/local/openresty/lualib/?.lua;;";
    
    server {
        listen 80;
        server_name images.example.com;
        
        # ============ 文字水印 ============
        location ~ ^/watermark/text/(.+)$ {
            set $image_path /var/www/images/$1;
            
            content_by_lua_block {
                local magick = require("imagick")
                local img = magick.open(ngx.var.image_path)
                
                if not img then
                    ngx.status = 404
                    ngx.say("Image not found")
                    return
                end
                
                -- 获取图片尺寸
                local width, height = img:get_width(), img:get_height()
                
                -- 创建水印文字
                local text = ngx.var.arg_text or "© Example.com"
                local font_size = tonumber(ngx.var.arg_size) or 24
                local opacity = tonumber(ngx.var.arg_opacity) or 0.5
                
                -- 添加文字水印
                img:set_font("Arial")
                img:set_font_size(font_size)
                img:set_gravity("SouthEast")  -- 右下角
                img:annotate(text, 10, 10, 0, opacity)
                
                -- 输出图片
                ngx.header["Content-Type"] = "image/jpeg"
                ngx.print(img:get_blob())
                
                img:destroy()
            }
        }
        
        # ============ 带背景的文字水印 ============
        location ~ ^/watermark/label/(.+)$ {
            set $image_path /var/www/images/$1;
            
            content_by_lua_block {
                local magick = require("imagick")
                local img = magick.open(ngx.var.image_path)
                
                if not img then
                    ngx.status = 404
                    return
                end
                
                local width, height = img:get_width(), img:get_height()
                local text = ngx.var.arg_text or "SAMPLE"
                
                -- 创建带背景的标签
                local label_width = #text * 12 + 20
                local label_height = 30
                
                -- 绘制半透明背景
                img:set_fill_color("rgba(0,0,0,0.6)")
                img:rectangle(
                    width - label_width - 10,
                    height - label_height - 10,
                    width - 10,
                    height - 10
                )
                
                -- 绘制文字
                img:set_fill_color("white")
                img:set_font_size(16)
                img:set_gravity("SouthEast")
                img:annotate(text, 15, 15)
                
                ngx.header["Content-Type"] = "image/jpeg"
                ngx.print(img:get_blob())
                img:destroy()
            }
        }
    }
}
```

#### 图片水印

```nginx
server {
    listen 80;
    server_name images.example.com;
    
    # ============ 图片水印 ============
    location ~ ^/watermark/image/(.+)$ {
        set $image_path /var/www/images/$1;
        set $watermark_path /var/www/watermarks/logo.png;
        
        content_by_lua_block {
            local magick = require("imagick")
            
            -- 打开原图
            local img = magick.open(ngx.var.image_path)
            if not img then
                ngx.status = 404
                ngx.say("Image not found")
                return
            end
            
            -- 打开水印图片
            local watermark = magick.open(ngx.var.watermark_path)
            if not watermark then
                ngx.header["Content-Type"] = "image/jpeg"
                ngx.print(img:get_blob())
                img:destroy()
                return
            end
            
            -- 获取尺寸
            local img_w, img_h = img:get_width(), img:get_height()
            local wm_w, wm_h = watermark:get_width(), watermark:get_height()
            
            -- 根据参数确定位置
            local position = ngx.var.arg_pos or "southeast"
            local margin = tonumber(ngx.var.arg_margin) or 10
            local opacity = tonumber(ngx.var.arg_opacity) or 0.8
            
            -- 计算水印位置
            local x, y = 0, 0
            if position == "southeast" or position == "se" then
                x = img_w - wm_w - margin
                y = img_h - wm_h - margin
            elseif position == "southwest" or position == "sw" then
                x = margin
                y = img_h - wm_h - margin
            elseif position == "northeast" or position == "ne" then
                x = img_w - wm_w - margin
                y = margin
            elseif position == "northwest" or position == "nw" then
                x = margin
                y = margin
            elseif position == "center" then
                x = (img_w - wm_w) / 2
                y = (img_h - wm_h) / 2
            end
            
            -- 设置水印透明度
            watermark:set_opacity(opacity)
            
            -- 合成图片
            img:composite(watermark, x, y, "Over")
            
            -- 输出
            ngx.header["Content-Type"] = "image/jpeg"
            ngx.print(img:get_blob())
            
            img:destroy()
            watermark:destroy()
        }
    }
    
    # ============ 平铺水印 ============
    location ~ ^/watermark/tile/(.+)$ {
        set $image_path /var/www/images/$1;
        set $watermark_path /var/www/watermarks/tile-logo.png;
        
        content_by_lua_block {
            local magick = require("imagick")
            
            local img = magick.open(ngx.var.image_path)
            local watermark = magick.open(ngx.var.watermark_path)
            
            if not img or not watermark then
                ngx.status = 404
                return
            end
            
            local img_w, img_h = img:get_width(), img:get_height()
            local wm_w, wm_h = watermark:get_width(), watermark:get_height()
            
            -- 设置水印透明度
            watermark:set_opacity(0.3)
            
            -- 平铺水印
            local spacing = 50  -- 水印间距
            for x = 0, img_w, wm_w + spacing do
                for y = 0, img_h, wm_h + spacing do
                    img:composite(watermark, x, y, "Over")
                end
            end
            
            ngx.header["Content-Type"] = "image/jpeg"
            ngx.print(img:get_blob())
            
            img:destroy()
            watermark:destroy()
        }
    }
}
```

### 使用外部程序处理 (ImageMagick)

通过 Nginx 调用 ImageMagick 命令行工具处理图片。

```nginx
server {
    listen 80;
    server_name images.example.com;
    
    # ============ 使用 proxy_pass 调用处理服务 ============
    location ~ ^/process/(.+)$ {
        # 转发到图片处理服务
        proxy_pass http://127.0.0.1:8080/process/$1$is_args$args;
        proxy_cache image_cache;
        proxy_cache_valid 200 7d;
    }
}

# 图片处理服务 (可以是 Python/Node.js/Go 等)
# 示例 Python Flask 服务:
```

```python
# image_processor.py
from flask import Flask, send_file, request
from PIL import Image, ImageDraw, ImageFont
import io
import os

app = Flask(__name__)
IMAGE_DIR = "/var/www/images"
WATERMARK_PATH = "/var/www/watermarks/logo.png"

@app.route("/process/<path:filename>")
def process_image(filename):
    filepath = os.path.join(IMAGE_DIR, filename)
    if not os.path.exists(filepath):
        return "Not found", 404
    
    # 打开图片
    img = Image.open(filepath)
    
    # 获取参数
    width = request.args.get("w", type=int)
    height = request.args.get("h", type=int)
    watermark = request.args.get("watermark", "")
    text = request.args.get("text", "")
    
    # 缩放
    if width or height:
        if width and height:
            img = img.resize((width, height), Image.LANCZOS)
        elif width:
            ratio = width / img.width
            img = img.resize((width, int(img.height * ratio)), Image.LANCZOS)
        elif height:
            ratio = height / img.height
            img = img.resize((int(img.width * ratio), height), Image.LANCZOS)
    
    # 图片水印
    if watermark and os.path.exists(WATERMARK_PATH):
        wm = Image.open(WATERMARK_PATH).convert("RGBA")
        # 调整水印透明度
        wm.putalpha(int(255 * 0.5))
        # 计算位置（右下角）
        pos = (img.width - wm.width - 10, img.height - wm.height - 10)
        img.paste(wm, pos, wm)
    
    # 文字水印
    if text:
        draw = ImageDraw.Draw(img)
        font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 24)
        # 计算文字位置
        bbox = draw.textbbox((0, 0), text, font=font)
        text_width = bbox[2] - bbox[0]
        text_height = bbox[3] - bbox[1]
        pos = (img.width - text_width - 10, img.height - text_height - 10)
        # 绘制阴影
        draw.text((pos[0]+1, pos[1]+1), text, font=font, fill=(0, 0, 0, 128))
        # 绘制文字
        draw.text(pos, text, font=font, fill=(255, 255, 255, 200))
    
    # 输出
    output = io.BytesIO()
    img.save(output, format="JPEG", quality=85)
    output.seek(0)
    
    return send_file(output, mimetype="image/jpeg")

if __name__ == "__main__":
    app.run(host="127.0.0.1", port=8080)
```

### 缓存配置

图片处理是 CPU 密集型操作，必须配置缓存。

```nginx
http {
    # ============ 代理缓存配置 ============
    proxy_cache_path /var/cache/nginx/images
        levels=1:2
        keys_zone=image_cache:100m
        max_size=10g
        inactive=7d
        use_temp_path=off;
    
    server {
        listen 80;
        server_name images.example.com;
        
        # 缓存处理后的图片
        location ~ ^/thumb/ {
            # 启用缓存
            proxy_cache image_cache;
            proxy_cache_key "$uri$is_args$args";
            proxy_cache_valid 200 7d;
            proxy_cache_valid 404 1m;
            
            # 添加缓存状态头
            add_header X-Cache-Status $upstream_cache_status;
            
            # 处理图片
            image_filter resize 300 200;
            image_filter_buffer 10M;
            
            alias /var/www/images/;
        }
        
        # ============ 浏览器缓存 ============
        location ~* \.(jpg|jpeg|png|gif|webp)$ {
            expires 30d;
            add_header Cache-Control "public, immutable";
            add_header Vary "Accept";
            
            root /var/www/images;
        }
    }
}
```

### 完整配置示例

```nginx
# /etc/nginx/conf.d/image-server.conf

# 缓存配置
proxy_cache_path /var/cache/nginx/images
    levels=1:2
    keys_zone=img_cache:100m
    max_size=20g
    inactive=30d
    use_temp_path=off;

# 限流配置
limit_req_zone $binary_remote_addr zone=img_limit:10m rate=10r/s;

server {
    listen 80;
    server_name images.example.com;
    
    root /var/www/images;
    
    # 访问日志
    access_log /var/log/nginx/images.access.log;
    error_log /var/log/nginx/images.error.log;
    
    # 客户端限制
    client_max_body_size 50M;
    
    # ============ 原图访问 ============
    location /original/ {
        alias /var/www/images/;
        
        # 防盗链
        valid_referers none blocked server_names *.example.com;
        if ($invalid_referer) {
            return 403;
        }
        
        # 浏览器缓存
        expires 30d;
        add_header Cache-Control "public";
    }
    
    # ============ 缩略图服务 ============
    # /thumb/300x200/path/to/image.jpg
    location ~ ^/thumb/(\d+)x(\d+)/(.+)$ {
        set $w $1;
        set $h $2;
        set $img $3;
        
        # 限流
        limit_req zone=img_limit burst=20 nodelay;
        
        # 缓存
        proxy_cache img_cache;
        proxy_cache_key "thumb_${w}x${h}_$img";
        proxy_cache_valid 200 30d;
        add_header X-Cache $upstream_cache_status;
        
        # 限制最大尺寸
        if ($w > 2000) {
            return 400;
        }
        if ($h > 2000) {
            return 400;
        }
        
        # 图片处理
        image_filter resize $w $h;
        image_filter_jpeg_quality 85;
        image_filter_buffer 20M;
        image_filter_interlace on;
        
        # 错误处理
        error_page 415 = /error/unsupported.jpg;
        
        alias /var/www/images/$img;
    }
    
    # ============ 裁剪服务 ============
    # /crop/200x200/path/to/image.jpg
    location ~ ^/crop/(\d+)x(\d+)/(.+)$ {
        set $w $1;
        set $h $2;
        set $img $3;
        
        proxy_cache img_cache;
        proxy_cache_key "crop_${w}x${h}_$img";
        proxy_cache_valid 200 30d;
        
        image_filter crop $w $h;
        image_filter_jpeg_quality 90;
        image_filter_buffer 20M;
        
        alias /var/www/images/$img;
    }
    
    # ============ 动态水印 ============
    # /watermark/path/to/image.jpg?text=Copyright
    location ~ ^/watermark/(.+)$ {
        set $img $1;
        
        # 限流（水印处理更耗资源）
        limit_req zone=img_limit burst=5 nodelay;
        
        proxy_cache img_cache;
        proxy_cache_key "wm_$img_$arg_text";
        proxy_cache_valid 200 7d;
        
        # 转发到 Lua 处理或外部服务
        proxy_pass http://127.0.0.1:8080/watermark/$img$is_args$args;
    }
    
    # ============ WebP 自动转换 ============
    location ~ ^/webp/(.+)\.(jpg|jpeg|png)$ {
        set $img $1.$2;
        
        # 检查浏览器是否支持 WebP
        if ($http_accept ~* "webp") {
            # 尝试返回 WebP 版本
            rewrite ^ /webp-internal/$img last;
        }
        
        # 返回原图
        alias /var/www/images/$img;
    }
    
    location /webp-internal/ {
        internal;
        
        # 检查 WebP 文件是否存在
        try_files /webp/$uri.webp /original/$uri =404;
        
        add_header Vary Accept;
        expires 30d;
    }
    
    # ============ 错误页面 ============
    location /error/ {
        internal;
        alias /var/www/images/errors/;
    }
    
    # ============ 健康检查 ============
    location /health {
        return 200 "OK";
        add_header Content-Type text/plain;
    }
}
```

### 性能优化建议

```nginx
# 1. 启用 sendfile
sendfile on;
tcp_nopush on;
tcp_nodelay on;

# 2. 调整缓冲区
image_filter_buffer 20M;          # 图片处理缓冲区
client_body_buffer_size 10M;       # 客户端请求体缓冲区

# 3. 限制处理尺寸
# 在应用层限制最大处理尺寸，防止资源耗尽
if ($w > 2000) { return 400; }

# 4. 使用缓存
proxy_cache_path ... max_size=20g inactive=30d;

# 5. 限流保护
limit_req_zone $binary_remote_addr zone=img_limit:10m rate=10r/s;

# 6. 预生成缩略图
# 对于热门图片，可以预先生成缩略图存储

# 7. CDN 加速
# 将处理后的图片推送到 CDN

# 8. 异步处理
# 对于复杂处理，使用消息队列异步处理
```

### 常用图片处理 URL 格式

```
# 缩放
/thumb/300x200/image.jpg          # 指定宽高
/thumb/300w/image.jpg             # 只指定宽度
/thumb/h200/image.jpg             # 只指定高度

# 裁剪
/crop/200x200/image.jpg           # 中心裁剪
/crop/200x200/nw/image.jpg        # 左上角裁剪

# 水印
/watermark/image.jpg?text=©2024   # 文字水印
/watermark/image.jpg?logo=1       # 图片水印
/watermark/image.jpg?pos=se       # 指定位置

# 质量
/thumb/300x200/image.jpg?q=85     # 指定质量

# 格式转换
/convert/webp/image.jpg           # 转换为 WebP
/convert/png/image.jpg            # 转换为 PNG

# 组合操作
/thumb/300x200/image.jpg?watermark=1&q=90
```

---

## 调试与测试

### 配置测试命令

```bash
# 测试配置语法
nginx -t

# 测试并显示完整配置
nginx -T

# 重载配置
nginx -s reload

# 查看编译参数
nginx -V

# 查看帮助
nginx -h
```

### 调试技巧

```nginx
# 1. 打印变量到响应头
add_header X-Debug-URI $uri;
add_header X-Debug-Args $args;
add_header X-Debug-Host $host;

# 2. 打印变量到日志
log_format debug '$remote_addr - $request - uri=$uri args=$args';
access_log /var/log/nginx/debug.log debug;

# 3. 调试 rewrite
rewrite_log on;
error_log /var/log/nginx/rewrite.log notice;

# 4. 返回调试信息
location /debug {
    default_type text/plain;
    return 200 "uri: $uri\nargs: $args\nhost: $host\n";
}
```

---

## 参考资源

- [Nginx 官方文档](https://nginx.org/en/docs/)
- [Nginx 指令索引](https://nginx.org/en/docs/dirindex.html)
- [Nginx 变量索引](https://nginx.org/en/docs/varindex.html)
- [Nginx 资源 - GitHub](https://github.com/fcambus/nginx-resources)

---

## 总结

| 主题 | 关键点 |
|------|--------|
| if | 仅用于 return、rewrite、set |
| set | 用于变量赋值和条件组合 |
| map | 大量条件映射的最佳选择 |
| rewrite | 复杂 URL 改写 |
| location | 注意匹配优先级 |
| 正则 | 尽量使用前缀匹配提高性能 |

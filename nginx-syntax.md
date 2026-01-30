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

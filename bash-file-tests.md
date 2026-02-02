# Bash 文件测试详解

本文档详细介绍 Bash 中文件存在性检查及各种文件测试操作。

## 目录

1. [文件存在性测试](#文件存在性测试)
2. [文件类型测试](#文件类型测试)
3. [文件权限测试](#文件权限测试)
4. [文件比较测试](#文件比较测试)
5. [字符串与数值测试](#字符串与数值测试)
6. [实用脚本示例](#实用脚本示例)

---

## 文件存在性测试

### 基本语法

```bash
# 方法1: test 命令
test -e /path/to/file

# 方法2: [ ] 语法（等同于 test）
[ -e /path/to/file ]

# 方法3: [[ ]] 语法（Bash 扩展，推荐）
[[ -e /path/to/file ]]
```

### 文件存在检查

```bash
#!/bin/bash

FILE="/path/to/file"

# ============ 检查文件是否存在 ============

# -e: 文件存在（任何类型）
if [[ -e "$FILE" ]]; then
    echo "文件存在"
fi

# -f: 存在且是普通文件
if [[ -f "$FILE" ]]; then
    echo "是普通文件"
fi

# -d: 存在且是目录
if [[ -d "$FILE" ]]; then
    echo "是目录"
fi

# ============ 文件不存在 ============

if [[ ! -e "$FILE" ]]; then
    echo "文件不存在"
fi

# ============ 存在时执行操作 ============

# 单行写法
[[ -f "$FILE" ]] && echo "文件存在"

# 不存在时执行
[[ -f "$FILE" ]] || echo "文件不存在"

# 存在/不存在分别处理
[[ -f "$FILE" ]] && echo "存在" || echo "不存在"
```

### 多文件检查

```bash
#!/bin/bash

# 检查多个文件
FILES=("/etc/passwd" "/etc/shadow" "/etc/hosts")

for file in "${FILES[@]}"; do
    if [[ -f "$file" ]]; then
        echo "✓ $file 存在"
    else
        echo "✗ $file 不存在"
    fi
done

# 检查所有文件都存在
all_exist=true
for file in "${FILES[@]}"; do
    if [[ ! -f "$file" ]]; then
        all_exist=false
        break
    fi
done

if $all_exist; then
    echo "所有文件都存在"
fi

# 检查至少一个文件存在
any_exist=false
for file in "${FILES[@]}"; do
    if [[ -f "$file" ]]; then
        any_exist=true
        break
    fi
done
```

---

## 文件类型测试

### 完整的文件类型测试表

| 操作符 | 说明 | 示例 |
|--------|------|------|
| `-e` | 文件存在 | `[[ -e file ]]` |
| `-f` | 普通文件 | `[[ -f file ]]` |
| `-d` | 目录 | `[[ -d dir ]]` |
| `-L` / `-h` | 符号链接 | `[[ -L link ]]` |
| `-b` | 块设备文件 | `[[ -b /dev/sda ]]` |
| `-c` | 字符设备文件 | `[[ -c /dev/tty ]]` |
| `-p` | 命名管道 (FIFO) | `[[ -p pipe ]]` |
| `-S` | Socket 文件 | `[[ -S socket ]]` |

### 使用示例

```bash
#!/bin/bash

check_file_type() {
    local path="$1"
    
    if [[ ! -e "$path" ]]; then
        echo "$path: 不存在"
        return 1
    fi
    
    if [[ -f "$path" ]]; then
        echo "$path: 普通文件"
    elif [[ -d "$path" ]]; then
        echo "$path: 目录"
    elif [[ -L "$path" ]]; then
        echo "$path: 符号链接 -> $(readlink "$path")"
    elif [[ -b "$path" ]]; then
        echo "$path: 块设备"
    elif [[ -c "$path" ]]; then
        echo "$path: 字符设备"
    elif [[ -p "$path" ]]; then
        echo "$path: 命名管道"
    elif [[ -S "$path" ]]; then
        echo "$path: Socket"
    else
        echo "$path: 未知类型"
    fi
}

# 测试
check_file_type "/etc/passwd"
check_file_type "/tmp"
check_file_type "/dev/null"
check_file_type "/dev/sda"
```

### 符号链接特殊处理

```bash
#!/bin/bash

LINK="/path/to/symlink"

# -L 检查是否为符号链接
if [[ -L "$LINK" ]]; then
    echo "是符号链接"
    
    # 获取链接目标
    TARGET=$(readlink "$LINK")
    echo "指向: $TARGET"
    
    # 检查目标是否存在
    if [[ -e "$LINK" ]]; then
        echo "链接目标存在"
    else
        echo "链接目标不存在（断链）"
    fi
fi

# 注意: -f 会跟随符号链接
# 如果链接指向普通文件，-f 返回 true
if [[ -f "$LINK" ]]; then
    echo "链接指向普通文件"
fi
```

---

## 文件权限测试

### 权限测试表

| 操作符 | 说明 | 示例 |
|--------|------|------|
| `-r` | 可读 | `[[ -r file ]]` |
| `-w` | 可写 | `[[ -w file ]]` |
| `-x` | 可执行 | `[[ -x file ]]` |
| `-u` | 设置了 SUID | `[[ -u file ]]` |
| `-g` | 设置了 SGID | `[[ -g file ]]` |
| `-k` | 设置了 Sticky Bit | `[[ -k dir ]]` |
| `-O` | 当前用户是所有者 | `[[ -O file ]]` |
| `-G` | 当前用户的组是所有组 | `[[ -G file ]]` |

### 使用示例

```bash
#!/bin/bash

FILE="/path/to/file"

# ============ 基本权限检查 ============

if [[ -r "$FILE" ]]; then
    echo "可读"
fi

if [[ -w "$FILE" ]]; then
    echo "可写"
fi

if [[ -x "$FILE" ]]; then
    echo "可执行"
fi

# ============ 组合检查 ============

# 可读且可写
if [[ -r "$FILE" && -w "$FILE" ]]; then
    echo "可读写"
fi

# 可读或可写
if [[ -r "$FILE" || -w "$FILE" ]]; then
    echo "至少可读或可写"
fi

# ============ 权限检查函数 ============

check_permissions() {
    local file="$1"
    local perms=""
    
    [[ -r "$file" ]] && perms+="r"
    [[ -w "$file" ]] && perms+="w"
    [[ -x "$file" ]] && perms+="x"
    
    echo "$file: $perms"
}

check_permissions "/etc/passwd"
check_permissions "/etc/shadow"
check_permissions "/bin/bash"
```

### 特殊权限检查

```bash
#!/bin/bash

# SUID - 以文件所有者权限执行
if [[ -u "/usr/bin/passwd" ]]; then
    echo "passwd 设置了 SUID"
fi

# SGID - 以文件所属组权限执行
if [[ -g "/usr/bin/wall" ]]; then
    echo "wall 设置了 SGID"
fi

# Sticky Bit - 目录中只有所有者能删除文件
if [[ -k "/tmp" ]]; then
    echo "/tmp 设置了 Sticky Bit"
fi

# 查找 SUID 文件
find /usr/bin -perm -4000 2>/dev/null
```

---

## 文件比较测试

### 文件属性测试

| 操作符 | 说明 | 示例 |
|--------|------|------|
| `-s` | 文件存在且大小 > 0 | `[[ -s file ]]` |
| `-t fd` | fd 是终端 | `[[ -t 0 ]]` |
| `-N` | 文件在上次读取后被修改 | `[[ -N file ]]` |

### 文件比较操作符

| 操作符 | 说明 | 示例 |
|--------|------|------|
| `-nt` | 文件1比文件2新 (newer than) | `[[ f1 -nt f2 ]]` |
| `-ot` | 文件1比文件2旧 (older than) | `[[ f1 -ot f2 ]]` |
| `-ef` | 同一文件（硬链接） | `[[ f1 -ef f2 ]]` |

### 使用示例

```bash
#!/bin/bash

# ============ 文件大小检查 ============

FILE="/var/log/syslog"

# 文件非空
if [[ -s "$FILE" ]]; then
    echo "文件有内容"
else
    echo "文件为空或不存在"
fi

# 获取文件大小
SIZE=$(stat -c%s "$FILE" 2>/dev/null)
echo "文件大小: $SIZE 字节"

# ============ 文件时间比较 ============

FILE1="file1.txt"
FILE2="file2.txt"

# 比较修改时间
if [[ "$FILE1" -nt "$FILE2" ]]; then
    echo "$FILE1 比 $FILE2 新"
elif [[ "$FILE1" -ot "$FILE2" ]]; then
    echo "$FILE1 比 $FILE2 旧"
else
    echo "两个文件同样新"
fi

# ============ 硬链接检查 ============

# 创建硬链接测试
touch original.txt
ln original.txt hardlink.txt

if [[ "original.txt" -ef "hardlink.txt" ]]; then
    echo "是同一文件（硬链接）"
fi

# ============ 终端检查 ============

# 检查标准输入是否是终端
if [[ -t 0 ]]; then
    echo "在终端中运行"
else
    echo "从管道或文件输入"
fi

# 检查标准输出是否是终端
if [[ -t 1 ]]; then
    echo "输出到终端"
else
    echo "输出被重定向"
fi
```

---

## 字符串与数值测试

### 字符串测试

| 操作符 | 说明 | 示例 |
|--------|------|------|
| `-z` | 字符串长度为 0 | `[[ -z "$str" ]]` |
| `-n` | 字符串长度不为 0 | `[[ -n "$str" ]]` |
| `=` / `==` | 字符串相等 | `[[ "$a" == "$b" ]]` |
| `!=` | 字符串不相等 | `[[ "$a" != "$b" ]]` |
| `<` | 字典序小于 | `[[ "$a" < "$b" ]]` |
| `>` | 字典序大于 | `[[ "$a" > "$b" ]]` |
| `=~` | 正则匹配 | `[[ "$str" =~ ^[0-9]+$ ]]` |

```bash
#!/bin/bash

STR="hello"

# 空字符串检查
if [[ -z "$STR" ]]; then
    echo "字符串为空"
fi

if [[ -n "$STR" ]]; then
    echo "字符串非空"
fi

# 字符串比较
if [[ "$STR" == "hello" ]]; then
    echo "字符串相等"
fi

# 正则匹配
EMAIL="user@example.com"
if [[ "$EMAIL" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    echo "有效的邮箱格式"
fi

# 通配符匹配（仅 [[ ]] 支持）
if [[ "$STR" == h* ]]; then
    echo "以 h 开头"
fi
```

### 数值测试

| 操作符 | 说明 | 示例 |
|--------|------|------|
| `-eq` | 等于 | `[[ $a -eq $b ]]` |
| `-ne` | 不等于 | `[[ $a -ne $b ]]` |
| `-lt` | 小于 | `[[ $a -lt $b ]]` |
| `-le` | 小于等于 | `[[ $a -le $b ]]` |
| `-gt` | 大于 | `[[ $a -gt $b ]]` |
| `-ge` | 大于等于 | `[[ $a -ge $b ]]` |

```bash
#!/bin/bash

A=10
B=20

# 数值比较
if [[ $A -lt $B ]]; then
    echo "$A 小于 $B"
fi

# 使用 (( )) 进行算术比较（更直观）
if (( A < B )); then
    echo "$A 小于 $B"
fi

if (( A >= 5 && A <= 15 )); then
    echo "$A 在 5-15 范围内"
fi
```

---

## 实用脚本示例

### 1. 安全删除脚本

```bash
#!/bin/bash

safe_delete() {
    local file="$1"
    
    # 检查参数
    if [[ -z "$file" ]]; then
        echo "错误: 请指定文件" >&2
        return 1
    fi
    
    # 检查文件是否存在
    if [[ ! -e "$file" ]]; then
        echo "错误: 文件不存在: $file" >&2
        return 1
    fi
    
    # 检查是否是目录
    if [[ -d "$file" ]]; then
        echo "警告: $file 是目录，使用 -r 选项删除"
        read -p "确认删除目录? [y/N] " confirm
        if [[ "$confirm" =~ ^[Yy]$ ]]; then
            rm -rf "$file"
            echo "目录已删除"
        fi
    else
        rm "$file"
        echo "文件已删除: $file"
    fi
}

safe_delete "$1"
```

### 2. 配置文件检查脚本

```bash
#!/bin/bash

CONFIG_DIR="/etc/myapp"
REQUIRED_FILES=(
    "config.yaml"
    "database.yaml"
    "secrets.env"
)

check_config() {
    local missing=()
    local readonly_files=()
    
    # 检查目录
    if [[ ! -d "$CONFIG_DIR" ]]; then
        echo "错误: 配置目录不存在: $CONFIG_DIR"
        exit 1
    fi
    
    # 检查必需文件
    for file in "${REQUIRED_FILES[@]}"; do
        local path="$CONFIG_DIR/$file"
        
        if [[ ! -f "$path" ]]; then
            missing+=("$file")
        elif [[ ! -r "$path" ]]; then
            readonly_files+=("$file")
        fi
    done
    
    # 报告结果
    if [[ ${#missing[@]} -gt 0 ]]; then
        echo "缺少配置文件:"
        printf '  - %s\n' "${missing[@]}"
    fi
    
    if [[ ${#readonly_files[@]} -gt 0 ]]; then
        echo "无法读取的文件:"
        printf '  - %s\n' "${readonly_files[@]}"
    fi
    
    if [[ ${#missing[@]} -eq 0 && ${#readonly_files[@]} -eq 0 ]]; then
        echo "✓ 所有配置文件检查通过"
        return 0
    else
        return 1
    fi
}

check_config
```

### 3. 日志轮转脚本

```bash
#!/bin/bash

LOG_FILE="/var/log/myapp/app.log"
MAX_SIZE=$((10 * 1024 * 1024))  # 10MB
MAX_BACKUPS=5

rotate_log() {
    # 检查日志文件是否存在
    if [[ ! -f "$LOG_FILE" ]]; then
        echo "日志文件不存在"
        return 0
    fi
    
    # 检查文件大小
    local size=$(stat -c%s "$LOG_FILE" 2>/dev/null || echo 0)
    
    if [[ $size -lt $MAX_SIZE ]]; then
        echo "日志大小 ($size bytes) 未超过阈值"
        return 0
    fi
    
    echo "开始轮转日志..."
    
    # 删除最旧的备份
    if [[ -f "${LOG_FILE}.${MAX_BACKUPS}" ]]; then
        rm "${LOG_FILE}.${MAX_BACKUPS}"
    fi
    
    # 重命名现有备份
    for ((i=MAX_BACKUPS-1; i>=1; i--)); do
        if [[ -f "${LOG_FILE}.$i" ]]; then
            mv "${LOG_FILE}.$i" "${LOG_FILE}.$((i+1))"
        fi
    done
    
    # 轮转当前日志
    mv "$LOG_FILE" "${LOG_FILE}.1"
    
    # 创建新日志文件
    touch "$LOG_FILE"
    
    echo "日志轮转完成"
}

rotate_log
```

### 4. 目录同步脚本

```bash
#!/bin/bash

sync_directories() {
    local src="$1"
    local dest="$2"
    
    # 验证源目录
    if [[ ! -d "$src" ]]; then
        echo "错误: 源目录不存在: $src" >&2
        return 1
    fi
    
    if [[ ! -r "$src" ]]; then
        echo "错误: 源目录不可读: $src" >&2
        return 1
    fi
    
    # 验证/创建目标目录
    if [[ ! -d "$dest" ]]; then
        echo "创建目标目录: $dest"
        mkdir -p "$dest" || {
            echo "错误: 无法创建目标目录" >&2
            return 1
        }
    fi
    
    if [[ ! -w "$dest" ]]; then
        echo "错误: 目标目录不可写: $dest" >&2
        return 1
    fi
    
    # 同步文件
    echo "同步: $src -> $dest"
    rsync -av --delete "$src/" "$dest/"
    
    echo "同步完成"
}

sync_directories "$1" "$2"
```

### 5. 服务依赖检查脚本

```bash
#!/bin/bash

# 检查所需文件和目录
check_dependencies() {
    local errors=0
    
    # 必需的可执行文件
    local required_commands=(
        "docker"
        "docker-compose"
        "curl"
        "jq"
    )
    
    # 必需的配置文件
    local required_files=(
        "./docker-compose.yml"
        "./.env"
    )
    
    # 必需的目录
    local required_dirs=(
        "./data"
        "./logs"
        "./config"
    )
    
    echo "=== 检查命令 ==="
    for cmd in "${required_commands[@]}"; do
        if command -v "$cmd" &>/dev/null; then
            echo "✓ $cmd: $(command -v "$cmd")"
        else
            echo "✗ $cmd: 未找到"
            ((errors++))
        fi
    done
    
    echo ""
    echo "=== 检查文件 ==="
    for file in "${required_files[@]}"; do
        if [[ -f "$file" ]]; then
            if [[ -r "$file" ]]; then
                echo "✓ $file"
            else
                echo "✗ $file: 不可读"
                ((errors++))
            fi
        else
            echo "✗ $file: 不存在"
            ((errors++))
        fi
    done
    
    echo ""
    echo "=== 检查目录 ==="
    for dir in "${required_dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            if [[ -w "$dir" ]]; then
                echo "✓ $dir"
            else
                echo "✗ $dir: 不可写"
                ((errors++))
            fi
        else
            echo "- $dir: 不存在，正在创建..."
            mkdir -p "$dir" && echo "  ✓ 创建成功" || {
                echo "  ✗ 创建失败"
                ((errors++))
            }
        fi
    done
    
    echo ""
    if [[ $errors -eq 0 ]]; then
        echo "✓ 所有依赖检查通过"
        return 0
    else
        echo "✗ 发现 $errors 个问题"
        return 1
    fi
}

check_dependencies
```

### 6. 等待文件出现

```bash
#!/bin/bash

wait_for_file() {
    local file="$1"
    local timeout="${2:-60}"  # 默认60秒
    local interval="${3:-1}"  # 默认1秒检查一次
    
    local elapsed=0
    
    echo "等待文件: $file (超时: ${timeout}秒)"
    
    while [[ ! -f "$file" ]]; do
        if [[ $elapsed -ge $timeout ]]; then
            echo "超时: 文件未出现"
            return 1
        fi
        
        sleep "$interval"
        ((elapsed += interval))
        echo -ne "\r等待中... ${elapsed}秒"
    done
    
    echo -e "\n文件已出现: $file"
    return 0
}

# 使用示例
wait_for_file "/tmp/ready.flag" 30 2
```

### 7. 创建文件（如果不存在）

```bash
#!/bin/bash

# 安全创建文件
create_if_not_exists() {
    local file="$1"
    local content="${2:-}"
    
    if [[ -e "$file" ]]; then
        echo "文件已存在: $file"
        return 0
    fi
    
    # 确保父目录存在
    local dir=$(dirname "$file")
    if [[ ! -d "$dir" ]]; then
        mkdir -p "$dir" || {
            echo "错误: 无法创建目录: $dir" >&2
            return 1
        }
    fi
    
    # 创建文件
    if [[ -n "$content" ]]; then
        echo "$content" > "$file"
    else
        touch "$file"
    fi
    
    echo "文件已创建: $file"
}

# 使用示例
create_if_not_exists "/tmp/myapp/config.json" '{"debug": true}'
```

---

## 快速参考表

### 文件测试操作符

| 操作符 | 描述 |
|--------|------|
| `-e file` | 文件存在 |
| `-f file` | 普通文件 |
| `-d file` | 目录 |
| `-L file` | 符号链接 |
| `-b file` | 块设备 |
| `-c file` | 字符设备 |
| `-p file` | 命名管道 |
| `-S file` | Socket |
| `-r file` | 可读 |
| `-w file` | 可写 |
| `-x file` | 可执行 |
| `-s file` | 文件大小 > 0 |
| `-u file` | SUID 位已设置 |
| `-g file` | SGID 位已设置 |
| `-k file` | Sticky 位已设置 |
| `-O file` | 当前用户是所有者 |
| `-G file` | 当前用户组是所属组 |
| `-N file` | 自上次读取后被修改 |
| `-t fd` | 文件描述符是终端 |
| `f1 -nt f2` | f1 比 f2 新 |
| `f1 -ot f2` | f1 比 f2 旧 |
| `f1 -ef f2` | 同一文件（硬链接） |

### 常用模式

```bash
# 文件存在
[[ -f "$file" ]] && echo "存在"

# 文件不存在
[[ ! -f "$file" ]] && echo "不存在"

# 存在时执行，不存在时执行其他
[[ -f "$file" ]] && process_file || create_file

# 目录不存在则创建
[[ -d "$dir" ]] || mkdir -p "$dir"

# 文件可读且非空
[[ -r "$file" && -s "$file" ]] && cat "$file"

# 检查命令是否存在
command -v docker &>/dev/null && echo "Docker 已安装"
type -P docker &>/dev/null && echo "Docker 已安装"
which docker &>/dev/null && echo "Docker 已安装"
```

---

## 注意事项

### [ ] vs [[ ]]

```bash
# [ ] - POSIX 兼容，但功能有限
[ -f "$file" ]

# [[ ]] - Bash 扩展，更强大
# 优势:
# 1. 不需要转义特殊字符
# 2. 支持 && 和 ||
# 3. 支持 =~ 正则匹配
# 4. 支持通配符匹配

# 示例区别
file="my file.txt"

# [ ] 必须引用变量
[ -f "$file" ]

# [[ ]] 可以不引用（但建议引用）
[[ -f $file ]]

# 正则匹配只有 [[ ]] 支持
[[ "$str" =~ ^[0-9]+$ ]]
```

### 引号的重要性

```bash
FILE=""

# 错误: 会报语法错误
[ -f $FILE ]  # 展开为 [ -f ]

# 正确: 使用引号
[ -f "$FILE" ]  # 展开为 [ -f "" ]

# [[ ]] 更宽容但仍建议使用引号
[[ -f $FILE ]]   # 可以工作
[[ -f "$FILE" ]] # 更安全
```

### 符号链接行为

```bash
# -f 会跟随符号链接
ln -s /etc/passwd mylink
[[ -f mylink ]]  # true（目标是文件）

# -L 检查是否是符号链接本身
[[ -L mylink ]]  # true

# -e 检查链接目标是否存在
ln -s /nonexistent broken_link
[[ -L broken_link ]]  # true（是链接）
[[ -e broken_link ]]  # false（目标不存在）
```

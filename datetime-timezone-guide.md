# 日期时间、UTC、时区与本地化详解

本文档详细介绍日期时间处理的核心概念，包括 UTC、时区、本地化，以及在 Go、Python、TypeScript 中的实践。

## 目录

1. [核心概念](#核心概念)
2. [Go 时间处理](#go-时间处理)
3. [Python 时间处理](#python-时间处理)
4. [TypeScript/JavaScript 时间处理](#typescriptjavascript-时间处理)
5. [数据库时间存储](#数据库时间存储)
6. [API 时间传输](#api-时间传输)
7. [最佳实践](#最佳实践)

---

## 核心概念

### UTC (协调世界时)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              UTC 概念                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  UTC (Coordinated Universal Time) - 协调世界时                          │
│                                                                         │
│  • 全球统一的时间标准，不受夏令时影响                                    │
│  • 以原子钟为基准，与 GMT (格林威治标准时间) 几乎相同                    │
│  • 时区偏移量以 UTC 为基准计算                                          │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │                        UTC 时间线                            │       │
│  │  ←─────────────────────|─────────────────────→               │       │
│  │                      UTC+0                                   │       │
│  │        UTC-8          |          UTC+8                       │       │
│  │    (洛杉矶 PST)       |       (北京 CST)                     │       │
│  │                       |                                      │       │
│  │    00:00 UTC-8  =  08:00 UTC  =  16:00 UTC+8                │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 时区 (Timezone)

```
常用时区标识符 (IANA Time Zone Database):

┌──────────────────┬────────────┬─────────────────────┐
│ 时区标识符        │ UTC 偏移   │ 说明                │
├──────────────────┼────────────┼─────────────────────┤
│ UTC              │ +00:00     │ 协调世界时          │
│ America/New_York │ -05:00/-04 │ 美国东部 (有夏令时) │
│ America/Los_Angeles│ -08:00/-07│ 美国西部 (有夏令时) │
│ Europe/London    │ +00:00/+01 │ 英国 (有夏令时)     │
│ Europe/Paris     │ +01:00/+02 │ 法国 (有夏令时)     │
│ Asia/Shanghai    │ +08:00     │ 中国 (无夏令时)     │
│ Asia/Tokyo       │ +09:00     │ 日本 (无夏令时)     │
│ Asia/Seoul       │ +09:00     │ 韩国 (无夏令时)     │
│ Asia/Singapore   │ +08:00     │ 新加坡 (无夏令时)   │
│ Australia/Sydney │ +10:00/+11 │ 澳大利亚 (有夏令时) │
└──────────────────┴────────────┴─────────────────────┘

注意：
- 使用 IANA 时区名称 (如 Asia/Shanghai) 而非缩写 (如 CST)
- CST 可能表示多个时区：中国标准时间、美国中部标准时间等
- 夏令时 (DST) 会导致偏移量变化
```

### Unix 时间戳

```
Unix Timestamp (Unix 时间戳):

定义：从 1970-01-01 00:00:00 UTC 到现在的秒数

┌─────────────────────────────────────────────────────────────┐
│ 1970-01-01 00:00:00 UTC                                     │
│         │                                                   │
│         ▼                                                   │
│    timestamp = 0                                            │
│                                                             │
│ 2024-01-15 12:30:45 UTC                                     │
│         │                                                   │
│         ▼                                                   │
│    timestamp = 1705321845                                   │
│                                                             │
│ 单位变体:                                                   │
│   • 秒级时间戳:      1705321845                             │
│   • 毫秒级时间戳:    1705321845000                          │
│   • 微秒级时间戳:    1705321845000000                       │
│   • 纳秒级时间戳:    1705321845000000000                    │
└─────────────────────────────────────────────────────────────┘

优点：
- 与时区无关，全球统一
- 便于计算时间差
- 存储空间小 (一个整数)
```

### ISO 8601 日期格式

```
ISO 8601 标准格式:

┌─────────────────────────────────────────────────────────────┐
│ 完整格式: YYYY-MM-DDTHH:mm:ss.sssZ                          │
│                                                             │
│ 示例:                                                       │
│   2024-01-15T12:30:45Z           # UTC 时间                 │
│   2024-01-15T12:30:45+08:00      # 带时区偏移               │
│   2024-01-15T12:30:45.123Z       # 带毫秒                   │
│   2024-01-15T12:30:45.123456Z    # 带微秒                   │
│   2024-01-15                     # 仅日期                   │
│   12:30:45                       # 仅时间                   │
│                                                             │
│ 各部分说明:                                                 │
│   YYYY  - 4位年份                                           │
│   MM    - 2位月份 (01-12)                                   │
│   DD    - 2位日期 (01-31)                                   │
│   T     - 日期时间分隔符                                    │
│   HH    - 2位小时 (00-23)                                   │
│   mm    - 2位分钟 (00-59)                                   │
│   ss    - 2位秒数 (00-59)                                   │
│   .sss  - 毫秒/微秒/纳秒                                    │
│   Z     - UTC 时区 (Zulu time)                              │
│   +/-HH:mm - 时区偏移                                       │
└─────────────────────────────────────────────────────────────┘
```

### 本地化 (Locale)

```
Locale 本地化:

定义：根据地区/语言习惯格式化日期、数字、货币等

┌─────────────────────────────────────────────────────────────┐
│ 同一时间在不同 Locale 下的显示:                             │
│                                                             │
│ 时间: 2024-01-15 14:30:00                                   │
│                                                             │
│ zh-CN (中国):     2024年1月15日 下午2:30                    │
│ en-US (美国):     January 15, 2024, 2:30 PM                 │
│ en-GB (英国):     15 January 2024, 14:30                    │
│ de-DE (德国):     15. Januar 2024, 14:30                    │
│ ja-JP (日本):     2024年1月15日 14:30                       │
│ ko-KR (韩国):     2024년 1월 15일 오후 2:30                 │
│ ar-SA (阿拉伯):   ١٥ يناير ٢٠٢٤، ٢:٣٠ م                     │
│                                                             │
│ Locale 标识符格式: language-REGION                          │
│   zh-CN  中文-中国                                          │
│   zh-TW  中文-台湾                                          │
│   en-US  英语-美国                                          │
│   en-GB  英语-英国                                          │
└─────────────────────────────────────────────────────────────┘
```

---

## Go 时间处理

### time 包基础

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // ============ 获取当前时间 ============
    now := time.Now()                    // 本地时间
    utcNow := time.Now().UTC()           // UTC 时间
    
    fmt.Println("本地时间:", now)         // 2024-01-15 20:30:45.123456789 +0800 CST
    fmt.Println("UTC时间:", utcNow)       // 2024-01-15 12:30:45.123456789 +0000 UTC
    
    // ============ 创建指定时间 ============
    // time.Date(year, month, day, hour, min, sec, nsec, loc)
    t1 := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)
    t2 := time.Date(2024, 1, 15, 20, 30, 45, 0, time.Local)
    
    // ============ 时间戳 ============
    timestamp := now.Unix()              // 秒级时间戳
    timestampMilli := now.UnixMilli()    // 毫秒级时间戳
    timestampMicro := now.UnixMicro()    // 微秒级时间戳
    timestampNano := now.UnixNano()      // 纳秒级时间戳
    
    // 时间戳转时间
    t3 := time.Unix(timestamp, 0)                    // 秒
    t4 := time.UnixMilli(timestampMilli)             // 毫秒
    t5 := time.UnixMicro(timestampMicro)             // 微秒
    
    // ============ 获取时间组件 ============
    year := now.Year()                   // 年
    month := now.Month()                 // 月 (time.Month 类型)
    day := now.Day()                     // 日
    hour := now.Hour()                   // 时
    minute := now.Minute()               // 分
    second := now.Second()               // 秒
    nanosecond := now.Nanosecond()       // 纳秒
    weekday := now.Weekday()             // 星期 (time.Weekday 类型)
    yearDay := now.YearDay()             // 一年中的第几天
    
    // ============ 时间比较 ============
    t1.Before(t2)                        // t1 是否在 t2 之前
    t1.After(t2)                         // t1 是否在 t2 之后
    t1.Equal(t2)                         // t1 是否等于 t2
    
    // ============ 时间计算 ============
    // 加法
    future := now.Add(24 * time.Hour)              // 加 24 小时
    future = now.AddDate(1, 2, 3)                  // 加 1年2月3天
    
    // 减法
    past := now.Add(-24 * time.Hour)               // 减 24 小时
    
    // 时间差
    duration := t2.Sub(t1)                         // 返回 time.Duration
    
    // ============ Duration ============
    d := 2*time.Hour + 30*time.Minute + 45*time.Second
    fmt.Println(d.Hours())                // 2.5125
    fmt.Println(d.Minutes())              // 150.75
    fmt.Println(d.Seconds())              // 9045
    fmt.Println(d.String())               // "2h30m45s"
}
```

### 时区处理

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // ============ 加载时区 ============
    // 使用 IANA 时区名称
    shanghai, _ := time.LoadLocation("Asia/Shanghai")
    tokyo, _ := time.LoadLocation("Asia/Tokyo")
    newYork, _ := time.LoadLocation("America/New_York")
    london, _ := time.LoadLocation("Europe/London")
    
    // 固定偏移时区
    cst := time.FixedZone("CST", 8*60*60)  // UTC+8
    
    // 预定义时区
    local := time.Local                     // 系统本地时区
    utc := time.UTC                         // UTC 时区
    
    // ============ 时区转换 ============
    now := time.Now()
    
    // 转换到不同时区
    shanghaiTime := now.In(shanghai)
    tokyoTime := now.In(tokyo)
    newYorkTime := now.In(newYork)
    utcTime := now.UTC()
    
    fmt.Println("上海:", shanghaiTime)     // 2024-01-15 20:30:45 +0800 CST
    fmt.Println("东京:", tokyoTime)        // 2024-01-15 21:30:45 +0900 JST
    fmt.Println("纽约:", newYorkTime)      // 2024-01-15 07:30:45 -0500 EST
    fmt.Println("UTC:", utcTime)           // 2024-01-15 12:30:45 +0000 UTC
    
    // ============ 在指定时区创建时间 ============
    // 注意：这表示在该时区的当地时间
    t1 := time.Date(2024, 1, 15, 12, 0, 0, 0, shanghai)  // 北京时间 12:00
    t2 := time.Date(2024, 1, 15, 12, 0, 0, 0, utc)       // UTC 12:00
    
    fmt.Println(t1.UTC())  // 2024-01-15 04:00:00 +0000 UTC
    fmt.Println(t2.UTC())  // 2024-01-15 12:00:00 +0000 UTC
    
    // ============ 获取时区信息 ============
    name, offset := now.Zone()
    fmt.Printf("时区: %s, 偏移: %d秒 (%d小时)\n", name, offset, offset/3600)
    
    // ============ 判断夏令时 ============
    // 纽约在夏季
    summer := time.Date(2024, 7, 15, 12, 0, 0, 0, newYork)
    // 纽约在冬季
    winter := time.Date(2024, 1, 15, 12, 0, 0, 0, newYork)
    
    summerName, summerOffset := summer.Zone()
    winterName, winterOffset := winter.Zone()
    
    fmt.Printf("夏季: %s (UTC%+d)\n", summerName, summerOffset/3600)  // EDT (UTC-4)
    fmt.Printf("冬季: %s (UTC%+d)\n", winterName, winterOffset/3600)  // EST (UTC-5)
}
```

### 时间格式化与解析

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    now := time.Now()
    
    // ============ Go 的格式化布局 ============
    // Go 使用参考时间: Mon Jan 2 15:04:05 MST 2006
    // 记忆方法: 1 2 3 4 5 6 7 (月 日 时 分 秒 年 时区)
    
    // 常用格式
    fmt.Println(now.Format("2006-01-02"))                    // 2024-01-15
    fmt.Println(now.Format("2006-01-02 15:04:05"))           // 2024-01-15 20:30:45
    fmt.Println(now.Format("2006/01/02 15:04:05"))           // 2024/01/15 20:30:45
    fmt.Println(now.Format("2006年01月02日 15:04:05"))        // 2024年01月15日 20:30:45
    fmt.Println(now.Format("15:04:05"))                      // 20:30:45
    fmt.Println(now.Format("3:04 PM"))                       // 8:30 PM
    fmt.Println(now.Format("Mon, 02 Jan 2006"))              // Mon, 15 Jan 2024
    fmt.Println(now.Format("Monday, January 2, 2006"))       // Monday, January 15, 2024
    
    // ISO 8601
    fmt.Println(now.Format(time.RFC3339))                    // 2024-01-15T20:30:45+08:00
    fmt.Println(now.Format(time.RFC3339Nano))                // 2024-01-15T20:30:45.123456789+08:00
    fmt.Println(now.UTC().Format("2006-01-02T15:04:05Z"))    // 2024-01-15T12:30:45Z
    
    // 预定义格式
    fmt.Println(now.Format(time.Layout))                     // 01/02 03:04:05PM '06 -0700
    fmt.Println(now.Format(time.ANSIC))                      // Mon Jan 15 20:30:45 2024
    fmt.Println(now.Format(time.UnixDate))                   // Mon Jan 15 20:30:45 CST 2024
    fmt.Println(now.Format(time.RubyDate))                   // Mon Jan 15 20:30:45 +0800 2024
    fmt.Println(now.Format(time.RFC822))                     // 15 Jan 24 20:30 CST
    fmt.Println(now.Format(time.RFC1123))                    // Mon, 15 Jan 2024 20:30:45 CST
    fmt.Println(now.Format(time.Kitchen))                    // 8:30PM
    
    // ============ 格式化占位符 ============
    /*
    年: 2006, 06
    月: 01, 1, Jan, January
    日: 02, 2, _2 (空格填充)
    时: 15 (24小时), 03, 3 (12小时)
    分: 04, 4
    秒: 05, 5
    毫秒: .000, .999 (尾随零)
    微秒: .000000, .999999
    纳秒: .000000000, .999999999
    AM/PM: PM, pm
    时区: MST, -0700, -07:00, Z0700, Z07:00
    星期: Mon, Monday
    */
    
    // ============ 解析时间字符串 ============
    // time.Parse(layout, value)
    t1, _ := time.Parse("2006-01-02", "2024-01-15")
    t2, _ := time.Parse("2006-01-02 15:04:05", "2024-01-15 20:30:45")
    t3, _ := time.Parse(time.RFC3339, "2024-01-15T20:30:45+08:00")
    
    // 注意：Parse 解析的时间默认是 UTC（如果字符串中没有时区信息）
    fmt.Println(t1)  // 2024-01-15 00:00:00 +0000 UTC
    
    // 在指定时区解析
    shanghai, _ := time.LoadLocation("Asia/Shanghai")
    t4, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-01-15 20:30:45", shanghai)
    fmt.Println(t4)  // 2024-01-15 20:30:45 +0800 CST
    
    // ============ 自定义时区偏移格式解析 ============
    t5, _ := time.Parse("2006-01-02T15:04:05-07:00", "2024-01-15T20:30:45+08:00")
    fmt.Println(t5)  // 2024-01-15 20:30:45 +0800 CST
}
```

### 实用工具函数

```go
package timeutil

import (
    "time"
)

// ============ 获取一天的开始和结束 ============
func StartOfDay(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// ============ 获取一周的开始和结束 ============
func StartOfWeek(t time.Time) time.Time {
    weekday := int(t.Weekday())
    if weekday == 0 {
        weekday = 7 // 周日
    }
    return StartOfDay(t.AddDate(0, 0, -weekday+1)) // 周一开始
}

func EndOfWeek(t time.Time) time.Time {
    return EndOfDay(StartOfWeek(t).AddDate(0, 0, 6))
}

// ============ 获取一月的开始和结束 ============
func StartOfMonth(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func EndOfMonth(t time.Time) time.Time {
    return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// ============ 获取一年的开始和结束 ============
func StartOfYear(t time.Time) time.Time {
    return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

func EndOfYear(t time.Time) time.Time {
    return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// ============ 判断是否同一天 ============
func IsSameDay(t1, t2 time.Time) bool {
    y1, m1, d1 := t1.Date()
    y2, m2, d2 := t2.Date()
    return y1 == y2 && m1 == m2 && d1 == d2
}

// ============ 计算两个时间之间的天数 ============
func DaysBetween(t1, t2 time.Time) int {
    t1 = StartOfDay(t1)
    t2 = StartOfDay(t2)
    return int(t2.Sub(t1).Hours() / 24)
}

// ============ 时区转换工具 ============
type TimeConverter struct {
    locations map[string]*time.Location
}

func NewTimeConverter() *TimeConverter {
    tc := &TimeConverter{
        locations: make(map[string]*time.Location),
    }
    // 预加载常用时区
    for _, tz := range []string{
        "UTC", "Asia/Shanghai", "Asia/Tokyo", 
        "America/New_York", "Europe/London",
    } {
        loc, _ := time.LoadLocation(tz)
        tc.locations[tz] = loc
    }
    return tc
}

func (tc *TimeConverter) Convert(t time.Time, timezone string) (time.Time, error) {
    loc, ok := tc.locations[timezone]
    if !ok {
        var err error
        loc, err = time.LoadLocation(timezone)
        if err != nil {
            return t, err
        }
        tc.locations[timezone] = loc
    }
    return t.In(loc), nil
}

// ============ 相对时间描述 ============
func RelativeTime(t time.Time) string {
    now := time.Now()
    diff := now.Sub(t)
    
    if diff < 0 {
        diff = -diff
        return formatFuture(diff)
    }
    return formatPast(diff)
}

func formatPast(d time.Duration) string {
    switch {
    case d < time.Minute:
        return "刚刚"
    case d < time.Hour:
        return fmt.Sprintf("%d 分钟前", int(d.Minutes()))
    case d < 24*time.Hour:
        return fmt.Sprintf("%d 小时前", int(d.Hours()))
    case d < 48*time.Hour:
        return "昨天"
    case d < 7*24*time.Hour:
        return fmt.Sprintf("%d 天前", int(d.Hours()/24))
    case d < 30*24*time.Hour:
        return fmt.Sprintf("%d 周前", int(d.Hours()/(24*7)))
    case d < 365*24*time.Hour:
        return fmt.Sprintf("%d 个月前", int(d.Hours()/(24*30)))
    default:
        return fmt.Sprintf("%d 年前", int(d.Hours()/(24*365)))
    }
}
```

---

## Python 时间处理

### datetime 模块基础

```python
from datetime import datetime, date, time, timedelta, timezone
from zoneinfo import ZoneInfo  # Python 3.9+
import time as time_module

# ============ 获取当前时间 ============
now = datetime.now()                    # 本地时间（naive）
utc_now = datetime.now(timezone.utc)    # UTC 时间（aware）
utc_now = datetime.utcnow()             # UTC 时间（naive，不推荐）

print(f"本地时间: {now}")                # 2024-01-15 20:30:45.123456
print(f"UTC时间: {utc_now}")             # 2024-01-15 12:30:45.123456+00:00

# ============ 创建指定时间 ============
dt = datetime(2024, 1, 15, 12, 30, 45)                    # naive datetime
dt = datetime(2024, 1, 15, 12, 30, 45, tzinfo=timezone.utc)  # aware datetime

d = date(2024, 1, 15)                   # 日期
t = time(12, 30, 45)                    # 时间

# ============ 时间戳 ============
timestamp = now.timestamp()              # 浮点数秒级时间戳
timestamp_int = int(now.timestamp())     # 整数秒级时间戳
timestamp_ms = int(now.timestamp() * 1000)  # 毫秒级

# 时间戳转 datetime
dt = datetime.fromtimestamp(timestamp)                     # 本地时间
dt = datetime.fromtimestamp(timestamp, tz=timezone.utc)    # UTC 时间
dt = datetime.utcfromtimestamp(timestamp)                  # UTC（naive，不推荐）

# ============ 获取时间组件 ============
year = now.year
month = now.month
day = now.day
hour = now.hour
minute = now.minute
second = now.second
microsecond = now.microsecond
weekday = now.weekday()      # 0=周一, 6=周日
isoweekday = now.isoweekday()  # 1=周一, 7=周日

# ============ 时间比较 ============
dt1 = datetime(2024, 1, 15)
dt2 = datetime(2024, 1, 16)

print(dt1 < dt2)    # True
print(dt1 == dt2)   # False
print(dt1 != dt2)   # True

# ============ 时间计算 ============
# 加法
future = now + timedelta(days=1, hours=2, minutes=30)

# 减法
past = now - timedelta(weeks=1)

# 时间差
diff = dt2 - dt1    # timedelta 对象
print(diff.days)           # 1
print(diff.seconds)        # 0
print(diff.total_seconds()) # 86400.0
```

### 时区处理

```python
from datetime import datetime, timezone, timedelta
from zoneinfo import ZoneInfo  # Python 3.9+

# ============ 创建时区 ============
# 使用 zoneinfo (推荐，Python 3.9+)
shanghai = ZoneInfo("Asia/Shanghai")
tokyo = ZoneInfo("Asia/Tokyo")
new_york = ZoneInfo("America/New_York")
london = ZoneInfo("Europe/London")
utc = timezone.utc

# 固定偏移时区
cst = timezone(timedelta(hours=8))      # UTC+8

# ============ 创建带时区的时间 ============
# 方法1: 创建时指定
dt = datetime(2024, 1, 15, 12, 0, 0, tzinfo=shanghai)

# 方法2: 给 naive datetime 添加时区信息
naive_dt = datetime(2024, 1, 15, 12, 0, 0)
aware_dt = naive_dt.replace(tzinfo=shanghai)  # 假设原时间是该时区的

# ============ 时区转换 ============
utc_time = datetime.now(timezone.utc)

# 转换到其他时区
shanghai_time = utc_time.astimezone(shanghai)
tokyo_time = utc_time.astimezone(tokyo)
new_york_time = utc_time.astimezone(new_york)

print(f"UTC: {utc_time}")
print(f"上海: {shanghai_time}")
print(f"东京: {tokyo_time}")
print(f"纽约: {new_york_time}")

# ============ 获取时区信息 ============
dt = datetime.now(shanghai)
print(dt.tzname())      # CST
print(dt.utcoffset())   # 8:00:00
print(dt.dst())         # 夏令时偏移

# ============ Naive vs Aware ============
naive = datetime.now()           # 没有时区信息
aware = datetime.now(timezone.utc)  # 有时区信息

print(naive.tzinfo)   # None
print(aware.tzinfo)   # UTC

# 检查是否有时区
def is_aware(dt):
    return dt.tzinfo is not None and dt.tzinfo.utcoffset(dt) is not None

# ============ 统一转换为 UTC ============
def to_utc(dt):
    """将任意 datetime 转换为 UTC"""
    if dt.tzinfo is None:
        # naive datetime，假设是本地时间
        dt = dt.astimezone()  # 先转为本地时区
    return dt.astimezone(timezone.utc)

# ============ 夏令时处理 ============
new_york = ZoneInfo("America/New_York")

# 冬季时间 (EST, UTC-5)
winter = datetime(2024, 1, 15, 12, 0, 0, tzinfo=new_york)
print(f"冬季: {winter} ({winter.tzname()})")  # EST

# 夏季时间 (EDT, UTC-4)
summer = datetime(2024, 7, 15, 12, 0, 0, tzinfo=new_york)
print(f"夏季: {summer} ({summer.tzname()})")  # EDT
```

### 时间格式化与解析

```python
from datetime import datetime, timezone
from zoneinfo import ZoneInfo

now = datetime.now()

# ============ strftime 格式化 ============
print(now.strftime("%Y-%m-%d"))                    # 2024-01-15
print(now.strftime("%Y-%m-%d %H:%M:%S"))           # 2024-01-15 20:30:45
print(now.strftime("%Y/%m/%d %H:%M:%S"))           # 2024/01/15 20:30:45
print(now.strftime("%Y年%m月%d日 %H:%M:%S"))        # 2024年01月15日 20:30:45
print(now.strftime("%H:%M:%S"))                    # 20:30:45
print(now.strftime("%I:%M %p"))                    # 08:30 PM
print(now.strftime("%A, %B %d, %Y"))               # Monday, January 15, 2024
print(now.strftime("%a, %d %b %Y"))                # Mon, 15 Jan 2024

# ISO 8601
print(now.isoformat())                             # 2024-01-15T20:30:45.123456
print(now.astimezone(timezone.utc).isoformat())    # 2024-01-15T12:30:45.123456+00:00

# ============ 格式化占位符 ============
"""
%Y  4位年份 (2024)
%y  2位年份 (24)
%m  月份 (01-12)
%d  日期 (01-31)
%H  24小时制小时 (00-23)
%I  12小时制小时 (01-12)
%M  分钟 (00-59)
%S  秒 (00-59)
%f  微秒 (000000-999999)
%p  AM/PM
%z  UTC偏移 (+0800)
%Z  时区名称 (CST)
%a  星期缩写 (Mon)
%A  星期全称 (Monday)
%b  月份缩写 (Jan)
%B  月份全称 (January)
%j  一年中的第几天 (001-366)
%U  一年中的第几周 (周日为第一天)
%W  一年中的第几周 (周一为第一天)
%w  星期几 (0=周日)
"""

# ============ strptime 解析 ============
dt1 = datetime.strptime("2024-01-15", "%Y-%m-%d")
dt2 = datetime.strptime("2024-01-15 20:30:45", "%Y-%m-%d %H:%M:%S")
dt3 = datetime.strptime("2024-01-15T20:30:45+08:00", "%Y-%m-%dT%H:%M:%S%z")

# ISO 格式解析 (Python 3.7+)
dt4 = datetime.fromisoformat("2024-01-15T20:30:45")
dt5 = datetime.fromisoformat("2024-01-15T20:30:45+08:00")

# ============ 注意事项 ============
# strptime 解析的结果是 naive datetime
dt = datetime.strptime("2024-01-15 12:00:00", "%Y-%m-%d %H:%M:%S")
print(dt.tzinfo)  # None

# 需要手动添加时区
shanghai = ZoneInfo("Asia/Shanghai")
dt = dt.replace(tzinfo=shanghai)
```

### dateutil 库

```python
# pip install python-dateutil
from dateutil import parser, tz, relativedelta
from datetime import datetime

# ============ 智能解析 ============
# dateutil 可以解析各种格式的日期字符串
dt1 = parser.parse("2024-01-15")
dt2 = parser.parse("January 15, 2024")
dt3 = parser.parse("15/01/2024")  # 注意：默认月/日/年
dt4 = parser.parse("15/01/2024", dayfirst=True)  # 日/月/年
dt5 = parser.parse("2024-01-15T20:30:45+08:00")
dt6 = parser.parse("Mon Jan 15 20:30:45 CST 2024")

# ============ 时区处理 ============
utc = tz.UTC
shanghai = tz.gettz("Asia/Shanghai")
local = tz.tzlocal()  # 系统本地时区

# 转换时区
dt = datetime.now(utc)
dt_shanghai = dt.astimezone(shanghai)

# ============ 相对时间计算 ============
from dateutil.relativedelta import relativedelta

now = datetime.now()

# 加减时间
next_month = now + relativedelta(months=1)
last_year = now - relativedelta(years=1)
next_weekday = now + relativedelta(weekday=0)  # 下一个周一

# 复杂计算
result = now + relativedelta(
    years=1,
    months=2,
    days=3,
    hours=4,
    minutes=5,
    seconds=6
)

# 计算两个日期之间的差值
dt1 = datetime(2024, 1, 15)
dt2 = datetime(2025, 3, 20)
diff = relativedelta(dt2, dt1)
print(f"{diff.years}年 {diff.months}月 {diff.days}天")  # 1年 2月 5天
```

---

## TypeScript/JavaScript 时间处理

### 原生 Date 对象

```typescript
// ============ 获取当前时间 ============
const now = new Date();                          // 本地时间
const utcNow = new Date().toISOString();         // UTC ISO 字符串

console.log(now);                                // Mon Jan 15 2024 20:30:45 GMT+0800
console.log(utcNow);                             // 2024-01-15T12:30:45.123Z

// ============ 创建指定时间 ============
const dt1 = new Date(2024, 0, 15);               // 月份从 0 开始！
const dt2 = new Date(2024, 0, 15, 12, 30, 45);
const dt3 = new Date("2024-01-15T12:30:45");     // ISO 格式
const dt4 = new Date("2024-01-15T12:30:45Z");    // UTC
const dt5 = new Date("2024-01-15T12:30:45+08:00"); // 带时区

// ============ 时间戳 ============
const timestamp = now.getTime();                  // 毫秒级时间戳
const timestampSec = Math.floor(now.getTime() / 1000);  // 秒级

// 时间戳转 Date
const dt = new Date(timestamp);
const dt2 = new Date(timestampSec * 1000);

// 当前时间戳
const currentTimestamp = Date.now();

// ============ 获取时间组件 ============
const year = now.getFullYear();
const month = now.getMonth();          // 0-11！
const date = now.getDate();            // 1-31
const day = now.getDay();              // 0=周日, 6=周六
const hours = now.getHours();
const minutes = now.getMinutes();
const seconds = now.getSeconds();
const milliseconds = now.getMilliseconds();

// UTC 版本
const utcYear = now.getUTCFullYear();
const utcMonth = now.getUTCMonth();
const utcHours = now.getUTCHours();

// ============ 设置时间组件 ============
const dt = new Date();
dt.setFullYear(2025);
dt.setMonth(5);                        // 6月
dt.setDate(20);
dt.setHours(14, 30, 0, 0);             // 时、分、秒、毫秒

// ============ 时间比较 ============
const dt1 = new Date("2024-01-15");
const dt2 = new Date("2024-01-16");

console.log(dt1 < dt2);                // true
console.log(dt1.getTime() === dt2.getTime());  // false

// ============ 时间计算 ============
const now = new Date();

// 加一天
const tomorrow = new Date(now);
tomorrow.setDate(tomorrow.getDate() + 1);

// 加一个月
const nextMonth = new Date(now);
nextMonth.setMonth(nextMonth.getMonth() + 1);

// 时间差（毫秒）
const diff = dt2.getTime() - dt1.getTime();
const diffDays = diff / (1000 * 60 * 60 * 24);
```

### 时区处理

```typescript
// ============ 获取时区信息 ============
const now = new Date();

// 时区偏移（分钟）
const offset = now.getTimezoneOffset();  // -480 表示 UTC+8
const offsetHours = -offset / 60;        // 8

// 时区名称
const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
console.log(timeZone);  // "Asia/Shanghai"

// ============ 格式化到指定时区 ============
const now = new Date();

// 使用 toLocaleString
const options: Intl.DateTimeFormatOptions = {
  timeZone: "America/New_York",
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hour12: false,
};

console.log(now.toLocaleString("en-US", options));
// 01/15/2024, 07:30:45

console.log(now.toLocaleString("zh-CN", { timeZone: "Asia/Shanghai" }));
// 2024/1/15 20:30:45

console.log(now.toLocaleString("ja-JP", { timeZone: "Asia/Tokyo" }));
// 2024/1/15 21:30:45

// ============ Intl.DateTimeFormat ============
const formatter = new Intl.DateTimeFormat("zh-CN", {
  timeZone: "Asia/Shanghai",
  year: "numeric",
  month: "long",
  day: "numeric",
  weekday: "long",
  hour: "numeric",
  minute: "numeric",
  second: "numeric",
});

console.log(formatter.format(now));
// 2024年1月15日星期一 20:30:45

// 格式化为各部分
const parts = formatter.formatToParts(now);
console.log(parts);
// [{ type: "year", value: "2024" }, { type: "literal", value: "年" }, ...]

// ============ 时区转换工具函数 ============
function convertTimezone(date: Date, targetTimezone: string): string {
  return date.toLocaleString("en-US", {
    timeZone: targetTimezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
}

// 获取指定时区的当前时间
function getTimeInTimezone(timezone: string): Date {
  const str = new Date().toLocaleString("en-US", { timeZone: timezone });
  return new Date(str);
}
```

### 本地化格式

```typescript
const now = new Date();

// ============ 不同 locale 的日期格式 ============
console.log(now.toLocaleDateString("zh-CN"));    // 2024/1/15
console.log(now.toLocaleDateString("en-US"));    // 1/15/2024
console.log(now.toLocaleDateString("en-GB"));    // 15/01/2024
console.log(now.toLocaleDateString("de-DE"));    // 15.1.2024
console.log(now.toLocaleDateString("ja-JP"));    // 2024/1/15
console.log(now.toLocaleDateString("ko-KR"));    // 2024. 1. 15.

// ============ 完整格式选项 ============
const options: Intl.DateTimeFormatOptions = {
  // 日期
  year: "numeric",      // "numeric" | "2-digit"
  month: "long",        // "numeric" | "2-digit" | "long" | "short" | "narrow"
  day: "numeric",       // "numeric" | "2-digit"
  weekday: "long",      // "long" | "short" | "narrow"
  
  // 时间
  hour: "2-digit",      // "numeric" | "2-digit"
  minute: "2-digit",
  second: "2-digit",
  fractionalSecondDigits: 3,  // 毫秒位数
  
  // 格式
  hour12: false,        // 12/24 小时制
  timeZone: "Asia/Shanghai",
  timeZoneName: "short", // "short" | "long" | "shortOffset" | "longOffset"
};

console.log(now.toLocaleString("zh-CN", options));
// 2024年1月15日星期一 20:30:45 GMT+8

// ============ 相对时间格式化 ============
const rtf = new Intl.RelativeTimeFormat("zh-CN", { numeric: "auto" });

console.log(rtf.format(-1, "day"));     // 昨天
console.log(rtf.format(-2, "day"));     // 2天前
console.log(rtf.format(1, "day"));      // 明天
console.log(rtf.format(-1, "week"));    // 上周
console.log(rtf.format(-1, "month"));   // 上个月
console.log(rtf.format(-1, "year"));    // 去年

// ============ 时间长度格式化 ============
const dtf = new Intl.DurationFormat("zh-CN", { style: "long" });
// 注意：DurationFormat 是较新的 API，可能需要 polyfill
```

### day.js 库

```typescript
// npm install dayjs
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";
import relativeTime from "dayjs/plugin/relativeTime";
import "dayjs/locale/zh-cn";

// 加载插件
dayjs.extend(utc);
dayjs.extend(timezone);
dayjs.extend(relativeTime);

// 设置默认语言
dayjs.locale("zh-cn");

// ============ 基本用法 ============
const now = dayjs();
const dt = dayjs("2024-01-15");
const dt2 = dayjs("2024-01-15T12:30:45+08:00");

// ============ 格式化 ============
now.format("YYYY-MM-DD");              // 2024-01-15
now.format("YYYY-MM-DD HH:mm:ss");     // 2024-01-15 20:30:45
now.format("YYYY年MM月DD日");           // 2024年01月15日
now.format("dddd");                    // 星期一
now.format();                          // ISO 8601

// ============ 时区处理 ============
// 当前时区
const local = dayjs();

// 转换到其他时区
const shanghai = dayjs().tz("Asia/Shanghai");
const tokyo = dayjs().tz("Asia/Tokyo");
const newYork = dayjs().tz("America/New_York");

// UTC 时间
const utcTime = dayjs.utc();

// 在指定时区解析
const dt = dayjs.tz("2024-01-15 12:00:00", "Asia/Shanghai");

// ============ 时间计算 ============
now.add(1, "day");
now.add(1, "week");
now.add(1, "month");
now.add(1, "year");
now.subtract(1, "hour");

// 开始/结束
now.startOf("day");
now.startOf("month");
now.startOf("year");
now.endOf("day");

// ============ 比较 ============
dayjs("2024-01-15").isBefore("2024-01-16");
dayjs("2024-01-15").isAfter("2024-01-14");
dayjs("2024-01-15").isSame("2024-01-15");
dayjs("2024-01-15").isSame("2024-01-16", "month");  // 同月

// ============ 相对时间 ============
dayjs("2024-01-14").fromNow();         // 1天前
dayjs("2024-01-16").toNow();           // 1天后
dayjs("2024-01-10").from(dayjs("2024-01-15"));  // 5天前

// ============ 时间戳 ============
now.unix();                            // 秒级
now.valueOf();                         // 毫秒级
dayjs.unix(1705321845);                // 从时间戳创建
```

---

## 数据库时间存储

### MySQL

```sql
-- ============ 时间类型 ============
CREATE TABLE events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    
    -- DATE: 只存日期 (YYYY-MM-DD)
    event_date DATE,
    
    -- TIME: 只存时间 (HH:MM:SS)
    event_time TIME,
    
    -- DATETIME: 日期时间，不带时区 (YYYY-MM-DD HH:MM:SS)
    -- 范围: 1000-01-01 到 9999-12-31
    created_at DATETIME,
    
    -- TIMESTAMP: 日期时间，内部存储为 UTC
    -- 范围: 1970-01-01 到 2038-01-19
    -- 会自动转换时区
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 带微秒精度
    precise_time DATETIME(6),
    precise_ts TIMESTAMP(6)
);

-- ============ DATETIME vs TIMESTAMP ============
-- DATETIME: 存什么就是什么，不转换
-- TIMESTAMP: 存储时转为 UTC，读取时转为当前时区

-- 设置会话时区
SET time_zone = '+08:00';
SET time_zone = 'Asia/Shanghai';

-- 查看当前时区
SELECT @@global.time_zone, @@session.time_zone;

-- ============ 时间函数 ============
SELECT NOW();                          -- 当前本地时间
SELECT UTC_TIMESTAMP();                -- 当前 UTC 时间
SELECT CURRENT_TIMESTAMP;              -- 同 NOW()
SELECT UNIX_TIMESTAMP();               -- 当前时间戳
SELECT FROM_UNIXTIME(1705321845);      -- 时间戳转 DATETIME

-- 时区转换
SELECT CONVERT_TZ('2024-01-15 12:00:00', '+00:00', '+08:00');
SELECT CONVERT_TZ(NOW(), @@session.time_zone, '+00:00');  -- 转 UTC

-- 格式化
SELECT DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s');
SELECT DATE_FORMAT(NOW(), '%Y年%m月%d日');

-- ============ 推荐做法 ============
-- 1. 使用 DATETIME 存储，应用层处理时区
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    created_at DATETIME NOT NULL,  -- 存储 UTC 时间
    timezone VARCHAR(50) DEFAULT 'UTC'  -- 用户时区偏好
);

-- 2. 或使用 TIMESTAMP（自动 UTC 转换）
CREATE TABLE logs (
    id BIGINT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### PostgreSQL

```sql
-- ============ 时间类型 ============
CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    
    -- timestamp without time zone (等同于 timestamp)
    -- 不带时区信息，存什么就是什么
    created_at TIMESTAMP,
    
    -- timestamp with time zone (等同于 timestamptz)
    -- 内部存储为 UTC，显示时转换为当前时区
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- 日期和时间
    event_date DATE,
    event_time TIME,
    event_time_tz TIME WITH TIME ZONE,
    
    -- 时间间隔
    duration INTERVAL
);

-- ============ 时区设置 ============
-- 设置时区
SET timezone = 'Asia/Shanghai';
SET timezone = 'UTC';

-- 查看时区
SHOW timezone;

-- ============ 时区转换 ============
-- AT TIME ZONE 用于转换时区
SELECT NOW() AT TIME ZONE 'UTC';
SELECT NOW() AT TIME ZONE 'Asia/Shanghai';

-- timestamp -> timestamptz (添加时区信息)
SELECT '2024-01-15 12:00:00'::timestamp AT TIME ZONE 'Asia/Shanghai';

-- timestamptz -> timestamp (转换到指定时区)
SELECT NOW() AT TIME ZONE 'America/New_York';

-- ============ 时间函数 ============
SELECT NOW();                          -- 当前时间 (带时区)
SELECT CURRENT_TIMESTAMP;              -- 同上
SELECT LOCALTIMESTAMP;                 -- 本地时间 (不带时区)
SELECT EXTRACT(EPOCH FROM NOW());      -- Unix 时间戳
SELECT TO_TIMESTAMP(1705321845);       -- 时间戳转时间

-- 格式化
SELECT TO_CHAR(NOW(), 'YYYY-MM-DD HH24:MI:SS');
SELECT TO_CHAR(NOW(), 'YYYY"年"MM"月"DD"日"');

-- ============ 推荐做法 ============
-- 始终使用 TIMESTAMPTZ
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### MongoDB

```javascript
// MongoDB 使用 BSON Date 类型，内部存储为 UTC 毫秒时间戳

// ============ 插入时间 ============
db.events.insertOne({
    name: "Event 1",
    // Date 对象会自动转为 UTC 存储
    createdAt: new Date(),
    // ISODate 是 Date 的别名
    updatedAt: ISODate("2024-01-15T12:30:45Z"),
    // 时间戳
    timestamp: new Date().getTime()
});

// ============ 查询时间 ============
// 范围查询
db.events.find({
    createdAt: {
        $gte: ISODate("2024-01-01T00:00:00Z"),
        $lt: ISODate("2024-02-01T00:00:00Z")
    }
});

// 今天的数据
const today = new Date();
today.setHours(0, 0, 0, 0);
const tomorrow = new Date(today);
tomorrow.setDate(tomorrow.getDate() + 1);

db.events.find({
    createdAt: { $gte: today, $lt: tomorrow }
});

// ============ 聚合中的时间处理 ============
db.events.aggregate([
    {
        $project: {
            // 提取时间部分
            year: { $year: "$createdAt" },
            month: { $month: "$createdAt" },
            day: { $dayOfMonth: "$createdAt" },
            hour: { $hour: "$createdAt" },
            
            // 时区转换
            localTime: {
                $dateToString: {
                    format: "%Y-%m-%d %H:%M:%S",
                    date: "$createdAt",
                    timezone: "Asia/Shanghai"
                }
            }
        }
    }
]);

// 按天分组统计
db.events.aggregate([
    {
        $group: {
            _id: {
                $dateToString: {
                    format: "%Y-%m-%d",
                    date: "$createdAt",
                    timezone: "Asia/Shanghai"
                }
            },
            count: { $sum: 1 }
        }
    }
]);
```

---

## API 时间传输

### 最佳实践

```
API 时间传输推荐格式:

1. ISO 8601 字符串 (推荐)
   "2024-01-15T12:30:45Z"           # UTC
   "2024-01-15T20:30:45+08:00"      # 带时区

2. Unix 时间戳
   1705321845                       # 秒级
   1705321845000                    # 毫秒级

规范建议:
┌─────────────────────────────────────────────────────────────┐
│ 1. 服务端统一使用 UTC 存储和处理                            │
│ 2. API 传输使用 ISO 8601 格式，始终带时区信息               │
│ 3. 客户端根据用户时区进行显示转换                           │
│ 4. 响应头中包含服务器时间，便于时间同步                     │
└─────────────────────────────────────────────────────────────┘
```

### 请求/响应示例

```json
// ============ 请求示例 ============
// POST /api/events
{
    "name": "Meeting",
    "startTime": "2024-01-15T14:00:00+08:00",
    "endTime": "2024-01-15T16:00:00+08:00",
    "timezone": "Asia/Shanghai"
}

// ============ 响应示例 ============
{
    "id": 1,
    "name": "Meeting",
    "startTime": "2024-01-15T06:00:00Z",
    "endTime": "2024-01-15T08:00:00Z",
    "timezone": "Asia/Shanghai",
    "createdAt": "2024-01-15T05:30:00Z",
    "updatedAt": "2024-01-15T05:30:00Z"
}

// ============ 列表响应 ============
{
    "data": [...],
    "meta": {
        "serverTime": "2024-01-15T12:30:45Z",
        "timezone": "UTC"
    }
}
```

### Go API 示例

```go
// 自定义 JSON 时间格式
type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
    stamp := time.Time(t).UTC().Format(time.RFC3339)
    return []byte(`"` + stamp + `"`), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
    str := strings.Trim(string(data), `"`)
    parsed, err := time.Parse(time.RFC3339, str)
    if err != nil {
        return err
    }
    *t = JSONTime(parsed)
    return nil
}

// 请求/响应结构
type EventRequest struct {
    Name      string   `json:"name"`
    StartTime JSONTime `json:"startTime"`
    Timezone  string   `json:"timezone"`
}

type EventResponse struct {
    ID        int64    `json:"id"`
    Name      string   `json:"name"`
    StartTime JSONTime `json:"startTime"`
    CreatedAt JSONTime `json:"createdAt"`
}
```

### Python API 示例

```python
from datetime import datetime, timezone
from pydantic import BaseModel, field_validator
from typing import Optional

class EventRequest(BaseModel):
    name: str
    start_time: datetime
    timezone: str = "UTC"
    
    @field_validator("start_time", mode="before")
    @classmethod
    def parse_datetime(cls, v):
        if isinstance(v, str):
            dt = datetime.fromisoformat(v.replace("Z", "+00:00"))
            return dt.astimezone(timezone.utc)
        return v

class EventResponse(BaseModel):
    id: int
    name: str
    start_time: datetime
    created_at: datetime
    
    class Config:
        json_encoders = {
            datetime: lambda v: v.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")
        }
```

---

## 最佳实践

### 总结

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         时间处理最佳实践                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  存储层:                                                                │
│  ├─ 始终使用 UTC 时间存储                                               │
│  ├─ 数据库选择: TIMESTAMP (MySQL) 或 TIMESTAMPTZ (PostgreSQL)          │
│  └─ 存储用户时区偏好，用于显示转换                                      │
│                                                                         │
│  应用层:                                                                │
│  ├─ 内部处理统一使用 UTC                                                │
│  ├─ 时区转换只在输入/输出边界进行                                       │
│  ├─ 使用 IANA 时区名称 (Asia/Shanghai) 而非缩写 (CST)                  │
│  └─ 注意夏令时的影响                                                    │
│                                                                         │
│  API 层:                                                                │
│  ├─ 使用 ISO 8601 格式传输                                              │
│  ├─ 始终包含时区信息                                                    │
│  ├─ 响应中返回 UTC 时间                                                 │
│  └─ 提供服务器时间用于客户端同步                                        │
│                                                                         │
│  客户端:                                                                │
│  ├─ 根据用户时区显示本地时间                                            │
│  ├─ 发送请求时转换为 UTC 或带时区的 ISO 格式                           │
│  └─ 使用 Intl API 进行本地化格式化                                      │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 常见问题

```
Q: 应该存储 UTC 还是本地时间？
A: 始终存储 UTC。本地时间会因夏令时、时区变更等产生歧义。

Q: 时间戳还是日期字符串？
A: 时间戳便于计算和存储，ISO 8601 字符串可读性好。API 推荐 ISO 8601。

Q: 如何处理用户的时区？
A: 存储用户的时区偏好，在显示时转换。不要试图"猜测"用户时区。

Q: 如何处理跨时区的定时任务？
A: 使用 UTC 时间定义任务，或明确指定时区（如 "每天北京时间 9:00"）。

Q: 如何处理"每月最后一天"这样的相对日期？
A: 使用语言提供的日期计算库，不要手动计算。

Q: 前端应该如何显示时间？
A: 使用 Intl.DateTimeFormat 根据用户 locale 格式化，提供相对时间（如"3小时前"）。
```

### 检查清单

```
□ 服务端是否统一使用 UTC？
□ 数据库时间类型是否正确？
□ API 是否使用 ISO 8601 格式？
□ 是否处理了夏令时？
□ 用户时区偏好是否保存？
□ 前端是否正确转换显示？
□ 日志时间是否使用 UTC？
□ 定时任务时区是否明确？
```

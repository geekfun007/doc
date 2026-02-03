# GORM 实用指南与底层详解

本文档详细介绍 Go 语言 ORM 库 GORM 的使用方法和底层实现原理。

## 目录

1. [快速开始](#快速开始)
2. [模型定义](#模型定义)
3. [CRUD 操作](#crud-操作)
4. [GORM 查询语法与占位符](#gorm-查询语法与占位符)
5. [查询详解](#查询详解)
6. [关联关系](#关联关系)
7. [事务处理](#事务处理)
8. [钩子函数](#钩子函数)
9. [高级特性](#高级特性)
10. [性能优化](#性能优化)
11. [底层原理](#底层原理)

---

## 快速开始

### 安装

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
go get -u gorm.io/driver/postgres
go get -u gorm.io/driver/sqlite
```

### 连接数据库

```go
package main

import (
    "log"
    "time"

    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// MySQL 连接
func connectMySQL() (*gorm.DB, error) {
    dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }
    
    // 获取底层 sql.DB 进行连接池配置
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    // 连接池配置
    sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
    sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
    sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
    sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接最大生命周期
    
    return db, nil
}

// PostgreSQL 连接
func connectPostgres() (*gorm.DB, error) {
    dsn := "host=localhost user=postgres password=password dbname=mydb port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// 带完整配置的连接
func connectWithConfig() (*gorm.DB, error) {
    dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    
    return gorm.Open(mysql.New(mysql.Config{
        DSN:                       dsn,
        DefaultStringSize:         256,   // string 类型默认长度
        DisableDatetimePrecision:  true,  // 禁用 datetime 精度
        DontSupportRenameIndex:    true,  // 重命名索引时删除再创建
        DontSupportRenameColumn:   true,  // 用 change 重命名列
        SkipInitializeWithVersion: false, // 根据版本自动配置
    }), &gorm.Config{
        SkipDefaultTransaction: true,                   // 跳过默认事务
        NamingStrategy: schema.NamingStrategy{
            TablePrefix:   "t_",                        // 表名前缀
            SingularTable: true,                        // 使用单数表名
        },
        DisableForeignKeyConstraintWhenMigrating: true, // 迁移时禁用外键
        PrepareStmt:                              true, // 预编译语句
    })
}
```

### 基本使用

```go
// 定义模型
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"size:100;not null"`
    Email     string         `gorm:"uniqueIndex;size:200"`
    Age       int            `gorm:"default:18"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除
}

func main() {
    db, err := connectMySQL()
    if err != nil {
        log.Fatal(err)
    }
    
    // 自动迁移
    db.AutoMigrate(&User{})
    
    // 创建
    user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
    db.Create(&user)
    
    // 查询
    var result User
    db.First(&result, user.ID)
    
    // 更新
    db.Model(&result).Update("Age", 26)
    
    // 删除
    db.Delete(&result)
}
```

---

## 模型定义

### 字段标签

```go
type Product struct {
    // 主键
    ID uint `gorm:"primaryKey;autoIncrement"`
    
    // 字符串
    Name        string `gorm:"type:varchar(100);not null;comment:商品名称"`
    Description string `gorm:"type:text"`
    Code        string `gorm:"size:50;uniqueIndex:idx_code"`
    
    // 数字
    Price    float64 `gorm:"type:decimal(10,2);default:0"`
    Stock    int     `gorm:"default:0;check:stock >= 0"`
    Priority int     `gorm:"index;default:0"`
    
    // 布尔
    IsActive bool `gorm:"default:true"`
    
    // 时间
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
    PublishAt *time.Time     // 可空时间
    
    // JSON
    Metadata datatypes.JSON `gorm:"type:json"`
    Tags     datatypes.JSON `gorm:"type:json"`
    
    // 忽略字段
    TempData string `gorm:"-"`              // 完全忽略
    ReadOnly string `gorm:"->"`             // 只读
    WriteOnly string `gorm:"<-"`            // 只写
    CreateOnly string `gorm:"<-:create"`    // 只在创建时写入
    UpdateOnly string `gorm:"<-:update"`    // 只在更新时写入
    
    // 外键
    CategoryID uint
    Category   Category `gorm:"foreignKey:CategoryID;references:ID"`
}
```

### 常用标签速查

| 标签 | 说明 |
|------|------|
| `primaryKey` | 主键 |
| `autoIncrement` | 自增 |
| `size:100` | 字段大小 |
| `type:varchar(100)` | 指定类型 |
| `not null` | 非空 |
| `default:0` | 默认值 |
| `uniqueIndex` | 唯一索引 |
| `index` | 普通索引 |
| `index:idx_name` | 命名索引 |
| `index:,composite:idx_name` | 复合索引 |
| `comment:注释` | 字段注释 |
| `check:age > 0` | 检查约束 |
| `embedded` | 嵌入结构 |
| `embeddedPrefix:prefix_` | 嵌入前缀 |
| `foreignKey:FieldName` | 外键 |
| `references:ID` | 引用字段 |
| `constraint:OnUpdate:CASCADE` | 约束 |

### 嵌入结构

```go
// 通用字段
type BaseModel struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 审计字段
type AuditModel struct {
    CreatedBy uint `gorm:"index"`
    UpdatedBy uint
}

// 组合使用
type Article struct {
    BaseModel            // 匿名嵌入
    AuditModel           // 匿名嵌入
    Title   string
    Content string
}

// 带前缀的嵌入
type Author struct {
    Name  string
    Email string
}

type Book struct {
    ID     uint
    Title  string
    Author Author `gorm:"embedded;embeddedPrefix:author_"`
    // 生成字段：author_name, author_email
}
```

### 自定义类型

```go
import (
    "database/sql/driver"
    "encoding/json"
    "errors"
)

// 自定义 JSON 类型
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
    if s == nil {
        return nil, nil
    }
    return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
    if value == nil {
        *s = nil
        return nil
    }
    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("failed to scan StringArray")
    }
    return json.Unmarshal(bytes, s)
}

// 自定义枚举类型
type Status int

const (
    StatusPending Status = iota
    StatusActive
    StatusInactive
)

func (s Status) Value() (driver.Value, error) {
    return int64(s), nil
}

func (s *Status) Scan(value interface{}) error {
    if value == nil {
        *s = StatusPending
        return nil
    }
    switch v := value.(type) {
    case int64:
        *s = Status(v)
    case []byte:
        *s = Status(v[0])
    default:
        return errors.New("failed to scan Status")
    }
    return nil
}

// 使用
type Post struct {
    ID     uint
    Title  string
    Tags   StringArray `gorm:"type:json"`
    Status Status      `gorm:"type:tinyint;default:0"`
}
```

---

## CRUD 操作

### Create 创建

```go
// 创建单条记录
user := User{Name: "Alice", Email: "alice@example.com"}
result := db.Create(&user)

fmt.Println(user.ID)             // 返回插入的 ID
fmt.Println(result.Error)        // 错误信息
fmt.Println(result.RowsAffected) // 影响行数

// 创建并指定字段
db.Select("Name", "Email").Create(&user)
// INSERT INTO users (name, email) VALUES (...)

// 创建并忽略字段
db.Omit("Age", "CreatedAt").Create(&user)

// 批量创建
users := []User{
    {Name: "Alice"},
    {Name: "Bob"},
    {Name: "Carol"},
}
db.Create(&users)

// 批量创建（分批）
db.CreateInBatches(&users, 100) // 每批 100 条

// Map 创建
db.Model(&User{}).Create(map[string]interface{}{
    "Name":  "Alice",
    "Email": "alice@example.com",
})

// 使用 SQL 表达式
db.Model(&User{}).Create(map[string]interface{}{
    "Name":      "Alice",
    "CreatedAt": gorm.Expr("NOW()"),
})

// Upsert（冲突时更新）
db.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "email"}},
    DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
}).Create(&user)
// INSERT INTO users ... ON DUPLICATE KEY UPDATE name=VALUES(name), updated_at=VALUES(updated_at)
```

### Read 查询

```go
// 查询单条
var user User

// 根据主键查询
db.First(&user, 1)                    // SELECT * FROM users WHERE id = 1 LIMIT 1
db.First(&user, "id = ?", 1)          // 同上
db.First(&user, User{Name: "Alice"})  // SELECT * FROM users WHERE name = 'Alice' LIMIT 1

// 获取一条记录（不指定排序）
db.Take(&user)

// 获取最后一条
db.Last(&user)

// 查询多条
var users []User
db.Find(&users)                           // SELECT * FROM users
db.Find(&users, []int{1, 2, 3})           // SELECT * FROM users WHERE id IN (1, 2, 3)
db.Find(&users, "name LIKE ?", "%Alice%") // SELECT * FROM users WHERE name LIKE '%Alice%'

// 查询到 map
var result map[string]interface{}
db.Model(&User{}).First(&result, 1)

var results []map[string]interface{}
db.Model(&User{}).Find(&results)

// 查询到指定结构
type UserDTO struct {
    Name  string
    Email string
}
var dto UserDTO
db.Model(&User{}).First(&dto, 1)

// FirstOrInit（未找到则初始化）
db.FirstOrInit(&user, User{Name: "Alice"})
db.Where(User{Name: "Alice"}).Attrs(User{Age: 20}).FirstOrInit(&user)

// FirstOrCreate（未找到则创建）
db.FirstOrCreate(&user, User{Name: "Alice"})
db.Where(User{Name: "Alice"}).Attrs(User{Age: 20}).FirstOrCreate(&user)

// 检查记录是否存在
var exists bool
db.Model(&User{}).Select("1").Where("id = ?", 1).Find(&exists)

// 使用 Exists
err := db.Model(&User{}).Where("id = ?", 1).First(&User{}).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
    // 记录不存在
}
```

### Update 更新

```go
// 更新单个字段
db.Model(&user).Update("Name", "NewName")
// UPDATE users SET name = 'NewName', updated_at = '...' WHERE id = 1

// 更新多个字段（结构体，只更新非零值）
db.Model(&user).Updates(User{Name: "Alice", Age: 0}) // Age 不会更新
// UPDATE users SET name = 'Alice', updated_at = '...' WHERE id = 1

// 更新多个字段（map，包括零值）
db.Model(&user).Updates(map[string]interface{}{
    "Name": "Alice",
    "Age":  0, // 会更新为 0
})

// 使用 Select 指定更新字段
db.Model(&user).Select("Name", "Age").Updates(User{Name: "Alice", Age: 0})

// 使用 Omit 排除字段
db.Model(&user).Omit("Name").Updates(User{Name: "Alice", Age: 18})

// 批量更新
db.Model(&User{}).Where("status = ?", "inactive").Update("status", "active")
db.Model(&User{}).Where("1 = 1").Updates(map[string]interface{}{"status": "active"})

// 使用表达式
db.Model(&product).Update("price", gorm.Expr("price * ? + ?", 1.1, 10))
// UPDATE products SET price = price * 1.1 + 10 WHERE id = 1

db.Model(&product).Updates(map[string]interface{}{
    "price": gorm.Expr("price + ?", 10),
    "stock": gorm.Expr("stock - ?", 1),
})

// 子查询更新
db.Model(&user).Update("company_name", db.Model(&Company{}).Select("name").Where("companies.id = users.company_id"))

// Save（保存所有字段）
user.Name = "NewName"
user.Age = 0
db.Save(&user) // 更新所有字段，包括零值
```

### Delete 删除

```go
// 根据主键删除
db.Delete(&User{}, 1)
db.Delete(&User{}, []int{1, 2, 3})

// 根据条件删除
db.Delete(&User{}, "name = ?", "Alice")
db.Where("name = ?", "Alice").Delete(&User{})

// 软删除（需要 DeletedAt 字段）
// DELETE 会变成 UPDATE SET deleted_at = NOW()
db.Delete(&user)

// 查询包含软删除的记录
db.Unscoped().Where("name = ?", "Alice").Find(&users)

// 永久删除
db.Unscoped().Delete(&user)

// 批量删除
db.Where("status = ?", "inactive").Delete(&User{})

// 阻止全表删除
db.Delete(&User{}) // 会报错：WHERE conditions required
db.Where("1 = 1").Delete(&User{}) // 显式删除全表
```

---

## GORM 查询语法与占位符

### 占位符 `?` 详解

GORM 使用 `?` 作为参数占位符，支持多种查询方式：

```go
// ============ 基本占位符 ============

// 单个 ?
db.Where("name = ?", "Alice").Find(&users)
// SQL: SELECT * FROM users WHERE name = 'Alice'

// 多个 ?（按顺序匹配）
db.Where("name = ? AND age = ?", "Alice", 18).Find(&users)
// SQL: SELECT * FROM users WHERE name = 'Alice' AND age = 18

// IN 查询（? 自动展开切片）
db.Where("id IN ?", []int{1, 2, 3}).Find(&users)
// SQL: SELECT * FROM users WHERE id IN (1, 2, 3)

// BETWEEN
db.Where("age BETWEEN ? AND ?", 18, 30).Find(&users)
// SQL: SELECT * FROM users WHERE age BETWEEN 18 AND 30

// LIKE
db.Where("name LIKE ?", "%Alice%").Find(&users)
// SQL: SELECT * FROM users WHERE name LIKE '%Alice%'

// IS NULL（不需要占位符）
db.Where("deleted_at IS NULL").Find(&users)

// ============ 避免 SQL 注入 ============

// ✅ 安全：使用占位符
name := userInput
db.Where("name = ?", name).Find(&users)

// ❌ 危险：字符串拼接
db.Where("name = '" + name + "'").Find(&users) // SQL 注入风险！

// ❌ 危险：fmt.Sprintf
db.Where(fmt.Sprintf("name = '%s'", name)).Find(&users) // SQL 注入风险！
```

### 命名参数

```go
// 使用 @name 命名参数
db.Where("name = @name AND age = @age", sql.Named("name", "Alice"), sql.Named("age", 18)).Find(&users)

// 使用 map 命名参数
db.Where("name = @name AND age = @age", map[string]interface{}{
    "name": "Alice",
    "age":  18,
}).Find(&users)

// 在 Raw SQL 中使用
db.Raw("SELECT * FROM users WHERE name = @name", sql.Named("name", "Alice")).Scan(&users)

db.Raw("SELECT * FROM users WHERE name = @name", map[string]interface{}{
    "name": "Alice",
}).Scan(&users)
```

### 结构体与 Map 条件

```go
// ============ 结构体条件 ============

// 结构体作为条件（只使用非零值字段）
db.Where(&User{Name: "Alice", Age: 18}).Find(&users)
// SQL: SELECT * FROM users WHERE name = 'Alice' AND age = 18

// ⚠️ 零值字段会被忽略
db.Where(&User{Name: "Alice", Age: 0}).Find(&users)
// SQL: SELECT * FROM users WHERE name = 'Alice'
// Age = 0 被忽略！

// 指定要查询的字段（包括零值）
db.Where(&User{Name: "Alice", Age: 0}, "Name", "Age").Find(&users)
// SQL: SELECT * FROM users WHERE name = 'Alice' AND age = 0

// 指定字段的另一种方式
db.Where(&User{Name: "Alice"}, "Age").Find(&users)
// SQL: SELECT * FROM users WHERE age = 0 (只使用 Age 字段)

// ============ Map 条件 ============

// Map 条件（包括零值）
db.Where(map[string]interface{}{
    "name": "Alice",
    "age":  0,  // 会作为条件
}).Find(&users)
// SQL: SELECT * FROM users WHERE name = 'Alice' AND age = 0

// Map 与切片
db.Where(map[string]interface{}{
    "name": []string{"Alice", "Bob"},  // 自动转为 IN
}).Find(&users)
// SQL: SELECT * FROM users WHERE name IN ('Alice', 'Bob')
```

### 内联条件

```go
// Find 的内联条件
db.Find(&users, "name = ?", "Alice")
// 等价于
db.Where("name = ?", "Alice").Find(&users)

// First 的内联条件
db.First(&user, "name = ?", "Alice")

// 主键查询
db.First(&user, 1)                    // SELECT * FROM users WHERE id = 1
db.First(&user, "id = ?", 1)          // 同上
db.Find(&users, []int{1, 2, 3})       // SELECT * FROM users WHERE id IN (1, 2, 3)

// 结构体内联条件
db.Find(&users, User{Name: "Alice"})
// SELECT * FROM users WHERE name = 'Alice'

// Map 内联条件
db.Find(&users, map[string]interface{}{"name": "Alice", "age": 18})
```

### 子查询语法

```go
// ============ WHERE 子查询 ============

// IN 子查询
subQuery := db.Model(&Order{}).Select("user_id").Where("amount > ?", 1000)
db.Where("id IN (?)", subQuery).Find(&users)
// SQL: SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE amount > 1000)

// EXISTS 子查询
db.Where("EXISTS (?)", db.Model(&Order{}).Select("1").Where("orders.user_id = users.id")).Find(&users)
// SQL: SELECT * FROM users WHERE EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)

// 比较子查询
db.Where("age > (?)", db.Model(&User{}).Select("AVG(age)")).Find(&users)
// SQL: SELECT * FROM users WHERE age > (SELECT AVG(age) FROM users)

// ============ FROM 子查询 ============

// 子查询作为表
subQuery := db.Model(&Order{}).Select("user_id, SUM(amount) as total").Group("user_id")
db.Table("(?) as order_totals", subQuery).Where("total > ?", 1000).Find(&results)
// SQL: SELECT * FROM (SELECT user_id, SUM(amount) as total FROM orders GROUP BY user_id) as order_totals WHERE total > 1000

// ============ SELECT 子查询 ============

// 子查询作为字段
db.Model(&User{}).Select(
    "users.*",
    "(?) as order_count", db.Model(&Order{}).Select("COUNT(*)").Where("orders.user_id = users.id"),
    "(?) as total_amount", db.Model(&Order{}).Select("COALESCE(SUM(amount), 0)").Where("orders.user_id = users.id"),
).Find(&results)
```

### 表达式与函数

```go
// ============ gorm.Expr 表达式 ============

// 更新时使用表达式
db.Model(&product).Update("price", gorm.Expr("price * ?", 1.1))
// SQL: UPDATE products SET price = price * 1.1 WHERE id = ?

db.Model(&product).Update("stock", gorm.Expr("stock - ?", 1))
// SQL: UPDATE products SET stock = stock - 1 WHERE id = ?

// 多字段表达式更新
db.Model(&product).Updates(map[string]interface{}{
    "price": gorm.Expr("price * ? + ?", 1.1, 10),
    "stock": gorm.Expr("stock - ?", 1),
})

// 查询中使用表达式
db.Where("amount > ?", gorm.Expr("(SELECT AVG(amount) FROM orders)")).Find(&orders)

// 创建时使用表达式
db.Model(&User{}).Create(map[string]interface{}{
    "name":       "Alice",
    "created_at": gorm.Expr("NOW()"),
    "uuid":       gorm.Expr("UUID()"),
})

// ============ 数据库函数 ============

// 日期函数
db.Where("DATE(created_at) = ?", "2024-01-01").Find(&orders)
db.Where("YEAR(created_at) = ? AND MONTH(created_at) = ?", 2024, 1).Find(&orders)

// 字符串函数
db.Where("LOWER(name) = ?", "alice").Find(&users)
db.Where("LENGTH(name) > ?", 5).Find(&users)
db.Where("name LIKE CONCAT('%', ?, '%')", keyword).Find(&users)

// 数学函数
db.Where("ROUND(price, 2) = ?", 99.99).Find(&products)

// 聚合函数（需要配合 Select）
db.Model(&Order{}).Select("COUNT(*) as count, SUM(amount) as total").Scan(&result)
```

### Clause 语法

```go
import "gorm.io/gorm/clause"

// ============ ON CONFLICT（Upsert）============

// MySQL: ON DUPLICATE KEY UPDATE
db.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "email"}},
    DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
}).Create(&user)
// SQL: INSERT INTO users ... ON DUPLICATE KEY UPDATE name=VALUES(name), updated_at=VALUES(updated_at)

// 全部更新
db.Clauses(clause.OnConflict{
    UpdateAll: true,
}).Create(&user)

// 什么都不做
db.Clauses(clause.OnConflict{
    DoNothing: true,
}).Create(&user)
// SQL: INSERT INTO users ... ON DUPLICATE KEY UPDATE id=id

// 条件更新
db.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "email"}},
    DoUpdates: clause.Assignments(map[string]interface{}{
        "name":       gorm.Expr("VALUES(name)"),
        "updated_at": gorm.Expr("NOW()"),
    }),
}).Create(&user)

// ============ RETURNING ============

// PostgreSQL: RETURNING
db.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "created_at"}}}).Create(&user)
// SQL: INSERT INTO users ... RETURNING id, created_at

// 返回所有字段
db.Clauses(clause.Returning{}).Create(&user)
// SQL: INSERT INTO users ... RETURNING *

// ============ Locking ============

// FOR UPDATE
db.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&users)
// SQL: SELECT * FROM users FOR UPDATE

// FOR SHARE
db.Clauses(clause.Locking{Strength: "SHARE"}).Find(&users)
// SQL: SELECT * FROM users FOR SHARE

// NOWAIT
db.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).Find(&users)
// SQL: SELECT * FROM users FOR UPDATE NOWAIT

// SKIP LOCKED
db.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Find(&users)
// SQL: SELECT * FROM users FOR UPDATE SKIP LOCKED

// 指定表
db.Clauses(clause.Locking{
    Strength: "UPDATE",
    Table:    clause.Table{Name: clause.CurrentTable},
}).Joins("Profile").Find(&users)
// SQL: SELECT * FROM users ... FOR UPDATE OF users

// ============ Hints ============

// 索引提示
db.Clauses(hints.UseIndex("idx_name")).Find(&users)
// SQL: SELECT * FROM users USE INDEX (idx_name)

db.Clauses(hints.ForceIndex("idx_name")).Find(&users)
// SQL: SELECT * FROM users FORCE INDEX (idx_name)

// 优化器提示
db.Clauses(hints.New("MAX_EXECUTION_TIME(1000)")).Find(&users)
// SQL: SELECT /*+ MAX_EXECUTION_TIME(1000) */ * FROM users
```

### 条件组合

```go
// ============ Where / Or / Not 组合 ============

// AND 组合
db.Where("name = ?", "Alice").Where("age > ?", 18).Find(&users)
// SQL: SELECT * FROM users WHERE name = 'Alice' AND age > 18

// OR 组合
db.Where("name = ?", "Alice").Or("name = ?", "Bob").Find(&users)
// SQL: SELECT * FROM users WHERE name = 'Alice' OR name = 'Bob'

// NOT 条件
db.Not("name = ?", "Alice").Find(&users)
// SQL: SELECT * FROM users WHERE NOT name = 'Alice'

db.Not(User{Name: "Alice"}).Find(&users)
// SQL: SELECT * FROM users WHERE name != 'Alice'

db.Not(map[string]interface{}{"name": []string{"Alice", "Bob"}}).Find(&users)
// SQL: SELECT * FROM users WHERE name NOT IN ('Alice', 'Bob')

// ============ 复杂条件组合 ============

// 使用 Group 组合条件
db.Where(
    db.Where("name = ?", "Alice").Or("name = ?", "Bob"),
).Where("age > ?", 18).Find(&users)
// SQL: SELECT * FROM users WHERE (name = 'Alice' OR name = 'Bob') AND age > 18

// 使用 Session 构建子条件
condition1 := db.Session(&gorm.Session{NewDB: true}).Where("name = ?", "Alice").Or("name = ?", "Bob")
condition2 := db.Session(&gorm.Session{NewDB: true}).Where("status = ?", "active")

db.Where(condition1).Where(condition2).Find(&users)
// SQL: SELECT * FROM users WHERE (name = 'Alice' OR name = 'Bob') AND status = 'active'

// ============ 动态条件构建 ============

func BuildQuery(db *gorm.DB, filters map[string]interface{}) *gorm.DB {
    if name, ok := filters["name"]; ok && name != "" {
        db = db.Where("name LIKE ?", "%"+name.(string)+"%")
    }
    if status, ok := filters["status"]; ok && status != "" {
        db = db.Where("status = ?", status)
    }
    if minAge, ok := filters["min_age"]; ok {
        db = db.Where("age >= ?", minAge)
    }
    if maxAge, ok := filters["max_age"]; ok {
        db = db.Where("age <= ?", maxAge)
    }
    if ids, ok := filters["ids"]; ok {
        db = db.Where("id IN ?", ids)
    }
    return db
}

// 使用
query := BuildQuery(db.Model(&User{}), map[string]interface{}{
    "name":    "Alice",
    "status":  "active",
    "min_age": 18,
})
query.Find(&users)
```

### 特殊查询语法

```go
// ============ DISTINCT ============
db.Distinct("name", "age").Find(&users)
// SQL: SELECT DISTINCT name, age FROM users

db.Model(&User{}).Distinct().Count(&count)
// SQL: SELECT COUNT(DISTINCT id) FROM users

// ============ GROUP BY / HAVING ============
db.Model(&Order{}).
    Select("user_id, SUM(amount) as total, COUNT(*) as count").
    Group("user_id").
    Having("total > ? AND count > ?", 1000, 5).
    Find(&results)
// SQL: SELECT user_id, SUM(amount) as total, COUNT(*) as count 
//      FROM orders GROUP BY user_id HAVING total > 1000 AND count > 5

// ============ ORDER BY ============
db.Order("created_at DESC, name ASC").Find(&users)
db.Order("FIELD(status, 'pending', 'active', 'inactive')").Find(&users)
db.Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).Find(&users)

// ============ LIMIT / OFFSET ============
db.Limit(10).Offset(20).Find(&users)
// SQL: SELECT * FROM users LIMIT 10 OFFSET 20

// 取消 Limit
db.Limit(10).Find(&users1).Limit(-1).Find(&users2)

// ============ 联合查询 UNION ============
db.Raw("? UNION ?",
    db.Model(&User{}).Select("name", "email").Where("status = ?", "active"),
    db.Model(&Admin{}).Select("name", "email").Where("status = ?", "active"),
).Scan(&results)

// ============ Pluck 获取单列 ============
var names []string
db.Model(&User{}).Pluck("name", &names)
// SQL: SELECT name FROM users

var ids []int
db.Model(&User{}).Where("status = ?", "active").Pluck("id", &ids)

// ============ Scan 到指定结构 ============
type Result struct {
    Name  string
    Total int64
}
var results []Result
db.Model(&Order{}).Select("users.name, SUM(orders.amount) as total").
    Joins("LEFT JOIN users ON users.id = orders.user_id").
    Group("users.name").
    Scan(&results)

// ============ 智能选择字段 ============
type APIUser struct {
    ID   uint
    Name string
}
// 只查询结构体中的字段
db.Model(&User{}).Find(&[]APIUser{})
// SQL: SELECT id, name FROM users
```

### 常见查询模式

```go
// ============ 分页查询 ============
type Pagination struct {
    Page     int
    PageSize int
    Total    int64
    Data     interface{}
}

func Paginate(db *gorm.DB, page, pageSize int, dest interface{}) (*Pagination, error) {
    var total int64
    
    // 计算总数
    if err := db.Count(&total).Error; err != nil {
        return nil, err
    }
    
    // 分页查询
    offset := (page - 1) * pageSize
    if err := db.Offset(offset).Limit(pageSize).Find(dest).Error; err != nil {
        return nil, err
    }
    
    return &Pagination{
        Page:     page,
        PageSize: pageSize,
        Total:    total,
        Data:     dest,
    }, nil
}

// ============ 存在性检查 ============
func Exists(db *gorm.DB, model interface{}, query interface{}, args ...interface{}) (bool, error) {
    var count int64
    err := db.Model(model).Where(query, args...).Limit(1).Count(&count).Error
    return count > 0, err
}

// 使用
exists, _ := Exists(db, &User{}, "email = ?", "alice@example.com")

// ============ 获取或创建 ============
// FirstOrCreate
var user User
db.Where(User{Email: "alice@example.com"}).
    Attrs(User{Name: "Alice", Age: 18}).  // 创建时使用
    FirstOrCreate(&user)

// FirstOrInit（只初始化不创建）
db.Where(User{Email: "alice@example.com"}).
    Attrs(User{Name: "Alice"}).
    Assign(User{Age: 20}).  // 无论找到与否都赋值
    FirstOrInit(&user)

// ============ 批量更新并返回 ============
var updatedUsers []User
db.Model(&User{}).
    Where("status = ?", "inactive").
    Update("status", "active").
    Find(&updatedUsers, "status = ?", "active")

// ============ 安全删除检查 ============
func SafeDelete(db *gorm.DB, model interface{}, id uint) error {
    // 先检查是否存在
    result := db.First(model, id)
    if result.Error != nil {
        return result.Error
    }
    
    // 执行删除
    return db.Delete(model, id).Error
}
```

---

## 查询详解

### 条件查询

```go
// Where
db.Where("name = ?", "Alice").Find(&users)
db.Where("name = ? AND age >= ?", "Alice", 18).Find(&users)
db.Where("name IN ?", []string{"Alice", "Bob"}).Find(&users)
db.Where("name LIKE ?", "%Alice%").Find(&users)
db.Where("created_at BETWEEN ? AND ?", startTime, endTime).Find(&users)
db.Where("name = ? OR email = ?", "Alice", "alice@example.com").Find(&users)

// 结构体条件（只使用非零值）
db.Where(&User{Name: "Alice", Age: 18}).Find(&users)
// SELECT * FROM users WHERE name = 'Alice' AND age = 18

// 指定结构体查询字段
db.Where(&User{Name: "Alice"}, "Name", "Age").Find(&users)
// 即使 Age 是零值也会作为条件

// Map 条件
db.Where(map[string]interface{}{"name": "Alice", "age": 18}).Find(&users)

// Not
db.Not("name = ?", "Alice").Find(&users)
db.Not(User{Name: "Alice"}).Find(&users)
db.Not(map[string]interface{}{"name": "Alice"}).Find(&users)

// Or
db.Where("name = ?", "Alice").Or("name = ?", "Bob").Find(&users)
db.Where("name = ?", "Alice").Or(User{Name: "Bob"}).Find(&users)

// 内联条件
db.Find(&users, "name = ?", "Alice")
db.Find(&users, User{Name: "Alice"})
```

### 选择字段

```go
// Select
db.Select("name", "email").Find(&users)
db.Select([]string{"name", "email"}).Find(&users)
db.Select("name, email").Find(&users)

// 表达式
db.Select("name", "COALESCE(age, 0) AS age").Find(&users)

// Distinct
db.Distinct("name").Find(&users)
db.Distinct("name", "age").Find(&users)

// 排除字段
db.Omit("password", "salt").Find(&users)
```

### 排序与分页

```go
// Order
db.Order("created_at DESC").Find(&users)
db.Order("created_at DESC, name ASC").Find(&users)
db.Order("FIELD(status, 'pending', 'active', 'inactive')").Find(&users) // 自定义排序

// 多次 Order
db.Order("created_at DESC").Order("name ASC").Find(&users)

// Limit & Offset
db.Limit(10).Find(&users)
db.Limit(10).Offset(20).Find(&users)

// 取消 Limit
db.Limit(10).Find(&users1).Limit(-1).Find(&users2) // users2 无 limit

// 分页封装
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        if page <= 0 {
            page = 1
        }
        if pageSize <= 0 {
            pageSize = 10
        }
        offset := (page - 1) * pageSize
        return db.Offset(offset).Limit(pageSize)
    }
}

// 使用
db.Scopes(Paginate(1, 10)).Find(&users)
```

### 分组与聚合

```go
// Group & Having
type Result struct {
    Country string
    Count   int64
}

var results []Result
db.Model(&User{}).Select("country, COUNT(*) as count").
    Group("country").
    Having("count > ?", 10).
    Find(&results)

// 多列分组
db.Model(&Order{}).
    Select("YEAR(created_at) as year, MONTH(created_at) as month, SUM(amount) as total").
    Group("YEAR(created_at), MONTH(created_at)").
    Find(&results)

// 聚合函数
var count int64
db.Model(&User{}).Where("status = ?", "active").Count(&count)

var sum float64
db.Model(&Order{}).Select("SUM(amount)").Where("user_id = ?", 1).Scan(&sum)

type Stats struct {
    Count int64
    Sum   float64
    Avg   float64
    Max   float64
    Min   float64
}
var stats Stats
db.Model(&Order{}).Select("COUNT(*) as count, SUM(amount) as sum, AVG(amount) as avg, MAX(amount) as max, MIN(amount) as min").Scan(&stats)
```

### 锁

```go
// 悲观锁 - FOR UPDATE
db.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&users)
// SELECT * FROM users FOR UPDATE

// 共享锁 - FOR SHARE
db.Clauses(clause.Locking{Strength: "SHARE"}).Find(&users)
// SELECT * FROM users FOR SHARE

// 跳过锁定的行
db.Clauses(clause.Locking{
    Strength: "UPDATE",
    Options:  "NOWAIT",
}).Find(&users)
// SELECT * FROM users FOR UPDATE NOWAIT

db.Clauses(clause.Locking{
    Strength: "UPDATE",
    Options:  "SKIP LOCKED",
}).Find(&users)
// SELECT * FROM users FOR UPDATE SKIP LOCKED
```

### 子查询

```go
// 子查询作为条件
subQuery := db.Model(&Order{}).Select("user_id").Where("amount > ?", 1000)
db.Where("id IN (?)", subQuery).Find(&users)
// SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE amount > 1000)

// 子查询作为表
db.Table("(?) as u", db.Model(&User{}).Select("name", "age")).Where("age > ?", 18).Find(&results)

// 子查询作为字段
db.Model(&User{}).Select(
    "users.*",
    "(?) as order_count",
    db.Model(&Order{}).Select("COUNT(*)").Where("orders.user_id = users.id"),
).Find(&results)

// FROM 子查询
subQuery := db.Model(&Order{}).Select("user_id, SUM(amount) as total").Group("user_id")
db.Table("(?) as order_totals", subQuery).Where("total > ?", 1000).Find(&results)
```

### 原生 SQL

```go
// 原生查询
db.Raw("SELECT * FROM users WHERE name = ?", "Alice").Scan(&users)

// 原生执行
db.Exec("UPDATE users SET name = ? WHERE id = ?", "Alice", 1)

// 命名参数
db.Raw("SELECT * FROM users WHERE name = @name", sql.Named("name", "Alice")).Scan(&users)
db.Raw("SELECT * FROM users WHERE name = @name", map[string]interface{}{"name": "Alice"}).Scan(&users)

// 行迭代
rows, _ := db.Model(&User{}).Where("status = ?", "active").Rows()
defer rows.Close()

for rows.Next() {
    var user User
    db.ScanRows(rows, &user)
    // 处理 user
}

// 获取 SQL
stmt := db.Session(&gorm.Session{DryRun: true}).First(&user, 1).Statement
sql := stmt.SQL.String()
vars := stmt.Vars
```

---

## 关联关系

### 一对一 (Has One)

```go
// User 有一个 Profile
type User struct {
    ID      uint
    Name    string
    Profile Profile // Has One
}

type Profile struct {
    ID     uint
    UserID uint   // 外键
    Bio    string
    Avatar string
}

// 查询预加载
db.Preload("Profile").Find(&users)

// 关联创建
user := User{
    Name: "Alice",
    Profile: Profile{Bio: "Developer"},
}
db.Create(&user)

// 关联查询
var profile Profile
db.Model(&user).Association("Profile").Find(&profile)

// 关联更新
db.Model(&user).Association("Profile").Replace(&Profile{Bio: "New Bio"})

// 删除关联
db.Model(&user).Association("Profile").Delete(&profile)

// 清空关联
db.Model(&user).Association("Profile").Clear()

// 关联计数
count := db.Model(&user).Association("Profile").Count()
```

### 一对多 (Has Many)

```go
// User 有多个 Order
type User struct {
    ID     uint
    Name   string
    Orders []Order // Has Many
}

type Order struct {
    ID     uint
    UserID uint // 外键
    Amount float64
}

// 预加载
db.Preload("Orders").Find(&users)

// 带条件的预加载
db.Preload("Orders", "amount > ?", 100).Find(&users)

// 预加载排序
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC").Limit(5)
}).Find(&users)

// 嵌套预加载
db.Preload("Orders.Items").Find(&users)

// 关联添加
db.Model(&user).Association("Orders").Append(&Order{Amount: 100})

// 关联替换
db.Model(&user).Association("Orders").Replace(&[]Order{{Amount: 100}, {Amount: 200}})
```

### 多对多 (Many To Many)

```go
// User 和 Role 多对多
type User struct {
    ID    uint
    Name  string
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    ID    uint
    Name  string
    Users []User `gorm:"many2many:user_roles;"`
}

// 自定义连接表
type User struct {
    ID    uint
    Name  string
    Roles []Role `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID"`
}

// 带额外字段的连接表
type UserRole struct {
    UserID    uint `gorm:"primaryKey"`
    RoleID    uint `gorm:"primaryKey"`
    CreatedAt time.Time
    ExpiredAt *time.Time
}

// 预加载
db.Preload("Roles").Find(&users)

// 关联操作
db.Model(&user).Association("Roles").Append(&role)
db.Model(&user).Association("Roles").Delete(&role)
db.Model(&user).Association("Roles").Replace(&[]Role{role1, role2})
db.Model(&user).Association("Roles").Clear()
```

### 属于 (Belongs To)

```go
// Order 属于 User
type Order struct {
    ID     uint
    UserID uint // 外键
    User   User // Belongs To
    Amount float64
}

type User struct {
    ID   uint
    Name string
}

// 预加载
db.Preload("User").Find(&orders)

// Joins 预加载（单条 SQL）
db.Joins("User").Find(&orders)

// 带条件的 Joins
db.Joins("User", db.Where(&User{Name: "Alice"})).Find(&orders)
```

### 多态关联

```go
// 多态关联
type Comment struct {
    ID            uint
    Content       string
    CommentableID uint
    CommentableType string
}

type Post struct {
    ID       uint
    Title    string
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}

type Video struct {
    ID       uint
    Title    string
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}

// 查询
db.Preload("Comments").Find(&posts)
```

### 关联模式

```go
// 关联模式
assoc := db.Model(&user).Association("Orders")

// 查找
var orders []Order
assoc.Find(&orders)

// 添加
assoc.Append(&Order{Amount: 100})

// 替换
assoc.Replace(&newOrders)

// 删除
assoc.Delete(&order)

// 清空
assoc.Clear()

// 计数
count := assoc.Count()

// 带条件查找
assoc.Find(&orders, "amount > ?", 100)
```

---

## 事务处理

### 自动事务

```go
// 使用 Transaction 方法
err := db.Transaction(func(tx *gorm.DB) error {
    // 在事务中执行操作
    if err := tx.Create(&user).Error; err != nil {
        return err // 返回错误会回滚事务
    }
    
    if err := tx.Create(&order).Error; err != nil {
        return err
    }
    
    return nil // 返回 nil 提交事务
})

// 嵌套事务
db.Transaction(func(tx *gorm.DB) error {
    tx.Create(&user1)
    
    tx.Transaction(func(tx2 *gorm.DB) error {
        tx2.Create(&user2)
        return errors.New("rollback user2") // 只回滚 user2
    })
    
    return nil // 提交 user1
})
```

### 手动事务

```go
// 开始事务
tx := db.Begin()

// 检查错误
if tx.Error != nil {
    return tx.Error
}

// 使用 defer 确保事务结束
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// 执行操作
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&order).Error; err != nil {
    tx.Rollback()
    return err
}

// 提交事务
return tx.Commit().Error
```

### SavePoint

```go
tx := db.Begin()

tx.Create(&user1)

// 创建保存点
tx.SavePoint("sp1")

tx.Create(&user2)

// 回滚到保存点
tx.RollbackTo("sp1") // user2 被回滚，user1 保留

tx.Commit() // 只提交 user1
```

### 事务选项

```go
// 设置隔离级别
tx := db.Begin(&sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
})

// 禁用默认事务（全局）
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    SkipDefaultTransaction: true,
})

// 临时启用事务
db.Session(&gorm.Session{SkipDefaultTransaction: false}).Create(&user)
```

---

## 钩子函数

### 钩子类型

```go
type User struct {
    ID       uint
    Name     string
    Password string
}

// 创建钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 密码加密
    if u.Password != "" {
        hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
        if err != nil {
            return err
        }
        u.Password = string(hashed)
    }
    return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
    // 发送欢迎邮件
    go sendWelcomeEmail(u.Email)
    return nil
}

// 查询钩子
func (u *User) AfterFind(tx *gorm.DB) error {
    // 查询后处理
    return nil
}

// 更新钩子
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    // 验证逻辑
    return nil
}

func (u *User) AfterUpdate(tx *gorm.DB) error {
    // 清除缓存
    return nil
}

// 删除钩子
func (u *User) BeforeDelete(tx *gorm.DB) error {
    // 检查是否可以删除
    return nil
}

func (u *User) AfterDelete(tx *gorm.DB) error {
    // 清理关联数据
    return nil
}

// 保存钩子（Create 和 Update 都会触发）
func (u *User) BeforeSave(tx *gorm.DB) error {
    return nil
}

func (u *User) AfterSave(tx *gorm.DB) error {
    return nil
}
```

### 钩子执行顺序

```
创建: BeforeSave -> BeforeCreate -> 插入数据 -> AfterCreate -> AfterSave
更新: BeforeSave -> BeforeUpdate -> 更新数据 -> AfterUpdate -> AfterSave
删除: BeforeDelete -> 删除数据 -> AfterDelete
查询: 查询数据 -> AfterFind
```

### 跳过钩子

```go
// 跳过所有钩子
db.Session(&gorm.Session{SkipHooks: true}).Create(&user)

// 使用 UpdateColumn/UpdateColumns 跳过钩子和时间追踪
db.Model(&user).UpdateColumn("name", "Alice")
db.Model(&user).UpdateColumns(map[string]interface{}{"name": "Alice", "age": 18})
```

---

## 高级特性

### Scopes（作用域）

```go
// 定义 Scope
func ActiveUsers(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "active")
}

func AgeGreaterThan(age int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("age > ?", age)
    }
}

func OrderByCreatedAt(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC")
}

// 分页 Scope
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        offset := (page - 1) * pageSize
        return db.Offset(offset).Limit(pageSize)
    }
}

// 使用 Scope
db.Scopes(ActiveUsers).Find(&users)
db.Scopes(ActiveUsers, AgeGreaterThan(18)).Find(&users)
db.Scopes(ActiveUsers, OrderByCreatedAt, Paginate(1, 10)).Find(&users)

// 动态 Scope
func FilterByStatus(status string) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        if status != "" {
            return db.Where("status = ?", status)
        }
        return db
    }
}
```

### 批处理

```go
// FindInBatches 批量查询
var users []User
db.Where("status = ?", "active").FindInBatches(&users, 100, func(tx *gorm.DB, batch int) error {
    for _, user := range users {
        // 处理每个用户
    }
    // 返回 error 停止处理
    return nil
})

// 批量创建
db.CreateInBatches(&users, 100)

// 使用 Rows 迭代
rows, _ := db.Model(&User{}).Where("status = ?", "active").Rows()
defer rows.Close()

for rows.Next() {
    var user User
    db.ScanRows(rows, &user)
    // 处理
}
```

### 数据库迁移

```go
// 自动迁移
db.AutoMigrate(&User{}, &Product{}, &Order{})

// 迁移器
m := db.Migrator()

// 表操作
m.CreateTable(&User{})
m.HasTable(&User{})
m.DropTable(&User{})
m.RenameTable(&User{}, &UserNew{})

// 列操作
m.AddColumn(&User{}, "Age")
m.DropColumn(&User{}, "Age")
m.AlterColumn(&User{}, "Age")
m.HasColumn(&User{}, "Age")
m.RenameColumn(&User{}, "Age", "UserAge")

// 索引操作
m.CreateIndex(&User{}, "Name")
m.CreateIndex(&User{}, "idx_name")
m.DropIndex(&User{}, "idx_name")
m.HasIndex(&User{}, "idx_name")
m.RenameIndex(&User{}, "idx_name", "idx_new_name")

// 约束操作
m.CreateConstraint(&User{}, "fk_users_company")
m.DropConstraint(&User{}, "fk_users_company")
m.HasConstraint(&User{}, "fk_users_company")
```

### 多数据库

```go
// 读写分离
import "gorm.io/plugin/dbresolver"

db.Use(dbresolver.Register(dbresolver.Config{
    Sources:  []gorm.Dialector{mysql.Open(masterDSN)},
    Replicas: []gorm.Dialector{mysql.Open(slaveDSN1), mysql.Open(slaveDSN2)},
    Policy:   dbresolver.RandomPolicy{},
}).Register(dbresolver.Config{
    Sources:  []gorm.Dialector{mysql.Open(orderMasterDSN)},
    Replicas: []gorm.Dialector{mysql.Open(orderSlaveDSN)},
}, &Order{}, &OrderItem{}))

// 手动切换
// 使用写库
db.Clauses(dbresolver.Write).First(&user)

// 使用读库
db.Clauses(dbresolver.Read).First(&user)

// 多数据库
db1, _ := gorm.Open(mysql.Open(dsn1))
db2, _ := gorm.Open(mysql.Open(dsn2))

// 根据条件选择
func GetDB(tenant string) *gorm.DB {
    switch tenant {
    case "tenant1":
        return db1
    case "tenant2":
        return db2
    default:
        return db1
    }
}
```

### Context

```go
// 设置 Context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

db.WithContext(ctx).Find(&users)

// 全局设置
db = db.WithContext(ctx)

// 配合 Session
db.Session(&gorm.Session{Context: ctx}).Find(&users)

// 使用 Context 传递值
type contextKey string

ctx := context.WithValue(context.Background(), contextKey("user_id"), 123)
db.WithContext(ctx).Create(&record)

// 在钩子中获取
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if userID, ok := tx.Statement.Context.Value(contextKey("user_id")).(int); ok {
        u.CreatedBy = userID
    }
    return nil
}
```

---

## 性能优化

### 预编译语句

```go
// 开启预编译
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    PrepareStmt: true,
})

// 关闭预编译
db.Session(&gorm.Session{PrepareStmt: false})
```

### 避免 N+1 查询

```go
// 问题：N+1 查询
var users []User
db.Find(&users)
for _, user := range users {
    var orders []Order
    db.Where("user_id = ?", user.ID).Find(&orders) // N 次查询
}

// 解决：使用 Preload
db.Preload("Orders").Find(&users) // 2 次查询

// 使用 Joins（单次查询）
db.Joins("Profile").Find(&users)

// 预加载优化
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
    return db.Select("id", "user_id", "amount") // 只查需要的字段
}).Find(&users)
```

### 批量操作

```go
// 批量插入
db.CreateInBatches(&users, 100)

// 批量更新
db.Model(&User{}).Where("status = ?", "inactive").Updates(map[string]interface{}{"status": "active"})

// 批量删除
db.Where("status = ?", "deleted").Delete(&User{})
```

### 索引优化

```go
type User struct {
    ID     uint   `gorm:"primaryKey"`
    Name   string `gorm:"index"`
    Email  string `gorm:"uniqueIndex"`
    Status string `gorm:"index:idx_status_created,priority:1"`
    CreatedAt time.Time `gorm:"index:idx_status_created,priority:2"`
}

// 复合索引
type Order struct {
    ID        uint
    UserID    uint `gorm:"index:idx_user_status,priority:1"`
    Status    string `gorm:"index:idx_user_status,priority:2"`
    CreatedAt time.Time `gorm:"index:idx_created"`
}
```

### 连接池调优

```go
sqlDB, _ := db.DB()

// 最大空闲连接数
sqlDB.SetMaxIdleConns(10)

// 最大打开连接数
sqlDB.SetMaxOpenConns(100)

// 连接最大生命周期
sqlDB.SetConnMaxLifetime(time.Hour)

// 空闲连接最大生命周期
sqlDB.SetConnMaxIdleTime(10 * time.Minute)
```

### 日志优化

```go
// 关闭日志
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Silent),
})

// 只记录慢查询
newLogger := logger.New(
    log.New(os.Stdout, "\r\n", log.LstdFlags),
    logger.Config{
        SlowThreshold:             200 * time.Millisecond,
        LogLevel:                  logger.Warn,
        IgnoreRecordNotFoundError: true,
        Colorful:                  true,
    },
)

db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: newLogger,
})
```

---

## 底层原理

### GORM 架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           GORM 架构                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                        用户代码层                                 │   │
│  │  db.Where("name = ?", "Alice").First(&user)                     │   │
│  └───────────────────────────────┬─────────────────────────────────┘   │
│                                  │                                      │
│  ┌───────────────────────────────▼─────────────────────────────────┐   │
│  │                       GORM API 层                                │   │
│  │  Create / Find / Update / Delete / Preload / Joins ...          │   │
│  └───────────────────────────────┬─────────────────────────────────┘   │
│                                  │                                      │
│  ┌───────────────────────────────▼─────────────────────────────────┐   │
│  │                       Chain 层 (链式调用)                         │   │
│  │  Where / Select / Order / Limit / Offset ...                    │   │
│  │  每次调用返回新的 *gorm.DB 实例                                   │   │
│  └───────────────────────────────┬─────────────────────────────────┘   │
│                                  │                                      │
│  ┌───────────────────────────────▼─────────────────────────────────┐   │
│  │                     Statement 层                                 │   │
│  │  存储查询条件、表名、模型信息等                                    │   │
│  └───────────────────────────────┬─────────────────────────────────┘   │
│                                  │                                      │
│  ┌───────────────────────────────▼─────────────────────────────────┐   │
│  │                     Callbacks 层                                 │   │
│  │  BeforeCreate / Create / AfterCreate ...                        │   │
│  │  钩子链式执行                                                     │   │
│  └───────────────────────────────┬─────────────────────────────────┘   │
│                                  │                                      │
│  ┌───────────────────────────────▼─────────────────────────────────┐   │
│  │                    Clause 层 (SQL 构建)                          │   │
│  │  SELECT / FROM / WHERE / ORDER BY / LIMIT ...                   │   │
│  │  构建最终 SQL 语句                                                │   │
│  └───────────────────────────────┬─────────────────────────────────┘   │
│                                  │                                      │
│  ┌───────────────────────────────▼─────────────────────────────────┐   │
│  │                    Dialector 层 (数据库方言)                      │   │
│  │  MySQL / PostgreSQL / SQLite / SQL Server                       │   │
│  └───────────────────────────────┬─────────────────────────────────┘   │
│                                  │                                      │
│  ┌───────────────────────────────▼─────────────────────────────────┐   │
│  │                    database/sql                                  │   │
│  │  Go 标准库数据库接口                                              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 核心结构体

```go
// gorm.DB 核心结构
type DB struct {
    *Config
    Error        error
    RowsAffected int64
    Statement    *Statement
    clone        int
}

// Statement 存储查询状态
type Statement struct {
    *DB
    TableExpr            *clause.Expr
    Table                string
    Model                interface{}
    Unscoped             bool
    Dest                 interface{}
    ReflectValue         reflect.Value
    Clauses              map[string]clause.Clause
    BuildClauses         []string
    Distinct             bool
    Selects              []string
    Omits                []string
    Joins                []join
    Preloads             map[string][]interface{}
    Settings             sync.Map
    ConnPool             ConnPool
    Schema               *schema.Schema
    Context              context.Context
    RaiseErrorOnNotFound bool
    SkipHooks            bool
    SQL                  strings.Builder
    Vars                 []interface{}
    CurDestIndex         int
    attrs                []interface{}
    assigns              []interface{}
    scopes               []func(*DB) *DB
}

// Schema 模型元信息
type Schema struct {
    Name                   string
    ModelType              reflect.Type
    Table                  string
    PrioritizedPrimaryField *Field
    PrimaryFields          []*Field
    PrimaryFieldDBNames    []string
    Fields                 []*Field
    FieldsByName           map[string]*Field
    FieldsByDBName         map[string]*Field
    Relationships          Relationships
    // ...
}
```

### 链式调用原理

```go
// 每次链式调用都会 clone DB
func (db *DB) Where(query interface{}, args ...interface{}) (tx *DB) {
    tx = db.getInstance() // clone
    if conds := tx.Statement.BuildCondition(query, args...); len(conds) > 0 {
        tx.Statement.AddClause(clause.Where{Exprs: conds})
    }
    return
}

// clone 实现
func (db *DB) getInstance() *DB {
    if db.clone > 0 {
        tx := &DB{Config: db.Config, Error: db.Error}
        if db.clone == 1 {
            // clone 整个 Statement
            tx.Statement = &Statement{
                DB:       tx,
                ConnPool: db.Statement.ConnPool,
                Context:  db.Statement.Context,
                Clauses:  map[string]clause.Clause{},
                Vars:     make([]interface{}, 0, 8),
            }
        } else {
            // 共享 Statement（用于 Session）
            tx.Statement = db.Statement
        }
        return tx
    }
    return db
}
```

### Callbacks 机制

```go
// 注册 Callback
func (db *DB) Callback() *callbacks {
    return db.callbacks
}

// Callback 处理器
type callbacks struct {
    processors map[string]*processor
}

// 处理器
type processor struct {
    db        *DB
    Clauses   []string
    fns       []func(*DB)
    callbacks []*callback
}

// 执行 Callbacks
func (p *processor) Execute(db *DB) *DB {
    // 排序 callbacks
    for _, fn := range p.fns {
        fn(db)
        if db.Error != nil && !db.Statement.RaiseErrorOnNotFound {
            break
        }
    }
    return db
}

// 默认 Callbacks
// Create:
//   callbacks.Register("gorm:begin_transaction", BeginTransaction)
//   callbacks.Register("gorm:before_create", BeforeCreate)
//   callbacks.Register("gorm:save_before_associations", SaveBeforeAssociations)
//   callbacks.Register("gorm:create", Create)
//   callbacks.Register("gorm:save_after_associations", SaveAfterAssociations)
//   callbacks.Register("gorm:after_create", AfterCreate)
//   callbacks.Register("gorm:commit_or_rollback_transaction", CommitOrRollbackTransaction)
```

### SQL 构建流程

```go
// Clause 接口
type Clause interface {
    Name() string
    Build(Builder)
}

// 构建 SQL
func (stmt *Statement) Build(clauses ...string) {
    for _, name := range clauses {
        if c, ok := stmt.Clauses[name]; ok {
            c.Build(stmt)
        }
    }
}

// WHERE Clause 示例
type Where struct {
    Exprs []Expression
}

func (where Where) Build(builder Builder) {
    builder.WriteString("WHERE ")
    for idx, expr := range where.Exprs {
        if idx > 0 {
            builder.WriteString(" AND ")
        }
        expr.Build(builder)
    }
}

// Expression 接口
type Expression interface {
    Build(Builder)
}

// Eq 表达式
type Eq struct {
    Column interface{}
    Value  interface{}
}

func (eq Eq) Build(builder Builder) {
    builder.WriteQuoted(eq.Column)
    builder.WriteString(" = ")
    builder.AddVar(builder, eq.Value)
}
```

### Schema 解析

```go
// 解析模型获取 Schema
func Parse(dest interface{}, cacheStore *sync.Map, namer Namer) (*Schema, error) {
    // 获取类型
    modelType := reflect.TypeOf(dest)
    if modelType.Kind() == reflect.Ptr {
        modelType = modelType.Elem()
    }
    
    // 从缓存获取
    if v, ok := cacheStore.Load(modelType); ok {
        return v.(*Schema), nil
    }
    
    // 创建 Schema
    schema := &Schema{
        Name:         modelType.Name(),
        ModelType:    modelType,
        FieldsByName: map[string]*Field{},
        FieldsByDBName: map[string]*Field{},
        Relationships: Relationships{},
    }
    
    // 解析字段
    for i := 0; i < modelType.NumField(); i++ {
        field := modelType.Field(i)
        // 解析 gorm tag
        // 创建 Field
        // 处理关联关系
    }
    
    // 缓存
    cacheStore.Store(modelType, schema)
    
    return schema, nil
}
```

### 连接池管理

```go
// ConnPool 接口
type ConnPool interface {
    PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// GORM 使用 database/sql 的连接池
// 连接池配置直接作用于 sql.DB

func (db *DB) DB() (*sql.DB, error) {
    connPool := db.ConnPool
    if dbConnector, ok := connPool.(interface{ GetDBConnector() *sql.DB }); ok {
        return dbConnector.GetDBConnector(), nil
    }
    if sqldb, ok := connPool.(*sql.DB); ok {
        return sqldb, nil
    }
    return nil, ErrInvalidDB
}
```

### 插件机制

```go
// Plugin 接口
type Plugin interface {
    Name() string
    Initialize(*DB) error
}

// 注册插件
func (db *DB) Use(plugin Plugin) error {
    name := plugin.Name()
    if _, ok := db.Plugins[name]; ok {
        return ErrRegistered
    }
    if err := plugin.Initialize(db); err != nil {
        return err
    }
    db.Plugins[name] = plugin
    return nil
}

// 示例：Prometheus 插件
type Prometheus struct {
    // ...
}

func (p *Prometheus) Name() string {
    return "prometheus"
}

func (p *Prometheus) Initialize(db *DB) error {
    // 注册 callbacks 收集指标
    db.Callback().Create().After("gorm:create").Register("prometheus:after_create", p.after)
    db.Callback().Query().After("gorm:query").Register("prometheus:after_query", p.after)
    return nil
}
```

---

## 总结

### 常用操作速查

| 操作 | 方法 |
|------|------|
| 创建 | `db.Create(&user)` |
| 批量创建 | `db.CreateInBatches(&users, 100)` |
| 查询单条 | `db.First(&user, id)` |
| 查询多条 | `db.Find(&users)` |
| 条件查询 | `db.Where("name = ?", name).Find(&users)` |
| 更新 | `db.Model(&user).Update("name", "new")` |
| 批量更新 | `db.Model(&User{}).Where(...).Updates(...)` |
| 删除 | `db.Delete(&user)` |
| 软删除 | 自动（有 DeletedAt 字段） |
| 预加载 | `db.Preload("Orders").Find(&users)` |
| 事务 | `db.Transaction(func(tx *gorm.DB) error {...})` |
| 原生 SQL | `db.Raw("SELECT ...").Scan(&result)` |

### 最佳实践

1. 使用 `Preload` 避免 N+1 查询
2. 使用 `CreateInBatches` 批量插入
3. 使用 `Scopes` 复用查询逻辑
4. 配置合理的连接池参数
5. 使用 `PrepareStmt` 提高性能
6. 生产环境关闭详细日志
7. 使用 Context 控制超时
8. 正确处理软删除场景

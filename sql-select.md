# SQL SELECT 语法详解

本文档详细介绍 SQL SELECT 语句的完整语法，包括 `SELECT 1`、子查询、窗口函数等高级用法。

## 目录

1. [SELECT 基础语法](#select-基础语法)
2. [SELECT 1 详解](#select-1-详解)
3. [WHERE 子句](#where-子句)
4. [JOIN 连接](#join-连接)
5. [GROUP BY 与聚合](#group-by-与聚合)
6. [ORDER BY 排序](#order-by-排序)
7. [LIMIT 与分页](#limit-与分页)
8. [子查询](#子查询)
9. [窗口函数](#窗口函数)
10. [CTE 公共表表达式](#cte-公共表表达式)
11. [集合操作](#集合操作)
12. [性能优化](#性能优化)

---

## SELECT 基础语法

### 完整语法结构

```sql
SELECT [ALL | DISTINCT | DISTINCTROW]
    [HIGH_PRIORITY]
    [STRAIGHT_JOIN]
    [SQL_SMALL_RESULT | SQL_BIG_RESULT]
    [SQL_BUFFER_RESULT]
    [SQL_CACHE | SQL_NO_CACHE]
    [SQL_CALC_FOUND_ROWS]
    select_expr [, select_expr ...]
    [FROM table_references
      [PARTITION partition_list]
    [WHERE where_condition]
    [GROUP BY {col_name | expr | position}
      [ASC | DESC], ... [WITH ROLLUP]]
    [HAVING where_condition]
    [ORDER BY {col_name | expr | position}
      [ASC | DESC], ...]
    [LIMIT {[offset,] row_count | row_count OFFSET offset}]
    [INTO OUTFILE 'file_name'
      [CHARACTER SET charset_name]
      export_options
      | INTO DUMPFILE 'file_name'
      | INTO var_name [, var_name]]
    [FOR UPDATE | LOCK IN SHARE MODE]
```

### 执行顺序

```
┌─────────────────────────────────────────────────────────────────────┐
│                      SQL 执行顺序                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   书写顺序                          执行顺序                         │
│   ────────                          ────────                        │
│   1. SELECT                         1. FROM                         │
│   2. FROM                           2. ON                           │
│   3. JOIN                           3. JOIN                         │
│   4. ON                             4. WHERE                        │
│   5. WHERE                          5. GROUP BY                     │
│   6. GROUP BY                       6. HAVING                       │
│   7. HAVING                         7. SELECT                       │
│   8. ORDER BY                       8. DISTINCT                     │
│   9. LIMIT                          9. ORDER BY                     │
│                                     10. LIMIT                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 基本查询示例

```sql
-- 查询所有列
SELECT * FROM users;

-- 查询指定列
SELECT id, name, email FROM users;

-- 使用别名
SELECT 
    id AS user_id,
    name AS user_name,
    email AS user_email
FROM users AS u;

-- 去重
SELECT DISTINCT department FROM employees;

-- 计算列
SELECT 
    name,
    price,
    quantity,
    price * quantity AS total
FROM order_items;

-- 条件表达式
SELECT 
    name,
    score,
    CASE 
        WHEN score >= 90 THEN 'A'
        WHEN score >= 80 THEN 'B'
        WHEN score >= 70 THEN 'C'
        WHEN score >= 60 THEN 'D'
        ELSE 'F'
    END AS grade
FROM students;
```

---

## SELECT 1 详解

### 什么是 SELECT 1

`SELECT 1` 是一个返回常量值 `1` 的查询，不访问任何表。

```sql
-- 返回单行单列，值为 1
SELECT 1;

-- 结果：
-- +---+
-- | 1 |
-- +---+
-- | 1 |
-- +---+
```

### SELECT 1 的常见用途

#### 1. 数据库连接测试

```sql
-- 测试数据库连接是否正常
SELECT 1;

-- 或使用数据库特定语法
SELECT 1 FROM DUAL;  -- Oracle
SELECT 1;            -- MySQL, PostgreSQL, SQL Server
```

#### 2. EXISTS 子查询

```sql
-- 检查是否存在符合条件的记录
-- SELECT 1 比 SELECT * 更高效，因为不需要读取实际数据

-- 查询有订单的用户
SELECT * FROM users u
WHERE EXISTS (
    SELECT 1 FROM orders o WHERE o.user_id = u.id
);

-- 等价于（但 EXISTS + SELECT 1 通常更优）
SELECT * FROM users u
WHERE EXISTS (
    SELECT * FROM orders o WHERE o.user_id = u.id
);

-- 查询没有订单的用户
SELECT * FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM orders o WHERE o.user_id = u.id
);
```

#### 3. IF EXISTS 检查

```sql
-- 检查表是否存在数据
IF EXISTS (SELECT 1 FROM users WHERE status = 'active')
BEGIN
    PRINT 'Active users exist'
END

-- MySQL 中
SELECT IF(EXISTS(SELECT 1 FROM users WHERE status = 'active'), 'Yes', 'No');
```

#### 4. 条件计数

```sql
-- 统计满足条件的记录数
SELECT COUNT(1) FROM users WHERE status = 'active';

-- COUNT(1) vs COUNT(*) vs COUNT(column)
-- COUNT(1): 统计行数，1 是常量
-- COUNT(*): 统计行数，包括 NULL
-- COUNT(column): 统计非 NULL 的行数
```

#### 5. 生成序列

```sql
-- 使用 UNION 生成数字序列
SELECT 1 AS n UNION ALL
SELECT 2 UNION ALL
SELECT 3 UNION ALL
SELECT 4 UNION ALL
SELECT 5;

-- 配合递归 CTE 生成序列
WITH RECURSIVE numbers AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM numbers WHERE n < 100
)
SELECT n FROM numbers;
```

#### 6. 行转列辅助

```sql
-- 使用 SELECT 1 配合 CASE 进行行转列
SELECT 
    user_id,
    SUM(CASE WHEN month = 1 THEN amount ELSE 0 END) AS jan,
    SUM(CASE WHEN month = 2 THEN amount ELSE 0 END) AS feb,
    SUM(CASE WHEN month = 3 THEN amount ELSE 0 END) AS mar
FROM sales
GROUP BY user_id;
```

### SELECT 1 vs SELECT * 性能

```sql
-- EXISTS 中使用 SELECT 1（推荐）
SELECT * FROM users u
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id);

-- EXISTS 中使用 SELECT *
SELECT * FROM users u
WHERE EXISTS (SELECT * FROM orders o WHERE o.user_id = u.id);

-- 注意：现代数据库优化器通常会自动优化
-- 但使用 SELECT 1 可以：
-- 1. 明确表达意图（只关心存在性）
-- 2. 避免不必要的列解析
-- 3. 代码更清晰
```

### 其他常量查询

```sql
-- SELECT 任意常量
SELECT 1;
SELECT 'Hello';
SELECT 1 + 1;
SELECT CURRENT_DATE;
SELECT NOW();
SELECT UUID();

-- 生成多列常量
SELECT 1 AS col1, 'text' AS col2, NOW() AS col3;

-- 结合 FROM 生成多行
SELECT 1 FROM users;  -- 返回与 users 表行数相同的 1
```

---

## WHERE 子句

### 比较运算符

```sql
-- 基本比较
SELECT * FROM products WHERE price > 100;
SELECT * FROM products WHERE price >= 100;
SELECT * FROM products WHERE price < 100;
SELECT * FROM products WHERE price <= 100;
SELECT * FROM products WHERE price = 100;
SELECT * FROM products WHERE price != 100;  -- 或 <>

-- BETWEEN（包含边界）
SELECT * FROM products WHERE price BETWEEN 100 AND 200;
-- 等价于
SELECT * FROM products WHERE price >= 100 AND price <= 200;

-- IN
SELECT * FROM users WHERE country IN ('China', 'Japan', 'Korea');

-- NOT IN
SELECT * FROM users WHERE country NOT IN ('USA', 'UK');

-- LIKE 模糊匹配
SELECT * FROM users WHERE name LIKE 'John%';     -- 以 John 开头
SELECT * FROM users WHERE name LIKE '%son';      -- 以 son 结尾
SELECT * FROM users WHERE name LIKE '%oh%';      -- 包含 oh
SELECT * FROM users WHERE name LIKE 'J_hn';      -- J 后面一个任意字符后面是 hn
SELECT * FROM users WHERE name LIKE 'J%n';       -- J 开头 n 结尾

-- ESCAPE 转义
SELECT * FROM products WHERE name LIKE '%10\%%' ESCAPE '\';  -- 包含 10%

-- IS NULL / IS NOT NULL
SELECT * FROM users WHERE deleted_at IS NULL;
SELECT * FROM users WHERE deleted_at IS NOT NULL;

-- 安全等于（MySQL）- 可以比较 NULL
SELECT * FROM users WHERE deleted_at <=> NULL;
```

### 逻辑运算符

```sql
-- AND
SELECT * FROM users 
WHERE status = 'active' AND role = 'admin';

-- OR
SELECT * FROM users 
WHERE role = 'admin' OR role = 'moderator';

-- NOT
SELECT * FROM users 
WHERE NOT (status = 'inactive');

-- 组合（注意优先级，AND 高于 OR）
SELECT * FROM users 
WHERE (role = 'admin' OR role = 'moderator') 
  AND status = 'active';

-- 复杂条件
SELECT * FROM orders
WHERE (status = 'pending' AND created_at < DATE_SUB(NOW(), INTERVAL 7 DAY))
   OR (status = 'processing' AND created_at < DATE_SUB(NOW(), INTERVAL 3 DAY));
```

### 高级条件

```sql
-- REGEXP / RLIKE 正则匹配
SELECT * FROM users WHERE email REGEXP '^[a-z]+@gmail\\.com$';

-- PostgreSQL 正则
SELECT * FROM users WHERE email ~ '^[a-z]+@gmail\.com$';
SELECT * FROM users WHERE email ~* '^[a-z]+@gmail\.com$';  -- 不区分大小写

-- JSON 字段查询（MySQL 5.7+）
SELECT * FROM users WHERE JSON_EXTRACT(preferences, '$.theme') = 'dark';
SELECT * FROM users WHERE preferences->>'$.theme' = 'dark';

-- JSON 字段查询（PostgreSQL）
SELECT * FROM users WHERE preferences->>'theme' = 'dark';
SELECT * FROM users WHERE preferences @> '{"theme": "dark"}';

-- 数组包含（PostgreSQL）
SELECT * FROM posts WHERE tags @> ARRAY['mysql'];
SELECT * FROM posts WHERE 'mysql' = ANY(tags);
```

---

## JOIN 连接

### JOIN 类型图解

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          JOIN 类型                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   INNER JOIN                    LEFT JOIN                               │
│   ┌─────┬─────┐                 ┌─────┬─────┐                          │
│   │     │█████│                 │█████│█████│                          │
│   │  A  │█████│ B               │█████│█████│ B                        │
│   │     │█████│                 │█████│█████│                          │
│   └─────┴─────┘                 └─────┴─────┘                          │
│   只返回匹配的行                 返回左表所有行                           │
│                                                                         │
│   RIGHT JOIN                    FULL OUTER JOIN                         │
│   ┌─────┬─────┐                 ┌─────┬─────┐                          │
│   │█████│█████│                 │█████│█████│                          │
│   │█████│█████│ B               │█████│█████│ B                        │
│   │█████│█████│                 │█████│█████│                          │
│   └─────┴─────┘                 └─────┴─────┘                          │
│   返回右表所有行                 返回两表所有行                           │
│                                                                         │
│   CROSS JOIN                                                            │
│   返回笛卡尔积（A的每行 × B的每行）                                       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### JOIN 语法示例

```sql
-- INNER JOIN（内连接）
SELECT u.name, o.order_id, o.total
FROM users u
INNER JOIN orders o ON u.id = o.user_id;

-- LEFT JOIN（左连接）
SELECT u.name, o.order_id, o.total
FROM users u
LEFT JOIN orders o ON u.id = o.user_id;

-- LEFT JOIN 排除右表（查找没有订单的用户）
SELECT u.name
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE o.id IS NULL;

-- RIGHT JOIN（右连接）
SELECT u.name, o.order_id
FROM users u
RIGHT JOIN orders o ON u.id = o.user_id;

-- FULL OUTER JOIN（全外连接）- MySQL 不直接支持
-- PostgreSQL / SQL Server
SELECT u.name, o.order_id
FROM users u
FULL OUTER JOIN orders o ON u.id = o.user_id;

-- MySQL 模拟 FULL OUTER JOIN
SELECT u.name, o.order_id
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
UNION
SELECT u.name, o.order_id
FROM users u
RIGHT JOIN orders o ON u.id = o.user_id;

-- CROSS JOIN（交叉连接）
SELECT u.name, p.product_name
FROM users u
CROSS JOIN products p;

-- 自连接
SELECT e.name AS employee, m.name AS manager
FROM employees e
LEFT JOIN employees m ON e.manager_id = m.id;

-- 多表连接
SELECT 
    u.name,
    o.order_id,
    p.product_name,
    oi.quantity
FROM users u
INNER JOIN orders o ON u.id = o.user_id
INNER JOIN order_items oi ON o.id = oi.order_id
INNER JOIN products p ON oi.product_id = p.id;
```

### JOIN 条件

```sql
-- 使用 ON
SELECT * FROM users u
JOIN orders o ON u.id = o.user_id;

-- 使用 USING（当列名相同时）
SELECT * FROM users u
JOIN orders o USING (user_id);

-- 多条件 JOIN
SELECT * FROM users u
JOIN orders o ON u.id = o.user_id AND o.status = 'completed';

-- 注意：ON 中的条件 vs WHERE 中的条件
-- 对于 INNER JOIN 结果相同
-- 对于 LEFT JOIN 结果不同

-- LEFT JOIN + ON 条件（保留左表所有行）
SELECT u.name, o.order_id
FROM users u
LEFT JOIN orders o ON u.id = o.user_id AND o.status = 'completed';

-- LEFT JOIN + WHERE 条件（过滤结果）
SELECT u.name, o.order_id
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE o.status = 'completed' OR o.status IS NULL;
```

---

## GROUP BY 与聚合

### 聚合函数

```sql
-- COUNT
SELECT COUNT(*) FROM users;                    -- 所有行数
SELECT COUNT(email) FROM users;                -- 非 NULL 的 email 数
SELECT COUNT(DISTINCT country) FROM users;     -- 不同国家数

-- SUM
SELECT SUM(amount) FROM orders;
SELECT SUM(DISTINCT amount) FROM orders;       -- 去重后求和

-- AVG
SELECT AVG(price) FROM products;
SELECT AVG(DISTINCT price) FROM products;

-- MAX / MIN
SELECT MAX(created_at) FROM orders;
SELECT MIN(price), MAX(price) FROM products;

-- GROUP_CONCAT（MySQL）/ STRING_AGG（PostgreSQL）
SELECT 
    user_id,
    GROUP_CONCAT(product_name SEPARATOR ', ') AS products
FROM order_items
GROUP BY user_id;

-- PostgreSQL
SELECT 
    user_id,
    STRING_AGG(product_name, ', ') AS products
FROM order_items
GROUP BY user_id;
```

### GROUP BY 用法

```sql
-- 基本分组
SELECT country, COUNT(*) AS user_count
FROM users
GROUP BY country;

-- 多列分组
SELECT country, city, COUNT(*) AS user_count
FROM users
GROUP BY country, city;

-- 分组后过滤（HAVING）
SELECT country, COUNT(*) AS user_count
FROM users
GROUP BY country
HAVING COUNT(*) > 100;

-- WHERE vs HAVING
-- WHERE: 分组前过滤行
-- HAVING: 分组后过滤组
SELECT country, COUNT(*) AS user_count
FROM users
WHERE status = 'active'       -- 先过滤活跃用户
GROUP BY country
HAVING COUNT(*) > 10;         -- 再过滤用户数 > 10 的国家

-- 使用表达式分组
SELECT 
    YEAR(created_at) AS year,
    MONTH(created_at) AS month,
    COUNT(*) AS order_count
FROM orders
GROUP BY YEAR(created_at), MONTH(created_at);

-- 或使用列别名位置（MySQL）
SELECT 
    YEAR(created_at) AS year,
    MONTH(created_at) AS month,
    COUNT(*) AS order_count
FROM orders
GROUP BY 1, 2;
```

### WITH ROLLUP

```sql
-- 生成小计和总计
SELECT 
    COALESCE(country, 'Total') AS country,
    COALESCE(city, 'Subtotal') AS city,
    COUNT(*) AS user_count
FROM users
GROUP BY country, city WITH ROLLUP;

-- 结果示例：
-- | country | city     | user_count |
-- |---------|----------|------------|
-- | China   | Beijing  | 100        |
-- | China   | Shanghai | 200        |
-- | China   | Subtotal | 300        |  -- 中国小计
-- | Japan   | Tokyo    | 150        |
-- | Japan   | Subtotal | 150        |  -- 日本小计
-- | Total   | Subtotal | 450        |  -- 总计
```

### GROUPING SETS（PostgreSQL / SQL Server）

```sql
-- 多种分组方式
SELECT country, city, COUNT(*)
FROM users
GROUP BY GROUPING SETS (
    (country, city),  -- 按 country 和 city 分组
    (country),        -- 只按 country 分组
    ()                -- 总计
);

-- CUBE：所有可能的组合
SELECT country, city, COUNT(*)
FROM users
GROUP BY CUBE (country, city);

-- ROLLUP：层级汇总
SELECT country, city, COUNT(*)
FROM users
GROUP BY ROLLUP (country, city);
```

---

## ORDER BY 排序

### 基本排序

```sql
-- 升序（默认）
SELECT * FROM users ORDER BY name ASC;
SELECT * FROM users ORDER BY name;  -- 默认升序

-- 降序
SELECT * FROM users ORDER BY created_at DESC;

-- 多列排序
SELECT * FROM users 
ORDER BY country ASC, created_at DESC;

-- 使用列位置
SELECT name, email, created_at FROM users
ORDER BY 3 DESC;  -- 按第 3 列排序

-- 使用别名
SELECT name, YEAR(created_at) AS year
FROM users
ORDER BY year DESC;

-- 使用表达式
SELECT * FROM products
ORDER BY price * quantity DESC;
```

### NULL 值排序

```sql
-- MySQL 中 NULL 排在最前面（升序）或最后面（降序）
SELECT * FROM users ORDER BY deleted_at ASC;   -- NULL 在前
SELECT * FROM users ORDER BY deleted_at DESC;  -- NULL 在后

-- PostgreSQL / Oracle 控制 NULL 位置
SELECT * FROM users ORDER BY deleted_at ASC NULLS LAST;
SELECT * FROM users ORDER BY deleted_at DESC NULLS FIRST;

-- MySQL 模拟 NULLS LAST
SELECT * FROM users 
ORDER BY deleted_at IS NULL, deleted_at ASC;
```

### 条件排序

```sql
-- CASE 表达式排序
SELECT * FROM users
ORDER BY 
    CASE status
        WHEN 'urgent' THEN 1
        WHEN 'high' THEN 2
        WHEN 'normal' THEN 3
        WHEN 'low' THEN 4
        ELSE 5
    END;

-- FIELD 函数（MySQL）
SELECT * FROM users
ORDER BY FIELD(status, 'urgent', 'high', 'normal', 'low');

-- 自定义排序：置顶某些记录
SELECT * FROM posts
ORDER BY 
    CASE WHEN is_pinned = 1 THEN 0 ELSE 1 END,
    created_at DESC;
```

### 随机排序

```sql
-- MySQL
SELECT * FROM users ORDER BY RAND() LIMIT 10;

-- PostgreSQL
SELECT * FROM users ORDER BY RANDOM() LIMIT 10;

-- SQL Server
SELECT TOP 10 * FROM users ORDER BY NEWID();

-- 优化：大表随机查询
SELECT * FROM users
WHERE id >= (SELECT FLOOR(RAND() * (SELECT MAX(id) FROM users)))
ORDER BY id
LIMIT 10;
```

---

## LIMIT 与分页

### 基本语法

```sql
-- MySQL / PostgreSQL / SQLite
SELECT * FROM users LIMIT 10;           -- 前 10 行
SELECT * FROM users LIMIT 10 OFFSET 20; -- 跳过 20 行，取 10 行
SELECT * FROM users LIMIT 20, 10;       -- MySQL 简写，等价于上面

-- SQL Server
SELECT TOP 10 * FROM users;
SELECT * FROM users 
ORDER BY id
OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY;

-- Oracle
SELECT * FROM users WHERE ROWNUM <= 10;
-- Oracle 12c+
SELECT * FROM users 
ORDER BY id
OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY;
```

### 分页查询

```sql
-- 基础分页（页码从 1 开始）
-- page_size = 10, page_number = 3
SELECT * FROM users
ORDER BY id
LIMIT 10 OFFSET 20;  -- (page_number - 1) * page_size

-- 带总数的分页
SELECT SQL_CALC_FOUND_ROWS * FROM users
ORDER BY id
LIMIT 10 OFFSET 20;

SELECT FOUND_ROWS() AS total;  -- 获取总数

-- 优化：使用子查询获取总数
SELECT 
    (SELECT COUNT(*) FROM users) AS total,
    u.*
FROM users u
ORDER BY id
LIMIT 10 OFFSET 20;
```

### 游标分页（更高效）

```sql
-- 基于游标的分页（避免大 OFFSET）
-- 第一页
SELECT * FROM users
ORDER BY id
LIMIT 10;

-- 下一页（基于上一页最后一条记录的 id）
SELECT * FROM users
WHERE id > 10  -- last_id
ORDER BY id
LIMIT 10;

-- 多列排序的游标分页
SELECT * FROM users
WHERE (created_at, id) > ('2024-01-01', 100)
ORDER BY created_at, id
LIMIT 10;
```

### Top-N 查询

```sql
-- 每组取前 N 条
-- 方法 1：子查询
SELECT * FROM orders o1
WHERE (
    SELECT COUNT(*) FROM orders o2
    WHERE o2.user_id = o1.user_id AND o2.created_at > o1.created_at
) < 3
ORDER BY user_id, created_at DESC;

-- 方法 2：窗口函数
SELECT * FROM (
    SELECT 
        *,
        ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at DESC) AS rn
    FROM orders
) t
WHERE rn <= 3;
```

---

## 子查询

### 子查询类型

```sql
-- 1. 标量子查询（返回单个值）
SELECT 
    name,
    (SELECT AVG(price) FROM products) AS avg_price
FROM products;

-- 2. 列子查询（返回一列）
SELECT * FROM users
WHERE id IN (SELECT user_id FROM orders WHERE total > 1000);

-- 3. 行子查询（返回一行）
SELECT * FROM users
WHERE (name, email) = (SELECT name, email FROM admins WHERE id = 1);

-- 4. 表子查询（返回多行多列）
SELECT * FROM (
    SELECT user_id, SUM(total) AS total_amount
    FROM orders
    GROUP BY user_id
) AS user_totals
WHERE total_amount > 10000;
```

### 相关子查询

```sql
-- 相关子查询：内部查询依赖外部查询
SELECT * FROM users u
WHERE EXISTS (
    SELECT 1 FROM orders o 
    WHERE o.user_id = u.id AND o.total > 1000
);

-- 等价的 JOIN 写法（通常性能更好）
SELECT DISTINCT u.* FROM users u
INNER JOIN orders o ON u.id = o.user_id AND o.total > 1000;

-- 相关子查询计算
SELECT 
    p.name,
    p.price,
    (SELECT AVG(price) FROM products p2 WHERE p2.category_id = p.category_id) AS category_avg
FROM products p;
```

### 子查询位置

```sql
-- SELECT 子句中
SELECT 
    name,
    (SELECT COUNT(*) FROM orders WHERE user_id = users.id) AS order_count
FROM users;

-- FROM 子句中（派生表）
SELECT * FROM (
    SELECT user_id, COUNT(*) AS cnt
    FROM orders
    GROUP BY user_id
) AS order_counts
WHERE cnt > 5;

-- WHERE 子句中
SELECT * FROM users
WHERE id IN (SELECT user_id FROM orders);

-- HAVING 子句中
SELECT user_id, COUNT(*) AS cnt
FROM orders
GROUP BY user_id
HAVING COUNT(*) > (SELECT AVG(order_count) FROM user_stats);

-- JOIN 中
SELECT u.name, o.total
FROM users u
INNER JOIN (
    SELECT user_id, SUM(amount) AS total
    FROM orders
    GROUP BY user_id
) o ON u.id = o.user_id;
```

### 子查询操作符

```sql
-- IN / NOT IN
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders);
SELECT * FROM users WHERE id NOT IN (SELECT user_id FROM blocked_users);

-- EXISTS / NOT EXISTS
SELECT * FROM users u
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id);

SELECT * FROM users u
WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id);

-- ANY / SOME（满足任一条件）
SELECT * FROM products WHERE price > ANY (SELECT price FROM competitor_products);
-- 等价于
SELECT * FROM products WHERE price > (SELECT MIN(price) FROM competitor_products);

-- ALL（满足所有条件）
SELECT * FROM products WHERE price > ALL (SELECT price FROM competitor_products);
-- 等价于
SELECT * FROM products WHERE price > (SELECT MAX(price) FROM competitor_products);
```

---

## 窗口函数

### 窗口函数语法

```sql
function_name(expression) OVER (
    [PARTITION BY partition_expression, ...]
    [ORDER BY sort_expression [ASC | DESC], ...]
    [frame_clause]
)
```

### 排名函数

```sql
-- ROW_NUMBER()：行号，无重复
-- RANK()：排名，有间隔
-- DENSE_RANK()：排名，无间隔

SELECT 
    name,
    department,
    salary,
    ROW_NUMBER() OVER (ORDER BY salary DESC) AS row_num,
    RANK() OVER (ORDER BY salary DESC) AS rank,
    DENSE_RANK() OVER (ORDER BY salary DESC) AS dense_rank
FROM employees;

-- 结果示例（salary: 100, 100, 90, 80）：
-- | name  | salary | row_num | rank | dense_rank |
-- |-------|--------|---------|------|------------|
-- | Alice | 100    | 1       | 1    | 1          |
-- | Bob   | 100    | 2       | 1    | 1          |
-- | Carol | 90     | 3       | 3    | 2          |
-- | David | 80     | 4       | 4    | 3          |

-- 分组排名
SELECT 
    name,
    department,
    salary,
    RANK() OVER (PARTITION BY department ORDER BY salary DESC) AS dept_rank
FROM employees;

-- NTILE：分桶
SELECT 
    name,
    salary,
    NTILE(4) OVER (ORDER BY salary DESC) AS quartile  -- 分成 4 组
FROM employees;
```

### 聚合窗口函数

```sql
-- 累计求和
SELECT 
    date,
    amount,
    SUM(amount) OVER (ORDER BY date) AS running_total
FROM sales;

-- 移动平均
SELECT 
    date,
    amount,
    AVG(amount) OVER (
        ORDER BY date 
        ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
    ) AS moving_avg_7d
FROM sales;

-- 分组聚合
SELECT 
    department,
    name,
    salary,
    SUM(salary) OVER (PARTITION BY department) AS dept_total,
    AVG(salary) OVER (PARTITION BY department) AS dept_avg,
    salary - AVG(salary) OVER (PARTITION BY department) AS diff_from_avg
FROM employees;

-- 百分比
SELECT 
    name,
    salary,
    salary * 100.0 / SUM(salary) OVER () AS pct_of_total,
    salary * 100.0 / SUM(salary) OVER (PARTITION BY department) AS pct_of_dept
FROM employees;
```

### 偏移函数

```sql
-- LAG：前一行
-- LEAD：后一行
SELECT 
    date,
    amount,
    LAG(amount, 1) OVER (ORDER BY date) AS prev_amount,
    LEAD(amount, 1) OVER (ORDER BY date) AS next_amount,
    amount - LAG(amount, 1) OVER (ORDER BY date) AS diff_from_prev
FROM sales;

-- LAG/LEAD 带默认值
SELECT 
    date,
    amount,
    LAG(amount, 1, 0) OVER (ORDER BY date) AS prev_amount  -- 默认值 0
FROM sales;

-- FIRST_VALUE / LAST_VALUE
SELECT 
    date,
    amount,
    FIRST_VALUE(amount) OVER (ORDER BY date) AS first_amount,
    LAST_VALUE(amount) OVER (
        ORDER BY date
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    ) AS last_amount
FROM sales;

-- NTH_VALUE
SELECT 
    name,
    salary,
    NTH_VALUE(name, 2) OVER (ORDER BY salary DESC) AS second_highest_earner
FROM employees;
```

### 窗口框架

```sql
-- 框架语法
ROWS | RANGE BETWEEN frame_start AND frame_end

-- frame_start / frame_end 可以是：
-- UNBOUNDED PRECEDING  - 分区的第一行
-- n PRECEDING          - 当前行前 n 行
-- CURRENT ROW          - 当前行
-- n FOLLOWING          - 当前行后 n 行
-- UNBOUNDED FOLLOWING  - 分区的最后一行

-- 示例
SELECT 
    date,
    amount,
    -- 累计（从第一行到当前行）
    SUM(amount) OVER (ORDER BY date ROWS UNBOUNDED PRECEDING) AS cumulative,
    
    -- 滑动窗口（前 3 行到后 3 行）
    AVG(amount) OVER (ORDER BY date ROWS BETWEEN 3 PRECEDING AND 3 FOLLOWING) AS centered_avg,
    
    -- 后续所有行
    SUM(amount) OVER (ORDER BY date ROWS BETWEEN CURRENT ROW AND UNBOUNDED FOLLOWING) AS remaining
FROM sales;
```

---

## CTE 公共表表达式

### 基本 CTE

```sql
-- 定义 CTE
WITH active_users AS (
    SELECT * FROM users WHERE status = 'active'
)
SELECT * FROM active_users WHERE country = 'China';

-- 多个 CTE
WITH 
    active_users AS (
        SELECT * FROM users WHERE status = 'active'
    ),
    user_orders AS (
        SELECT user_id, COUNT(*) AS order_count, SUM(total) AS total_amount
        FROM orders
        GROUP BY user_id
    )
SELECT 
    u.name,
    o.order_count,
    o.total_amount
FROM active_users u
LEFT JOIN user_orders o ON u.id = o.user_id;

-- CTE 引用其他 CTE
WITH 
    base_data AS (
        SELECT * FROM sales WHERE year = 2024
    ),
    monthly_totals AS (
        SELECT month, SUM(amount) AS total
        FROM base_data
        GROUP BY month
    ),
    with_running_total AS (
        SELECT 
            month,
            total,
            SUM(total) OVER (ORDER BY month) AS running_total
        FROM monthly_totals
    )
SELECT * FROM with_running_total;
```

### 递归 CTE

```sql
-- 递归 CTE 语法
WITH RECURSIVE cte_name AS (
    -- 锚点成员（初始查询）
    SELECT ...
    UNION ALL
    -- 递归成员（引用 CTE 自身）
    SELECT ... FROM cte_name WHERE ...
)
SELECT * FROM cte_name;

-- 示例：生成数字序列
WITH RECURSIVE numbers AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM numbers WHERE n < 100
)
SELECT n FROM numbers;

-- 示例：生成日期序列
WITH RECURSIVE dates AS (
    SELECT DATE('2024-01-01') AS date
    UNION ALL
    SELECT DATE_ADD(date, INTERVAL 1 DAY) FROM dates WHERE date < '2024-12-31'
)
SELECT date FROM dates;

-- 示例：组织架构树
WITH RECURSIVE org_tree AS (
    -- 锚点：顶级管理者
    SELECT id, name, manager_id, 1 AS level, CAST(name AS CHAR(1000)) AS path
    FROM employees
    WHERE manager_id IS NULL
    
    UNION ALL
    
    -- 递归：下级员工
    SELECT e.id, e.name, e.manager_id, t.level + 1, CONCAT(t.path, ' > ', e.name)
    FROM employees e
    INNER JOIN org_tree t ON e.manager_id = t.id
)
SELECT * FROM org_tree ORDER BY path;

-- 示例：物料清单（BOM）
WITH RECURSIVE bom AS (
    SELECT id, name, parent_id, 1 AS quantity, 1 AS level
    FROM parts
    WHERE parent_id IS NULL
    
    UNION ALL
    
    SELECT p.id, p.name, p.parent_id, p.quantity * b.quantity, b.level + 1
    FROM parts p
    INNER JOIN bom b ON p.parent_id = b.id
)
SELECT * FROM bom;
```

---

## 集合操作

### UNION / UNION ALL

```sql
-- UNION：合并并去重
SELECT name, email FROM customers
UNION
SELECT name, email FROM suppliers;

-- UNION ALL：合并不去重（性能更好）
SELECT name, email FROM customers
UNION ALL
SELECT name, email FROM suppliers;

-- 多表合并
SELECT 'customer' AS type, name FROM customers
UNION ALL
SELECT 'supplier' AS type, name FROM suppliers
UNION ALL
SELECT 'employee' AS type, name FROM employees;
```

### INTERSECT

```sql
-- 交集：两个查询都有的记录
SELECT user_id FROM orders_2023
INTERSECT
SELECT user_id FROM orders_2024;

-- MySQL 模拟 INTERSECT
SELECT DISTINCT user_id FROM orders_2023
WHERE user_id IN (SELECT user_id FROM orders_2024);
```

### EXCEPT / MINUS

```sql
-- 差集：在第一个查询但不在第二个查询的记录
-- PostgreSQL / SQL Server
SELECT user_id FROM orders_2023
EXCEPT
SELECT user_id FROM orders_2024;

-- Oracle
SELECT user_id FROM orders_2023
MINUS
SELECT user_id FROM orders_2024;

-- MySQL 模拟 EXCEPT
SELECT DISTINCT user_id FROM orders_2023
WHERE user_id NOT IN (SELECT user_id FROM orders_2024);
```

---

## 性能优化

### 索引优化

```sql
-- 创建合适的索引
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_orders_user_date ON orders(user_id, created_at);

-- 复合索引顺序（最左前缀原则）
-- 索引 (a, b, c) 可以用于：
-- WHERE a = ?
-- WHERE a = ? AND b = ?
-- WHERE a = ? AND b = ? AND c = ?
-- 但不能用于：
-- WHERE b = ?
-- WHERE b = ? AND c = ?

-- 覆盖索引（避免回表）
CREATE INDEX idx_orders_covering ON orders(user_id, status, total);
SELECT status, total FROM orders WHERE user_id = 1;  -- 只查索引列

-- 查看执行计划
EXPLAIN SELECT * FROM users WHERE status = 'active';
EXPLAIN ANALYZE SELECT * FROM users WHERE status = 'active';  -- PostgreSQL
```

### 查询优化技巧

```sql
-- 1. 避免 SELECT *
SELECT id, name, email FROM users;  -- 只查需要的列

-- 2. 使用 EXISTS 代替 IN（大数据集）
-- 较慢
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders);
-- 较快
SELECT * FROM users u WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id);

-- 3. 避免在索引列上使用函数
-- 不走索引
SELECT * FROM users WHERE YEAR(created_at) = 2024;
-- 走索引
SELECT * FROM users WHERE created_at >= '2024-01-01' AND created_at < '2025-01-01';

-- 4. 避免隐式类型转换
-- user_id 是 INT 类型
SELECT * FROM users WHERE user_id = '123';  -- 可能不走索引
SELECT * FROM users WHERE user_id = 123;    -- 走索引

-- 5. LIMIT 优化
-- 较慢（大 OFFSET）
SELECT * FROM users ORDER BY id LIMIT 1000000, 10;
-- 较快（游标分页）
SELECT * FROM users WHERE id > 1000000 ORDER BY id LIMIT 10;

-- 6. 延迟关联
-- 较慢
SELECT * FROM orders ORDER BY created_at DESC LIMIT 1000000, 10;
-- 较快
SELECT o.* FROM orders o
INNER JOIN (
    SELECT id FROM orders ORDER BY created_at DESC LIMIT 1000000, 10
) t ON o.id = t.id;

-- 7. 批量操作代替循环
-- 较慢
INSERT INTO users (name) VALUES ('a');
INSERT INTO users (name) VALUES ('b');
INSERT INTO users (name) VALUES ('c');
-- 较快
INSERT INTO users (name) VALUES ('a'), ('b'), ('c');

-- 8. 使用 UNION ALL 代替 UNION（如果不需要去重）
SELECT * FROM table1 UNION ALL SELECT * FROM table2;
```

### 统计与诊断

```sql
-- MySQL 慢查询日志
SET GLOBAL slow_query_log = 1;
SET GLOBAL long_query_time = 1;  -- 超过 1 秒记录

-- 查看表统计信息
SHOW TABLE STATUS LIKE 'users';

-- 查看索引使用情况
SHOW INDEX FROM users;

-- MySQL 性能模式
SELECT * FROM performance_schema.events_statements_summary_by_digest
ORDER BY SUM_TIMER_WAIT DESC LIMIT 10;

-- PostgreSQL 统计信息
SELECT * FROM pg_stat_user_tables WHERE relname = 'users';
SELECT * FROM pg_stat_user_indexes WHERE relname = 'users';
```

---

## 总结

### SELECT 语句速查

```sql
SELECT [DISTINCT] columns
FROM table
[JOIN other_table ON condition]
[WHERE condition]
[GROUP BY columns [HAVING condition]]
[ORDER BY columns [ASC|DESC]]
[LIMIT count [OFFSET offset]]
```

### 常用技巧

| 场景 | 推荐方法 |
|------|---------|
| 检查存在性 | `EXISTS (SELECT 1 ...)` |
| 分页查询 | 游标分页 > OFFSET 分页 |
| 去重 | `DISTINCT` 或 `GROUP BY` |
| 排名 | `ROW_NUMBER()` / `RANK()` |
| 累计 | 窗口函数 `SUM() OVER()` |
| 层级查询 | 递归 CTE |
| 复杂逻辑 | CTE 分步处理 |

### 性能原则

1. 只查需要的列，避免 `SELECT *`
2. 使用合适的索引
3. 避免在索引列上使用函数
4. 大数据集用 `EXISTS` 代替 `IN`
5. 使用游标分页代替大 OFFSET
6. 使用 `EXPLAIN` 分析执行计划

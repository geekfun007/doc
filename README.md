# Kitex + Semi Form 参数校验最优解

全链路参数校验方案：以 Thrift IDL 为**唯一事实来源（Single Source of Truth）**，前端 Semi Design Form 与后端 Kitex Server 共享同一套校验规则，实现"一次定义，两端生效"。

## 目录

- [整体架构](#整体架构)
- [后端：IDL 定义校验规则](#后端idl-定义校验规则)
- [后端：Kitex Middleware 自动校验](#后端kitex-middleware-自动校验)
- [前端：Semi Form 校验规则](#前端semi-form-校验规则)
- [打通前后端：IDL → 前端规则自动生成](#打通前后端idl--前端规则自动生成)
- [完整示例](#完整示例)
- [最佳实践](#最佳实践)

---

## 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    Thrift IDL (唯一事实来源)               │
│  struct CreateUserReq {                                  │
│      1: string name  (vt.min_size="1", vt.max_size="50")│
│      2: string email (vt.pattern="...")                  │
│      3: i32    age   (vt.ge="0", vt.le="150")           │
│  }                                                       │
└────────────┬──────────────────────────┬──────────────────┘
             │                          │
     thrift-gen-validator         IDL → JSON 脚本
             │                          │
             ▼                          ▼
┌────────────────────┐     ┌────────────────────────┐
│   Kitex Server     │     │   Semi Design Form     │
│                    │     │                        │
│  Middleware 调用    │     │  rules 自动生成         │
│  req.IsValid()     │     │  required / pattern /  │
│  → 统一错误码返回   │     │  validator(val)        │
└────────────────────┘     └────────────────────────┘
```

核心思想：
1. **IDL 是规则的唯一来源**，后端通过 `thrift-gen-validator` 自动生成 Go 校验代码，前端通过脚本/工具从 IDL 提取规则生成 Semi Form `rules`。
2. **后端校验是安全底线**，通过 Kitex Middleware 对所有入站请求自动执行 `IsValid()`，无需在每个 handler 中手写。
3. **前端校验是体验优化**，Semi Form 的 `rules` 在用户输入时即时反馈，减少无效请求。

---

## 后端：IDL 定义校验规则

### 安装 thrift-gen-validator

```bash
go install github.com/cloudwego/thrift-gen-validator@latest
```

### IDL 校验注解语法

在 `.thrift` 文件中使用 `vt.{约束类型} = "值"` 格式添加校验注解：

```thrift
// user.thrift
namespace go user

enum Gender {
    UNKNOWN = 0
    MALE    = 1
    FEMALE  = 2
}

struct CreateUserReq {
    // 字符串：长度、正则、前缀
    1: required string name      (vt.min_size = "1", vt.max_size = "50")
    2: required string email     (vt.pattern = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")
    3: optional string phone     (vt.pattern = "^1[3-9]\\d{9}$")

    // 数值：范围
    4: required i32 age          (vt.ge = "0", vt.le = "150")
    5: optional double score     (vt.ge = "0", vt.le = "100")

    // 枚举：仅允许已定义值
    6: required Gender gender    (vt.defined_only = "true")

    // 集合：长度 + 元素约束
    7: optional list<string> tags (vt.max_size = "10", vt.elem.max_size = "20")

    // 嵌套结构
    8: optional Address address
}

struct Address {
    1: required string province  (vt.min_size = "1")
    2: required string city      (vt.min_size = "1")
    3: required string detail    (vt.min_size = "1", vt.max_size = "200")
    4: optional string zipCode   (vt.pattern = "^\\d{6}$")
}

struct CreateUserResp {
    1: i32    code
    2: string message
    3: i64    userId
}

service UserService {
    CreateUserResp CreateUser(1: CreateUserReq req)
}
```

### 常用注解速查

| 注解 | 适用类型 | 说明 | 示例 |
|------|---------|------|------|
| `vt.const` | 数值/字符串 | 常量约束 | `vt.const = "abc"` |
| `vt.lt` / `vt.le` | 数值 | 小于 / 小于等于 | `vt.lt = "100"` |
| `vt.gt` / `vt.ge` | 数值 | 大于 / 大于等于 | `vt.ge = "0"` |
| `vt.in` | 数值/字符串 | 枚举值（可多次出现） | `vt.in = "A"`, `vt.in = "B"` |
| `vt.not_in` | 数值/字符串 | 排除值 | `vt.not_in = "root"` |
| `vt.min_size` | 字符串/集合/map | 最小长度 | `vt.min_size = "1"` |
| `vt.max_size` | 字符串/集合/map | 最大长度 | `vt.max_size = "50"` |
| `vt.pattern` | 字符串 | 正则匹配 | `vt.pattern = "^\\d+$"` |
| `vt.prefix` | 字符串 | 前缀约束 | `vt.prefix = "usr_"` |
| `vt.suffix` | 字符串 | 后缀约束 | `vt.suffix = ".png"` |
| `vt.contains` | 字符串 | 包含子串 | `vt.contains = "@"` |
| `vt.not_contains` | 字符串 | 不包含子串 | `vt.not_contains = " "` |
| `vt.elem` | list/set 元素 | 元素级约束 | `vt.elem.gt = "0"` |
| `vt.key` / `vt.value` | map 键/值 | map 键值约束 | `vt.key.min_size = "1"` |
| `vt.defined_only` | enum | 仅允许已定义值 | `vt.defined_only = "true"` |
| `vt.not_nil` | optional 指针 | 不允许 nil | `vt.not_nil = "true"` |

### 生成校验代码

```bash
kitex --thrift-plugin validator -service user -module github.com/yourorg/yourproject user.thrift
```

生成后会在 `kitex_gen/` 目录产生 `*-validator.go` 文件，每个 struct 自动拥有 `IsValid() error` 方法。

---

## 后端：Kitex Middleware 自动校验

不要在每个 handler 里手写 `req.IsValid()`，而是通过 **Kitex Server Middleware** 统一拦截：

```go
package middleware

import (
    "context"
    "fmt"

    "github.com/cloudwego/kitex/pkg/endpoint"
    "github.com/cloudwego/kitex/pkg/kerrors"
)

// Validatable 是 thrift-gen-validator 为所有 struct 生成的接口
type Validatable interface {
    IsValid() error
}

// ValidationMiddleware 自动校验所有入站请求参数
func ValidationMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
    return func(ctx context.Context, req, resp interface{}) error {
        if v, ok := req.(Validatable); ok {
            if err := v.IsValid(); err != nil {
                return kerrors.NewBizStatusError(400, fmt.Sprintf("参数校验失败: %v", err))
            }
        }
        return next(ctx, req, resp)
    }
}
```

在 Server 启动时注册：

```go
package main

import (
    "github.com/cloudwego/kitex/server"
    "github.com/yourorg/yourproject/middleware"
    user "github.com/yourorg/yourproject/kitex_gen/user/userservice"
)

func main() {
    svr := user.NewServer(
        new(UserServiceImpl),
        server.WithMiddleware(middleware.ValidationMiddleware),
        // ... 其他选项
    )
    if err := svr.Run(); err != nil {
        panic(err)
    }
}
```

这样做的好处：
- **零遗漏**：所有 RPC 方法的入参都会被自动校验，新增方法无需额外处理
- **关注点分离**：handler 只关心业务逻辑，校验由框架层统一处理
- **统一错误码**：通过 `kerrors.NewBizStatusError` 返回标准化的业务错误码

---

## 前端：Semi Form 校验规则

Semi Design 的 `<Form.Field>` 通过 `rules` 属性声明校验规则，支持以下能力：

### 基础用法

```tsx
import { Form, Button } from '@douyinfe/semi-ui';

const CreateUserForm: React.FC = () => {
    const handleSubmit = (values: Record<string, any>) => {
        // 调用 RPC / HTTP 接口
        console.log('提交数据:', values);
    };

    return (
        <Form onSubmit={handleSubmit}>
            <Form.Input
                field="name"
                label="姓名"
                rules={[
                    { required: true, message: '姓名不能为空' },
                    { min: 1, max: 50, message: '姓名长度需在 1-50 个字符之间' },
                ]}
                trigger="blur"
            />

            <Form.Input
                field="email"
                label="邮箱"
                rules={[
                    { required: true, message: '邮箱不能为空' },
                    {
                        pattern: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
                        message: '请输入合法的邮箱地址',
                    },
                ]}
                trigger="blur"
            />

            <Form.Input
                field="phone"
                label="手机号"
                rules={[
                    {
                        pattern: /^1[3-9]\d{9}$/,
                        message: '请输入合法的手机号',
                    },
                ]}
                trigger="blur"
            />

            <Form.InputNumber
                field="age"
                label="年龄"
                rules={[
                    { required: true, message: '年龄不能为空' },
                    { type: 'number', min: 0, max: 150, message: '年龄需在 0-150 之间' },
                ]}
                trigger="blur"
            />

            <Form.Select
                field="gender"
                label="性别"
                rules={[{ required: true, message: '请选择性别' }]}
            >
                <Form.Select.Option value={1}>男</Form.Select.Option>
                <Form.Select.Option value={2}>女</Form.Select.Option>
            </Form.Select>

            <Form.TagInput
                field="tags"
                label="标签"
                rules={[
                    {
                        validator: (rule, value) => {
                            if (value && value.length > 10) {
                                return Promise.reject('最多添加 10 个标签');
                            }
                            if (value && value.some((t: string) => t.length > 20)) {
                                return Promise.reject('每个标签最多 20 个字符');
                            }
                            return Promise.resolve();
                        },
                    },
                ]}
            />

            <Button htmlType="submit" type="primary">
                提交
            </Button>
        </Form>
    );
};
```

### Semi Form rules 属性速查

| 属性 | 类型 | 说明 |
|------|------|------|
| `required` | `boolean` | 是否必填 |
| `message` | `string` | 校验失败时的提示信息 |
| `type` | `string` | 值类型（`string` / `number` / `array` / `email` / `url` 等） |
| `pattern` | `RegExp` | 正则校验 |
| `min` / `max` | `number` | 字符串长度或数值范围 |
| `len` | `number` | 精确长度 |
| `enum` | `any[]` | 值必须在枚举列表中 |
| `whitespace` | `boolean` | 纯空白是否视为空 |
| `transform` | `(value) => value` | 校验前对值做变换 |
| `validator` | `(rule, value) => Promise` | 自定义校验函数 |
| `asyncValidator` | `(rule, value) => Promise` | 异步校验（如后端去重查询） |

`trigger` 属性控制校验触发时机：`'blur'`（失焦时）、`'change'`（值变化时）、`'custom'`（手动触发）、`'mount'`（挂载时）。

---

## 打通前后端：IDL → 前端规则自动生成

手工在前端维护一份与 IDL 一致的 rules 既繁琐又容易遗漏。推荐的方案是编写一个脚本，解析 IDL 中的 `vt.*` 注解，自动生成前端可用的 Semi Form rules。

### 方案一：Node.js 脚本（推荐）

```typescript
// scripts/gen-form-rules.ts
// 解析 thrift IDL 中的 vt.* 注解，生成 Semi Form rules

import * as fs from 'fs';
import * as path from 'path';

interface VtAnnotation {
    minSize?: number;
    maxSize?: number;
    gt?: number;
    ge?: number;
    lt?: number;
    le?: number;
    pattern?: string;
    prefix?: string;
    suffix?: string;
    contains?: string;
    in?: string[];
    definedOnly?: boolean;
    notNil?: boolean;
    elem?: Partial<VtAnnotation>;
}

interface FieldDef {
    name: string;
    type: string;
    required: boolean;
    annotations: VtAnnotation;
}

interface SemiFormRule {
    required?: boolean;
    message?: string;
    type?: string;
    pattern?: RegExp | string;
    min?: number;
    max?: number;
    enum?: any[];
    validator?: string; // 生成为函数字符串
}

function parseAnnotations(raw: string): VtAnnotation {
    const result: VtAnnotation = {};
    const regex = /vt\.(\w+(?:\.\w+)?)\s*=\s*"([^"]*)"/g;
    let match;

    while ((match = regex.exec(raw)) !== null) {
        const [, key, value] = match;
        switch (key) {
            case 'min_size': result.minSize = parseInt(value); break;
            case 'max_size': result.maxSize = parseInt(value); break;
            case 'gt': result.gt = parseFloat(value); break;
            case 'ge': result.ge = parseFloat(value); break;
            case 'lt': result.lt = parseFloat(value); break;
            case 'le': result.le = parseFloat(value); break;
            case 'pattern': result.pattern = value; break;
            case 'prefix': result.prefix = value; break;
            case 'suffix': result.suffix = value; break;
            case 'contains': result.contains = value; break;
            case 'defined_only': result.definedOnly = value === 'true'; break;
            case 'not_nil': result.notNil = value === 'true'; break;
            case 'in':
                if (!result.in) result.in = [];
                result.in.push(value);
                break;
            case 'elem.gt':
            case 'elem.ge':
            case 'elem.lt':
            case 'elem.le':
            case 'elem.max_size':
            case 'elem.min_size':
                if (!result.elem) result.elem = {};
                const elemKey = key.split('.')[1];
                (result.elem as any)[toCamelCase(elemKey)] = parseFloat(value) || value;
                break;
        }
    }
    return result;
}

function toCamelCase(s: string): string {
    return s.replace(/_([a-z])/g, (_, c) => c.toUpperCase());
}

function annotationsToRules(field: FieldDef): SemiFormRule[] {
    const rules: SemiFormRule[] = [];
    const { annotations: a, required: isRequired, name, type } = field;

    if (isRequired || a.notNil) {
        rules.push({ required: true, message: `${name} 不能为空` });
    }

    const isString = type === 'string' || type === 'binary';
    const isNumeric = ['i8', 'i16', 'i32', 'i64', 'double', 'byte'].includes(type);

    if (isString && (a.minSize !== undefined || a.maxSize !== undefined)) {
        rules.push({
            min: a.minSize,
            max: a.maxSize,
            message: `${name} 长度需在 ${a.minSize ?? 0}-${a.maxSize ?? '∞'} 之间`,
        });
    }

    if (isNumeric) {
        const min = a.ge ?? (a.gt !== undefined ? a.gt + 1 : undefined);
        const max = a.le ?? (a.lt !== undefined ? a.lt - 1 : undefined);
        if (min !== undefined || max !== undefined) {
            rules.push({
                type: 'number',
                min,
                max,
                message: `${name} 需在 ${min ?? '-∞'}-${max ?? '∞'} 之间`,
            });
        }
    }

    if (a.pattern) {
        rules.push({
            pattern: a.pattern,
            message: `${name} 格式不正确`,
        });
    }

    if (a.in && a.in.length > 0) {
        rules.push({
            enum: a.in,
            message: `${name} 必须是以下值之一: ${a.in.join(', ')}`,
        });
    }

    return rules;
}

// 使用示例：
// npx ts-node scripts/gen-form-rules.ts idl/user.thrift > src/generated/userFormRules.ts
```

### 方案二：利用 Kitex 生成的 Go struct tag 做 JSON Schema 桥接

```
Thrift IDL
    │
    ├──► thrift-gen-validator → Go IsValid()
    │
    └──► 自定义 Go 工具 → JSON Schema → 前端 ajv / Semi rules
```

编写一个 Go 小工具，反射读取生成的 struct tag 或直接解析 IDL AST（使用 `github.com/cloudwego/thriftgo` 的 parser），输出 JSON Schema 文件。前端使用 `ajv` 或自定义转换层将 JSON Schema 映射为 Semi Form rules。

### 推荐工程目录结构

```
project/
├── idl/
│   └── user.thrift              # IDL 唯一事实来源
├── kitex_gen/                   # Kitex 生成代码 (含 validator)
├── scripts/
│   └── gen-form-rules.ts        # IDL → Semi rules 脚本
├── server/
│   ├── main.go
│   ├── handler.go
│   └── middleware/
│       └── validation.go        # 通用校验 Middleware
└── web/
    └── src/
        ├── generated/
        │   └── userFormRules.ts  # 自动生成的校验规则
        └── pages/
            └── CreateUser.tsx    # 使用生成规则的 Form
```

---

## 完整示例

### 1. IDL 定义 (`idl/user.thrift`)

```thrift
namespace go user

struct CreateUserReq {
    1: required string name  (vt.min_size = "1", vt.max_size = "50")
    2: required string email (vt.pattern = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")
    3: required i32    age   (vt.ge = "0", vt.le = "150")
}

struct CreateUserResp {
    1: i32    code
    2: string message
}

service UserService {
    CreateUserResp CreateUser(1: CreateUserReq req)
}
```

### 2. 后端 Middleware (`server/middleware/validation.go`)

```go
package middleware

import (
    "context"
    "fmt"

    "github.com/cloudwego/kitex/pkg/endpoint"
    "github.com/cloudwego/kitex/pkg/kerrors"
)

type Validatable interface {
    IsValid() error
}

func ValidationMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
    return func(ctx context.Context, req, resp interface{}) error {
        if v, ok := req.(Validatable); ok {
            if err := v.IsValid(); err != nil {
                return kerrors.NewBizStatusError(400, fmt.Sprintf("参数校验失败: %v", err))
            }
        }
        return next(ctx, req, resp)
    }
}
```

### 3. Handler 保持纯净 (`server/handler.go`)

```go
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *user.CreateUserReq) (*user.CreateUserResp, error) {
    // 无需手写校验，Middleware 已处理
    // 只关注业务逻辑
    userId, err := s.userRepo.Create(ctx, req)
    if err != nil {
        return nil, err
    }
    return &user.CreateUserResp{Code: 0, Message: "success", UserId: userId}, nil
}
```

### 4. 前端自动生成的规则 (`web/src/generated/userFormRules.ts`)

```typescript
// 由 scripts/gen-form-rules.ts 从 idl/user.thrift 自动生成，请勿手动修改

export const createUserReqRules = {
    name: [
        { required: true, message: '姓名不能为空' },
        { min: 1, max: 50, message: '姓名长度需在 1-50 个字符之间' },
    ],
    email: [
        { required: true, message: '邮箱不能为空' },
        {
            pattern: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
            message: '请输入合法的邮箱地址',
        },
    ],
    age: [
        { required: true, message: '年龄不能为空' },
        { type: 'number' as const, min: 0, max: 150, message: '年龄需在 0-150 之间' },
    ],
};
```

### 5. 前端 Form 消费规则 (`web/src/pages/CreateUser.tsx`)

```tsx
import { Form, Button, Toast } from '@douyinfe/semi-ui';
import { createUserReqRules } from '../generated/userFormRules';

const CreateUser: React.FC = () => {
    const handleSubmit = async (values: Record<string, any>) => {
        try {
            const resp = await fetch('/api/user/create', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(values),
            });
            const data = await resp.json();
            if (data.code === 0) {
                Toast.success('创建成功');
            } else {
                Toast.error(data.message);
            }
        } catch {
            Toast.error('网络错误');
        }
    };

    return (
        <Form onSubmit={handleSubmit} labelPosition="left" labelWidth="80px">
            <Form.Input
                field="name"
                label="姓名"
                rules={createUserReqRules.name}
                trigger="blur"
            />
            <Form.Input
                field="email"
                label="邮箱"
                rules={createUserReqRules.email}
                trigger="blur"
            />
            <Form.InputNumber
                field="age"
                label="年龄"
                rules={createUserReqRules.age}
                trigger="blur"
            />
            <Button htmlType="submit" type="primary" style={{ marginTop: 16 }}>
                提交
            </Button>
        </Form>
    );
};

export default CreateUser;
```

---

## 最佳实践

### 1. 校验分层原则

| 层级 | 职责 | 工具 | 触发时机 |
|------|------|------|---------|
| **前端 Form** | 用户体验优化，即时反馈 | Semi Form `rules` | `blur` / `change` |
| **后端 Middleware** | 安全底线，防止绕过前端 | `IsValid()` + Middleware | 每次 RPC 调用 |
| **数据库约束** | 最终保障，数据完整性 | DDL `CHECK` / `NOT NULL` | 写入时 |

**永远不要只依赖前端校验**——用户可以跳过前端直接调接口。后端校验是安全底线，必须存在。

### 2. 错误信息标准化

后端返回结构化的校验错误，方便前端展示：

```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type BizError struct {
    Code   int               `json:"code"`
    Msg    string            `json:"msg"`
    Errors []ValidationError `json:"errors,omitempty"`
}
```

前端接收后可以直接映射到对应字段：

```typescript
const handleServerErrors = (errors: Array<{field: string; message: string}>, formApi: any) => {
    errors.forEach(({ field, message }) => {
        formApi.setError(field, message);
    });
};
```

### 3. 校验触发时机选择

```
用户输入
  │
  ├─ 首次输入 → 不校验（不打扰用户）
  │
  ├─ 失焦 (blur) → 触发校验（推荐大多数场景）
  │
  ├─ 提交 (submit) → 全量校验
  │
  └─ 修正后 → 实时校验 (change)，快速清除错误
```

Semi Form 推荐 `trigger="blur"` 作为默认策略，提交时自动做全量校验。

### 4. 异步校验（如唯一性检查）

IDL 注解无法表达"用户名是否已存在"这类业务校验，需要在前端使用 `asyncValidator`：

```tsx
<Form.Input
    field="username"
    label="用户名"
    rules={[
        ...generatedRules.username,
        {
            asyncValidator: async (rule, value) => {
                if (!value) return;
                const resp = await fetch(`/api/user/check?username=${value}`);
                const { exists } = await resp.json();
                if (exists) {
                    throw new Error('用户名已被占用');
                }
            },
        },
    ]}
    trigger="blur"
/>
```

### 5. CI 流程集成

在 CI 中加入校验规则同步检查，确保 IDL 变更后前端规则同步更新：

```yaml
# .github/workflows/validate-rules.yml
- name: 检查前端规则是否与 IDL 同步
  run: |
    npx ts-node scripts/gen-form-rules.ts idl/user.thrift > /tmp/rules.ts
    diff /tmp/rules.ts web/src/generated/userFormRules.ts
```

### 6. 总结

| 方面 | 最优解 |
|------|--------|
| 规则来源 | Thrift IDL `vt.*` 注解，唯一事实来源 |
| 后端校验 | `thrift-gen-validator` 生成 `IsValid()` + Kitex Middleware 自动调用 |
| 前端校验 | 脚本从 IDL 提取注解 → 生成 Semi Form `rules` |
| 错误反馈 | 前端 `trigger="blur"` 即时反馈 + 后端结构化错误映射到字段 |
| 异步校验 | Semi Form `asyncValidator` 处理业务级校验（唯一性等） |
| 一致性保障 | CI 流程自动检查 IDL ↔ 前端规则同步 |

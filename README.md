# Protocol Buffers (.proto) 详解

Protocol Buffers（简称 protobuf）是 Google 开发的一种语言无关、平台无关、可扩展的序列化结构数据的方法。它比 XML 更小、更快、更简单。

## 目录

1. [基本概念](#基本概念)
2. [语法版本](#语法版本)
3. [基本语法结构](#基本语法结构)
4. [数据类型](#数据类型)
5. [字段规则](#字段规则)
6. [高级特性](#高级特性)
7. [服务定义 (gRPC)](#服务定义-grpc)
8. [最佳实践](#最佳实践)
9. [示例](#示例)

---

## 基本概念

### 什么是 Protocol Buffers?

Protocol Buffers 是一种：
- **序列化格式**：将结构化数据转换为字节流
- **接口定义语言 (IDL)**：定义数据结构和服务接口
- **代码生成工具**：自动生成多种编程语言的代码

### 主要优势

| 特性 | Protocol Buffers | JSON | XML |
|------|-----------------|------|-----|
| 大小 | 小 (二进制) | 中等 | 大 |
| 解析速度 | 快 | 中等 | 慢 |
| 可读性 | 低 (需工具) | 高 | 高 |
| Schema | 强制 | 可选 | 可选 |
| 向后兼容 | 优秀 | 一般 | 一般 |

---

## 语法版本

Protocol Buffers 有两个主要版本：

### proto2 (旧版本)
```protobuf
syntax = "proto2";

message Person {
  required string name = 1;
  optional int32 age = 2;
  repeated string emails = 3;
}
```

### proto3 (推荐使用)
```protobuf
syntax = "proto3";

message Person {
  string name = 1;
  int32 age = 2;
  repeated string emails = 3;
}
```

**主要区别：**
- proto3 移除了 `required` 和 `optional` 关键字
- proto3 所有字段默认都是可选的
- proto3 标量类型有默认值（不能区分"未设置"和"默认值"）
- proto3 移除了 `default` 选项
- proto3 支持 JSON 映射

---

## 基本语法结构

### 1. 文件结构

```protobuf
// 语法声明（必须在第一行非空非注释行）
syntax = "proto3";

// 包名声明
package mycompany.myproject;

// 导入其他 proto 文件
import "google/protobuf/timestamp.proto";
import "other_protos/common.proto";

// 选项设置
option java_package = "com.mycompany.myproject";
option go_package = "github.com/mycompany/myproject/pb";

// 消息定义
message MyMessage {
  // 字段定义
}

// 枚举定义
enum MyEnum {
  // 枚举值
}

// 服务定义
service MyService {
  // RPC 方法
}
```

### 2. 消息定义 (Message)

消息是 protobuf 的核心构建块：

```protobuf
message SearchRequest {
  string query = 1;           // 字段名 = 字段编号
  int32 page_number = 2;
  int32 results_per_page = 3;
}
```

**字段编号规则：**
- 必须是正整数
- 1-15 使用 1 字节编码（常用字段应使用）
- 16-2047 使用 2 字节编码
- 最大值为 2²⁹-1 (536,870,911)
- 19000-19999 保留给 protobuf 实现
- 一旦使用，**永远不要更改**

### 3. 嵌套消息

```protobuf
message SearchResponse {
  message Result {
    string url = 1;
    string title = 2;
    repeated string snippets = 3;
  }
  
  repeated Result results = 1;
  int32 total_count = 2;
}

// 在其他消息中引用嵌套消息
message AnotherMessage {
  SearchResponse.Result result = 1;
}
```

---

## 数据类型

### 标量类型 (Scalar Types)

| .proto 类型 | 说明 | Go 类型 | Java 类型 | Python 类型 |
|------------|------|---------|-----------|-------------|
| `double` | 双精度浮点 | float64 | double | float |
| `float` | 单精度浮点 | float32 | float | float |
| `int32` | 变长编码整数 | int32 | int | int |
| `int64` | 变长编码整数 | int64 | long | int/long |
| `uint32` | 变长编码无符号整数 | uint32 | int | int/long |
| `uint64` | 变长编码无符号整数 | uint64 | long | int/long |
| `sint32` | 变长编码有符号整数(负数更高效) | int32 | int | int |
| `sint64` | 变长编码有符号整数(负数更高效) | int64 | long | int/long |
| `fixed32` | 固定4字节(大数更高效) | uint32 | int | int/long |
| `fixed64` | 固定8字节(大数更高效) | uint64 | long | int/long |
| `sfixed32` | 固定4字节有符号 | int32 | int | int |
| `sfixed64` | 固定8字节有符号 | int64 | long | int/long |
| `bool` | 布尔值 | bool | boolean | bool |
| `string` | UTF-8 字符串 | string | String | str |
| `bytes` | 任意字节序列 | []byte | ByteString | bytes |

### 默认值

在 proto3 中，标量类型的默认值：
- 数字类型: `0`
- bool: `false`
- string: `""`（空字符串）
- bytes: `空字节序列`
- enum: 第一个枚举值（必须是 0）
- message: 语言特定的 null/nil

### 枚举类型 (Enum)

```protobuf
enum Status {
  STATUS_UNSPECIFIED = 0;  // proto3 要求第一个值必须是 0
  STATUS_PENDING = 1;
  STATUS_APPROVED = 2;
  STATUS_REJECTED = 3;
}

message Order {
  string id = 1;
  Status status = 2;
}
```

**枚举别名：**
```protobuf
enum EnumAllowingAlias {
  option allow_alias = true;
  EAA_UNSPECIFIED = 0;
  EAA_STARTED = 1;
  EAA_RUNNING = 1;  // 别名，与 STARTED 相同
  EAA_FINISHED = 2;
}
```

### 复合类型

#### Map 类型
```protobuf
message Project {
  string name = 1;
  map<string, string> labels = 2;        // key: string, value: string
  map<int32, User> users = 3;            // key: int32, value: User message
}
```

**限制：**
- key 只能是整数或字符串类型
- value 可以是任何类型（除了另一个 map）
- map 不能使用 `repeated`

#### Any 类型
```protobuf
import "google/protobuf/any.proto";

message ErrorStatus {
  string message = 1;
  repeated google.protobuf.Any details = 2;
}
```

#### Oneof 类型
```protobuf
message SampleMessage {
  oneof test_oneof {
    string name = 4;
    SubMessage sub_message = 9;
  }
}
```
- 同时只能设置其中一个字段
- 设置任何成员会自动清除其他成员

#### Wrapper Types (可空类型)
```protobuf
import "google/protobuf/wrappers.proto";

message Person {
  string name = 1;
  google.protobuf.Int32Value age = 2;        // 可以区分 0 和未设置
  google.protobuf.StringValue nickname = 3;  // 可以区分 "" 和未设置
}
```

---

## 字段规则

### proto3 字段规则

| 规则 | 说明 |
|------|------|
| singular (默认) | 0 或 1 个该字段 |
| `optional` | 明确表示可选，可追踪是否设置 |
| `repeated` | 0 到多个（有序数组） |
| `map` | 键值对集合 |

### optional 关键字 (proto3)

```protobuf
syntax = "proto3";

message Person {
  string name = 1;           // 隐式可选
  optional int32 age = 2;    // 显式可选，可以检测是否设置
}
```

使用 `optional` 可以区分：
- 字段未设置
- 字段设置为默认值（如 0）

### reserved 关键字

防止字段编号或名称被重用：

```protobuf
message Foo {
  reserved 2, 15, 9 to 11;           // 保留字段编号
  reserved "foo", "bar";              // 保留字段名
  
  string name = 1;
  // int32 old_field = 2;  // 错误！2 已被保留
}
```

---

## 高级特性

### 1. 导入 (Import)

```protobuf
// 标准导入
import "myproject/other_protos.proto";

// 公开导入（导入的内容可被再次导入此文件的文件访问）
import public "myproject/public_dependency.proto";

// 弱导入（文件不存在时不报错）
import weak "myproject/optional.proto";
```

### 2. 包 (Package)

```protobuf
package foo.bar;

message Open {
  // ...
}
```

在其他文件中使用：
```protobuf
message Foo {
  foo.bar.Open open = 1;
}
```

### 3. 选项 (Options)

#### 文件级选项
```protobuf
option java_package = "com.example.foo";
option java_outer_classname = "Ponycopter";
option java_multiple_files = true;
option go_package = "github.com/example/foo/pb";
option optimize_for = SPEED;  // CODE_SIZE, LITE_RUNTIME
option cc_enable_arenas = true;
option objc_class_prefix = "GPB";
```

#### 消息级选项
```protobuf
message Foo {
  option message_set_wire_format = true;
  option deprecated = true;
}
```

#### 字段级选项
```protobuf
message Bar {
  int32 old_field = 1 [deprecated = true];
  repeated int32 samples = 2 [packed = true];
  string json_name = 3 [json_name = "jsonFieldName"];
}
```

### 4. 自定义选项

```protobuf
import "google/protobuf/descriptor.proto";

extend google.protobuf.FieldOptions {
  optional string my_field_option = 50001;
}

message MyMessage {
  string name = 1 [(my_field_option) = "custom value"];
}
```

---

## 服务定义 (gRPC)

Protocol Buffers 常与 gRPC 一起使用定义 RPC 服务：

```protobuf
syntax = "proto3";

package helloworld;

// 服务定义
service Greeter {
  // 一元 RPC
  rpc SayHello (HelloRequest) returns (HelloReply);
  
  // 服务端流式 RPC
  rpc LotsOfReplies (HelloRequest) returns (stream HelloReply);
  
  // 客户端流式 RPC
  rpc LotsOfGreetings (stream HelloRequest) returns (HelloReply);
  
  // 双向流式 RPC
  rpc BidiHello (stream HelloRequest) returns (stream HelloReply);
}

message HelloRequest {
  string name = 1;
}

message HelloReply {
  string message = 1;
}
```

### RPC 类型说明

| 类型 | 请求 | 响应 | 使用场景 |
|------|------|------|---------|
| 一元 | 单个 | 单个 | 普通请求/响应 |
| 服务端流式 | 单个 | 流 | 下载大文件、实时更新 |
| 客户端流式 | 流 | 单个 | 上传大文件、聚合数据 |
| 双向流式 | 流 | 流 | 聊天、实时协作 |

---

## 最佳实践

### 1. 命名规范

```protobuf
// 文件名：使用 snake_case
// user_service.proto

// 包名：使用小写点分隔
package mycompany.myproject.v1;

// 消息名：使用 PascalCase
message UserRequest {
  // 字段名：使用 snake_case
  string user_name = 1;
  int32 page_size = 2;
}

// 枚举名：使用 PascalCase
enum UserStatus {
  // 枚举值：使用 SCREAMING_SNAKE_CASE，并加前缀
  USER_STATUS_UNSPECIFIED = 0;
  USER_STATUS_ACTIVE = 1;
  USER_STATUS_INACTIVE = 2;
}

// 服务名：使用 PascalCase
service UserService {
  // 方法名：使用 PascalCase
  rpc GetUser (GetUserRequest) returns (GetUserResponse);
  rpc ListUsers (ListUsersRequest) returns (ListUsersResponse);
}
```

### 2. API 版本控制

```protobuf
// v1/user.proto
package mycompany.user.v1;

message User {
  string id = 1;
  string name = 2;
}

// v2/user.proto
package mycompany.user.v2;

message User {
  string id = 1;
  string first_name = 2;
  string last_name = 3;  // v2 新增字段
}
```

### 3. 向后兼容性原则

**DO (可以做):**
- 添加新字段
- 添加新的枚举值
- 添加新的消息类型
- 添加新的服务方法
- 将字段标记为 `deprecated`

**DON'T (不要做):**
- 更改字段编号
- 更改字段类型（除非类型兼容）
- 重命名字段（影响 JSON 序列化）
- 删除字段（使用 `reserved` 代替）
- 更改枚举值的编号

### 4. 常用设计模式

#### 请求/响应模式
```protobuf
message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  User user = 1;
}

message ListUsersRequest {
  int32 page_size = 1;
  string page_token = 2;
  string filter = 3;
}

message ListUsersResponse {
  repeated User users = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}
```

#### 标准错误详情
```protobuf
import "google/protobuf/any.proto";

message Status {
  int32 code = 1;
  string message = 2;
  repeated google.protobuf.Any details = 3;
}
```

---

## 示例

### 完整示例：电商订单系统

```protobuf
syntax = "proto3";

package ecommerce.order.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";

option go_package = "github.com/mycompany/ecommerce/order/v1;orderv1";
option java_package = "com.mycompany.ecommerce.order.v1";
option java_multiple_files = true;

// ============ 枚举 ============

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_CONFIRMED = 2;
  ORDER_STATUS_SHIPPED = 3;
  ORDER_STATUS_DELIVERED = 4;
  ORDER_STATUS_CANCELLED = 5;
}

enum PaymentMethod {
  PAYMENT_METHOD_UNSPECIFIED = 0;
  PAYMENT_METHOD_CREDIT_CARD = 1;
  PAYMENT_METHOD_DEBIT_CARD = 2;
  PAYMENT_METHOD_PAYPAL = 3;
  PAYMENT_METHOD_BANK_TRANSFER = 4;
}

// ============ 消息 ============

message Money {
  string currency_code = 1;  // ISO 4217 货币代码
  int64 units = 2;           // 整数部分
  int32 nanos = 3;           // 小数部分 (0-999,999,999)
}

message Address {
  string street = 1;
  string city = 2;
  string state = 3;
  string country = 4;
  string postal_code = 5;
}

message OrderItem {
  string product_id = 1;
  string product_name = 2;
  int32 quantity = 3;
  Money unit_price = 4;
  Money total_price = 5;
  map<string, string> attributes = 6;  // 如：颜色、尺寸
}

message Order {
  string order_id = 1;
  string customer_id = 2;
  
  repeated OrderItem items = 3;
  
  OrderStatus status = 4;
  PaymentMethod payment_method = 5;
  
  Address shipping_address = 6;
  Address billing_address = 7;
  
  Money subtotal = 8;
  Money shipping_fee = 9;
  Money tax = 10;
  Money total = 11;
  
  google.protobuf.StringValue coupon_code = 12;  // 可选优惠券
  google.protobuf.Int32Value discount_percent = 13;
  
  google.protobuf.Timestamp created_at = 14;
  google.protobuf.Timestamp updated_at = 15;
  
  map<string, string> metadata = 16;
}

// ============ 服务 ============

service OrderService {
  // 创建订单
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  
  // 获取订单
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
  
  // 列出订单
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
  
  // 更新订单状态
  rpc UpdateOrderStatus(UpdateOrderStatusRequest) returns (UpdateOrderStatusResponse);
  
  // 取消订单
  rpc CancelOrder(CancelOrderRequest) returns (CancelOrderResponse);
  
  // 订单状态变更流（服务端流式）
  rpc WatchOrderStatus(WatchOrderStatusRequest) returns (stream OrderStatusUpdate);
}

// ============ 请求/响应消息 ============

message CreateOrderRequest {
  string customer_id = 1;
  repeated OrderItem items = 2;
  Address shipping_address = 3;
  Address billing_address = 4;
  PaymentMethod payment_method = 5;
  optional string coupon_code = 6;
}

message CreateOrderResponse {
  Order order = 1;
}

message GetOrderRequest {
  string order_id = 1;
}

message GetOrderResponse {
  Order order = 1;
}

message ListOrdersRequest {
  string customer_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  
  // 过滤条件
  repeated OrderStatus status_filter = 4;
  google.protobuf.Timestamp created_after = 5;
  google.protobuf.Timestamp created_before = 6;
}

message ListOrdersResponse {
  repeated Order orders = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

message UpdateOrderStatusRequest {
  string order_id = 1;
  OrderStatus new_status = 2;
  optional string reason = 3;
}

message UpdateOrderStatusResponse {
  Order order = 1;
}

message CancelOrderRequest {
  string order_id = 1;
  string cancellation_reason = 2;
}

message CancelOrderResponse {
  Order order = 1;
  bool refund_initiated = 2;
}

message WatchOrderStatusRequest {
  string order_id = 1;
}

message OrderStatusUpdate {
  string order_id = 1;
  OrderStatus previous_status = 2;
  OrderStatus current_status = 3;
  google.protobuf.Timestamp timestamp = 4;
  optional string message = 5;
}
```

---

## 编译与代码生成

### 安装 protoc 编译器

```bash
# macOS
brew install protobuf

# Ubuntu/Debian
apt-get install protobuf-compiler

# 验证安装
protoc --version
```

### 代码生成示例

```bash
# Go
protoc --go_out=. --go-grpc_out=. proto/*.proto

# Python
protoc --python_out=. --grpc_python_out=. proto/*.proto

# Java
protoc --java_out=. --grpc-java_out=. proto/*.proto

# 多语言同时生成
protoc \
  --go_out=./gen/go \
  --go-grpc_out=./gen/go \
  --python_out=./gen/python \
  --grpc_python_out=./gen/python \
  -I ./proto \
  ./proto/*.proto
```

---

## 参考资源

- [Protocol Buffers 官方文档](https://developers.google.com/protocol-buffers)
- [Protocol Buffers Language Guide (proto3)](https://developers.google.com/protocol-buffers/docs/proto3)
- [gRPC 官方文档](https://grpc.io/docs/)
- [Google API 设计指南](https://cloud.google.com/apis/design)
- [Buf - 现代化 Protobuf 工具](https://buf.build/)

---

## 总结

Protocol Buffers 是一个强大的数据序列化工具，具有以下特点：

1. **高效**：二进制格式，体积小，解析快
2. **类型安全**：强类型 Schema 定义
3. **跨语言**：支持多种编程语言
4. **向后兼容**：良好的版本演进支持
5. **代码生成**：自动生成序列化/反序列化代码
6. **生态丰富**：与 gRPC、Envoy 等工具深度集成

掌握 Protocol Buffers 对于构建高性能微服务架构、API 设计和数据通信至关重要。

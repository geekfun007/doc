# TypeScript Object vs Map 详解

本文档详细对比 TypeScript/JavaScript 中 Object 和 Map 的区别、使用场景和最佳实践。

## 目录

1. [基本概念](#基本概念)
2. [语法对比](#语法对比)
3. [核心区别](#核心区别)
4. [性能对比](#性能对比)
5. [类型定义](#类型定义)
6. [使用场景](#使用场景)
7. [常见操作对比](#常见操作对比)
8. [最佳实践](#最佳实践)

---

## 基本概念

### Object

Object 是 JavaScript 的基础数据结构，用于存储键值对。

```typescript
// 对象字面量
const obj: { [key: string]: number } = {
  apple: 1,
  banana: 2,
};

// Record 类型
const prices: Record<string, number> = {
  apple: 1.5,
  banana: 0.5,
};
```

### Map

Map 是 ES6 引入的集合类型，专门用于存储键值对。

```typescript
// Map 构造函数
const map = new Map<string, number>();
map.set('apple', 1);
map.set('banana', 2);

// 从数组初始化
const map2 = new Map<string, number>([
  ['apple', 1],
  ['banana', 2],
]);
```

---

## 语法对比

### 创建

```typescript
// ============ Object ============
// 字面量
const obj1 = {};
const obj2 = { key: 'value' };

// 构造函数
const obj3 = new Object();
const obj4 = Object.create(null);  // 无原型对象

// ============ Map ============
// 构造函数
const map1 = new Map();
const map2 = new Map([['key', 'value']]);

// 从对象创建 Map
const obj = { a: 1, b: 2 };
const map3 = new Map(Object.entries(obj));

// 从 Map 创建对象
const map4 = new Map([['a', 1], ['b', 2]]);
const obj5 = Object.fromEntries(map4);
```

### 增删改查

```typescript
// ============ Object ============
const obj: Record<string, number> = {};

// 添加/修改
obj.key = 1;
obj['key'] = 1;

// 获取
const value = obj.key;
const value2 = obj['key'];

// 删除
delete obj.key;

// 检查键是否存在
'key' in obj;
obj.hasOwnProperty('key');
Object.hasOwn(obj, 'key');  // ES2022+

// 获取所有键
Object.keys(obj);

// 获取所有值
Object.values(obj);

// 获取所有键值对
Object.entries(obj);

// ============ Map ============
const map = new Map<string, number>();

// 添加/修改
map.set('key', 1);

// 获取
const value = map.get('key');

// 删除
map.delete('key');

// 清空
map.clear();

// 检查键是否存在
map.has('key');

// 获取大小
map.size;

// 获取所有键
map.keys();

// 获取所有值
map.values();

// 获取所有键值对
map.entries();
```

### 遍历

```typescript
const obj: Record<string, number> = { a: 1, b: 2, c: 3 };
const map = new Map<string, number>([['a', 1], ['b', 2], ['c', 3]]);

// ============ Object 遍历 ============

// for...in（包含原型链属性，需要 hasOwnProperty 检查）
for (const key in obj) {
  if (obj.hasOwnProperty(key)) {
    console.log(key, obj[key]);
  }
}

// Object.keys()
Object.keys(obj).forEach(key => {
  console.log(key, obj[key]);
});

// Object.entries()
Object.entries(obj).forEach(([key, value]) => {
  console.log(key, value);
});

// for...of + Object.entries()
for (const [key, value] of Object.entries(obj)) {
  console.log(key, value);
}

// ============ Map 遍历 ============

// forEach
map.forEach((value, key) => {
  console.log(key, value);
});

// for...of（默认迭代 entries）
for (const [key, value] of map) {
  console.log(key, value);
}

// for...of + keys()
for (const key of map.keys()) {
  console.log(key);
}

// for...of + values()
for (const value of map.values()) {
  console.log(value);
}

// for...of + entries()
for (const [key, value] of map.entries()) {
  console.log(key, value);
}
```

---

## 核心区别

### 1. 键的类型

```typescript
// ============ Object: 键只能是 string 或 symbol ============
const obj: any = {};

obj['string'] = 1;        // ✅ 字符串键
obj[Symbol('id')] = 2;    // ✅ Symbol 键
obj[1] = 3;               // ⚠️ 数字被转换为字符串 "1"
obj[{}] = 4;              // ⚠️ 对象被转换为 "[object Object]"
obj[true] = 5;            // ⚠️ 布尔被转换为 "true"

console.log(Object.keys(obj));  // ["string", "1", "[object Object]", "true"]

// ============ Map: 键可以是任何类型 ============
const map = new Map<any, number>();

map.set('string', 1);     // ✅ 字符串键
map.set(Symbol('id'), 2); // ✅ Symbol 键
map.set(1, 3);            // ✅ 数字键（保持为数字）
map.set({}, 4);           // ✅ 对象键
map.set(true, 5);         // ✅ 布尔键
map.set(null, 6);         // ✅ null 键
map.set(undefined, 7);    // ✅ undefined 键
map.set(NaN, 8);          // ✅ NaN 键（NaN === NaN 在 Map 中为 true）

// 对象作为键
const objKey1 = { id: 1 };
const objKey2 = { id: 1 };
map.set(objKey1, 'value1');
map.set(objKey2, 'value2');  // 不同的对象，不同的键

console.log(map.get(objKey1));  // 'value1'
console.log(map.get(objKey2));  // 'value2'
console.log(map.size);          // 至少 9 个键
```

### 2. 键的顺序

```typescript
// ============ Object: 键的顺序不保证（ES6+ 有部分保证） ============
const obj: Record<string, number> = {};
obj['b'] = 1;
obj['a'] = 2;
obj['2'] = 3;
obj['1'] = 4;

// ES6+ 规则：
// 1. 整数键按升序排列
// 2. 字符串键按插入顺序
// 3. Symbol 键按插入顺序
console.log(Object.keys(obj));  // ["1", "2", "b", "a"]

// ============ Map: 始终保持插入顺序 ============
const map = new Map<string, number>();
map.set('b', 1);
map.set('a', 2);
map.set('2', 3);
map.set('1', 4);

console.log([...map.keys()]);  // ["b", "a", "2", "1"]
```

### 3. 大小获取

```typescript
// ============ Object: 需要计算 ============
const obj = { a: 1, b: 2, c: 3 };
const size = Object.keys(obj).length;  // 3

// ============ Map: 直接获取 ============
const map = new Map([['a', 1], ['b', 2], ['c', 3]]);
const size = map.size;  // 3
```

### 4. 默认键

```typescript
// ============ Object: 有原型链属性 ============
const obj: Record<string, any> = {};

console.log(obj.toString);       // [Function: toString]
console.log('toString' in obj);  // true
console.log(obj.hasOwnProperty('toString'));  // false

// 安全的对象（无原型）
const safeObj = Object.create(null);
console.log(safeObj.toString);   // undefined

// ============ Map: 没有默认键 ============
const map = new Map();

console.log(map.get('toString'));  // undefined
console.log(map.has('toString'));  // false
```

### 5. 可迭代性

```typescript
// ============ Object: 不直接可迭代 ============
const obj = { a: 1, b: 2 };

// ❌ 错误：Object 不可迭代
// for (const [k, v] of obj) { }

// ✅ 需要转换
for (const [k, v] of Object.entries(obj)) {
  console.log(k, v);
}

// ============ Map: 直接可迭代 ============
const map = new Map([['a', 1], ['b', 2]]);

// ✅ Map 实现了 Symbol.iterator
for (const [k, v] of map) {
  console.log(k, v);
}

// 展开操作
const arr = [...map];  // [['a', 1], ['b', 2]]
```

### 6. 序列化

```typescript
// ============ Object: 直接支持 JSON ============
const obj = { a: 1, b: 'hello' };

// 序列化
const json = JSON.stringify(obj);  // '{"a":1,"b":"hello"}'

// 反序列化
const parsed = JSON.parse(json);

// ============ Map: 需要手动处理 ============
const map = new Map([['a', 1], ['b', 'hello']]);

// 序列化
const json = JSON.stringify([...map]);  // '[["a",1],["b","hello"]]'
// 或
const json2 = JSON.stringify(Object.fromEntries(map));  // '{"a":1,"b":"hello"}'

// 反序列化
const parsed = new Map(JSON.parse(json));

// 自定义序列化
function mapToJson<K, V>(map: Map<K, V>): string {
  return JSON.stringify({
    dataType: 'Map',
    value: [...map],
  });
}

function jsonToMap<K, V>(json: string): Map<K, V> {
  const parsed = JSON.parse(json);
  if (parsed.dataType === 'Map') {
    return new Map(parsed.value);
  }
  throw new Error('Invalid Map JSON');
}
```

---

## 性能对比

### 基准测试结果（大致参考）

| 操作 | Object | Map | 胜者 |
|------|--------|-----|------|
| 创建（小量数据） | 更快 | 较慢 | Object |
| 创建（大量数据） | 较慢 | 更快 | Map |
| 插入 | 较慢 | 更快 | Map |
| 读取 | 相近 | 相近 | 平局 |
| 删除 | 较慢 | 更快 | Map |
| 迭代 | 较慢 | 更快 | Map |
| 检查键存在 | 较慢 | 更快 | Map |
| 获取大小 | O(n) | O(1) | Map |
| 内存占用 | 较少 | 较多 | Object |

### 性能测试示例

```typescript
// 性能测试函数
function benchmark(name: string, fn: () => void, iterations = 100000): void {
  const start = performance.now();
  for (let i = 0; i < iterations; i++) {
    fn();
  }
  const end = performance.now();
  console.log(`${name}: ${(end - start).toFixed(2)}ms`);
}

const SIZE = 10000;

// ============ 插入性能 ============
benchmark('Object insert', () => {
  const obj: Record<string, number> = {};
  for (let i = 0; i < SIZE; i++) {
    obj[`key${i}`] = i;
  }
});

benchmark('Map insert', () => {
  const map = new Map<string, number>();
  for (let i = 0; i < SIZE; i++) {
    map.set(`key${i}`, i);
  }
});

// ============ 读取性能 ============
const testObj: Record<string, number> = {};
const testMap = new Map<string, number>();
for (let i = 0; i < SIZE; i++) {
  testObj[`key${i}`] = i;
  testMap.set(`key${i}`, i);
}

benchmark('Object read', () => {
  for (let i = 0; i < SIZE; i++) {
    const _ = testObj[`key${i}`];
  }
});

benchmark('Map read', () => {
  for (let i = 0; i < SIZE; i++) {
    const _ = testMap.get(`key${i}`);
  }
});

// ============ 删除性能 ============
benchmark('Object delete', () => {
  const obj: Record<string, number> = { ...testObj };
  for (let i = 0; i < SIZE; i++) {
    delete obj[`key${i}`];
  }
});

benchmark('Map delete', () => {
  const map = new Map(testMap);
  for (let i = 0; i < SIZE; i++) {
    map.delete(`key${i}`);
  }
});

// ============ 检查存在性能 ============
benchmark('Object has', () => {
  for (let i = 0; i < SIZE; i++) {
    const _ = `key${i}` in testObj;
  }
});

benchmark('Map has', () => {
  for (let i = 0; i < SIZE; i++) {
    const _ = testMap.has(`key${i}`);
  }
});
```

---

## 类型定义

### Object 类型

```typescript
// ============ 索引签名 ============
interface StringDict {
  [key: string]: number;
}

const dict: StringDict = {
  a: 1,
  b: 2,
};

// ============ Record 工具类型 ============
type Status = 'pending' | 'success' | 'error';
const statusCodes: Record<Status, number> = {
  pending: 100,
  success: 200,
  error: 500,
};

// ============ 混合类型 ============
interface Config {
  name: string;
  version: number;
  [key: string]: string | number;  // 索引签名
}

// ============ 只读对象 ============
const frozen: Readonly<Record<string, number>> = Object.freeze({
  a: 1,
  b: 2,
});
// frozen.a = 3;  // ❌ 错误

// ============ 部分可选 ============
interface User {
  id: number;
  name: string;
  email?: string;
}

const partialUser: Partial<User> = {
  name: 'Alice',
};
```

### Map 类型

```typescript
// ============ 基本 Map 类型 ============
const map1: Map<string, number> = new Map();
const map2: Map<number, string> = new Map();
const map3: Map<object, any> = new Map();

// ============ 联合类型键 ============
type Key = string | number;
const map4: Map<Key, string> = new Map();
map4.set('key', 'value');
map4.set(123, 'value');

// ============ 复杂类型 ============
interface User {
  id: number;
  name: string;
}

const userMap: Map<number, User> = new Map();
userMap.set(1, { id: 1, name: 'Alice' });

// ============ 只读 Map ============
type ReadonlyMap<K, V> = Omit<Map<K, V>, 'set' | 'delete' | 'clear'>;

function getReadonlyMap(): ReadonlyMap<string, number> {
  const map = new Map<string, number>();
  map.set('a', 1);
  return map;
}

const readonlyMap = getReadonlyMap();
// readonlyMap.set('b', 2);  // ❌ 类型错误

// ============ 自定义 Map 类 ============
class TypedMap<K extends string, V> extends Map<K, V> {
  getOrDefault(key: K, defaultValue: V): V {
    return this.has(key) ? this.get(key)! : defaultValue;
  }
  
  toObject(): Record<K, V> {
    return Object.fromEntries(this) as Record<K, V>;
  }
}
```

### WeakMap 类型

```typescript
// WeakMap: 键必须是对象，且是弱引用
const weakMap = new WeakMap<object, string>();

const obj1 = { id: 1 };
const obj2 = { id: 2 };

weakMap.set(obj1, 'value1');
weakMap.set(obj2, 'value2');

console.log(weakMap.get(obj1));  // 'value1'

// 当 obj1 没有其他引用时，会被垃圾回收
// weakMap 中对应的条目也会自动删除

// WeakMap 的限制：
// - 不可迭代（没有 keys(), values(), entries(), forEach()）
// - 没有 size 属性
// - 没有 clear() 方法
```

---

## 使用场景

### 使用 Object 的场景

```typescript
// ============ 1. 结构已知的配置对象 ============
interface AppConfig {
  apiUrl: string;
  timeout: number;
  debug: boolean;
}

const config: AppConfig = {
  apiUrl: 'https://api.example.com',
  timeout: 5000,
  debug: false,
};

// ============ 2. JSON 数据 ============
interface ApiResponse {
  status: number;
  data: {
    id: number;
    name: string;
  };
}

const response: ApiResponse = JSON.parse(jsonString);

// ============ 3. 简单的键值映射（字符串键） ============
const translations: Record<string, string> = {
  hello: '你好',
  goodbye: '再见',
};

// ============ 4. 枚举映射 ============
enum Status {
  Pending = 'pending',
  Active = 'active',
  Completed = 'completed',
}

const statusLabels: Record<Status, string> = {
  [Status.Pending]: '待处理',
  [Status.Active]: '进行中',
  [Status.Completed]: '已完成',
};

// ============ 5. 函数参数对象 ============
interface CreateUserOptions {
  name: string;
  email: string;
  role?: string;
}

function createUser(options: CreateUserOptions): void {
  // ...
}
```

### 使用 Map 的场景

```typescript
// ============ 1. 非字符串键 ============
// DOM 元素作为键
const elementData = new Map<HTMLElement, { clicks: number }>();
const button = document.querySelector('button')!;
elementData.set(button, { clicks: 0 });

// 对象作为键
interface User {
  id: number;
  name: string;
}

const userSessions = new Map<User, { token: string; expires: Date }>();
const user: User = { id: 1, name: 'Alice' };
userSessions.set(user, { token: 'abc123', expires: new Date() });

// ============ 2. 频繁增删操作 ============
class Cache<K, V> {
  private cache = new Map<K, V>();
  private maxSize: number;
  
  constructor(maxSize: number) {
    this.maxSize = maxSize;
  }
  
  set(key: K, value: V): void {
    if (this.cache.size >= this.maxSize) {
      // 删除最早的条目
      const firstKey = this.cache.keys().next().value;
      this.cache.delete(firstKey);
    }
    this.cache.set(key, value);
  }
  
  get(key: K): V | undefined {
    return this.cache.get(key);
  }
}

// ============ 3. 需要保持插入顺序 ============
const orderedSteps = new Map<number, string>();
orderedSteps.set(1, '第一步');
orderedSteps.set(2, '第二步');
orderedSteps.set(3, '第三步');

// 遍历时保持顺序
for (const [step, description] of orderedSteps) {
  console.log(`${step}: ${description}`);
}

// ============ 4. 需要频繁获取大小 ============
function processUntilEmpty<K, V>(map: Map<K, V>): void {
  while (map.size > 0) {  // O(1) 获取大小
    const [key] = map.keys().next().value;
    // 处理...
    map.delete(key);
  }
}

// ============ 5. 避免原型链污染 ============
// 用户输入作为键时更安全
const userInputMap = new Map<string, any>();

function setUserValue(key: string, value: any): void {
  // Map 不会有 __proto__ 等原型链问题
  userInputMap.set(key, value);
}

// ============ 6. 私有数据存储 ============
const privateData = new WeakMap<object, { secret: string }>();

class SecretHolder {
  constructor(secret: string) {
    privateData.set(this, { secret });
  }
  
  getSecret(): string {
    return privateData.get(this)!.secret;
  }
}

// ============ 7. 双向映射 ============
class BiMap<K, V> {
  private forward = new Map<K, V>();
  private reverse = new Map<V, K>();
  
  set(key: K, value: V): void {
    this.forward.set(key, value);
    this.reverse.set(value, key);
  }
  
  getByKey(key: K): V | undefined {
    return this.forward.get(key);
  }
  
  getByValue(value: V): K | undefined {
    return this.reverse.get(value);
  }
}
```

---

## 常见操作对比

### 合并

```typescript
// ============ Object 合并 ============
const obj1 = { a: 1, b: 2 };
const obj2 = { b: 3, c: 4 };

// 展开运算符
const merged1 = { ...obj1, ...obj2 };  // { a: 1, b: 3, c: 4 }

// Object.assign
const merged2 = Object.assign({}, obj1, obj2);

// ============ Map 合并 ============
const map1 = new Map([['a', 1], ['b', 2]]);
const map2 = new Map([['b', 3], ['c', 4]]);

// 展开运算符
const merged3 = new Map([...map1, ...map2]);  // a=1, b=3, c=4

// 手动合并
function mergeMaps<K, V>(...maps: Map<K, V>[]): Map<K, V> {
  const result = new Map<K, V>();
  for (const map of maps) {
    for (const [key, value] of map) {
      result.set(key, value);
    }
  }
  return result;
}
```

### 过滤

```typescript
// ============ Object 过滤 ============
const obj = { a: 1, b: 2, c: 3, d: 4 };

// 过滤值大于 2 的
const filtered1 = Object.fromEntries(
  Object.entries(obj).filter(([_, value]) => value > 2)
);  // { c: 3, d: 4 }

// ============ Map 过滤 ============
const map = new Map([['a', 1], ['b', 2], ['c', 3], ['d', 4]]);

const filtered2 = new Map(
  [...map].filter(([_, value]) => value > 2)
);  // Map { 'c' => 3, 'd' => 4 }

// 工具函数
function filterMap<K, V>(
  map: Map<K, V>,
  predicate: (value: V, key: K) => boolean
): Map<K, V> {
  const result = new Map<K, V>();
  for (const [key, value] of map) {
    if (predicate(value, key)) {
      result.set(key, value);
    }
  }
  return result;
}
```

### 映射转换

```typescript
// ============ Object 映射 ============
const obj = { a: 1, b: 2, c: 3 };

// 值翻倍
const mapped1 = Object.fromEntries(
  Object.entries(obj).map(([key, value]) => [key, value * 2])
);  // { a: 2, b: 4, c: 6 }

// ============ Map 映射 ============
const map = new Map([['a', 1], ['b', 2], ['c', 3]]);

const mapped2 = new Map(
  [...map].map(([key, value]) => [key, value * 2])
);  // Map { 'a' => 2, 'b' => 4, 'c' => 6 }

// 工具函数
function mapValues<K, V, U>(
  map: Map<K, V>,
  fn: (value: V, key: K) => U
): Map<K, U> {
  const result = new Map<K, U>();
  for (const [key, value] of map) {
    result.set(key, fn(value, key));
  }
  return result;
}
```

### 查找

```typescript
// ============ Object 查找 ============
const obj = { a: 1, b: 2, c: 3 };

// 查找值
const found1 = Object.entries(obj).find(([_, v]) => v === 2);  // ['b', 2]

// 查找键
const key1 = Object.keys(obj).find(k => obj[k] === 2);  // 'b'

// ============ Map 查找 ============
const map = new Map([['a', 1], ['b', 2], ['c', 3]]);

// 查找条目
const found2 = [...map].find(([_, v]) => v === 2);  // ['b', 2]

// 查找键
function findKey<K, V>(map: Map<K, V>, predicate: (v: V) => boolean): K | undefined {
  for (const [key, value] of map) {
    if (predicate(value)) return key;
  }
  return undefined;
}
```

### 分组

```typescript
interface Item {
  category: string;
  name: string;
  price: number;
}

const items: Item[] = [
  { category: 'fruit', name: 'apple', price: 1 },
  { category: 'fruit', name: 'banana', price: 0.5 },
  { category: 'vegetable', name: 'carrot', price: 0.8 },
];

// ============ 使用 Object 分组 ============
const groupedObj = items.reduce<Record<string, Item[]>>((acc, item) => {
  if (!acc[item.category]) {
    acc[item.category] = [];
  }
  acc[item.category].push(item);
  return acc;
}, {});

// 或使用 Object.groupBy (ES2024)
// const groupedObj = Object.groupBy(items, item => item.category);

// ============ 使用 Map 分组 ============
const groupedMap = items.reduce<Map<string, Item[]>>((acc, item) => {
  const group = acc.get(item.category) || [];
  group.push(item);
  acc.set(item.category, group);
  return acc;
}, new Map());

// 或使用 Map.groupBy (ES2024)
// const groupedMap = Map.groupBy(items, item => item.category);

// 工具函数
function groupBy<T, K>(items: T[], keyFn: (item: T) => K): Map<K, T[]> {
  const map = new Map<K, T[]>();
  for (const item of items) {
    const key = keyFn(item);
    const group = map.get(key) || [];
    group.push(item);
    map.set(key, group);
  }
  return map;
}
```

---

## 最佳实践

### 选择指南

```typescript
// ✅ 使用 Object 当：
// 1. 键是已知的、固定的字符串
// 2. 需要 JSON 序列化
// 3. 数据结构简单，性能不是关键问题
// 4. 与现有 API 兼容

interface UserSettings {
  theme: 'light' | 'dark';
  language: string;
  notifications: boolean;
}

// ✅ 使用 Map 当：
// 1. 键不是字符串（对象、函数等）
// 2. 需要保持插入顺序
// 3. 频繁增删操作
// 4. 需要快速获取大小
// 5. 键可能与 Object 原型冲突

const componentState = new Map<React.Component, { mounted: boolean }>();
```

### 工具函数库

```typescript
// ============ Object 工具函数 ============

/** 安全获取嵌套属性 */
function getPath<T>(obj: Record<string, any>, path: string, defaultValue?: T): T {
  const keys = path.split('.');
  let result: any = obj;
  for (const key of keys) {
    result = result?.[key];
    if (result === undefined) return defaultValue as T;
  }
  return result;
}

/** 深度冻结对象 */
function deepFreeze<T extends object>(obj: T): Readonly<T> {
  Object.freeze(obj);
  Object.getOwnPropertyNames(obj).forEach(prop => {
    const value = (obj as any)[prop];
    if (value && typeof value === 'object') {
      deepFreeze(value);
    }
  });
  return obj;
}

/** 对象深拷贝 */
function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') return obj;
  if (obj instanceof Date) return new Date(obj.getTime()) as T;
  if (obj instanceof Map) return new Map(obj) as T;
  if (obj instanceof Set) return new Set(obj) as T;
  if (Array.isArray(obj)) return obj.map(deepClone) as T;
  
  const cloned = {} as T;
  for (const key in obj) {
    if (obj.hasOwnProperty(key)) {
      cloned[key] = deepClone(obj[key]);
    }
  }
  return cloned;
}

// ============ Map 工具函数 ============

/** Map 转 Object */
function mapToObject<V>(map: Map<string, V>): Record<string, V> {
  return Object.fromEntries(map);
}

/** Object 转 Map */
function objectToMap<V>(obj: Record<string, V>): Map<string, V> {
  return new Map(Object.entries(obj));
}

/** 获取或设置默认值 */
function getOrSet<K, V>(map: Map<K, V>, key: K, defaultValue: V | (() => V)): V {
  if (map.has(key)) {
    return map.get(key)!;
  }
  const value = typeof defaultValue === 'function' 
    ? (defaultValue as () => V)() 
    : defaultValue;
  map.set(key, value);
  return value;
}

/** Map 转换为数组 */
function mapToArray<K, V, R>(map: Map<K, V>, fn: (key: K, value: V) => R): R[] {
  const result: R[] = [];
  for (const [key, value] of map) {
    result.push(fn(key, value));
  }
  return result;
}

/** 从数组创建 Map */
function arrayToMap<T, K, V>(
  array: T[],
  keyFn: (item: T) => K,
  valueFn: (item: T) => V = (item) => item as unknown as V
): Map<K, V> {
  const map = new Map<K, V>();
  for (const item of array) {
    map.set(keyFn(item), valueFn(item));
  }
  return map;
}
```

### 类型安全的 Map 包装器

```typescript
class TypeSafeMap<K, V> {
  private map = new Map<K, V>();
  
  constructor(entries?: readonly (readonly [K, V])[]) {
    if (entries) {
      for (const [key, value] of entries) {
        this.map.set(key, value);
      }
    }
  }
  
  get size(): number {
    return this.map.size;
  }
  
  set(key: K, value: V): this {
    this.map.set(key, value);
    return this;
  }
  
  get(key: K): V | undefined {
    return this.map.get(key);
  }
  
  getOrThrow(key: K): V {
    if (!this.map.has(key)) {
      throw new Error(`Key not found: ${String(key)}`);
    }
    return this.map.get(key)!;
  }
  
  getOrDefault(key: K, defaultValue: V): V {
    return this.map.has(key) ? this.map.get(key)! : defaultValue;
  }
  
  has(key: K): boolean {
    return this.map.has(key);
  }
  
  delete(key: K): boolean {
    return this.map.delete(key);
  }
  
  clear(): void {
    this.map.clear();
  }
  
  keys(): IterableIterator<K> {
    return this.map.keys();
  }
  
  values(): IterableIterator<V> {
    return this.map.values();
  }
  
  entries(): IterableIterator<[K, V]> {
    return this.map.entries();
  }
  
  forEach(callback: (value: V, key: K, map: this) => void): void {
    this.map.forEach((value, key) => callback(value, key, this));
  }
  
  map<U>(fn: (value: V, key: K) => U): TypeSafeMap<K, U> {
    const result = new TypeSafeMap<K, U>();
    for (const [key, value] of this.map) {
      result.set(key, fn(value, key));
    }
    return result;
  }
  
  filter(predicate: (value: V, key: K) => boolean): TypeSafeMap<K, V> {
    const result = new TypeSafeMap<K, V>();
    for (const [key, value] of this.map) {
      if (predicate(value, key)) {
        result.set(key, value);
      }
    }
    return result;
  }
  
  toObject(): K extends string ? Record<K, V> : never {
    return Object.fromEntries(this.map) as any;
  }
  
  toArray(): [K, V][] {
    return [...this.map];
  }
  
  [Symbol.iterator](): IterableIterator<[K, V]> {
    return this.map[Symbol.iterator]();
  }
}

// 使用示例
const users = new TypeSafeMap<number, { name: string }>([
  [1, { name: 'Alice' }],
  [2, { name: 'Bob' }],
]);

const user = users.getOrThrow(1);  // { name: 'Alice' }
const names = users.map(u => u.name);  // TypeSafeMap<number, string>
```

---

## 总结

| 特性 | Object | Map |
|------|--------|-----|
| 键类型 | string / Symbol | 任意类型 |
| 键顺序 | 部分保证 | 插入顺序 |
| 大小获取 | O(n) | O(1) |
| 默认键 | 有原型链 | 无 |
| 可迭代 | 需转换 | 原生支持 |
| JSON 支持 | 原生支持 | 需转换 |
| 性能（增删） | 较慢 | 更快 |
| 内存占用 | 较少 | 较多 |
| TypeScript | 索引签名/Record | Map<K, V> |

**选择建议：**
- 简单配置、JSON 数据 → Object
- 动态键值、频繁增删、非字符串键 → Map

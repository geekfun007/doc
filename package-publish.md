# 多语言包创建与发布指南

本文档详细介绍 TypeScript、Python、Go 和 Rust 包的创建、发布流程，以及自托管私有仓库的配置。

## 目录

1. [TypeScript/npm 包](#typescriptnpm-包)
2. [Python 包](#python-包)
3. [Go 模块](#go-模块)
4. [Rust Crate](#rust-crate)
5. [自托管私有仓库](#自托管私有仓库)
6. [CI/CD 自动发布](#cicd-自动发布)

---

## TypeScript/npm 包

### 1. 项目初始化

```bash
# 创建项目目录
mkdir my-ts-package && cd my-ts-package

# 初始化 npm 项目
npm init -y

# 安装 TypeScript 和相关依赖
npm install -D typescript @types/node ts-node

# 初始化 TypeScript 配置
npx tsc --init
```

### 2. 项目结构

```
my-ts-package/
├── src/
│   ├── index.ts          # 主入口
│   ├── utils/
│   │   └── helper.ts
│   └── types/
│       └── index.ts
├── dist/                  # 编译输出
├── tests/
│   └── index.test.ts
├── package.json
├── tsconfig.json
├── tsconfig.build.json
├── .npmignore
├── .gitignore
├── README.md
├── LICENSE
└── CHANGELOG.md
```

### 3. package.json 配置

```json
{
  "name": "@myorg/my-package",
  "version": "1.0.0",
  "description": "My awesome TypeScript package",
  "main": "dist/index.js",
  "module": "dist/index.mjs",
  "types": "dist/index.d.ts",
  "exports": {
    ".": {
      "import": "./dist/index.mjs",
      "require": "./dist/index.js",
      "types": "./dist/index.d.ts"
    },
    "./utils": {
      "import": "./dist/utils/index.mjs",
      "require": "./dist/utils/index.js",
      "types": "./dist/utils/index.d.ts"
    }
  },
  "files": [
    "dist",
    "README.md",
    "LICENSE"
  ],
  "scripts": {
    "build": "tsup src/index.ts --format cjs,esm --dts --clean",
    "build:tsc": "tsc -p tsconfig.build.json",
    "test": "vitest run",
    "test:watch": "vitest",
    "lint": "eslint src --ext .ts",
    "prepublishOnly": "npm run build && npm run test",
    "version": "npm run build"
  },
  "keywords": ["typescript", "utility"],
  "author": "Your Name <your@email.com>",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/username/my-package.git"
  },
  "bugs": {
    "url": "https://github.com/username/my-package/issues"
  },
  "homepage": "https://github.com/username/my-package#readme",
  "engines": {
    "node": ">=16.0.0"
  },
  "publishConfig": {
    "access": "public",
    "registry": "https://registry.npmjs.org/"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "tsup": "^8.0.0",
    "vitest": "^1.0.0",
    "@types/node": "^20.0.0",
    "eslint": "^8.0.0",
    "@typescript-eslint/eslint-plugin": "^6.0.0",
    "@typescript-eslint/parser": "^6.0.0"
  },
  "peerDependencies": {
    "typescript": ">=4.7.0"
  },
  "peerDependenciesMeta": {
    "typescript": {
      "optional": true
    }
  }
}
```

### 4. tsconfig.json 配置

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020"],
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "moduleResolution": "node",
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "isolatedModules": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "tests"]
}
```

### 5. 源代码示例

```typescript
// src/index.ts
export * from './types';
export * from './utils/helper';

export interface Config {
  debug?: boolean;
  timeout?: number;
}

export class MyPackage {
  private config: Config;

  constructor(config: Config = {}) {
    this.config = {
      debug: false,
      timeout: 5000,
      ...config,
    };
  }

  public greet(name: string): string {
    if (this.config.debug) {
      console.log(`[DEBUG] Greeting: ${name}`);
    }
    return `Hello, ${name}!`;
  }
}

export default MyPackage;
```

```typescript
// src/types/index.ts
export interface User {
  id: string;
  name: string;
  email: string;
}

export type UserRole = 'admin' | 'user' | 'guest';

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}
```

```typescript
// src/utils/helper.ts
export function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

export function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

export function deepClone<T>(obj: T): T {
  return JSON.parse(JSON.stringify(obj));
}
```

### 6. 发布到 npm

```bash
# 登录 npm
npm login

# 检查包名是否可用
npm search @myorg/my-package

# 查看将发布的文件
npm pack --dry-run

# 发布（公开包）
npm publish --access public

# 发布 beta 版本
npm version prerelease --preid=beta
npm publish --tag beta

# 发布到私有仓库
npm publish --registry https://npm.mycompany.com
```

### 7. .npmrc 配置

```ini
# ~/.npmrc 或项目根目录 .npmrc

# 默认 registry
registry=https://registry.npmjs.org/

# 私有 scope 使用私有仓库
@myorg:registry=https://npm.mycompany.com/
//npm.mycompany.com/:_authToken=${NPM_TOKEN}

# 代理设置（可选）
# proxy=http://proxy.company.com:8080
# https-proxy=http://proxy.company.com:8080
```

---

## Python 包

### 1. 项目初始化

```bash
# 创建项目目录
mkdir my-python-package && cd my-python-package

# 创建虚拟环境
python -m venv venv
source venv/bin/activate  # Linux/Mac
# venv\Scripts\activate   # Windows

# 安装构建工具
pip install build twine hatch
```

### 2. 项目结构

```
my-python-package/
├── src/
│   └── my_package/
│       ├── __init__.py
│       ├── core.py
│       ├── utils.py
│       └── py.typed          # PEP 561 类型标记
├── tests/
│   ├── __init__.py
│   ├── test_core.py
│   └── test_utils.py
├── docs/
│   └── index.md
├── pyproject.toml            # 主配置文件
├── setup.py                  # 可选，向后兼容
├── MANIFEST.in
├── README.md
├── LICENSE
├── CHANGELOG.md
└── .gitignore
```

### 3. pyproject.toml 配置

```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project]
name = "my-package"
version = "1.0.0"
description = "My awesome Python package"
readme = "README.md"
license = {file = "LICENSE"}
authors = [
    {name = "Your Name", email = "your@email.com"}
]
maintainers = [
    {name = "Your Name", email = "your@email.com"}
]
keywords = ["utility", "helper"]
classifiers = [
    "Development Status :: 4 - Beta",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Operating System :: OS Independent",
    "Programming Language :: Python :: 3",
    "Programming Language :: Python :: 3.9",
    "Programming Language :: Python :: 3.10",
    "Programming Language :: Python :: 3.11",
    "Programming Language :: Python :: 3.12",
    "Typing :: Typed",
]
requires-python = ">=3.9"
dependencies = [
    "requests>=2.28.0",
    "pydantic>=2.0.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.0.0",
    "pytest-cov>=4.0.0",
    "pytest-asyncio>=0.21.0",
    "mypy>=1.0.0",
    "ruff>=0.1.0",
    "black>=23.0.0",
]
docs = [
    "mkdocs>=1.5.0",
    "mkdocs-material>=9.0.0",
    "mkdocstrings[python]>=0.24.0",
]

[project.urls]
Homepage = "https://github.com/username/my-package"
Documentation = "https://my-package.readthedocs.io"
Repository = "https://github.com/username/my-package.git"
Issues = "https://github.com/username/my-package/issues"
Changelog = "https://github.com/username/my-package/blob/main/CHANGELOG.md"

[project.scripts]
my-cli = "my_package.cli:main"

[project.entry-points."my_package.plugins"]
plugin1 = "my_package.plugins.plugin1:Plugin1"

[tool.hatch.build.targets.sdist]
include = [
    "/src",
    "/tests",
]

[tool.hatch.build.targets.wheel]
packages = ["src/my_package"]

[tool.hatch.version]
path = "src/my_package/__init__.py"

# ============ 工具配置 ============

[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py"]
addopts = "-v --cov=my_package --cov-report=term-missing"

[tool.mypy]
python_version = "3.9"
strict = true
warn_return_any = true
warn_unused_ignores = true
disallow_untyped_defs = true

[tool.ruff]
target-version = "py39"
line-length = 88
select = [
    "E",   # pycodestyle errors
    "W",   # pycodestyle warnings
    "F",   # pyflakes
    "I",   # isort
    "B",   # flake8-bugbear
    "C4",  # flake8-comprehensions
    "UP",  # pyupgrade
]
ignore = ["E501"]

[tool.ruff.isort]
known-first-party = ["my_package"]

[tool.black]
line-length = 88
target-version = ["py39", "py310", "py311"]
include = '\.pyi?$'
```

### 4. 源代码示例

```python
# src/my_package/__init__.py
"""My awesome Python package."""

from .core import MyClass, process_data
from .utils import format_date, deep_merge

__version__ = "1.0.0"
__all__ = ["MyClass", "process_data", "format_date", "deep_merge"]
```

```python
# src/my_package/core.py
"""Core functionality."""

from typing import Any, Dict, List, Optional
from dataclasses import dataclass
from pydantic import BaseModel


class Config(BaseModel):
    """Configuration model."""
    
    debug: bool = False
    timeout: int = 5000
    retries: int = 3


@dataclass
class Result:
    """Result container."""
    
    success: bool
    data: Optional[Any] = None
    error: Optional[str] = None


class MyClass:
    """Main class of the package."""
    
    def __init__(self, config: Optional[Config] = None) -> None:
        """Initialize with optional config.
        
        Args:
            config: Configuration object
        """
        self.config = config or Config()
    
    def greet(self, name: str) -> str:
        """Generate a greeting message.
        
        Args:
            name: The name to greet
            
        Returns:
            A greeting string
        """
        if self.config.debug:
            print(f"[DEBUG] Greeting: {name}")
        return f"Hello, {name}!"
    
    def process(self, data: Dict[str, Any]) -> Result:
        """Process input data.
        
        Args:
            data: Input dictionary
            
        Returns:
            Result object with processed data
        """
        try:
            processed = {k: str(v).upper() for k, v in data.items()}
            return Result(success=True, data=processed)
        except Exception as e:
            return Result(success=False, error=str(e))


def process_data(items: List[Any]) -> List[Any]:
    """Process a list of items.
    
    Args:
        items: List of items to process
        
    Returns:
        Processed list
    """
    return [str(item).strip() for item in items if item is not None]
```

```python
# src/my_package/utils.py
"""Utility functions."""

from datetime import datetime
from typing import Any, Dict


def format_date(dt: datetime, fmt: str = "%Y-%m-%d") -> str:
    """Format a datetime object.
    
    Args:
        dt: Datetime to format
        fmt: Format string
        
    Returns:
        Formatted date string
    """
    return dt.strftime(fmt)


def deep_merge(base: Dict[str, Any], override: Dict[str, Any]) -> Dict[str, Any]:
    """Deep merge two dictionaries.
    
    Args:
        base: Base dictionary
        override: Dictionary with override values
        
    Returns:
        Merged dictionary
    """
    result = base.copy()
    
    for key, value in override.items():
        if (
            key in result
            and isinstance(result[key], dict)
            and isinstance(value, dict)
        ):
            result[key] = deep_merge(result[key], value)
        else:
            result[key] = value
    
    return result
```

### 5. 构建与发布

```bash
# 构建包
python -m build

# 检查包
twine check dist/*

# 发布到 TestPyPI（测试）
twine upload --repository testpypi dist/*

# 发布到 PyPI
twine upload dist/*

# 使用 token 发布
twine upload -u __token__ -p $PYPI_TOKEN dist/*
```

### 6. pip 配置 (~/.pip/pip.conf)

```ini
[global]
index-url = https://pypi.org/simple/
extra-index-url = https://pypi.mycompany.com/simple/
trusted-host = pypi.mycompany.com

[install]
trusted-host = pypi.mycompany.com
```

### 7. .pypirc 配置 (~/.pypirc)

```ini
[distutils]
index-servers =
    pypi
    testpypi
    private

[pypi]
repository = https://upload.pypi.org/legacy/
username = __token__
password = pypi-xxx...

[testpypi]
repository = https://test.pypi.org/legacy/
username = __token__
password = pypi-xxx...

[private]
repository = https://pypi.mycompany.com/
username = myuser
password = mypassword
```

---

## Go 模块

### 1. 项目初始化

```bash
# 创建项目目录
mkdir my-go-package && cd my-go-package

# 初始化 Go 模块
go mod init github.com/username/my-go-package

# 创建目录结构
mkdir -p cmd pkg internal examples
```

### 2. 项目结构

```
my-go-package/
├── cmd/
│   └── mycli/
│       └── main.go           # CLI 入口
├── pkg/
│   └── mypackage/
│       ├── mypackage.go      # 公开 API
│       ├── types.go
│       ├── utils.go
│       └── mypackage_test.go
├── internal/
│   └── helper/
│       └── helper.go         # 内部包
├── examples/
│   └── basic/
│       └── main.go
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── CHANGELOG.md
├── Makefile
└── .goreleaser.yaml
```

### 3. go.mod 配置

```go
module github.com/username/my-go-package

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/stretchr/testify v1.8.4
)

require (
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/spf13/pflag v1.0.5 // indirect
)
```

### 4. 源代码示例

```go
// pkg/mypackage/mypackage.go
package mypackage

import (
    "fmt"
    "sync"
)

// Version is the current version of the package.
const Version = "1.0.0"

// Config holds the configuration for MyPackage.
type Config struct {
    Debug   bool
    Timeout int
    Retries int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
    return Config{
        Debug:   false,
        Timeout: 5000,
        Retries: 3,
    }
}

// MyPackage is the main struct of the package.
type MyPackage struct {
    config Config
    mu     sync.RWMutex
}

// New creates a new MyPackage instance with the given config.
func New(config Config) *MyPackage {
    return &MyPackage{
        config: config,
    }
}

// NewWithDefaults creates a new MyPackage with default configuration.
func NewWithDefaults() *MyPackage {
    return New(DefaultConfig())
}

// Greet returns a greeting message for the given name.
func (m *MyPackage) Greet(name string) string {
    if m.config.Debug {
        fmt.Printf("[DEBUG] Greeting: %s\n", name)
    }
    return fmt.Sprintf("Hello, %s!", name)
}

// Process processes the input data and returns the result.
func (m *MyPackage) Process(data map[string]interface{}) (*Result, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    processed := make(map[string]string)
    for k, v := range data {
        processed[k] = fmt.Sprintf("%v", v)
    }
    
    return &Result{
        Success: true,
        Data:    processed,
    }, nil
}

// SetConfig updates the configuration.
func (m *MyPackage) SetConfig(config Config) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.config = config
}
```

```go
// pkg/mypackage/types.go
package mypackage

// Result represents the result of an operation.
type Result struct {
    Success bool
    Data    map[string]string
    Error   error
}

// User represents a user entity.
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Option is a function that modifies Config.
type Option func(*Config)

// WithDebug sets the debug mode.
func WithDebug(debug bool) Option {
    return func(c *Config) {
        c.Debug = debug
    }
}

// WithTimeout sets the timeout.
func WithTimeout(timeout int) Option {
    return func(c *Config) {
        c.Timeout = timeout
    }
}

// NewWithOptions creates a new MyPackage with functional options.
func NewWithOptions(opts ...Option) *MyPackage {
    config := DefaultConfig()
    for _, opt := range opts {
        opt(&config)
    }
    return New(config)
}
```

```go
// pkg/mypackage/utils.go
package mypackage

import (
    "encoding/json"
    "time"
)

// FormatDate formats a time.Time to a date string.
func FormatDate(t time.Time) string {
    return t.Format("2006-01-02")
}

// FormatDateTime formats a time.Time to a datetime string.
func FormatDateTime(t time.Time) string {
    return t.Format(time.RFC3339)
}

// ToJSON converts a value to JSON string.
func ToJSON(v interface{}) (string, error) {
    b, err := json.Marshal(v)
    if err != nil {
        return "", err
    }
    return string(b), nil
}

// FromJSON parses JSON string to a value.
func FromJSON[T any](s string) (T, error) {
    var result T
    err := json.Unmarshal([]byte(s), &result)
    return result, err
}
```

```go
// pkg/mypackage/mypackage_test.go
package mypackage

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
    pkg := NewWithDefaults()
    assert.NotNil(t, pkg)
}

func TestGreet(t *testing.T) {
    pkg := NewWithDefaults()
    
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"simple", "World", "Hello, World!"},
        {"with spaces", "Go Developer", "Hello, Go Developer!"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := pkg.Greet(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestProcess(t *testing.T) {
    pkg := NewWithDefaults()
    
    data := map[string]interface{}{
        "name": "test",
        "count": 42,
    }
    
    result, err := pkg.Process(data)
    require.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "test", result.Data["name"])
    assert.Equal(t, "42", result.Data["count"])
}

func TestWithOptions(t *testing.T) {
    pkg := NewWithOptions(
        WithDebug(true),
        WithTimeout(10000),
    )
    
    assert.NotNil(t, pkg)
}
```

### 5. 发布 Go 模块

```bash
# 确保代码已提交到 Git
git add .
git commit -m "feat: initial release"

# 创建版本标签
git tag v1.0.0
git push origin main
git push origin v1.0.0

# 让 pkg.go.dev 索引你的包
GOPROXY=proxy.golang.org go list -m github.com/username/my-go-package@v1.0.0

# 发布新版本
git tag v1.1.0
git push origin v1.1.0
```

### 6. 使用私有模块

```bash
# 设置 GOPRIVATE 环境变量
export GOPRIVATE=github.com/mycompany/*,gitlab.mycompany.com/*

# 设置 Git 认证
git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# 或使用 SSH
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

### 7. go.env 配置

```bash
# ~/.config/go/env 或通过 go env -w 设置

# 私有模块
GOPRIVATE=github.com/mycompany/*,*.mycompany.com

# 代理设置（中国用户）
GOPROXY=https://goproxy.cn,https://goproxy.io,direct

# 禁用校验（私有模块）
GONOSUMDB=github.com/mycompany/*

# 或配置私有代理
GOPROXY=https://goproxy.mycompany.com,https://proxy.golang.org,direct
```

---

## Rust Crate

### 1. 项目初始化

```bash
# 创建库项目
cargo new my-rust-crate --lib
cd my-rust-crate

# 或创建二进制项目
cargo new my-rust-cli

# 添加工作空间（可选）
mkdir -p crates
```

### 2. 项目结构

```
my-rust-crate/
├── src/
│   ├── lib.rs            # 库入口
│   ├── config.rs
│   ├── error.rs
│   ├── utils/
│   │   ├── mod.rs
│   │   └── helpers.rs
│   └── types/
│       ├── mod.rs
│       └── user.rs
├── examples/
│   └── basic.rs
├── benches/
│   └── benchmark.rs
├── tests/
│   └── integration_test.rs
├── Cargo.toml
├── Cargo.lock
├── README.md
├── LICENSE
├── CHANGELOG.md
└── .cargo/
    └── config.toml
```

### 3. Cargo.toml 配置

```toml
[package]
name = "my-rust-crate"
version = "1.0.0"
edition = "2021"
rust-version = "1.70"
authors = ["Your Name <your@email.com>"]
description = "My awesome Rust crate"
documentation = "https://docs.rs/my-rust-crate"
readme = "README.md"
homepage = "https://github.com/username/my-rust-crate"
repository = "https://github.com/username/my-rust-crate"
license = "MIT OR Apache-2.0"
keywords = ["utility", "helper", "tools"]
categories = ["development-tools", "utilities"]
exclude = [
    ".github/*",
    "benches/*",
    "tests/*",
]
include = [
    "src/**/*",
    "Cargo.toml",
    "README.md",
    "LICENSE*",
]

[package.metadata.docs.rs]
all-features = true
rustdoc-args = ["--cfg", "docsrs"]

# ============ 依赖 ============

[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
thiserror = "1.0"
tracing = "0.1"

# 可选依赖
tokio = { version = "1.0", features = ["full"], optional = true }
async-trait = { version = "0.1", optional = true }

[dev-dependencies]
tokio = { version = "1.0", features = ["full", "test-util"] }
criterion = { version = "0.5", features = ["html_reports"] }
pretty_assertions = "1.4"

[build-dependencies]
# 构建脚本依赖

# ============ Features ============

[features]
default = ["std"]
std = []
async = ["tokio", "async-trait"]
full = ["std", "async"]

# ============ Targets ============

[lib]
name = "my_rust_crate"
path = "src/lib.rs"

[[bin]]
name = "my-cli"
path = "src/bin/main.rs"
required-features = ["std"]

[[example]]
name = "basic"
path = "examples/basic.rs"

[[bench]]
name = "benchmark"
harness = false

# ============ Profile ============

[profile.release]
lto = true
codegen-units = 1
panic = "abort"
strip = true

[profile.dev]
opt-level = 0
debug = true

[profile.bench]
lto = true
```

### 4. 源代码示例

```rust
// src/lib.rs
//! My awesome Rust crate
//!
//! This crate provides utilities for doing awesome things.
//!
//! # Examples
//!
//! ```rust
//! use my_rust_crate::{MyStruct, Config};
//!
//! let config = Config::default();
//! let instance = MyStruct::new(config);
//! let greeting = instance.greet("World");
//! assert_eq!(greeting, "Hello, World!");
//! ```

#![cfg_attr(docsrs, feature(doc_cfg))]
#![warn(missing_docs)]
#![warn(rustdoc::missing_crate_level_docs)]

pub mod config;
pub mod error;
pub mod types;
pub mod utils;

pub use config::Config;
pub use error::{Error, Result};
pub use types::User;

use std::collections::HashMap;

/// The main struct of the crate.
#[derive(Debug, Clone)]
pub struct MyStruct {
    config: Config,
}

impl MyStruct {
    /// Creates a new instance with the given configuration.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use my_rust_crate::{MyStruct, Config};
    ///
    /// let my_struct = MyStruct::new(Config::default());
    /// ```
    pub fn new(config: Config) -> Self {
        Self { config }
    }

    /// Creates a new instance with default configuration.
    pub fn with_defaults() -> Self {
        Self::new(Config::default())
    }

    /// Generates a greeting message.
    ///
    /// # Arguments
    ///
    /// * `name` - The name to greet
    ///
    /// # Returns
    ///
    /// A greeting string
    pub fn greet(&self, name: &str) -> String {
        if self.config.debug {
            tracing::debug!("Greeting: {}", name);
        }
        format!("Hello, {}!", name)
    }

    /// Processes input data and returns the result.
    ///
    /// # Errors
    ///
    /// Returns an error if processing fails.
    pub fn process(&self, data: HashMap<String, String>) -> Result<ProcessResult> {
        let processed: HashMap<String, String> = data
            .into_iter()
            .map(|(k, v)| (k, v.to_uppercase()))
            .collect();

        Ok(ProcessResult {
            success: true,
            data: processed,
        })
    }
}

impl Default for MyStruct {
    fn default() -> Self {
        Self::with_defaults()
    }
}

/// Result of a processing operation.
#[derive(Debug, Clone)]
pub struct ProcessResult {
    /// Whether the operation was successful.
    pub success: bool,
    /// The processed data.
    pub data: HashMap<String, String>,
}

/// Builder for MyStruct.
#[derive(Debug, Default)]
pub struct MyStructBuilder {
    debug: bool,
    timeout: u64,
}

impl MyStructBuilder {
    /// Creates a new builder.
    pub fn new() -> Self {
        Self::default()
    }

    /// Sets the debug mode.
    pub fn debug(mut self, debug: bool) -> Self {
        self.debug = debug;
        self
    }

    /// Sets the timeout.
    pub fn timeout(mut self, timeout: u64) -> Self {
        self.timeout = timeout;
        self
    }

    /// Builds the MyStruct instance.
    pub fn build(self) -> MyStruct {
        MyStruct::new(Config {
            debug: self.debug,
            timeout: self.timeout,
            ..Config::default()
        })
    }
}

// 异步支持（可选 feature）
#[cfg(feature = "async")]
#[cfg_attr(docsrs, doc(cfg(feature = "async")))]
pub mod async_api {
    //! Async API for the crate.
    
    use super::*;
    
    impl MyStruct {
        /// Async version of process.
        pub async fn process_async(
            &self,
            data: HashMap<String, String>,
        ) -> Result<ProcessResult> {
            // 模拟异步操作
            tokio::time::sleep(std::time::Duration::from_millis(10)).await;
            self.process(data)
        }
    }
}
```

```rust
// src/config.rs
//! Configuration module.

use serde::{Deserialize, Serialize};

/// Configuration for MyStruct.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    /// Enable debug mode.
    pub debug: bool,
    /// Timeout in milliseconds.
    pub timeout: u64,
    /// Number of retries.
    pub retries: u32,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            debug: false,
            timeout: 5000,
            retries: 3,
        }
    }
}

impl Config {
    /// Creates a new Config with the given debug setting.
    pub fn with_debug(debug: bool) -> Self {
        Self {
            debug,
            ..Self::default()
        }
    }
    
    /// Loads config from JSON string.
    pub fn from_json(json: &str) -> crate::Result<Self> {
        serde_json::from_str(json).map_err(Into::into)
    }
    
    /// Serializes config to JSON string.
    pub fn to_json(&self) -> crate::Result<String> {
        serde_json::to_string_pretty(self).map_err(Into::into)
    }
}
```

```rust
// src/error.rs
//! Error types for the crate.

use thiserror::Error;

/// Result type for this crate.
pub type Result<T> = std::result::Result<T, Error>;

/// Error type for this crate.
#[derive(Debug, Error)]
pub enum Error {
    /// IO error.
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
    
    /// JSON serialization/deserialization error.
    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),
    
    /// Configuration error.
    #[error("Config error: {message}")]
    Config {
        /// Error message.
        message: String,
    },
    
    /// Invalid input error.
    #[error("Invalid input: {0}")]
    InvalidInput(String),
    
    /// Custom error with a message.
    #[error("{0}")]
    Custom(String),
}

impl Error {
    /// Creates a new config error.
    pub fn config(message: impl Into<String>) -> Self {
        Self::Config {
            message: message.into(),
        }
    }
    
    /// Creates a new custom error.
    pub fn custom(message: impl Into<String>) -> Self {
        Self::Custom(message.into())
    }
}
```

```rust
// src/types/mod.rs
//! Type definitions.

mod user;

pub use user::User;
```

```rust
// src/types/user.rs
//! User type.

use serde::{Deserialize, Serialize};

/// Represents a user.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct User {
    /// User ID.
    pub id: String,
    /// User name.
    pub name: String,
    /// User email.
    pub email: String,
}

impl User {
    /// Creates a new user.
    pub fn new(id: impl Into<String>, name: impl Into<String>, email: impl Into<String>) -> Self {
        Self {
            id: id.into(),
            name: name.into(),
            email: email.into(),
        }
    }
}
```

### 5. 发布到 crates.io

```bash
# 登录 crates.io
cargo login <your-api-token>

# 检查包
cargo package --list
cargo package

# 发布
cargo publish

# 发布到私有仓库
cargo publish --registry my-registry

# Dry run（不实际发布）
cargo publish --dry-run
```

### 6. .cargo/config.toml 配置

```toml
# .cargo/config.toml

# 私有仓库配置
[registries.my-registry]
index = "sparse+https://crates.mycompany.com/index/"
# 或 Git 索引
# index = "https://github.com/mycompany/crates-index"

[registry]
default = "crates-io"
# default = "my-registry"  # 使用私有仓库作为默认

# 认证（也可以用 cargo login）
[registries.my-registry]
token = "xxx"

# 网络设置
[net]
retry = 3
git-fetch-with-cli = true

# 代理设置
[http]
proxy = "http://proxy.company.com:8080"
timeout = 60

# 构建设置
[build]
jobs = 8
rustflags = ["-C", "target-cpu=native"]

# 目标平台
[target.x86_64-unknown-linux-gnu]
linker = "clang"
rustflags = ["-C", "link-arg=-fuse-ld=lld"]

# 别名
[alias]
b = "build"
t = "test"
r = "run"
c = "check"
```

---

## 自托管私有仓库

### 1. Verdaccio (npm 私有仓库)

```yaml
# docker-compose.yml
version: '3.8'

services:
  verdaccio:
    image: verdaccio/verdaccio:5
    container_name: verdaccio
    ports:
      - "4873:4873"
    volumes:
      - ./verdaccio/conf:/verdaccio/conf
      - ./verdaccio/storage:/verdaccio/storage
      - ./verdaccio/plugins:/verdaccio/plugins
    environment:
      - VERDACCIO_PORT=4873
    restart: unless-stopped
```

```yaml
# verdaccio/conf/config.yaml
storage: /verdaccio/storage
plugins: /verdaccio/plugins

web:
  title: My Private NPM Registry
  logo: /verdaccio/conf/logo.png

auth:
  htpasswd:
    file: /verdaccio/conf/htpasswd
    max_users: 100

uplinks:
  npmjs:
    url: https://registry.npmjs.org/
    timeout: 30s
    maxage: 2m
    cache: true

packages:
  '@mycompany/*':
    access: $authenticated
    publish: $authenticated
    unpublish: $authenticated

  '@*/*':
    access: $all
    publish: $authenticated
    proxy: npmjs

  '**':
    access: $all
    publish: $authenticated
    proxy: npmjs

middlewares:
  audit:
    enabled: true

logs:
  - { type: stdout, format: pretty, level: info }

listen: 0.0.0.0:4873
```

### 2. PyPI Server (Python 私有仓库)

```yaml
# docker-compose.yml
version: '3.8'

services:
  pypiserver:
    image: pypiserver/pypiserver:latest
    container_name: pypiserver
    ports:
      - "8080:8080"
    volumes:
      - ./packages:/data/packages
      - ./htpasswd:/data/.htpasswd
    command: run -P /data/.htpasswd -a update,download,list /data/packages
    restart: unless-stopped

  # 或使用 devpi
  devpi:
    image: devpi/devpi:latest
    container_name: devpi
    ports:
      - "3141:3141"
    volumes:
      - ./devpi-data:/data
    environment:
      - DEVPI_PASSWORD=admin123
    restart: unless-stopped
```

```bash
# 上传到私有 PyPI
twine upload --repository-url http://localhost:8080 dist/*

# pip 配置
pip install --index-url http://localhost:8080/simple/ my-package

# 或在 pip.conf 中配置
[global]
index-url = http://localhost:8080/simple/
trusted-host = localhost
```

### 3. Athens (Go 私有代理)

```yaml
# docker-compose.yml
version: '3.8'

services:
  athens:
    image: gomods/athens:latest
    container_name: athens
    ports:
      - "3000:3000"
    volumes:
      - ./athens-storage:/var/lib/athens
    environment:
      - ATHENS_DISK_STORAGE_ROOT=/var/lib/athens
      - ATHENS_STORAGE_TYPE=disk
      - GO_ENV=development
      - ATHENS_GO_BINARY_PATH=/usr/local/go/bin/go
    restart: unless-stopped
```

```bash
# 使用私有代理
export GOPROXY=http://localhost:3000,https://proxy.golang.org,direct
export GONOSUMDB=github.com/mycompany/*

# 或在 go.env 中配置
go env -w GOPROXY=http://localhost:3000,https://proxy.golang.org,direct
```

### 4. Cargo Registry (Rust 私有仓库)

```bash
# 使用 Kellnr - 轻量级 Cargo 私有仓库
docker run -d \
  --name kellnr \
  -p 8000:8000 \
  -v kellnr-data:/opt/kellnr/data \
  ghcr.io/kellnr/kellnr:latest
```

```toml
# .cargo/config.toml
[registries.kellnr]
index = "sparse+http://localhost:8000/api/v1/crates/"

[registry]
default = "kellnr"
```

### 5. Nexus Repository (统一仓库管理)

```yaml
# docker-compose.yml - Nexus 支持 npm, PyPI, Maven, Docker 等
version: '3.8'

services:
  nexus:
    image: sonatype/nexus3:latest
    container_name: nexus
    ports:
      - "8081:8081"
      - "8082:8082"  # Docker registry
      - "8083:8083"  # npm registry
    volumes:
      - ./nexus-data:/nexus-data
    environment:
      - INSTALL4J_ADD_VM_PARAMS=-Xms2g -Xmx2g -XX:MaxDirectMemorySize=2g
    restart: unless-stopped
```

### 6. GitLab Package Registry

```yaml
# .gitlab-ci.yml - 发布到 GitLab Package Registry

stages:
  - build
  - publish

# npm 包发布
publish-npm:
  stage: publish
  image: node:18
  script:
    - echo "@mycompany:registry=https://gitlab.com/api/v4/projects/${CI_PROJECT_ID}/packages/npm/" > .npmrc
    - echo "//gitlab.com/api/v4/projects/${CI_PROJECT_ID}/packages/npm/:_authToken=${CI_JOB_TOKEN}" >> .npmrc
    - npm publish
  only:
    - tags

# Python 包发布
publish-pypi:
  stage: publish
  image: python:3.11
  script:
    - pip install build twine
    - python -m build
    - TWINE_USERNAME=gitlab-ci-token TWINE_PASSWORD=${CI_JOB_TOKEN} twine upload --repository-url https://gitlab.com/api/v4/projects/${CI_PROJECT_ID}/packages/pypi dist/*
  only:
    - tags

# Go 模块（通过 Git tag）
publish-go:
  stage: publish
  image: golang:1.21
  script:
    - echo "Go modules are published via Git tags"
    - go list -m gitlab.com/mycompany/mypackage@${CI_COMMIT_TAG}
  only:
    - tags
```

---

## CI/CD 自动发布

### 1. GitHub Actions

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  # ============ TypeScript/npm ============
  release-npm:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          registry-url: 'https://registry.npmjs.org'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Build
        run: npm run build
      
      - name: Test
        run: npm test
      
      - name: Publish
        run: npm publish --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}

  # ============ Python ============
  release-pypi:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: |
          pip install build twine
          pip install -e ".[dev]"
      
      - name: Test
        run: pytest
      
      - name: Build
        run: python -m build
      
      - name: Publish to PyPI
        env:
          TWINE_USERNAME: __token__
          TWINE_PASSWORD: ${{ secrets.PYPI_TOKEN }}
        run: twine upload dist/*

  # ============ Go ============
  release-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Test
        run: go test -v ./...
      
      - name: Verify module
        run: |
          go mod verify
          go list -m github.com/${{ github.repository }}@${{ github.ref_name }}

  # ============ Rust ============
  release-rust:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Rust
        uses: dtolnay/rust-action@stable
      
      - name: Test
        run: cargo test --all-features
      
      - name: Publish to crates.io
        run: cargo publish
        env:
          CARGO_REGISTRY_TOKEN: ${{ secrets.CRATES_IO_TOKEN }}
```

### 2. 版本管理脚本

```bash
#!/bin/bash
# scripts/release.sh

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# 获取当前版本
get_current_version() {
    if [[ -f "package.json" ]]; then
        grep '"version"' package.json | head -1 | awk -F'"' '{print $4}'
    elif [[ -f "pyproject.toml" ]]; then
        grep '^version' pyproject.toml | head -1 | awk -F'"' '{print $2}'
    elif [[ -f "Cargo.toml" ]]; then
        grep '^version' Cargo.toml | head -1 | awk -F'"' '{print $2}'
    elif [[ -f "go.mod" ]]; then
        git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
    fi
}

# 更新版本
bump_version() {
    local version=$1
    
    # npm
    if [[ -f "package.json" ]]; then
        npm version "$version" --no-git-tag-version
    fi
    
    # Python
    if [[ -f "pyproject.toml" ]]; then
        sed -i "s/^version = .*/version = \"$version\"/" pyproject.toml
    fi
    
    # Rust
    if [[ -f "Cargo.toml" ]]; then
        sed -i "s/^version = .*/version = \"$version\"/" Cargo.toml
    fi
}

# 主逻辑
main() {
    local bump_type=${1:-patch}  # major, minor, patch
    
    current=$(get_current_version)
    echo -e "${GREEN}Current version: $current${NC}"
    
    # 计算新版本
    IFS='.' read -ra parts <<< "${current#v}"
    major=${parts[0]}
    minor=${parts[1]}
    patch=${parts[2]}
    
    case $bump_type in
        major)
            ((major++))
            minor=0
            patch=0
            ;;
        minor)
            ((minor++))
            patch=0
            ;;
        patch)
            ((patch++))
            ;;
        *)
            echo -e "${RED}Invalid bump type: $bump_type${NC}"
            exit 1
            ;;
    esac
    
    new_version="$major.$minor.$patch"
    echo -e "${GREEN}New version: $new_version${NC}"
    
    # 更新版本
    bump_version "$new_version"
    
    # Git 操作
    git add -A
    git commit -m "chore: bump version to $new_version"
    git tag -a "v$new_version" -m "Release v$new_version"
    
    echo -e "${GREEN}Done! Run 'git push && git push --tags' to publish${NC}"
}

main "$@"
```

### 3. 统一发布配置 (Monorepo)

```yaml
# .github/workflows/monorepo-release.yml
name: Monorepo Release

on:
  push:
    tags:
      - '*@*'  # 匹配 package-name@version 格式

jobs:
  detect-package:
    runs-on: ubuntu-latest
    outputs:
      package: ${{ steps.detect.outputs.package }}
      version: ${{ steps.detect.outputs.version }}
      language: ${{ steps.detect.outputs.language }}
    steps:
      - id: detect
        run: |
          TAG="${{ github.ref_name }}"
          PACKAGE="${TAG%@*}"
          VERSION="${TAG#*@}"
          
          # 检测语言
          if [[ -f "packages/$PACKAGE/package.json" ]]; then
            LANGUAGE="npm"
          elif [[ -f "packages/$PACKAGE/pyproject.toml" ]]; then
            LANGUAGE="python"
          elif [[ -f "packages/$PACKAGE/Cargo.toml" ]]; then
            LANGUAGE="rust"
          elif [[ -f "packages/$PACKAGE/go.mod" ]]; then
            LANGUAGE="go"
          fi
          
          echo "package=$PACKAGE" >> $GITHUB_OUTPUT
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          echo "language=$LANGUAGE" >> $GITHUB_OUTPUT

  release:
    needs: detect-package
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Release npm package
        if: needs.detect-package.outputs.language == 'npm'
        working-directory: packages/${{ needs.detect-package.outputs.package }}
        run: |
          npm ci
          npm run build
          npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
      
      - name: Release Python package
        if: needs.detect-package.outputs.language == 'python'
        working-directory: packages/${{ needs.detect-package.outputs.package }}
        run: |
          pip install build twine
          python -m build
          twine upload dist/*
        env:
          TWINE_USERNAME: __token__
          TWINE_PASSWORD: ${{ secrets.PYPI_TOKEN }}
      
      - name: Release Rust crate
        if: needs.detect-package.outputs.language == 'rust'
        working-directory: packages/${{ needs.detect-package.outputs.package }}
        run: cargo publish
        env:
          CARGO_REGISTRY_TOKEN: ${{ secrets.CRATES_IO_TOKEN }}
```

---

## 总结

| 语言 | 包管理器 | 公共仓库 | 配置文件 |
|------|---------|---------|---------|
| TypeScript | npm/yarn/pnpm | npmjs.com | package.json, .npmrc |
| Python | pip/poetry | pypi.org | pyproject.toml, .pypirc |
| Go | go mod | proxy.golang.org | go.mod, go.env |
| Rust | cargo | crates.io | Cargo.toml, .cargo/config.toml |

| 私有仓库方案 | 支持语言 | 部署方式 |
|-------------|---------|---------|
| Verdaccio | npm | Docker |
| PyPI Server / devpi | Python | Docker |
| Athens | Go | Docker |
| Kellnr | Rust | Docker |
| Nexus | 多语言 | Docker |
| GitLab Package Registry | 多语言 | SaaS/Self-hosted |
| GitHub Packages | 多语言 | SaaS |

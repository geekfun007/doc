# FastAPI 详解与实战

本文档详细介绍 Python 高性能 Web 框架 FastAPI 的使用方法和最佳实践。

## 目录

1. [框架概述](#框架概述)
2. [快速开始](#快速开始)
3. [路由与请求处理](#路由与请求处理)
4. [请求参数与验证](#请求参数与验证)
5. [响应处理](#响应处理)
6. [依赖注入](#依赖注入)
7. [中间件与异常处理](#中间件与异常处理)
8. [数据库集成](#数据库集成)
9. [认证与授权](#认证与授权)
10. [后台任务与WebSocket](#后台任务与websocket)
11. [测试与部署](#测试与部署)
12. [项目实战](#项目实战)

---

## 框架概述

### FastAPI 特点

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         FastAPI 核心特性                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐        │
│  │   高性能        │  │   类型提示      │  │   自动文档      │        │
│  │  Starlette +    │  │  Pydantic 验证  │  │  OpenAPI/Swagger│        │
│  │  Uvicorn        │  │  自动序列化     │  │  自动生成       │        │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘        │
│                                                                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐        │
│  │   异步支持      │  │   依赖注入      │  │   安全认证      │        │
│  │  async/await    │  │  自动解析       │  │  OAuth2/JWT     │        │
│  │  原生支持       │  │  可复用组件     │  │  内置支持       │        │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 技术栈

| 组件 | 说明 |
|------|------|
| Starlette | 轻量级 ASGI 框架，提供路由、中间件等 |
| Pydantic | 数据验证和序列化 |
| Uvicorn | 高性能 ASGI 服务器 |
| OpenAPI | 自动生成 API 文档 |
| Python 3.8+ | 利用类型提示特性 |

### 与其他框架对比

| 特性 | FastAPI | Flask | Django |
|------|---------|-------|--------|
| 性能 | 极高 | 中等 | 中等 |
| 异步支持 | 原生 | 需扩展 | 3.1+ |
| 类型检查 | 内置 | 无 | 无 |
| 自动文档 | 内置 | 需扩展 | 需扩展 |
| 学习曲线 | 中等 | 低 | 高 |
| 适用场景 | API 服务 | 小型应用 | 全栈应用 |

---

## 快速开始

### 安装

```bash
# 安装 FastAPI 和 ASGI 服务器
pip install fastapi uvicorn[standard]

# 或使用 uv
uv add fastapi uvicorn[standard]

# 可选依赖
pip install python-multipart  # 表单和文件上传
pip install python-jose[cryptography]  # JWT
pip install passlib[bcrypt]  # 密码哈希
pip install sqlalchemy  # ORM
pip install asyncpg  # 异步 PostgreSQL
pip install aiomysql  # 异步 MySQL
```

### Hello World

```python
# main.py
from fastapi import FastAPI

app = FastAPI()

@app.get("/")
async def root():
    return {"message": "Hello World"}

@app.get("/items/{item_id}")
async def read_item(item_id: int):
    return {"item_id": item_id}
```

### 运行

```bash
# 开发模式（自动重载）
uvicorn main:app --reload

# 生产模式
uvicorn main:app --host 0.0.0.0 --port 8000 --workers 4

# 访问
# API: http://localhost:8000
# Swagger 文档: http://localhost:8000/docs
# ReDoc 文档: http://localhost:8000/redoc
```

### 应用配置

```python
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI(
    title="My API",
    description="API 描述信息",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
)

# CORS 配置
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
```

---

## 路由与请求处理

### 基本路由

```python
from fastapi import FastAPI, APIRouter
from enum import Enum

app = FastAPI()

# 基本路由
@app.get("/")
async def root():
    return {"message": "GET"}

@app.post("/")
async def create():
    return {"message": "POST"}

@app.put("/")
async def update():
    return {"message": "PUT"}

@app.delete("/")
async def delete():
    return {"message": "DELETE"}

# 路由参数
@app.get("/users/{user_id}")
async def get_user(user_id: int):  # 自动类型转换和验证
    return {"user_id": user_id}

@app.get("/files/{file_path:path}")  # 路径参数（包含 /）
async def get_file(file_path: str):
    return {"file_path": file_path}

# 枚举路径参数
class ModelName(str, Enum):
    alexnet = "alexnet"
    resnet = "resnet"

@app.get("/models/{model_name}")
async def get_model(model_name: ModelName):
    return {"model": model_name.value}
```

### 路由分组 (APIRouter)

```python
# routers/users.py
from fastapi import APIRouter

router = APIRouter(
    prefix="/users",
    tags=["users"],
    responses={404: {"description": "Not found"}},
)

@router.get("/")
async def list_users():
    return [{"username": "alice"}]

@router.get("/{user_id}")
async def get_user(user_id: int):
    return {"user_id": user_id}

# main.py
from fastapi import FastAPI
from routers import users, items

app = FastAPI()
app.include_router(users.router)
app.include_router(items.router, prefix="/api/v1")
```

---

## 请求参数与验证

### 查询参数

```python
from fastapi import FastAPI, Query
from typing import Optional, List

app = FastAPI()

# 基本查询参数
@app.get("/items/")
async def list_items(
    skip: int = 0,
    limit: int = 10,
    q: Optional[str] = None
):
    return {"skip": skip, "limit": limit, "q": q}

# 使用 Query 进行验证
@app.get("/items/search")
async def search_items(
    q: str = Query(
        ...,                    # ... 表示必填
        min_length=3,
        max_length=50,
        regex="^[a-zA-Z0-9]+$",
        description="搜索关键词",
    ),
    page: int = Query(default=1, ge=1, le=100),
):
    return {"q": q, "page": page}

# 列表查询参数 ?tags=a&tags=b
@app.get("/items/filter")
async def filter_items(
    tags: List[str] = Query(default=[]),
):
    return {"tags": tags}
```

### 路径参数

```python
from fastapi import FastAPI, Path

app = FastAPI()

@app.get("/items/{item_id}")
async def get_item(
    item_id: int = Path(..., ge=1, le=10000, description="商品 ID"),
):
    return {"item_id": item_id}
```

### 请求体 (Pydantic)

```python
from fastapi import FastAPI, Body
from pydantic import BaseModel, Field, EmailStr, validator
from typing import Optional, List

app = FastAPI()

# 基本模型
class UserCreate(BaseModel):
    username: str
    email: EmailStr
    password: str

class UserResponse(BaseModel):
    id: int
    username: str
    email: EmailStr

    class Config:
        from_attributes = True

@app.post("/users/", response_model=UserResponse)
async def create_user(user: UserCreate):
    return user

# 高级验证
class ItemCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    price: float = Field(..., gt=0)
    tags: List[str] = Field(default_factory=list)

    @validator("name")
    def name_not_empty(cls, v):
        if not v.strip():
            raise ValueError("名称不能为空白")
        return v.strip()

    class Config:
        json_schema_extra = {
            "example": {
                "name": "iPhone 15",
                "price": 7999.00,
                "tags": ["electronics"],
            }
        }

# 嵌套模型
class Address(BaseModel):
    city: str
    street: str

class OrderItem(BaseModel):
    product_id: int = Field(..., gt=0)
    quantity: int = Field(..., gt=0, le=100)

class OrderCreate(BaseModel):
    user_id: int
    items: List[OrderItem] = Field(..., min_length=1)
    shipping_address: Address

@app.post("/orders/")
async def create_order(order: OrderCreate):
    total = sum(item.quantity for item in order.items)
    return {"order": order, "total_items": total}
```

### 请求头与 Cookie

```python
from fastapi import FastAPI, Header, Cookie
from typing import Optional

app = FastAPI()

@app.get("/headers/")
async def read_headers(
    user_agent: Optional[str] = Header(default=None),
    x_token: Optional[str] = Header(default=None, alias="X-Token"),
):
    return {"User-Agent": user_agent, "X-Token": x_token}

@app.get("/cookies/")
async def read_cookies(
    session_id: Optional[str] = Cookie(default=None),
):
    return {"session_id": session_id}
```

### 表单与文件上传

```python
from fastapi import FastAPI, Form, File, UploadFile
from typing import List

app = FastAPI()

# 表单数据
@app.post("/login/")
async def login(
    username: str = Form(...),
    password: str = Form(...),
):
    return {"username": username}

# 单文件上传
@app.post("/upload/")
async def upload_file(file: UploadFile = File(...)):
    contents = await file.read()
    return {"filename": file.filename, "size": len(contents)}

# 多文件上传
@app.post("/upload/multiple/")
async def upload_files(files: List[UploadFile] = File(...)):
    return [{"filename": f.filename} for f in files]

# 表单 + 文件
@app.post("/upload/with-form/")
async def upload_with_form(
    title: str = Form(...),
    file: UploadFile = File(...),
):
    return {"title": title, "filename": file.filename}
```

---

## 响应处理

### 响应模型

```python
from fastapi import FastAPI
from pydantic import BaseModel
from typing import List, Union
from datetime import datetime

app = FastAPI()

class UserResponse(BaseModel):
    id: int
    username: str
    created_at: datetime

    class Config:
        from_attributes = True

# 响应模型过滤
@app.post("/users/", response_model=UserResponse)
async def create_user(user: UserCreate):
    return {"id": 1, "username": user.username, "password": "secret", "created_at": datetime.now()}

# 排除未设置的字段
@app.get("/items/{item_id}", response_model=Item, response_model_exclude_unset=True)
async def get_item(item_id: int):
    return {"name": "Foo", "price": 10.5}

# 多种响应类型
class Cat(BaseModel):
    name: str
    meow: str

class Dog(BaseModel):
    name: str
    bark: str

@app.get("/animals/{id}", response_model=Union[Cat, Dog])
async def get_animal(id: int):
    if id % 2 == 0:
        return Cat(name="Whiskers", meow="meow")
    return Dog(name="Buddy", bark="woof")

# 列表响应
@app.get("/users/", response_model=List[UserResponse])
async def list_users():
    return [{"id": 1, "username": "alice", "created_at": datetime.now()}]
```

### 响应状态码与类型

```python
from fastapi import FastAPI, status, Response
from fastapi.responses import JSONResponse, HTMLResponse, RedirectResponse, FileResponse, StreamingResponse

app = FastAPI()

# 状态码
@app.post("/items/", status_code=status.HTTP_201_CREATED)
async def create_item(name: str):
    return {"name": name}

# 动态状态码
@app.get("/items/{item_id}")
async def get_item(item_id: int, response: Response):
    if item_id == 0:
        response.status_code = status.HTTP_404_NOT_FOUND
        return {"error": "Not found"}
    return {"item_id": item_id}

# HTML 响应
@app.get("/html/", response_class=HTMLResponse)
async def html_response():
    return "<h1>Hello World</h1>"

# 重定向
@app.get("/redirect/")
async def redirect():
    return RedirectResponse(url="/")

# 文件下载
@app.get("/download/")
async def download():
    return FileResponse("file.pdf", filename="download.pdf")

# 流式响应
@app.get("/stream/")
async def stream():
    async def generate():
        for i in range(10):
            yield f"data: {i}\n\n"
    return StreamingResponse(generate(), media_type="text/event-stream")
```

### 响应头与 Cookie

```python
from fastapi import FastAPI, Response

app = FastAPI()

@app.get("/headers/")
async def custom_headers(response: Response):
    response.headers["X-Custom-Header"] = "value"
    return {"message": "Hello"}

@app.get("/set-cookie/")
async def set_cookie(response: Response):
    response.set_cookie(key="session_id", value="abc123", httponly=True)
    return {"message": "Cookie set"}
```

---

## 依赖注入

### 基本依赖

```python
from fastapi import FastAPI, Depends, Query
from typing import Optional

app = FastAPI()

# 函数依赖
async def common_parameters(
    q: Optional[str] = None,
    skip: int = 0,
    limit: int = 100,
):
    return {"q": q, "skip": skip, "limit": limit}

@app.get("/items/")
async def list_items(commons: dict = Depends(common_parameters)):
    return commons

# 类依赖
class Pagination:
    def __init__(
        self,
        page: int = Query(default=1, ge=1),
        size: int = Query(default=10, ge=1, le=100),
    ):
        self.page = page
        self.size = size
        self.offset = (page - 1) * size

@app.get("/users/")
async def list_users(pagination: Pagination = Depends()):
    return {"page": pagination.page, "offset": pagination.offset}

# 数据库依赖
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

@app.get("/items/")
async def list_items(db: Session = Depends(get_db)):
    return db.query(Item).all()
```

### 认证依赖

```python
from fastapi import Depends, HTTPException, status, Header

async def get_token_header(x_token: str = Header(...)):
    if x_token != "secret-token":
        raise HTTPException(status_code=401, detail="Invalid token")
    return x_token

async def get_current_user(token: str = Depends(get_token_header)):
    return {"username": "alice", "token": token}

@app.get("/users/me/")
async def read_current_user(user: dict = Depends(get_current_user)):
    return user

# 路由级别依赖
router = APIRouter(dependencies=[Depends(get_token_header)])
```

---

## 中间件与异常处理

### 中间件

```python
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.gzip import GZipMiddleware
import time

app = FastAPI()

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# GZip 压缩
app.add_middleware(GZipMiddleware, minimum_size=1000)

# 自定义中间件
@app.middleware("http")
async def add_process_time(request: Request, call_next):
    start = time.time()
    response = await call_next(request)
    response.headers["X-Process-Time"] = str(time.time() - start)
    return response
```

### 异常处理

```python
from fastapi import FastAPI, HTTPException, Request, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError

app = FastAPI()

# HTTPException
@app.get("/items/{item_id}")
async def get_item(item_id: int):
    if item_id == 0:
        raise HTTPException(status_code=404, detail="Item not found")
    return {"item_id": item_id}

# 自定义异常
class BusinessError(Exception):
    def __init__(self, code: int, message: str):
        self.code = code
        self.message = message

@app.exception_handler(BusinessError)
async def business_error_handler(request: Request, exc: BusinessError):
    return JSONResponse(status_code=400, content={"code": exc.code, "message": exc.message})

# 覆盖验证错误
@app.exception_handler(RequestValidationError)
async def validation_handler(request: Request, exc: RequestValidationError):
    errors = [{"field": ".".join(str(x) for x in e["loc"]), "message": e["msg"]} for e in exc.errors()]
    return JSONResponse(status_code=422, content={"code": 422, "errors": errors})

# 全局异常
@app.exception_handler(Exception)
async def global_handler(request: Request, exc: Exception):
    return JSONResponse(status_code=500, content={"code": 500, "message": "Internal error"})
```

---

## 数据库集成

### MySQL 连接配置

```python
# database.py
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.orm import DeclarativeBase
from sqlalchemy.pool import QueuePool
import os

# ============ 数据库 URL 格式 ============
# MySQL (异步): mysql+aiomysql://user:password@host:port/dbname
# MySQL (同步): mysql+pymysql://user:password@host:port/dbname
# PostgreSQL: postgresql+asyncpg://user:password@host:port/dbname
# SQLite: sqlite+aiosqlite:///./test.db

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "mysql+aiomysql://root:password@localhost:3306/myapp?charset=utf8mb4"
)

# ============ 创建异步引擎 ============
engine = create_async_engine(
    DATABASE_URL,
    echo=True,                    # 打印 SQL（生产环境关闭）
    pool_size=5,                  # 连接池大小
    max_overflow=10,              # 超出 pool_size 后最多创建的连接
    pool_timeout=30,              # 获取连接超时时间
    pool_recycle=1800,            # 连接回收时间（秒）
    pool_pre_ping=True,           # 使用前检查连接是否有效
    poolclass=QueuePool,          # 连接池类型
)

# ============ 创建会话工厂 ============
async_session = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,       # 提交后不过期对象
    autocommit=False,
    autoflush=False,
)

# ============ 基类 ============
class Base(DeclarativeBase):
    pass

# ============ 依赖注入 ============
async def get_db() -> AsyncSession:
    async with async_session() as session:
        try:
            yield session
        finally:
            await session.close()

# ============ 初始化数据库 ============
async def init_db():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

async def close_db():
    await engine.dispose()

# ============ 在 FastAPI 中使用 ============
from contextlib import asynccontextmanager
from fastapi import FastAPI

@asynccontextmanager
async def lifespan(app: FastAPI):
    # 启动时
    await init_db()
    yield
    # 关闭时
    await close_db()

app = FastAPI(lifespan=lifespan)
```

### 模型定义

```python
# models/base.py
from sqlalchemy import Column, Integer, DateTime, func
from sqlalchemy.orm import declared_attr
from datetime import datetime

class TimestampMixin:
    """时间戳混入类"""
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

class SoftDeleteMixin:
    """软删除混入类"""
    deleted_at = Column(DateTime, nullable=True)
    
    @property
    def is_deleted(self) -> bool:
        return self.deleted_at is not None

# models/user.py
from sqlalchemy import Column, Integer, String, Boolean, Text, Enum, Index
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.mysql import BIGINT, TINYINT, JSON
import enum

class UserStatus(enum.Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"
    BANNED = "banned"

class User(Base, TimestampMixin, SoftDeleteMixin):
    __tablename__ = "users"
    __table_args__ = (
        Index("idx_username", "username"),
        Index("idx_email", "email"),
        Index("idx_status_created", "status", "created_at"),
        {"mysql_engine": "InnoDB", "mysql_charset": "utf8mb4"},
    )

    id = Column(BIGINT(unsigned=True), primary_key=True, autoincrement=True)
    username = Column(String(50), unique=True, nullable=False, comment="用户名")
    email = Column(String(100), unique=True, nullable=False, comment="邮箱")
    hashed_password = Column(String(255), nullable=False)
    nickname = Column(String(50), nullable=True)
    avatar = Column(String(255), nullable=True)
    phone = Column(String(20), unique=True, nullable=True)
    status = Column(Enum(UserStatus), default=UserStatus.ACTIVE, nullable=False)
    is_superuser = Column(Boolean, default=False)
    extra_data = Column(JSON, nullable=True, comment="额外数据")

    # 关系
    profile = relationship("UserProfile", back_populates="user", uselist=False, lazy="selectin")
    posts = relationship("Post", back_populates="author", lazy="selectin")
    orders = relationship("Order", back_populates="user", lazy="noload")

    def __repr__(self):
        return f"<User(id={self.id}, username={self.username})>"

# models/profile.py
class UserProfile(Base, TimestampMixin):
    __tablename__ = "user_profiles"

    id = Column(BIGINT(unsigned=True), primary_key=True)
    user_id = Column(BIGINT(unsigned=True), ForeignKey("users.id", ondelete="CASCADE"), unique=True, nullable=False)
    real_name = Column(String(50), nullable=True)
    id_card = Column(String(20), nullable=True)
    gender = Column(TINYINT, default=0, comment="0:未知 1:男 2:女")
    birthday = Column(DateTime, nullable=True)
    address = Column(String(255), nullable=True)
    bio = Column(Text, nullable=True)

    user = relationship("User", back_populates="profile")

# models/post.py
class Post(Base, TimestampMixin, SoftDeleteMixin):
    __tablename__ = "posts"
    __table_args__ = (
        Index("idx_author_created", "author_id", "created_at"),
    )

    id = Column(BIGINT(unsigned=True), primary_key=True)
    title = Column(String(200), nullable=False)
    content = Column(Text, nullable=True)
    author_id = Column(BIGINT(unsigned=True), ForeignKey("users.id"), nullable=False)
    view_count = Column(Integer, default=0)
    status = Column(TINYINT, default=1, comment="1:草稿 2:已发布")

    author = relationship("User", back_populates="posts")
    tags = relationship("Tag", secondary="post_tags", back_populates="posts")

# models/tag.py (多对多关系)
class Tag(Base):
    __tablename__ = "tags"

    id = Column(Integer, primary_key=True)
    name = Column(String(50), unique=True, nullable=False)

    posts = relationship("Post", secondary="post_tags", back_populates="tags")

# 关联表
post_tags = Table(
    "post_tags",
    Base.metadata,
    Column("post_id", BIGINT(unsigned=True), ForeignKey("posts.id", ondelete="CASCADE"), primary_key=True),
    Column("tag_id", Integer, ForeignKey("tags.id", ondelete="CASCADE"), primary_key=True),
)

# models/order.py
class Order(Base, TimestampMixin):
    __tablename__ = "orders"

    id = Column(BIGINT(unsigned=True), primary_key=True)
    order_no = Column(String(32), unique=True, nullable=False)
    user_id = Column(BIGINT(unsigned=True), ForeignKey("users.id"), nullable=False)
    total_amount = Column(Numeric(10, 2), nullable=False)
    status = Column(TINYINT, default=0, comment="0:待支付 1:已支付 2:已发货 3:已完成")
    paid_at = Column(DateTime, nullable=True)

    user = relationship("User", back_populates="orders")
    items = relationship("OrderItem", back_populates="order", lazy="selectin")

class OrderItem(Base):
    __tablename__ = "order_items"

    id = Column(BIGINT(unsigned=True), primary_key=True)
    order_id = Column(BIGINT(unsigned=True), ForeignKey("orders.id", ondelete="CASCADE"), nullable=False)
    product_id = Column(BIGINT(unsigned=True), nullable=False)
    product_name = Column(String(100), nullable=False)
    quantity = Column(Integer, nullable=False)
    price = Column(Numeric(10, 2), nullable=False)

    order = relationship("Order", back_populates="items")
```

### CRUD 操作

```python
# crud/base.py
from typing import Generic, TypeVar, Type, Optional, List, Any
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, update, delete, func
from sqlalchemy.orm import selectinload

ModelType = TypeVar("ModelType", bound=Base)
CreateSchemaType = TypeVar("CreateSchemaType", bound=BaseModel)
UpdateSchemaType = TypeVar("UpdateSchemaType", bound=BaseModel)

class CRUDBase(Generic[ModelType, CreateSchemaType, UpdateSchemaType]):
    def __init__(self, model: Type[ModelType]):
        self.model = model

    async def get(self, db: AsyncSession, id: Any) -> Optional[ModelType]:
        result = await db.execute(select(self.model).where(self.model.id == id))
        return result.scalar_one_or_none()

    async def get_multi(
        self, db: AsyncSession, *, skip: int = 0, limit: int = 100
    ) -> List[ModelType]:
        result = await db.execute(
            select(self.model).offset(skip).limit(limit)
        )
        return list(result.scalars().all())

    async def create(self, db: AsyncSession, *, obj_in: CreateSchemaType) -> ModelType:
        obj_data = obj_in.model_dump()
        db_obj = self.model(**obj_data)
        db.add(db_obj)
        await db.commit()
        await db.refresh(db_obj)
        return db_obj

    async def update(
        self, db: AsyncSession, *, db_obj: ModelType, obj_in: UpdateSchemaType
    ) -> ModelType:
        update_data = obj_in.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(db_obj, field, value)
        db.add(db_obj)
        await db.commit()
        await db.refresh(db_obj)
        return db_obj

    async def remove(self, db: AsyncSession, *, id: int) -> Optional[ModelType]:
        obj = await self.get(db, id)
        if obj:
            await db.delete(obj)
            await db.commit()
        return obj

    async def count(self, db: AsyncSession) -> int:
        result = await db.execute(select(func.count()).select_from(self.model))
        return result.scalar()

# crud/user.py
class UserCRUD(CRUDBase[User, UserCreate, UserUpdate]):
    async def get_by_username(self, db: AsyncSession, username: str) -> Optional[User]:
        result = await db.execute(
            select(User).where(User.username == username)
        )
        return result.scalar_one_or_none()

    async def get_by_email(self, db: AsyncSession, email: str) -> Optional[User]:
        result = await db.execute(
            select(User).where(User.email == email)
        )
        return result.scalar_one_or_none()

    async def get_with_profile(self, db: AsyncSession, user_id: int) -> Optional[User]:
        result = await db.execute(
            select(User)
            .options(selectinload(User.profile))
            .where(User.id == user_id)
        )
        return result.scalar_one_or_none()

    async def get_active_users(
        self, db: AsyncSession, *, skip: int = 0, limit: int = 100
    ) -> List[User]:
        result = await db.execute(
            select(User)
            .where(User.status == UserStatus.ACTIVE)
            .where(User.deleted_at.is_(None))
            .offset(skip)
            .limit(limit)
        )
        return list(result.scalars().all())

    async def search(
        self, db: AsyncSession, *, keyword: str, skip: int = 0, limit: int = 100
    ) -> tuple[List[User], int]:
        query = select(User).where(
            (User.username.contains(keyword)) | (User.nickname.contains(keyword))
        )
        
        # 获取总数
        count_result = await db.execute(
            select(func.count()).select_from(query.subquery())
        )
        total = count_result.scalar()
        
        # 获取数据
        result = await db.execute(query.offset(skip).limit(limit))
        users = list(result.scalars().all())
        
        return users, total

    async def soft_delete(self, db: AsyncSession, user_id: int) -> bool:
        result = await db.execute(
            update(User)
            .where(User.id == user_id)
            .values(deleted_at=datetime.utcnow())
        )
        await db.commit()
        return result.rowcount > 0

user_crud = UserCRUD(User)
```

### 复杂查询

```python
from sqlalchemy import select, func, and_, or_, case, desc, asc
from sqlalchemy.orm import selectinload, joinedload, contains_eager

# ============ 关联查询 ============
async def get_user_with_posts(db: AsyncSession, user_id: int) -> Optional[User]:
    """获取用户及其所有文章"""
    result = await db.execute(
        select(User)
        .options(selectinload(User.posts))
        .where(User.id == user_id)
    )
    return result.scalar_one_or_none()

async def get_posts_with_author(db: AsyncSession, skip: int = 0, limit: int = 10):
    """获取文章列表（包含作者信息）"""
    result = await db.execute(
        select(Post)
        .options(joinedload(Post.author))
        .where(Post.deleted_at.is_(None))
        .order_by(desc(Post.created_at))
        .offset(skip)
        .limit(limit)
    )
    return result.unique().scalars().all()

# ============ 聚合查询 ============
async def get_user_post_stats(db: AsyncSession):
    """获取每个用户的文章统计"""
    result = await db.execute(
        select(
            User.id,
            User.username,
            func.count(Post.id).label("post_count"),
            func.sum(Post.view_count).label("total_views"),
        )
        .join(Post, User.id == Post.author_id, isouter=True)
        .group_by(User.id)
        .order_by(desc("post_count"))
    )
    return result.all()

async def get_daily_stats(db: AsyncSession, days: int = 7):
    """获取每日注册统计"""
    result = await db.execute(
        select(
            func.date(User.created_at).label("date"),
            func.count(User.id).label("count"),
        )
        .where(User.created_at >= datetime.utcnow() - timedelta(days=days))
        .group_by(func.date(User.created_at))
        .order_by("date")
    )
    return result.all()

# ============ 条件查询 ============
async def search_users(
    db: AsyncSession,
    *,
    keyword: Optional[str] = None,
    status: Optional[UserStatus] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    skip: int = 0,
    limit: int = 100,
):
    """动态条件查询"""
    query = select(User)
    conditions = []
    
    if keyword:
        conditions.append(
            or_(
                User.username.contains(keyword),
                User.nickname.contains(keyword),
                User.email.contains(keyword),
            )
        )
    
    if status:
        conditions.append(User.status == status)
    
    if start_date:
        conditions.append(User.created_at >= start_date)
    
    if end_date:
        conditions.append(User.created_at <= end_date)
    
    # 排除已删除
    conditions.append(User.deleted_at.is_(None))
    
    if conditions:
        query = query.where(and_(*conditions))
    
    query = query.order_by(desc(User.created_at)).offset(skip).limit(limit)
    
    result = await db.execute(query)
    return result.scalars().all()

# ============ 子查询 ============
async def get_users_with_post_count(db: AsyncSession):
    """使用子查询获取用户及文章数"""
    post_count_subq = (
        select(Post.author_id, func.count(Post.id).label("post_count"))
        .group_by(Post.author_id)
        .subquery()
    )
    
    result = await db.execute(
        select(User, post_count_subq.c.post_count)
        .join(post_count_subq, User.id == post_count_subq.c.author_id, isouter=True)
    )
    return result.all()

# ============ CASE 表达式 ============
async def get_users_with_level(db: AsyncSession):
    """根据文章数计算用户等级"""
    post_count = func.count(Post.id)
    level = case(
        (post_count >= 100, "expert"),
        (post_count >= 50, "advanced"),
        (post_count >= 10, "intermediate"),
        else_="beginner"
    ).label("level")
    
    result = await db.execute(
        select(User.username, post_count.label("post_count"), level)
        .join(Post, User.id == Post.author_id, isouter=True)
        .group_by(User.id)
    )
    return result.all()

# ============ 原生 SQL ============
async def execute_raw_sql(db: AsyncSession):
    """执行原生 SQL"""
    from sqlalchemy import text
    
    result = await db.execute(
        text("""
            SELECT u.username, COUNT(p.id) as post_count
            FROM users u
            LEFT JOIN posts p ON u.id = p.author_id
            WHERE u.status = :status
            GROUP BY u.id
            ORDER BY post_count DESC
            LIMIT :limit
        """),
        {"status": "active", "limit": 10}
    )
    return result.fetchall()
```

### 事务处理

```python
from sqlalchemy.ext.asyncio import AsyncSession

# ============ 基本事务（自动提交/回滚）============
async def create_user_with_profile(
    db: AsyncSession,
    user_data: UserCreate,
    profile_data: ProfileCreate,
) -> User:
    """创建用户及其详情（事务）"""
    async with db.begin():
        # 创建用户
        user = User(**user_data.model_dump())
        db.add(user)
        await db.flush()  # 获取 user.id
        
        # 创建详情
        profile = UserProfile(user_id=user.id, **profile_data.model_dump())
        db.add(profile)
        
        # 事务自动提交
    
    await db.refresh(user)
    return user

# ============ 嵌套事务（保存点）============
async def create_order_with_items(
    db: AsyncSession,
    order_data: OrderCreate,
) -> Order:
    """创建订单（带保存点）"""
    async with db.begin():
        # 创建订单
        order = Order(
            order_no=generate_order_no(),
            user_id=order_data.user_id,
            total_amount=order_data.total_amount,
        )
        db.add(order)
        await db.flush()
        
        # 嵌套事务处理订单项
        async with db.begin_nested():
            for item in order_data.items:
                order_item = OrderItem(
                    order_id=order.id,
                    **item.model_dump()
                )
                db.add(order_item)
                
                # 扣减库存
                result = await db.execute(
                    update(Product)
                    .where(Product.id == item.product_id)
                    .where(Product.stock >= item.quantity)
                    .values(stock=Product.stock - item.quantity)
                )
                if result.rowcount == 0:
                    raise ValueError(f"Product {item.product_id} stock insufficient")
    
    return order

# ============ 手动事务控制 ============
async def transfer_balance(
    db: AsyncSession,
    from_user_id: int,
    to_user_id: int,
    amount: float,
):
    """转账（手动事务）"""
    try:
        # 检查余额
        from_user = await db.get(User, from_user_id)
        if from_user.balance < amount:
            raise ValueError("Insufficient balance")
        
        # 扣款
        from_user.balance -= amount
        
        # 入账
        to_user = await db.get(User, to_user_id)
        to_user.balance += amount
        
        await db.commit()
    except Exception as e:
        await db.rollback()
        raise e

# ============ 乐观锁 ============
async def update_with_version(
    db: AsyncSession,
    user_id: int,
    update_data: dict,
    version: int,
) -> bool:
    """乐观锁更新"""
    result = await db.execute(
        update(User)
        .where(User.id == user_id)
        .where(User.version == version)
        .values(**update_data, version=version + 1)
    )
    await db.commit()
    return result.rowcount > 0

# ============ 悲观锁 ============
async def update_with_lock(db: AsyncSession, user_id: int, update_data: dict):
    """悲观锁更新"""
    async with db.begin():
        result = await db.execute(
            select(User)
            .where(User.id == user_id)
            .with_for_update()  # SELECT ... FOR UPDATE
        )
        user = result.scalar_one_or_none()
        if user:
            for key, value in update_data.items():
                setattr(user, key, value)
```

### 数据库迁移 (Alembic)

```bash
# 安装
pip install alembic

# 初始化
alembic init alembic
```

```python
# alembic/env.py
from logging.config import fileConfig
from sqlalchemy import pool
from sqlalchemy.engine import Connection
from sqlalchemy.ext.asyncio import async_engine_from_config
from alembic import context
import asyncio

from app.database import Base
from app.models import *  # 导入所有模型

config = context.config
if config.config_file_name is not None:
    fileConfig(config.config_file_name)

target_metadata = Base.metadata

def run_migrations_offline() -> None:
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )
    with context.begin_transaction():
        context.run_migrations()

def do_run_migrations(connection: Connection) -> None:
    context.configure(connection=connection, target_metadata=target_metadata)
    with context.begin_transaction():
        context.run_migrations()

async def run_async_migrations() -> None:
    connectable = async_engine_from_config(
        config.get_section(config.config_ini_section, {}),
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )
    async with connectable.connect() as connection:
        await connection.run_sync(do_run_migrations)
    await connectable.dispose()

def run_migrations_online() -> None:
    asyncio.run(run_async_migrations())

if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
```

```ini
# alembic.ini
[alembic]
script_location = alembic
sqlalchemy.url = mysql+aiomysql://root:password@localhost:3306/myapp
```

```bash
# 常用命令
alembic revision --autogenerate -m "create users table"  # 生成迁移
alembic upgrade head                                      # 执行迁移
alembic downgrade -1                                      # 回滚一步
alembic history                                           # 查看历史
alembic current                                           # 当前版本
```

### API 路由完整示例

```python
# schemas/user.py
from pydantic import BaseModel, EmailStr, Field
from typing import Optional, List
from datetime import datetime
from enum import Enum

class UserStatus(str, Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"
    BANNED = "banned"

class UserBase(BaseModel):
    username: str = Field(..., min_length=3, max_length=50)
    email: EmailStr
    nickname: Optional[str] = None

class UserCreate(UserBase):
    password: str = Field(..., min_length=6)

class UserUpdate(BaseModel):
    nickname: Optional[str] = None
    avatar: Optional[str] = None
    status: Optional[UserStatus] = None

class UserResponse(UserBase):
    id: int
    status: UserStatus
    created_at: datetime

    class Config:
        from_attributes = True

class UserListResponse(BaseModel):
    items: List[UserResponse]
    total: int
    page: int
    size: int

# api/users.py
from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy.ext.asyncio import AsyncSession

router = APIRouter(prefix="/users", tags=["users"])

@router.post("/", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
async def create_user(
    user_in: UserCreate,
    db: AsyncSession = Depends(get_db),
):
    """创建用户"""
    # 检查用户名
    if await user_crud.get_by_username(db, user_in.username):
        raise HTTPException(status_code=400, detail="Username already exists")
    
    # 检查邮箱
    if await user_crud.get_by_email(db, user_in.email):
        raise HTTPException(status_code=400, detail="Email already registered")
    
    # 创建用户
    user = await user_crud.create(db, obj_in=user_in)
    return user

@router.get("/", response_model=UserListResponse)
async def list_users(
    page: int = Query(default=1, ge=1),
    size: int = Query(default=10, ge=1, le=100),
    keyword: Optional[str] = Query(default=None),
    status: Optional[UserStatus] = Query(default=None),
    db: AsyncSession = Depends(get_db),
):
    """获取用户列表"""
    skip = (page - 1) * size
    
    if keyword:
        users, total = await user_crud.search(db, keyword=keyword, skip=skip, limit=size)
    else:
        users = await user_crud.get_multi(db, skip=skip, limit=size)
        total = await user_crud.count(db)
    
    return UserListResponse(items=users, total=total, page=page, size=size)

@router.get("/{user_id}", response_model=UserResponse)
async def get_user(
    user_id: int,
    db: AsyncSession = Depends(get_db),
):
    """获取用户详情"""
    user = await user_crud.get(db, user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user

@router.put("/{user_id}", response_model=UserResponse)
async def update_user(
    user_id: int,
    user_in: UserUpdate,
    db: AsyncSession = Depends(get_db),
):
    """更新用户"""
    user = await user_crud.get(db, user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    
    user = await user_crud.update(db, db_obj=user, obj_in=user_in)
    return user

@router.delete("/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_user(
    user_id: int,
    db: AsyncSession = Depends(get_db),
):
    """删除用户（软删除）"""
    success = await user_crud.soft_delete(db, user_id)
    if not success:
        raise HTTPException(status_code=404, detail="User not found")
    return None
```

### ORM 最佳实践

```python
# ============ 1. 使用异步会话上下文管理器 ============
async def get_db():
    async with async_session() as session:
        try:
            yield session
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()

# ============ 2. 批量操作优化 ============
async def bulk_create_users(db: AsyncSession, users: List[UserCreate]):
    """批量创建用户"""
    db_users = [User(**u.model_dump()) for u in users]
    db.add_all(db_users)
    await db.commit()
    return db_users

async def bulk_update(db: AsyncSession, user_ids: List[int], status: str):
    """批量更新"""
    await db.execute(
        update(User).where(User.id.in_(user_ids)).values(status=status)
    )
    await db.commit()

# ============ 3. 分页优化（使用游标）============
async def get_users_cursor(
    db: AsyncSession,
    last_id: int = 0,
    limit: int = 100,
):
    """基于游标的分页（适合大数据量）"""
    result = await db.execute(
        select(User)
        .where(User.id > last_id)
        .order_by(User.id)
        .limit(limit)
    )
    return result.scalars().all()

# ============ 4. 预加载优化 ============
async def get_users_optimized(db: AsyncSession):
    """优化关联加载"""
    result = await db.execute(
        select(User)
        .options(
            selectinload(User.posts).selectinload(Post.tags),  # 嵌套预加载
            selectinload(User.profile),
        )
        .where(User.status == UserStatus.ACTIVE)
    )
    return result.unique().scalars().all()

# ============ 5. 只查询需要的字段 ============
async def get_user_names(db: AsyncSession):
    """只查询部分字段"""
    result = await db.execute(
        select(User.id, User.username, User.email)
    )
    return result.all()

# ============ 6. 使用 exists 检查 ============
async def user_exists(db: AsyncSession, username: str) -> bool:
    """检查用户是否存在"""
    result = await db.execute(
        select(func.count()).where(User.username == username)
    )
    return result.scalar() > 0

# ============ 7. 事务装饰器 ============
from functools import wraps

def transactional(func):
    @wraps(func)
    async def wrapper(db: AsyncSession, *args, **kwargs):
        async with db.begin():
            return await func(db, *args, **kwargs)
    return wrapper

@transactional
async def create_user_transactional(db: AsyncSession, user_in: UserCreate):
    user = User(**user_in.model_dump())
    db.add(user)
    await db.flush()
    return user
```

---

## 认证与授权

### JWT 认证

```python
from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from jose import JWTError, jwt
from passlib.context import CryptContext
from datetime import datetime, timedelta

SECRET_KEY = "your-secret-key"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

def verify_password(plain: str, hashed: str) -> bool:
    return pwd_context.verify(plain, hashed)

def hash_password(password: str) -> str:
    return pwd_context.hash(password)

def create_access_token(data: dict, expires_delta: timedelta = None) -> str:
    to_encode = data.copy()
    expire = datetime.utcnow() + (expires_delta or timedelta(minutes=15))
    to_encode.update({"exp": expire})
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)

async def get_current_user(token: str = Depends(oauth2_scheme)):
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        if username is None:
            raise credentials_exception
    except JWTError:
        raise credentials_exception
    return {"username": username}

@app.post("/token")
async def login(form_data: OAuth2PasswordRequestForm = Depends()):
    # 验证用户...
    access_token = create_access_token(
        data={"sub": form_data.username},
        expires_delta=timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES),
    )
    return {"access_token": access_token, "token_type": "bearer"}

@app.get("/users/me")
async def read_users_me(current_user: dict = Depends(get_current_user)):
    return current_user
```

### 角色权限

```python
from enum import Enum
from typing import List

class Role(str, Enum):
    ADMIN = "admin"
    USER = "user"

class RoleChecker:
    def __init__(self, allowed_roles: List[Role]):
        self.allowed_roles = allowed_roles

    def __call__(self, user = Depends(get_current_user)):
        if user.role not in self.allowed_roles:
            raise HTTPException(status_code=403, detail="Permission denied")
        return user

allow_admin = RoleChecker([Role.ADMIN])

@app.get("/admin/users")
async def admin_users(user = Depends(allow_admin)):
    return {"message": "Admin only"}
```

---

## 后台任务与WebSocket

### 后台任务

```python
from fastapi import FastAPI, BackgroundTasks

app = FastAPI()

def send_email(email: str, message: str):
    import time
    time.sleep(5)
    print(f"Email sent to {email}")

@app.post("/send-notification/")
async def send_notification(email: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(send_email, email, "Welcome!")
    return {"message": "Notification sent in background"}
```

### WebSocket

```python
from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from typing import List

app = FastAPI()

class ConnectionManager:
    def __init__(self):
        self.active_connections: List[WebSocket] = []

    async def connect(self, websocket: WebSocket):
        await websocket.accept()
        self.active_connections.append(websocket)

    def disconnect(self, websocket: WebSocket):
        self.active_connections.remove(websocket)

    async def broadcast(self, message: str):
        for connection in self.active_connections:
            await connection.send_text(message)

manager = ConnectionManager()

@app.websocket("/ws/{client_id}")
async def websocket_endpoint(websocket: WebSocket, client_id: int):
    await manager.connect(websocket)
    try:
        while True:
            data = await websocket.receive_text()
            await manager.broadcast(f"Client #{client_id}: {data}")
    except WebSocketDisconnect:
        manager.disconnect(websocket)
        await manager.broadcast(f"Client #{client_id} left")
```

---

## 测试与部署

### 单元测试

```python
from fastapi.testclient import TestClient
from main import app
import pytest

client = TestClient(app)

def test_read_root():
    response = client.get("/")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello World"}

def test_create_user():
    response = client.post("/users/", json={"username": "alice", "email": "alice@example.com", "password": "secret"})
    assert response.status_code == 201

# 异步测试
@pytest.mark.anyio
async def test_async():
    from httpx import AsyncClient
    async with AsyncClient(app=app, base_url="http://test") as ac:
        response = await ac.get("/")
    assert response.status_code == 200

# 依赖覆盖
def test_override():
    app.dependency_overrides[get_db] = lambda: {"db": "test"}
    response = client.get("/items/")
    app.dependency_overrides.clear()
```

### 部署配置

```dockerfile
# Dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql+asyncpg://postgres:password@db:5432/app
    depends_on:
      - db
  db:
    image: postgres:15
    environment:
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=app
    volumes:
      - postgres_data:/var/lib/postgresql/data
volumes:
  postgres_data:
```

---

## 项目实战

### 项目结构

```
myproject/
├── app/
│   ├── __init__.py
│   ├── main.py              # 应用入口
│   ├── config.py            # 配置
│   ├── database.py          # 数据库
│   ├── models/              # SQLAlchemy 模型
│   ├── schemas/             # Pydantic 模型
│   ├── crud/                # CRUD 操作
│   ├── api/                 # API 路由
│   │   ├── deps.py          # 依赖
│   │   └── v1/
│   └── core/                # 核心模块
├── tests/
├── alembic/                 # 数据库迁移
├── requirements.txt
├── Dockerfile
└── docker-compose.yml
```

### 配置管理

```python
# config.py
from pydantic_settings import BaseSettings
from functools import lru_cache

class Settings(BaseSettings):
    app_name: str = "My API"
    debug: bool = False
    database_url: str
    secret_key: str
    access_token_expire_minutes: int = 30

    class Config:
        env_file = ".env"

@lru_cache()
def get_settings():
    return Settings()
```

### 常用命令

```bash
# 开发
uvicorn app.main:app --reload

# 测试
pytest tests/ -v

# 数据库迁移
alembic revision --autogenerate -m "message"
alembic upgrade head

# 生产
uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4
```

---

## 总结

### FastAPI 速查

```python
# 路由
@app.get("/")
@app.post("/")
@app.put("/")
@app.delete("/")

# 参数
def func(
    path_param: int,                    # 路径参数
    query_param: str = Query(...),      # 查询参数
    body_param: Model = Body(...),      # 请求体
    header_param: str = Header(...),    # 请求头
    cookie_param: str = Cookie(...),    # Cookie
)

# 响应
@app.get("/", response_model=Model, status_code=201)

# 依赖注入
@app.get("/", dependencies=[Depends(func)])
def endpoint(dep = Depends(func)):
    pass

# 异常
raise HTTPException(status_code=404, detail="Not found")

# 后台任务
def endpoint(background_tasks: BackgroundTasks):
    background_tasks.add_task(func, arg)
```

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

### SQLAlchemy 异步

```python
# database.py
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.orm import DeclarativeBase

DATABASE_URL = "mysql+aiomysql://user:password@localhost:3306/dbname"

engine = create_async_engine(DATABASE_URL, echo=True)
async_session = async_sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)

class Base(DeclarativeBase):
    pass

# 依赖
async def get_db() -> AsyncSession:
    async with async_session() as session:
        yield session
```

### 模型定义

```python
from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Boolean
from sqlalchemy.orm import relationship
from datetime import datetime

class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    username = Column(String(50), unique=True, index=True)
    email = Column(String(100), unique=True, index=True)
    hashed_password = Column(String(255))
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    posts = relationship("Post", back_populates="author")

class Post(Base):
    __tablename__ = "posts"

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(200))
    content = Column(String(10000))
    author_id = Column(Integer, ForeignKey("users.id"))

    author = relationship("User", back_populates="posts")
```

### CRUD 操作

```python
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from sqlalchemy.orm import selectinload

class UserCRUD:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def create(self, username: str, email: str, password: str) -> User:
        user = User(username=username, email=email, hashed_password=password)
        self.db.add(user)
        await self.db.commit()
        await self.db.refresh(user)
        return user

    async def get_by_id(self, user_id: int) -> User | None:
        result = await self.db.execute(select(User).where(User.id == user_id))
        return result.scalar_one_or_none()

    async def get_with_posts(self, user_id: int) -> User | None:
        result = await self.db.execute(
            select(User).options(selectinload(User.posts)).where(User.id == user_id)
        )
        return result.scalar_one_or_none()

    async def list(self, skip: int = 0, limit: int = 100) -> list[User]:
        result = await self.db.execute(select(User).offset(skip).limit(limit))
        return result.scalars().all()
```

### API 路由

```python
from fastapi import APIRouter, Depends, HTTPException

router = APIRouter(prefix="/users", tags=["users"])

@router.post("/", response_model=UserResponse, status_code=201)
async def create_user(user_in: UserCreate, db: AsyncSession = Depends(get_db)):
    crud = UserCRUD(db)
    if await crud.get_by_username(user_in.username):
        raise HTTPException(status_code=400, detail="Username exists")
    return await crud.create(user_in.username, user_in.email, hash_password(user_in.password))

@router.get("/{user_id}", response_model=UserResponse)
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)):
    crud = UserCRUD(db)
    user = await crud.get_by_id(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
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

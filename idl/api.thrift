namespace go api

// ====================== Common ======================

struct BaseResp {
    1: i32    code    (api.body="code")
    2: string message (api.body="message")
    3: string data    (api.body="data")
}

// ====================== User ======================

struct RegisterReq {
    1: string username (api.body="username", api.vd="len($)>1 && len($)<33; msg:'username length must be 2-32'")
    2: string email    (api.body="email",    api.vd="len($)>4 && len($)<65; msg:'email length must be 5-64'")
    3: string password (api.body="password", api.vd="len($)>5 && len($)<129; msg:'password length must be 6-128'")
}

struct RegisterResp {
    1: i32    code    (api.body="code")
    2: string message (api.body="message")
    3: i64    user_id (api.body="user_id")
    4: string token   (api.body="token")
}

struct LoginReq {
    1: string username (api.body="username", api.vd="len($)>1; msg:'username is required'")
    2: string password (api.body="password", api.vd="len($)>5; msg:'password length must be >5'")
}

struct LoginResp {
    1: i32    code    (api.body="code")
    2: string message (api.body="message")
    3: i64    user_id (api.body="user_id")
    4: string token   (api.body="token")
}

struct GetUserReq {
    1: i64 id (api.path="id", api.vd="$>0; msg:'invalid user id'")
}

struct GetUserResp {
    1: i32    code      (api.body="code")
    2: string message   (api.body="message")
    3: i64    id        (api.body="id")
    4: string username  (api.body="username")
    5: string email     (api.body="email")
    6: string avatar    (api.body="avatar")
    7: string created_at(api.body="created_at")
    8: string updated_at(api.body="updated_at")
}

struct UpdateUserReq {
    1: i64    id       (api.path="id",       api.vd="$>0; msg:'invalid user id'")
    2: string username (api.body="username")
    3: string email    (api.body="email")
    4: string avatar   (api.body="avatar")
}

struct UpdateUserResp {
    1: i32    code    (api.body="code")
    2: string message (api.body="message")
}

// ====================== Article ======================

struct CreateArticleReq {
    1: string title   (api.body="title",   api.vd="len($)>0 && len($)<257; msg:'title length must be 1-256'")
    2: string content (api.body="content", api.vd="len($)>0; msg:'content is required'")
}

struct CreateArticleResp {
    1: i32    code       (api.body="code")
    2: string message    (api.body="message")
    3: i64    article_id (api.body="article_id")
}

struct GetArticleReq {
    1: i64 id (api.path="id", api.vd="$>0; msg:'invalid article id'")
}

struct GetArticleResp {
    1: i32    code        (api.body="code")
    2: string message     (api.body="message")
    3: i64    id          (api.body="id")
    4: string title       (api.body="title")
    5: string content     (api.body="content")
    6: i64    author_id   (api.body="author_id")
    7: string author_name (api.body="author_name")
    8: i32    status      (api.body="status")
    9: i64    view_count  (api.body="view_count")
    10: string created_at (api.body="created_at")
    11: string updated_at (api.body="updated_at")
}

struct ListArticleReq {
    1: i64 author_id (api.query="author_id")
    2: i32 page      (api.query="page",      api.vd="$>0; msg:'page must be > 0'")
    3: i32 page_size (api.query="page_size",  api.vd="$>0 && $<=100; msg:'page_size must be 1-100'")
}

struct ListArticleResp {
    1: i32    code    (api.body="code")
    2: string message (api.body="message")
    3: string data    (api.body="data")
}

struct UpdateArticleReq {
    1: i64    id      (api.path="id",      api.vd="$>0; msg:'invalid article id'")
    2: string title   (api.body="title")
    3: string content (api.body="content")
    4: i32    status  (api.body="status")
}

struct UpdateArticleResp {
    1: i32    code    (api.body="code")
    2: string message (api.body="message")
}

struct DeleteArticleReq {
    1: i64 id (api.path="id", api.vd="$>0; msg:'invalid article id'")
}

struct DeleteArticleResp {
    1: i32    code    (api.body="code")
    2: string message (api.body="message")
}

// ====================== Service ======================

service ApiService {
    RegisterResp      Register(1: RegisterReq req)          (api.post="/api/v1/user/register")
    LoginResp         Login(1: LoginReq req)                (api.post="/api/v1/user/login")
    GetUserResp       GetUser(1: GetUserReq req)            (api.get="/api/v1/user/:id")
    UpdateUserResp    UpdateUser(1: UpdateUserReq req)       (api.put="/api/v1/user/:id")

    CreateArticleResp CreateArticle(1: CreateArticleReq req) (api.post="/api/v1/article")
    GetArticleResp    GetArticle(1: GetArticleReq req)       (api.get="/api/v1/article/:id")
    ListArticleResp   ListArticle(1: ListArticleReq req)     (api.get="/api/v1/articles")
    UpdateArticleResp UpdateArticle(1: UpdateArticleReq req) (api.put="/api/v1/article/:id")
    DeleteArticleResp DeleteArticle(1: DeleteArticleReq req) (api.delete="/api/v1/article/:id")
}

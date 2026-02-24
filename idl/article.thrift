namespace go article

struct Article {
    1: i64    id
    2: string title
    3: string content
    4: i64    author_id
    5: string author_name
    6: i32    status       // 0=draft 1=published 2=deleted
    7: i64    view_count
    8: string created_at
    9: string updated_at
}

struct CreateArticleRequest {
    1: string title   (vt.min_size = "1", vt.max_size = "256")
    2: string content (vt.min_size = "1")
    3: i64    author_id
}

struct CreateArticleResponse {
    1: i64 article_id
}

struct GetArticleRequest {
    1: i64 article_id
}

struct GetArticleResponse {
    1: Article article
}

struct ListArticleRequest {
    1: i64 author_id
    2: i32 page      // starting from 1
    3: i32 page_size
}

struct ListArticleResponse {
    1: list<Article> articles
    2: i64 total
}

struct UpdateArticleRequest {
    1: i64    article_id
    2: string title
    3: string content
    4: i32    status
}

struct UpdateArticleResponse {
    1: bool success
}

struct DeleteArticleRequest {
    1: i64 article_id
}

struct DeleteArticleResponse {
    1: bool success
}

service ArticleService {
    CreateArticleResponse CreateArticle(1: CreateArticleRequest req)
    GetArticleResponse GetArticle(1: GetArticleRequest req)
    ListArticleResponse ListArticle(1: ListArticleRequest req)
    UpdateArticleResponse UpdateArticle(1: UpdateArticleRequest req)
    DeleteArticleResponse DeleteArticle(1: DeleteArticleRequest req)
}

package main

import (
	"context"
	"time"

	"byte.dance/kitex_gen/article"
	"byte.dance/article/dal"
)

type ArticleServiceImpl struct{}

func (s *ArticleServiceImpl) CreateArticle(ctx context.Context, req *article.CreateArticleRequest) (*article.CreateArticleResponse, error) {
	id, err := dal.GetStore().Create(req.Title, req.Content, req.AuthorId, "")
	if err != nil {
		return nil, err
	}
	return &article.CreateArticleResponse{ArticleId: id}, nil
}

func (s *ArticleServiceImpl) GetArticle(ctx context.Context, req *article.GetArticleRequest) (*article.GetArticleResponse, error) {
	a, err := dal.GetStore().GetByID(req.ArticleId)
	if err != nil {
		return nil, err
	}
	return &article.GetArticleResponse{
		Article: toArticle(a),
	}, nil
}

func (s *ArticleServiceImpl) ListArticle(ctx context.Context, req *article.ListArticleRequest) (*article.ListArticleResponse, error) {
	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	list, total := dal.GetStore().ListByAuthor(req.AuthorId, page, pageSize)
	articles := make([]*article.Article, 0, len(list))
	for _, a := range list {
		articles = append(articles, toArticle(a))
	}

	return &article.ListArticleResponse{
		Articles: articles,
		Total:    total,
	}, nil
}

func (s *ArticleServiceImpl) UpdateArticle(ctx context.Context, req *article.UpdateArticleRequest) (*article.UpdateArticleResponse, error) {
	err := dal.GetStore().Update(req.ArticleId, req.Title, req.Content, req.Status)
	if err != nil {
		return nil, err
	}
	return &article.UpdateArticleResponse{Success: true}, nil
}

func (s *ArticleServiceImpl) DeleteArticle(ctx context.Context, req *article.DeleteArticleRequest) (*article.DeleteArticleResponse, error) {
	err := dal.GetStore().Delete(req.ArticleId)
	if err != nil {
		return nil, err
	}
	return &article.DeleteArticleResponse{Success: true}, nil
}

func toArticle(m *dal.ArticleModel) *article.Article {
	return &article.Article{
		Id:         m.ID,
		Title:      m.Title,
		Content:    m.Content,
		AuthorId:   m.AuthorID,
		AuthorName: m.AuthorName,
		Status:     m.Status,
		ViewCount:  m.ViewCount,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  m.UpdatedAt.Format(time.RFC3339),
	}
}

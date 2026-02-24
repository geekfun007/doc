package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	"byte.dance/kitex_gen/article"
	"byte.dance/api/biz/rpc"
	"byte.dance/pkg/errno"
)

func CreateArticle(_ context.Context, c *app.RequestContext) {
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	uid := c.GetInt64("user_id")
	r, err := rpc.ArticleClient().CreateArticle(context.Background(), &article.CreateArticleRequest{
		Title:    req.Title,
		Content:  req.Content,
		AuthorId: uid,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp(errno.ServiceErr, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp(errno.Success, utils.H{
		"article_id": r.ArticleId,
	}))
}

func GetArticle(_ context.Context, c *app.RequestContext) {
	aid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	r, err := rpc.ArticleClient().GetArticle(context.Background(), &article.GetArticleRequest{
		ArticleId: aid,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp(errno.ServiceErr, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp(errno.Success, r.Article))
}

func ListArticle(_ context.Context, c *app.RequestContext) {
	authorID, _ := strconv.ParseInt(c.Query("author_id"), 10, 64)
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 32)

	r, err := rpc.ArticleClient().ListArticle(context.Background(), &article.ListArticleRequest{
		AuthorId: authorID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp(errno.ServiceErr, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp(errno.Success, utils.H{
		"articles": r.Articles,
		"total":    r.Total,
	}))
}

func UpdateArticle(_ context.Context, c *app.RequestContext) {
	aid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  int32  `json:"status"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	_, err = rpc.ArticleClient().UpdateArticle(context.Background(), &article.UpdateArticleRequest{
		ArticleId: aid,
		Title:     req.Title,
		Content:   req.Content,
		Status:    req.Status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp(errno.ServiceErr, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp(errno.Success, nil))
}

func DeleteArticle(_ context.Context, c *app.RequestContext) {
	aid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	_, err = rpc.ArticleClient().DeleteArticle(context.Background(), &article.DeleteArticleRequest{
		ArticleId: aid,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp(errno.ServiceErr, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp(errno.Success, nil))
}

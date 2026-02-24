package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	"byte.dance/api/biz/handler"
)

func Register(h *server.Hertz) {
	v1 := h.Group("/api/v1")

	// public
	v1.POST("/user/register", handler.Register)
	v1.POST("/user/login", handler.Login)
	v1.GET("/articles", handler.ListArticle)
	v1.GET("/article/:id", handler.GetArticle)
	v1.GET("/user/:id", handler.GetUser)

	// auth required
	auth := v1.Group("", handler.AuthMiddleware())
	auth.PUT("/user/:id", handler.UpdateUser)
	auth.POST("/article", handler.CreateArticle)
	auth.PUT("/article/:id", handler.UpdateArticle)
	auth.DELETE("/article/:id", handler.DeleteArticle)
}

package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"byte.dance/pkg/errno"
	"byte.dance/pkg/middleware"
)

func AuthMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		auth := string(c.GetHeader("Authorization"))
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, resp(errno.AuthErr, nil))
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := middleware.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, resp(errno.AuthErr, nil))
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next(ctx)
	}
}

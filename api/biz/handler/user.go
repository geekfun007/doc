package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	"byte.dance/kitex_gen/user"
	"byte.dance/api/biz/rpc"
	"byte.dance/pkg/errno"
)

func Register(_ context.Context, c *app.RequestContext) {
	var req struct {
		Username string `json:"username" vd:"len($)>1"`
		Email    string `json:"email"`
		Password string `json:"password" vd:"len($)>5"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	r, err := rpc.UserClient().Register(context.Background(), &user.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp(errno.ServiceErr, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp(errno.Success, utils.H{
		"user_id": r.UserId,
		"token":   r.Token,
	}))
}

func Login(_ context.Context, c *app.RequestContext) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	r, err := rpc.UserClient().Login(context.Background(), &user.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp(errno.ServiceErr, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp(errno.Success, utils.H{
		"user_id": r.UserId,
		"token":   r.Token,
	}))
}

func GetUser(_ context.Context, c *app.RequestContext) {
	uid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	r, err := rpc.UserClient().GetUser(context.Background(), &user.GetUserRequest{UserId: uid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp(errno.ServiceErr, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp(errno.Success, r.User))
}

func UpdateUser(_ context.Context, c *app.RequestContext) {
	uid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Avatar   string `json:"avatar"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, resp(errno.ParamErr, nil))
		return
	}

	_, err = rpc.UserClient().UpdateUser(context.Background(), &user.UpdateUserRequest{
		UserId:   uid,
		Username: req.Username,
		Email:    req.Email,
		Avatar:   req.Avatar,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp(errno.ServiceErr, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp(errno.Success, nil))
}

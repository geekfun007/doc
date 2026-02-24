package main

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"byte.dance/kitex_gen/user"
	"byte.dance/pkg/middleware"
	"byte.dance/user/dal"
)

type UserServiceImpl struct{}

func (s *UserServiceImpl) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u, err := dal.GetStore().Create(req.Username, req.Email, string(hashed))
	if err != nil {
		if errors.Is(err, dal.ErrUserExists) {
			return nil, err
		}
		return nil, err
	}

	token, err := middleware.GenerateToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &user.RegisterResponse{
		UserId: u.ID,
		Token:  token,
	}, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	u, err := dal.GetStore().GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("wrong password")
	}

	token, err := middleware.GenerateToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &user.LoginResponse{
		UserId: u.ID,
		Token:  token,
	}, nil
}

func (s *UserServiceImpl) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	u, err := dal.GetStore().GetByID(req.UserId)
	if err != nil {
		return nil, err
	}

	return &user.GetUserResponse{
		User: &user.User{
			Id:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			Avatar:    u.Avatar,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	err := dal.GetStore().Update(req.UserId, req.Username, req.Email, req.Avatar)
	if err != nil {
		return nil, err
	}

	return &user.UpdateUserResponse{Success: true}, nil
}
